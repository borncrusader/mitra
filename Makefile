.PHONY: build run-server clean dev test protogen help

BINARY_NAME=mitra
BUILD_DIR=bin

help:
	@echo "Available targets:"
	@echo "  build      - Build the server binary"
	@echo "  run-server - Build and run the server"
	@echo "  dev        - Run server in development mode"
	@echo "  test       - Run tests"
	@echo "  protogen   - Generate Go code from proto files"
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

protogen:
	@protoc --go_out=. --go_opt=paths=source_relative internal/proto/*.proto

clean:
	@rm -rf $(BUILD_DIR)
