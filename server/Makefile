.PHONY: build run clean dev

BINARY_NAME=mitra
BUILD_DIR=bin

build:
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) .

run: build
	@./$(BUILD_DIR)/$(BINARY_NAME) server

dev:
	@go run main.go

clean:
	@rm -rf $(BUILD_DIR)
