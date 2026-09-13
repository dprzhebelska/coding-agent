# Coding Agent
## Objectives
- Learn how coding agents work
- Learn how to develop in Go

## Running the app
Update model_config.json with your config first. Supported model families:
- google
- openai

<br>

```
export API_KEY="<your api key>"

make run
```

## Initial version with gemini
A simple agent that can uses google gemini and has access to read, write files and see the directory. 
![screenshot](readme_assets/initial%20gemini%20example.png)

## Todo
- Update the terminal title with the conversation summary
- Make the terminal scrollable
- Add skills
- Add hooks
- Support for MCP servers
- More tools: diff write, exit, grep...
- Interface enhancements
- Thinking
- Save conversation history
- streaming

## Bugs
- 