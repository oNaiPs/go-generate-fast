# Variables
GO := go
SRC_DIR := src
OUT_DIR := bin
TARGET := $(OUT_DIR)/go-generate-fast

# List of all Go source files
SOURCES := $(shell find $(SRC_DIR) -name "*.go")

# Pin the tools versions
GOTESTSUM := go run gotest.tools/gotestsum@v1.11.0
GOLANGCILINT := go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2

BUILD_TAGS :=
TEST_TAGS := test

# By default, commands are not printed, unless you pass V=1
ifndef V
.SILENT:
endif

build: $(TARGET) ## Build project

$(TARGET): $(SOURCES)
	mkdir -p $(OUT_DIR)
	echo "Building..."
	$(GO) build -o $(TARGET) --tags=$(BUILD_TAGS)
	echo "Build complete. Output: $(TARGET)"

test: ## Run tests
	echo "Running unit tests..."
	$(GOTESTSUM) --raw-command -- $(GO) -C $(SRC_DIR) test --tags=$(TEST_TAGS) -json -cover ./...

lint: ## Lints files
	echo "Linting..."
	$(GOLANGCILINT) run
lint-fix: ## Lints files, and fixes ones that are fixable 
	echo "Linting and fixing..."
	$(GOLANGCILINT) run --fix

clean: ## Cleans build files
	rm -rfv $(OUT_DIR)
	echo Done.

.PHONY: build clean test help

.DEFAULT_GOAL := help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

