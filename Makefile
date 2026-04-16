# Makefile for PulseCat
# =====================

# Project configuration
BINARY_NAME := pulsecat
CLIENT_BINARY_NAME := pulsekitten
VERSION := 0.1.0
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go configuration
GO := go
GOFLAGS := -v
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.CommitHash=$(COMMIT_HASH)"
GOFILES := $(shell find . -name "*.go" -type f ! -path "./vendor/*")

# Directories
BUILD_DIR := ./build
DIST_DIR := ./dist
COVERAGE_DIR := ./coverage
PROTO_DIR := ./api/v1
API_DIR := ./pkg/api/v1
PROTO_FILE := $(PROTO_DIR)/pulsecat.proto
PROTO_GO_FILE := $(API_DIR)/pulsecat.pb.go
PROTO_GRPC_GO_FILE := $(API_DIR)/pulsecat_grpc.pb.go

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build: $(BUILD_DIR)/$(BINARY_NAME) $(BUILD_DIR)/$(CLIENT_BINARY_NAME)

$(BUILD_DIR)/$(BINARY_NAME): $(GOFILES)
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/pulsecat

$(BUILD_DIR)/$(CLIENT_BINARY_NAME): $(GOFILES)
	@echo "Building $(CLIENT_BINARY_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(CLIENT_BINARY_NAME) ./cmd/pulsekitten

# Build for production (stripped, optimized)
.PHONY: build-prod
build-prod:
	@echo "Building production binary..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-prod ./cmd/pulsecat
	$(GO) build $(GOFLAGS) -ldflags="-s -w" -o $(BUILD_DIR)/$(CLIENT_BINARY_NAME)-prod ./cmd/pulsekitten

# Install the binary to $GOPATH/bin
.PHONY: install
install:
	$(GO) install $(GOFLAGS) .

# Run the application
.PHONY: run
run:
	$(GO) run $(GOFLAGS) cmd/pulsecat/main.go

# Run with development settings (short intervals for testing)
.PHONY: run-dev
run-dev:
	$(GO) run $(GOFLAGS) cmd/pulsecat/main.go -start 2 -frequency 1 -log-level debug

# Run with specific port
.PHONY: run-port
run-port:
	$(GO) run $(GOFLAGS) cmd/pulsecat/main.go -port 8080 -log-level info

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR) $(DIST_DIR) $(COVERAGE_DIR)
	rm -f $(BINARY_NAME) $(BINARY_NAME)-prod $(CLIENT_BINARY_NAME) $(CLIENT_BINARY_NAME)-prod
	rm -f coverage.out coverage.html

# Test targets
.PHONY: test
test:
	@echo "Running tests..."
	$(GO) test $(GOFLAGS) ./... -race

.PHONY: test-verbose
test-verbose:
	$(GO) test $(GOFLAGS) ./... -v -race

.PHONY: test-coverage
test-coverage:
	@mkdir -p $(COVERAGE_DIR)
	$(GO) test $(GOFLAGS) ./... -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic
	$(GO) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report generated: $(COVERAGE_DIR)/coverage.html"

# Linting and code quality
.PHONY: lint
lint:
	@echo "Running golangci-lint..."
	golangci-lint run

.PHONY: lint-fix
lint-fix:
	golangci-lint run --fix

.PHONY: fmt
fmt:
	@echo "Formatting Go code..."
	$(GO) fmt ./...

.PHONY: vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

.PHONY: tidy
tidy:
	@echo "Tidying go.mod..."
	$(GO) mod tidy

# Dependency management
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download

.PHONY: deps-update
deps-update:
	@echo "Updating dependencies..."
	$(GO) get -u ./...
	$(GO) mod tidy

# Protobuf generation
.PHONY: proto
proto: $(PROTO_GO_FILE) $(PROTO_GRPC_GO_FILE)

