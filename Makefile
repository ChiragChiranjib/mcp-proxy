.PHONY: run tidy build lint migrate migrate-run seed-servers migrate-up migrate-down migrate-status migrate-create seed

run:
	APP_ENV=dev MCP_MODE=streamable-http go run ./cmd/mcp-gateway

tidy:
	go mod tidy

build:
	GOOS=darwin GOARCH=amd64 go build -o bin/mcp-gateway ./cmd/mcp-gateway

build-migrate:
	go build -o bin/migrate ./cmd/migrate

lint:
	golangci-lint run ./...

# Goose helpers
migrate-up:
	APP_ENV=dev MCP_MODE=streamable-http go run ./cmd/migrate -dir migrations up

migrate-down:
	APP_ENV=dev MCP_MODE=streamable-http go run ./cmd/migrate -dir migrations down

migrate-status:
	APP_ENV=dev MCP_MODE=streamable-http go run ./cmd/migrate -dir migrations status

# Seed catalog mcp_servers from config/mcp_servers.json and exit
seed:
	APP_ENV=dev MCP_MODE=streamable-http go run ./cmd/seed -file config/mcp_servers.json