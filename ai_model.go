package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
)

type agent struct {
	client         *googleai.GoogleAI
	toolList       []llms.Tool
	messageHistory []llms.MessageContent
	model          string
}

type modelConfig struct {
	ApiKey      string `json:"apiKey"`
	Model       string `json:"model"`
	ModelFamily string `json:"modelFamily"`
}

func newAgent(ctx context.Context) *agent {
	log.Println("Initializing agent...")

	loadedModelConfig, _ := loadConfig()

	apiKey := os.Getenv(loadedModelConfig.ApiKey)
	model := loadedModelConfig.Model

	log.Println("initializing agent")

	// todo: initialize client based on model family

	client, err := googleai.New(ctx, googleai.WithAPIKey(apiKey), googleai.WithDefaultModel(model))

	if err != nil {
		log.Fatalln("Error creating new client: ", err)
	}

	systemPrompt := "You are a simple coding agent"

	history := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
	}

	a := &agent{client: client, toolList: getToolList(), messageHistory: history, model: model}

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

func (a *agent) makeResponse(prompt string, ctx context.Context) (string, error) {
	log.Println("Received prompt: ", prompt)
	a.messageHistory = append(a.messageHistory, llms.TextParts(llms.ChatMessageTypeHuman, prompt))
	log.Println("history: ", a.messageHistory)
	log.Println("toolList: ", a.toolList)
	response, err := a.client.GenerateContent(ctx, a.messageHistory, llms.WithTools(a.toolList))
	if err != nil {
		log.Printf("Error calling model: %v\n", err)
		return "", err
	}
	finalResponse := ""

	for range 10 {
		debugPrint(response)

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
			return finalResponse, nil
		}

		for _, tc := range respChoice.ToolCalls {
			funcCallResponse := handleFunctionCall(*tc.FunctionCall)

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
		response, err = a.client.GenerateContent(ctx, a.messageHistory, llms.WithTools(a.toolList))
	}

	return "", fmt.Errorf("reached the end of the loop")
}
