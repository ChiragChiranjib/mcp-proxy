.PHONY: run tidy build lint migrate migrate-run seed-servers

run:
	APP_ENV=dev MCP_MODE=streamable-http go run ./cmd/mcp-gateway

tidy:
	go mod tidy

build:
	GOOS=darwin GOARCH=amd64 go build -o bin/mcp-gateway ./cmd/mcp-gateway

lint:
	golangci-lint run ./...

migrate:
	go build -o bin/migrate ./cmd/migrate

migrate-run:
	APP_ENV=dev MCP_MODE=streamable-http go run ./cmd/migrate

# Seed catalog mcp_servers from config/mcp_servers.json and exit
seed-servers:
	APP_ENV=dev MCP_MODE=streamable-http go run ./cmd/mcp-gateway --seed-mcp-servers
