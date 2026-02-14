.PHONY: build run clean dev

BINARY_NAME=server
BUILD_DIR=bin

build:
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) main.go

run: build
	@./$(BUILD_DIR)/$(BINARY_NAME)

dev:
	@go run main.go

clean:
	@rm -rf $(BUILD_DIR)
