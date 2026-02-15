.PHONY: build run-server clean dev test protogen completion lint lint-fix help

BINARY_NAME=mitra
BUILD_DIR=bin

help:
	@echo "Available targets:"
	@echo "  build      - Build the server binary"
	@echo "  run-server - Build and run the server"
	@echo "  dev        - Run server in development mode"
	@echo "  test       - Run tests"
	@echo "  lint       - Run golangci-lint"
	@echo "  lint-fix   - Run golangci-lint with auto-fix"
	@echo "  protogen   - Generate Go code from proto files"
	@echo "  shell      - Generate shell artifacts like aliases and completion (zsh only)"
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

lint-fix:
	@golangci-lint run --fix

protogen:
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		internal/proto/*.proto

shell: build
	@./$(BUILD_DIR)/$(BINARY_NAME) completion zsh > shell/zsh.completion
	@echo "Zsh completion written to shell/zsh.completion"

clean:
	@rm -rf $(BUILD_DIR)
