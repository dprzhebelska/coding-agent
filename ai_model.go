package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"google.golang.org/genai"
)

type agent struct {
	client *genai.Client
	chat   *genai.Chat
	model  string
	config *genai.GenerateContentConfig
}

func newAgent(ctx context.Context) *agent {
	log.Println("Initializing agent...")

	apiKey := os.Getenv("GEMINI_API_KEY")

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})

	if err != nil {
		log.Fatalln("Error creating new client: ", err)
	}
	model := "gemini-3.6-flash"

	systemPrompt := "You are a simple coding agent"

	config := &genai.GenerateContentConfig{
		Tools: getToolList(),
		SystemInstruction: &genai.Content{Parts: []*genai.Part{
			{Text: systemPrompt},
		},
		},
	}

	chat, err := client.Chats.Create(ctx, model, config, nil)

	if err != nil {
		log.Fatalln("Error creating new chat: ", err)
	}

	a := &agent{client: client, chat: chat, model: model, config: config}

	log.Println("Agent initialized!")

	return a
}

func debugPrint[T any](r *T) {
	// Marshal the result to JSON.
	response, err := json.MarshalIndent(*r, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	// Log the output.
	log.Print("Received response: ", string(response))
}

func (a *agent) makeResponse(prompt string, ctx context.Context) string {
	log.Println("Received prompt: ", prompt)
	response, err := a.chat.SendMessage(ctx, genai.Part{Text: prompt})

	if err != nil {
		log.Fatalln("Error calling model: ", err)
		return "Error getting response, please try again"
	}

	for response != nil {
		debugPrint(response)
		for _, candidate := range response.Candidates {
			if candidate.Content != nil {
				for _, part := range candidate.Content.Parts {
					if part.Text != "" {
						return part.Text
					}
					if part.FunctionCall != nil {
						fn := part.FunctionCall

						result := handleFunctionCall(fn)

						if result != nil {
							response, err = a.chat.SendMessage(ctx, genai.Part{
								FunctionResponse: &genai.FunctionResponse{
									Name: fn.Name,
									// The response must be inside a map structure matching the key-value outputs
									Response: map[string]any{"result": result},
								},
							},
							)
							if err != nil {
								log.Fatalln("Error getting response from model after tool call: ", err)
								return "Error getting response, please try again"
							}

						} else {
							response = nil
						}
					}
				}
			}
		}
	}

	return "Function finished"
}

func responseHasFunctionCall(response *genai.GenerateContentResponse) bool {
	hasFunction := false
	for _, candidate := range response.Candidates {
		if candidate.Content != nil {
			for _, part := range candidate.Content.Parts {
				if part.FunctionCall != nil {
					hasFunction = true
				}
			}
		}
	}
	return hasFunction
}
