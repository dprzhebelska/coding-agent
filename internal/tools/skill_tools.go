package tools

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"gopkg.in/yaml.v2"
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
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "skillListSkills",
				Description: "List all available skills",
				Parameters: map[string]any{
					"type":       "object",
					"properties": map[string]any{},
				},
			},
		},
	}
}

func GetSkillList() (string, error) {

	skillFiles, err := os.ReadDir(skillDirectory)
	if err != nil {
		log.Println("Error listing path: ", skillDirectory, err)
		return "", err
	}

	res := make([]map[string]any, 0, len(skillFiles))

	for _, skillFile := range skillFiles {
		content, err := os.ReadFile(skillDirectory + "/" + skillFile.Name())
		if err != nil {
			log.Println("Error reading file: ", skillFile, err)
			return "", err
		}
		meta, err := parseFrontmatter(string(content))
		if err != nil {
			log.Println("Error parsing frontmatter: ", skillFile.Name(), err)
			continue
		}
		meta["filename"] = skillFile.Name()
		res = append(res, meta)
	}
	readable_response := "Here is a list of available skills: \n"
	for _, skill := range res {
		readable_response += fmt.Sprintf("- %s: %s\n    - filepath: %s", skill["name"], skill["description"], skill["filename"])
	}
	return readable_response, nil
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
			content, err := loadSkill(args.Filepath)
			if err != nil {
				result = map[string]any{"success": false, "error": "error calling tool"}
				log.Print("error calling loadskill tool: ", err)
			} else {
				result = map[string]any{"success": true, "content": content}
			}
		}
	case "skillListSkills":
		content, _ := GetSkillList()
		if err != nil {
			result = map[string]any{"success": false, "error": "error calling tool"}
			log.Print("error calling getSkill tool: ", err)
		} else {
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
	// Strip YAML frontmatter
	text := string(content)
	if strings.HasPrefix(text, "---") {
		end := strings.Index(text[3:], "---")
		if end != -1 {
			text = text[3+end+3:]
		}
	}
	return strings.TrimSpace(text), nil
}

func parseFrontmatter(content string) (map[string]any, error) {
	if !strings.HasPrefix(content, "---") {
		return nil, fmt.Errorf("no frontmatter found")
	}
	end := strings.Index(content[3:], "---")
	if end == -1 {
		return nil, fmt.Errorf("unterminated frontmatter")
	}
	yamlContent := content[3 : 3+end]

	var meta map[string]any
	if err := yaml.Unmarshal([]byte(yamlContent), &meta); err != nil {
		return nil, err
	}
	return meta, nil
}

// func main() {
// 	fmt.Print(loadSkill("SWEDEN.md"))
// }
