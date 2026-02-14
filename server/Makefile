.PHONY: build run clean dev

BINARY_NAME=mitra
BUILD_DIR=bin

build:
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/mitra

run: build
	@./$(BUILD_DIR)/$(BINARY_NAME) server

dev:
	@go run ./cmd/mitra

clean:
	@rm -rf $(BUILD_DIR)
