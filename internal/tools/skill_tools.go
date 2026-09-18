package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"

	"github.com/yuin/goldmark/parser"
)

const skillDirectory string = "./skills"

func GetSkillToolList() []llms.Tool {
	return []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "skillLoadSkill",
				Description: "Load a skill given a filepath",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"filepath": map[string]any{
							"type":        "string",
							"description": "The filepath of the skill to read",
						},
					},
					"required": []string{"filepath"},
				},
			},
		},
	}
}

func GetSkillList() ([]map[string]any, error) {

	skillFiles, err := os.ReadDir(skillDirectory)
	if err != nil {
		log.Println("Error listing path: ", skillDirectory, err)
		return nil, err
	}

	res := make([]map[string]any, len(skillFiles))

	for _, skillFile := range skillFiles {
		content, err := os.ReadFile(skillDirectory + "/" + skillFile.Name())
		if err != nil {
			log.Println("Error reading file: ", skillFile, err)
			return nil, err
		}
		md := goldmark.New(
			goldmark.WithExtensions(
				meta.Meta,
			),
		)

		var buf bytes.Buffer
		context := parser.NewContext()
		if err := md.Convert(content, &buf, parser.WithContext(context)); err != nil {
			log.Fatal(err)
		}

		skillMap := meta.Get(context)
		skillMap["filename"] = skillFile.Name()
		res = append(res, skillMap)
	}
	return res, nil
}

func HandleSkillFunctionCall(fn llms.FunctionCall) map[string]any {
	var result map[string]any
	var args struct {
		Filepath string `json:"filepath"`
	}
	rawArgs := fn.Arguments
	err := json.Unmarshal([]byte(rawArgs), &args)
	if err != nil {
		log.Printf("Error deserializing function call: %v\n", err)
		return result
	}
	switch fn.Name {
	case "skillLoadSkill":
		if args.Filepath != "" {
			content, _ := loadSkill(args.Filepath)
			result = map[string]any{"success": true, "content": content}
		}
	default:
		result = map[string]any{"success": false, "error": fmt.Sprintf("No function called %s", fn.Name)}
	}
	return result
}

func loadSkill(filename string) (string, error) {
	content, err := os.ReadFile(skillDirectory + "/" + filename)
	if err != nil {
		log.Println("Error reading file: ", filename, err)
		return "", err
	}
	md := goldmark.New(
		goldmark.WithExtensions(
			meta.Meta,
		),
	)

	var buf bytes.Buffer
	context := parser.NewContext()
	if err := md.Convert(content, &buf, parser.WithContext(context)); err != nil {
		log.Fatal(err)
	}

	skillBody := buf.String()
	return skillBody, nil
}
