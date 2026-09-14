package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
)

func getTuiToolList() []llms.Tool {
	return []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "tuiUpdateTitle",
				Description: "Update the terminal title",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"title": map[string]any{
							"type":        "string",
							"description": "The new title to update to",
						},
					},
					"required": []string{"title"},
				},
			},
		},
	}
}

func handleTuiFunctionCall(fn llms.FunctionCall) (map[string]any, map[string]any) {
	var result map[string]any
	var tuiUpdate map[string]any
	var args struct {
		Title string `json:"title"`
	}
	rawArgs := fn.Arguments
	err := json.Unmarshal([]byte(rawArgs), &args)
	if err != nil {
		log.Printf("Error deserializing function call: %v\n", err)
		return result, tuiUpdate
	}
	switch fn.Name {
	case "tuiUpdateTitle":
		if args.Title != "" {
			result = map[string]any{"success": true}
			tuiUpdate = map[string]any{"title": args.Title}
		}
	default:
		result = map[string]any{"success": false, "error": fmt.Sprintf("No function called %s", fn.Name)}
	}
	return result, tuiUpdate
}
