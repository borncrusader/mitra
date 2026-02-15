.PHONY: build run-server clean dev test protogen completion lint help

BINARY_NAME=mitra
BUILD_DIR=bin

help:
	@echo "Available targets:"
	@echo "  build      - Build the server binary"
	@echo "  run-server - Build and run the server"
	@echo "  dev        - Run server in development mode"
	@echo "  test       - Run tests"
	@echo "  lint       - Run golangci-lint"
	@echo "  protogen   - Generate Go code from proto files"
	@echo "  completion - Generate zsh completion file"
	@echo "  clean      - Remove build artifacts"

build:
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/mitra

run-server: build
	@./$(BUILD_DIR)/$(BINARY_NAME) server

dev:
	@go run ./cmd/mitra

test:
	@go test ./...

lint:
	@golangci-lint run

protogen:
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		internal/proto/*.proto

completion: build
	@./$(BUILD_DIR)/$(BINARY_NAME) completion zsh > .zsh.completion
	@echo "Zsh completion written to .zsh.completion"

clean:
	@rm -rf $(BUILD_DIR)
