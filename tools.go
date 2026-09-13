package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"google.golang.org/genai"
)

func getToolList() []*genai.Tool {
	return []*genai.Tool{
		{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				{
					Name:        "createNewFile",
					Description: "Create a new file given a filepath",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"filepath": {
								Type:        genai.TypeString,
								Description: "The filepath of file to create",
							},
						},
						Required: []string{"filepath"},
					},
				},
			},
		},
		{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				{
					Name:        "writeFile",
					Description: "Write to a file given a filepath. This will overwrite the whole file. Make sure to use the read tool first and only call this tool with the full file contents.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"filepath": {
								Type:        genai.TypeString,
								Description: "The filepath of file to write",
							},
							"contents": {
								Type:        genai.TypeString,
								Description: "The full contents of the file",
							},
						},
						Required: []string{"filepath"},
					},
				},
			},
		},
		{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				{
					Name:        "readFile",
					Description: "Read a file given a filepath",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"filepath": {
								Type:        genai.TypeString,
								Description: "The filepath of file to read",
							},
						},
						Required: []string{"filepath"},
					},
				},
			},
		},
		{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				{
					Name:        "pwd",
					Description: "Get the path of the present working directory",
				},
			},
		},
		{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				{
					Name:        "ls",
					Description: "List all of the files in a given directory",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"filepath": {
								Type:        genai.TypeString,
								Description: "The filepath of the directory to list",
							},
						},
						Required: []string{"filepath"},
					},
				},
			},
		},
	}
}

func handleFunctionCall(fn *genai.FunctionCall) map[string]any {
	var result map[string]any = nil
	args := fn.Args
	switch fn.Name {
	case "createNewFile":
		if args["filepath"] != nil {
			result = createNewFile(args["filepath"].(string))
		}
	case "writeFile":
		if args["filepath"] != nil {
			result = writeFile(args["filepath"].(string), args["contents"].(string))
		}
	case "readFile":
		if args["filepath"] != nil {
			result = readFile(args["filepath"].(string))
		}
	case "pwd":
		result = pwd()
	case "ls":
		if args["filepath"] != nil {
			result = ls(args["filepath"].(string))
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
