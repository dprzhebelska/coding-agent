build:
	go build -o coding_agent
	chmod +x coding_agent

run: build
	./coding_agent