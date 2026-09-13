package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/tmc/langchaingo/llms"
)

func getToolList() []llms.Tool {
	return []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "createNewFile",
				Description: "Create a new file given a filepath",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"filepath": map[string]any{
							"type":        "string",
							"description": "The filepath of file to create",
						},
					},
					"required": []string{"filepath"},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "writeFile",
				Description: "Write to a file given a filepath. This will overwrite the whole file. Make sure to use the read tool first and only call this tool with the full file contents.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"filepath": map[string]any{
							"type":        "string",
							"description": "The filepath of file to write",
						},
						"contents": map[string]any{
							"type":        "string",
							"description": "The full contents of the file",
						},
					},
					"required": []string{"filepath", "contents"},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "readFile",
				Description: "Read a file given a filepath",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"filepath": map[string]any{
							"type":        "string",
							"description": "The filepath of file to read",
						},
					},
					"required": []string{"filepath"},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "pwd",
				Description: "Get the path of the present working directory",
				Parameters: map[string]any{
					"type":       "object",
					"properties": map[string]any{},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "ls",
				Description: "List all of the files in a given directory",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"filepath": map[string]any{
							"type":        "string",
							"description": "The filepath of the directory to list",
						},
					},
					"required": []string{"filepath"},
				},
			},
		},
	}
}

func handleFunctionCall(fn llms.FunctionCall) map[string]any {
	var result map[string]any = nil
	var args struct {
		Filepath string `json:"filepath"`
		Contents string `json:"contents"`
	}
	rawArgs := fn.Arguments
	err := json.Unmarshal([]byte(rawArgs), &args)
	if err != nil {
		log.Fatalf("Error deserializing function call: %v\n", err)
		return result
	}
	log.Print(fn.Name, args)
	switch fn.Name {
	case "createNewFile":
		if args.Filepath != "" {
			result = createNewFile(args.Filepath)
		}
	case "writeFile":
		if args.Filepath != "" {
			result = writeFile(args.Filepath, args.Contents)
		}
	case "readFile":
		if args.Filepath != "" {
			result = readFile(args.Filepath)
		}
	case "pwd":
		result = pwd()
	case "ls":
		if args.Filepath != "" {
			result = ls(args.Filepath)
		}
	default:
		result = map[string]any{"success": false, "error": fmt.Sprintf("No function called %s", fn.Name)}
	}
	log.Printf("Function response: %v\n", result)
	return result
}

func createNewFile(fileName string) map[string]any {
	_, err := os.Create(fileName)
	if err != nil {
		log.Fatalln("Error creating file: ", fileName, err)
		return map[string]any{"success": false, "error": err}
	}
	return map[string]any{"success": true}
}

func writeFile(fileName string, contents string) map[string]any {
	bytes := []byte(contents)
	full_filepath, err := filepath.Abs(fileName)
	if err != nil {
		full_filepath = fileName
	}
	err = os.WriteFile(full_filepath, bytes, 0644)

	if err != nil {
		log.Fatalln("Error writing file: ", full_filepath, err)
		return map[string]any{"success": false, "error": err}
	}
	return map[string]any{"success": true}
}

func readFile(fileName string) map[string]any {
	bytes, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatalln("Error reading file: ", fileName, err)
		return map[string]any{"success": false, "error": err}
	}
	return map[string]any{"success": true, "contents": string(bytes)}
}

func ls(path string) map[string]any {
	files, err := os.ReadDir(path)
	if err != nil {
		log.Fatalln("Error listing path: ", path, err)
		return map[string]any{"success": false, "error": err}
	}
	return map[string]any{"success": true, "file_list": files}
}

func pwd() map[string]any {
	directory, err := os.Getwd()
	if err != nil {
		log.Fatalln("Error getting pwd: ", err)
		return map[string]any{"success": false, "error": err}
	}
	return map[string]any{"success": true, "directory": fmt.Sprint(directory)}
}

// func main() {
// 	fmt.Print(ls("/home/daria/Documents/go/coding-agent"))
// }