$(PROTO_GO_FILE) $(PROTO_GRPC_GO_FILE): $(PROTO_FILE)
	@echo "Generating Go code from protobuf..."
	@mkdir -p $(API_DIR)
	protoc --go_out=./pkg --go_opt=paths=source_relative \
	       --go-grpc_out=./pkg --go-grpc_opt=paths=source_relative \
	       $(PROTO_FILE)
	@echo "Protobuf code generated successfully"

# Force regeneration of protobuf files
.PHONY: proto-force
proto-force:
	@echo "Force regenerating Go code from protobuf..."
	@mkdir -p $(API_DIR)
	protoc --go_out=./pkg --go_opt=paths=source_relative \
	       --go-grpc_out=./pkg --go-grpc_opt=paths=source_relative \
	       $(PROTO_FILE)
	@echo "Protobuf code generated successfully"

# Code generation (for protobuf, if needed)
.PHONY: generate
generate: proto
	@echo "Generating code..."
	$(GO) generate ./...

# Cross-compilation for multiple platforms
.PHONY: cross-build
cross-build:
	@echo "Cross-compiling for multiple platforms..."
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/pulsecat
	GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/pulsecat
	GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/pulsecat
	GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/pulsecat
	GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/pulsecat
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(CLIENT_BINARY_NAME)-linux-amd64 ./cmd/pulsekitten
	GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(CLIENT_BINARY_NAME)-linux-arm64 ./cmd/pulsekitten
	GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(CLIENT_BINARY_NAME)-darwin-amd64 ./cmd/pulsekitten
	GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(CLIENT_BINARY_NAME)-darwin-arm64 ./cmd/pulsekitten
	GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(CLIENT_BINARY_NAME)-windows-amd64.exe ./cmd/pulsekitten

# Docker targets
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) -t $(BINARY_NAME):latest .

.PHONY: docker-run
docker-run:
	docker run -p 25225:25225 --rm $(BINARY_NAME):latest

.PHONY: docker-clean
docker-clean:
	docker rmi $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest 2>/dev/null || true

# Development workflow
.PHONY: dev
dev: deps fmt vet lint test build

.PHONY: ci
ci: deps fmt vet lint test build-prod

# Help target
.PHONY: help
help:
	@echo "PulseCat Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  all           - Default target, builds both server and client binaries"
	@echo "  build         - Build both server (pulsecat) and client (pulsekitten) binaries"
	@echo "  build-prod    - Build optimized production binaries"
	@echo "  install       - Install to GOPATH/bin"
	@echo "  run           - Run the server with default settings"
	@echo "  run-dev       - Run server with development settings (short intervals)"
	@echo "  run-port      - Run server on specific port (8080)"
	@echo "  clean         - Remove build artifacts"
	@echo ""
	@echo "Testing:"
	@echo "  test          - Run tests"
	@echo "  test-verbose  - Run tests with verbose output"
	@echo "  test-coverage - Generate test coverage report"
	@echo ""
	@echo "Code quality:"
	@echo "  lint          - Run golangci-lint"
	@echo "  lint-fix      - Run golangci-lint with auto-fix"
	@echo "  fmt           - Format Go code"
	@echo "  vet           - Run go vet"
	@echo "  tidy          - Tidy go.mod"
	@echo ""
	@echo "Dependencies:"
	@echo "  deps          - Download dependencies"
	@echo "  deps-update   - Update all dependencies"
	@echo ""
	@echo "Cross-compilation:"
	@echo "  cross-build   - Build for multiple platforms (both server and client)"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build  - Build Docker image for server"
	@echo "  docker-run    - Run Docker container"
	@echo "  docker-clean  - Remove Docker images"
	@echo ""
	@echo "Development:"
	@echo "  dev           - Full development workflow (deps, fmt, vet, lint, test, build)"
	@echo "  ci            - CI workflow (deps, fmt, vet, lint, test, build-prod)"
	@echo ""
	@echo "  help          - Show this help message"

# Default target
.DEFAULT_GOAL := help