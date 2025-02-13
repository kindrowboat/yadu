# Variables
BINARY_NAME=yadu
BUILD_DIR=build
INSTALL_DIR=/usr/local/bin

# Go related variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/$(BUILD_DIR)

# Build
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) cmd/yadu/main.go

# Install
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME)..."
	@sudo install -m 755 $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@yadu completion bash | sudo tee /etc/bash_completion.d/yadu > /dev/null
# Clean
.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

# Test
.PHONY: test
test:
	@go test -v ./...

# Test with coverage
.PHONY: coverage
coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

# Generate completions
.PHONY: completions
completions: build
	@mkdir -p $(BUILD_DIR)/completions
	@$(BUILD_DIR)/$(BINARY_NAME) completion bash > $(BUILD_DIR)/completions/yadu.bash
	@$(BUILD_DIR)/$(BINARY_NAME) completion zsh > $(BUILD_DIR)/completions/yadu.zsh
	@$(BUILD_DIR)/$(BINARY_NAME) completion fish > $(BUILD_DIR)/completions/yadu.fish

# Format
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Default target
.DEFAULT_GOAL := build