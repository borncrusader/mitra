.PHONY: build run-server clean dev help

BINARY_NAME=mitra
BUILD_DIR=bin

help:
	@echo "Available targets:"
	@echo "  build      - Build the server binary"
	@echo "  run-server - Build and run the server"
	@echo "  dev        - Run server in development mode"
	@echo "  clean      - Remove build artifacts"

build:
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/mitra

run-server: build
	@./$(BUILD_DIR)/$(BINARY_NAME) server

dev:
	@go run ./cmd/mitra

clean:
	@rm -rf $(BUILD_DIR)
