package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"os"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/openai"
)

type agent struct {
	llm            llms.Model
	toolList       []llms.Tool
	messageHistory []llms.MessageContent
	model          string
}

type modelConfig struct {
	ApiKey      string `json:"apiKey"`
	Model       string `json:"model"`
	ModelFamily string `json:"modelFamily"`
	URL         string `json:"url"`
}

func newAgent(ctx context.Context) *agent {
	log.Println("Initializing agent...")

	loadedModelConfig, _ := loadConfig()

	apiKey := os.Getenv(loadedModelConfig.ApiKey)
	model := loadedModelConfig.Model

	log.Println("initializing agent")

	// todo: initialize client based on model family

	var llm llms.Model
	var err error
	switch loadedModelConfig.ModelFamily {
	case "google":
		llm, err = googleai.New(ctx, googleai.WithAPIKey(apiKey), googleai.WithDefaultModel(model))
		if err != nil {
			log.Fatalln("Error creating new google llm: ", err)
		}
	case "openai":
		llm, err = openai.New(openai.WithBaseURL(loadedModelConfig.URL), openai.WithToken(apiKey), openai.WithModel(model))
		if err != nil {
			log.Fatalln("Error creating new openai llm: ", err)
		}
	}

	systemPrompt := "You are a simple coding agent. After you have exchanged a few messages with the user, use the tuiUpdateTitle tool to update the title of the session with a summary of the conversation. If the conversation direction changes, update the title then too. Don't update any files unless the user tells you to. Ask for permission first."

	history := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
	}

	a := &agent{llm: llm, toolList: append(getIoToolList(), getTuiToolList()...), messageHistory: history, model: model}

	log.Println("Agent initialized!")

	return a
}

func loadConfig() (modelConfig, error) {
	defaultConfig := modelConfig{ApiKey: "GEMINI_API_KEY", Model: "gemini-3.6-flash", ModelFamily: "google"}

	bytes, err := os.ReadFile("model_config.json")
	if err != nil {
		log.Printf("Error loading config: %v\nusing default: %v", err, defaultConfig)
		return defaultConfig, err
	}
	var loadedConfig modelConfig
	err = json.Unmarshal(bytes, &loadedConfig)
	if err != nil {
		log.Printf("Error loading config: %v\nusing default: %v", err, defaultConfig)
		return defaultConfig, err
	}
	log.Printf("loaded model config: %v", loadedConfig)
	return loadedConfig, nil
}

func debugPrint[T any](r *T) {
	// Marshal the result to JSON.
	response, err := json.MarshalIndent(*r, "", "  ")
	if err != nil {
		log.Println("Error printing: ", err)
	}
	// Log the output.
	log.Print("Received response: ", string(response))
}

func (a *agent) makeResponse(prompt string, ctx context.Context) (string, map[string]any, error) {
	log.Println("Received prompt: ", prompt)
	a.messageHistory = append(a.messageHistory, llms.TextParts(llms.ChatMessageTypeHuman, prompt))
	response, err := a.llm.GenerateContent(ctx, a.messageHistory, llms.WithTools(a.toolList))
	if err != nil {
		log.Printf("Error calling model: %v\n", err)
		return "", map[string]any{}, err
	}
	finalResponse := ""
	tuiUpdates := map[string]any{}

	for range 10 {
		debugPrint(response)

		if response == nil || len(response.Choices) == 0 {
			return finalResponse, tuiUpdates, nil
		}

		respChoice := response.Choices[0]
		content := respChoice.Content

		// add to the model's context
		aiResponseHistory := llms.TextParts(llms.ChatMessageTypeAI, content)
		if len(respChoice.ToolCalls) > 0 {
			for _, tc := range respChoice.ToolCalls {
				aiResponseHistory.Parts = append(aiResponseHistory.Parts, tc)
			}
		}
		a.messageHistory = append(a.messageHistory, aiResponseHistory)

		finalResponse = finalResponse + content
		// if no more tool calls, we are done
		if len(respChoice.ToolCalls) == 0 {
			return finalResponse, tuiUpdates, nil
		}

		var funcCallResponse map[string]any
		for _, tc := range respChoice.ToolCalls {
			if strings.HasPrefix(tc.FunctionCall.Name, "io") {
				funcCallResponse = handleIoFunctionCall(*tc.FunctionCall)
			} else if strings.HasPrefix(tc.FunctionCall.Name, "tui") {
				var tuiUpdate map[string]any
				funcCallResponse, tuiUpdate = handleTuiFunctionCall(*tc.FunctionCall)
				maps.Copy(tuiUpdates, tuiUpdate)
				log.Printf("tui updates: %v, %v", tuiUpdate, tuiUpdates)
			}

			toolResponse := llms.MessageContent{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{
						Name:    tc.FunctionCall.Name,
						Content: fmt.Sprint(funcCallResponse),
					},
				},
			}
			a.messageHistory = append(a.messageHistory, toolResponse)
		}
		response, err = a.llm.GenerateContent(ctx, a.messageHistory, llms.WithTools(a.toolList))
	}

	return "", tuiUpdates, fmt.Errorf("reached the end of the loop")
}
