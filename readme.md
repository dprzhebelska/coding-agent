# Coding Agent
## Objectives
- Learn how coding agents work
- Learn how to develop in Go

## Build
```
export GEMINI_API_KEY="<your api key>"
go build -o coding_agent
chmod +x coding_agent
./coding_agent
```

## Initial version with gemini
A simple agent that can uses google gemini and has access to read, write files and see the directory. 
![screenshot](readme_assets/initial%20gemini%20example.png)

## Todo
- Switch llm library to langchain, add support for other models
- (with higher API limits) update the terminal title with the conversation summary
- Add skills
- Add hooks
- Support for MCP servers
- More tools: diff write, exit, grep...
- Interface enhancements
- Thinking
- Save conversation history
- streaming

## Bugs
- hitting the API rate limits causes system exit instead of just printing out an error