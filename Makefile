.PHONY: run tidy build build-migrate lint migrate-up migrate-down migrate-status seed setup up down teardown stop

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

# Seed catalog mcp_servers from cmd/seed/data and exit
seed:
	APP_ENV=dev MCP_MODE=streamable-http go run ./cmd/seed -only servers

# Setup: migrate up, seed, and start server (foreground)
setup: migrate-up seed run

# Up: migrate up, seed, and start server in background
up:
	APP_ENV=dev MCP_MODE=streamable-http go run ./cmd/migrate -dir migrations up
	APP_ENV=dev MCP_MODE=streamable-http go run ./cmd/seed -only servers
	APP_ENV=dev MCP_MODE=streamable-http go run ./cmd/seed -only users
	APP_ENV=dev MCP_MODE=streamable-http nohup go run ./cmd/mcp-gateway > server.out 2>&1 & echo $$! > server.pid
	@echo "Server started with PID `cat server.pid`"

# Stop: stop background server if running
stop:
	@if [ -f server.pid ]; then \
		kill `cat server.pid` || true; \
		rm -f server.pid; \
		echo "Server stopped"; \
	else \
		echo "No server.pid found"; \
	fi

# Down: stop server and drop all tables (migration down)
down: stop
	APP_ENV=dev MCP_MODE=streamable-http go run ./cmd/migrate -dir migrations down

# Teardown: stop server, drop tables
teardown: down