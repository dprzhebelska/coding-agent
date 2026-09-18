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
- Move system prompt to its own file
- Add guardrails/blocks from editing files without permission
- Clean up Skills
- Add AGENT.md support
- Add hooks
- Streaming
- Support for MCP servers
- More tools: gh, diff write, exit, grep...
- Interface enhancements - select, shift+enter, copy/paste doesn't take up the whole terminal
- Thinking
- Save conversation history

## Bugs
- terminal title update doesn't actually update the title