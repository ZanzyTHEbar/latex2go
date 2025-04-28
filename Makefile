# Makefile for Thoth Network

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Go build flags
BUILD_FLAGS := -v -ldflags "-X 'version.BuildTime=$(shell date -u)'"
GOBUILD_FLAGS := $(BUILD_FLAGS) -o

# Detect the operating system and architecture
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
ifeq ($(GOOS), windows)
	GOOS := windows
else ifeq ($(GOOS), darwin)
	GOOS := darwin
else ifeq ($(GOOS), linux)
	GOOS := linux
endif

# Binary names
BINARY_BASE_NAME=latex2go
BINARY_NAME := $(BINARY_BASE_NAME)_$(GOOS)_$(GOARCH)

# Binary path
BINARY_PATH_PREFIX := ./bin
BINARY_PATH := $(BINARY_PATH_PREFIX)/$(GOOS)_$(GOARCH)

# Module path
MODULE_PATH := ./cmd/$(BINARY_BASE_NAME).go

# Default target
.DEFAULT_GOAL := build

# Temporary paths
TEMP_PATHS := /tmp/$(BINARY_BASE_NAME)_test_*

# Build the application
all: test build

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BINARY_PATH)
	@CGO_ENABLED=1 GOOS=$(GOOS) go build $(BUILD_FLAGS) -o $(BINARY_PATH)/$(BINARY_NAME) $(MODULE_PATH)
	@echo "Binary built at $(BINARY_PATH)/$(BINARY_NAME)"

test:
	$(GOTEST) -v ./...

run: build
	bin/$(BINARY_NAME_SERVER)

deps:
	$(GOMOD) download

tidy:
	$(GOMOD) tidy

lint: vet
	golangci-lint run ./...
	@echo "Linting completed."

vet:
	go vet ./...
	@echo "Vet completed."

# Generate mocks for testing
mocks:
	mockery --all --dir=internal --output=internal/mocks

# Clean targets
clean-all: proto-clean
	@echo "Cleaning all build artifacts..."
	@rm -rf $(BINARY_PATH_PREFIX)
	@go clean -cache -modcache -i -r
	@echo "All build artifacts cleaned successfully."

clean:
	@echo "Cleaning binary and temporary files..."
	@[ -f $(DEFAULT_DB_PATH) ] && { echo "Removing $(DEFAULT_DB_PATH)..."; rm -f $(DEFAULT_DB_PATH); } || true
	@[ -d $(BINARY_PATH_PREFIX) ] && { echo "Removing $(BINARY_PATH_PREFIX)..."; rm -rf $(BINARY_PATH_PREFIX); } || true
	@find /tmp -name "$(BINARY_BASE_NAME)_test_*" -type d -exec rm -rf {} + 2>/dev/null || true
	@echo "Binary and temporary files cleaned successfully."

# Detect the operating system
OS := $(shell uname -s)

# Define terminal commands based on OS
ifeq ($(OS),Linux)
    # Check for common Linux terminal emulators in order of preference
    ifeq ($(shell command -v gnome-terminal),)
        ifeq ($(shell command -v kitty),)
            ifeq ($(shell command -v konsole),)
                ifeq ($(shell command -v xterm),)
                    TERMINAL := echo "No terminal emulator found."
                else
                    TERMINAL := xterm
                endif
            else
                TERMINAL := konsole
            endif
        else
            TERMINAL := kitty
        endif
    else
        TERMINAL := gnome-terminal
    endif
else ifeq ($(OS),Darwin)  # macOS
    TERMINAL := open -a Terminal
else  # Assume Windows if not Linux or macOS
    TERMINAL := powershell -Command "Start-Process powershell -ArgumentList '-NoExit', '-Command', 'Get-Content %1 -Wait'"
endif

# Help target
help:
	@echo "Thoth Network Makefile Help"
	@echo "===================="
	@echo "Main targets:"
	@echo "  all                  - Build and test the application"
	@echo "  build                - Build the application"
	@echo "  test                 - Run tests"
	@echo "  run                  - Run the application"
	@echo ""
	@echo "Cleaning targets:"
	@echo "  clean                - Clean binary and temporary files"
	@echo "  clean-all            - Clean all build artifacts including caches"

# Mark all targets as phony (not associated with files)
.PHONY: all build docker-run docker-down test run dev stop-dev monitor-logs build-binary \
	proto proto-deps proto-lint proto-format proto-gen proto-breaking proto-clean proto-verify \
	clean clean-all help