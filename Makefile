.PHONY: build run-server clean dev help

help:
	@echo "Available targets:"
	@echo "  build      - Build the server binary"
	@echo "  run-server - Build and run the server"
	@echo "  dev        - Run server in development mode"
	@echo "  clean      - Remove build artifacts"

build:
	@cd server && $(MAKE) build

run-server:
	@cd server && $(MAKE) run-server

dev:
	@cd server && $(MAKE) dev

clean:
	@cd server && $(MAKE) clean
