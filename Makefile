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
BINARY_NAME_SERVER=$(BINARY_NAME)_server
BINARY_NAME_CLI=$(BINARY_NAME)_cli
DOCKER_IMAGE=$(BINARY_NAME)

# Binary path
BINARY_PATH_PREFIX := ./bin
BINARY_PATH := $(BINARY_PATH_PREFIX)/$(GOOS)_$(GOARCH)

# Module path
MODULE_PATH := ./cmd/$(BINARY_BASE_NAME).go

# Protocol buffer variables
PROTO_DIR := .
PROTO_GEN_DIR := pkgs/gen

# Default target
.DEFAULT_GOAL := build

# Temporary paths
TEMP_PATHS := /tmp/$(BINARY_BASE_NAME)_test_*

# Build the application
all: test build

build:
	@echo "Building..."
	@CGO_ENABLED=1 GOOS=linux go build -o main $(MODULE_PATH)

test:
	$(GOTEST) -v ./...

run: build
	./$(BINARY_NAME_SERVER)

deps:
	$(GOMOD) download

tidy:
	$(GOMOD) tidy

docker-build:
	docker build -t $(DOCKER_IMAGE) .

# Docker targets
docker-run:
	@echo "Starting Docker container..."
	@if docker compose up --build 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up --build; \
	fi

docker-down:
	@echo "Stopping Docker container..."
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

lint: vet
	golangci-lint run ./...
	@echo "Linting completed."

vet:
	go vet ./...
	@echo "Vet completed."

# Generate mocks for testing
mocks:
	mockery --all --dir=internal --output=internal/mocks

# Run server with hot reload

# Development environment targets
dev:
	@echo "Starting development environment..."
	@command -v air >/dev/null 2>&1 || { echo "Installing air..."; go install github.com/air-verse/air@latest; }
	@mkdir -p tmp
	@air -c .air.toml > $(BINARY_NAME).log 2>&1 & echo $$! > $(BINARY_NAME).pid
	@echo "Development environment started. Logs are in $(BINARY_NAME).log."
	@echo "Use 'make stop-dev' to stop the development environment."

stop-dev:
	@echo "Stopping development environment..."
	@if [ -f $(BINARY_NAME).pid ]; then \
		if ps -p $$(cat $(BINARY_NAME).pid) > /dev/null; then \
			echo "Stopping process..."; \
			kill $$(cat $(BINARY_NAME).pid) && rm $(BINARY_NAME).pid; \
		else \
			echo "$(BINARY_NAME) process not running. Removing stale PID file."; \
			rm $(BINARY_NAME).pid; \
		fi \
	else \
		echo "No PID file found. Development environment is not running."; \
	fi
	@echo "Development environment stopped."

monitor-logs:
	@if [ -f $(BINARY_NAME).log ]; then \
		echo "Monitoring logs..."; \
		$(TERMINAL) -e "tail -f $(BINARY_NAME).log" & \
	else \
		echo "Log file not found. Start the development environment first."; \
	fi
	@echo "Log monitoring stopped."
	@echo "Use 'make stop-dev' to stop the development environment."

# Build binary
build-binary:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BINARY_PATH)
	@go build $(BUILD_FLAGS) -o $(BINARY_PATH)/$(BINARY_NAME) $(MODULE_PATH)
	@echo "Binary built at $(BINARY_PATH)/$(BINARY_NAME)"

# Protocol buffer targets
proto-deps:
	@echo "Installing protocol buffer tools..."
	@go install github.com/bufbuild/buf/cmd/buf@latest
	@go get github.com/bufbuild/protovalidate-go
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	@echo "Protocol buffer tools installed successfully."

proto-lint:
	@echo "Linting protocol buffer files..."
	@cd $(PROTO_DIR) && \
	lint_output=$$(buf lint 2>&1); \
	if [ -n "$$lint_output" ]; then \
		echo "Error: Linting issues found!"; \
		echo "$$lint_output"; \
		exit 1; \
	else \
		echo "Linting passed successfully."; \
	fi

proto-format:
	@echo "Formatting protocol buffer files..."
	@cd $(PROTO_DIR) && buf format -w
	@echo "Protocol buffer files formatted successfully."

proto-gen:
	@echo "Generating code from protocol buffer files..."
	@cd $(PROTO_DIR) && \
	gen_output=$$(buf generate 2>&1); \
	if [ $$? -ne 0 ]; then \
		echo "Error generating code from protocol buffers!"; \
		echo "$$gen_output"; \
		exit 1; \
	else \
		echo "Protocol buffer code generated successfully."; \
	fi

proto-breaking:
	@echo "Checking for breaking changes in protocol buffer files..."
	@cd $(PROTO_DIR) && \
	if [ -d .git ]; then \
		break_output=$$(buf breaking --against '.git#branch=main' 2>&1); \
		if [ $$? -ne 0 ]; then \
			echo "Warning: Breaking changes detected!"; \
			echo "$$break_output"; \
		else \
			echo "No breaking changes detected."; \
		fi \
	else \
		echo "Warning: Git repository not found. Skipping breaking change check."; \
	fi

proto-clean:
	@echo "Cleaning generated protocol buffer files..."
	@find . -type d -name $(PROTO_GEN_DIR) -exec rm -rf {} + 2>/dev/null || true
	@echo "Protocol buffer generated files cleaned successfully."

# Main protocol buffer target that runs all steps
proto: proto-deps proto-format proto-lint proto-gen
	@echo "Protocol buffer processing complete."

# Verification target that doesn't generate files
proto-verify: proto-format proto-lint proto-breaking
	@echo "Protocol buffer verification complete."

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
	@echo "Development targets:"
	@echo "  dev                  - Start development environment with hot reloading"
	@echo "  stop-dev             - Stop development environment"
	@echo "  monitor-logs         - Monitor logs from development environment"
	@echo ""
	@echo "Docker targets:" 
	@echo "  docker-run           - Start Docker container"
	@echo "  docker-down          - Stop Docker container"
	@echo ""
	@echo "Protocol buffer targets:"
	@echo "  proto                - Run all protocol buffer steps (deps, format, lint, gen)"
	@echo "  proto-verify         - Verify protocol buffer files without generating code"
	@echo "  proto-deps           - Install protocol buffer tools"
	@echo "  proto-format         - Format protocol buffer files"
	@echo "  proto-lint           - Lint protocol buffer files"
	@echo "  proto-gen            - Generate code from protocol buffer files"
	@echo "  proto-breaking       - Check for breaking changes in protocol buffer files"
	@echo "  proto-clean          - Clean generated protocol buffer files"
	@echo ""
	@echo "Cleaning targets:"
	@echo "  clean                - Clean binary and temporary files"
	@echo "  clean-all            - Clean all build artifacts including caches"

# Mark all targets as phony (not associated with files)
.PHONY: all build docker-run docker-down test run dev stop-dev monitor-logs build-binary \
	proto proto-deps proto-lint proto-format proto-gen proto-breaking proto-clean proto-verify \
	clean clean-all help