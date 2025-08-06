# GopherTales Makefile
# Common development and deployment tasks

.PHONY: help build run test clean dev docker lint fmt vet deps check install uninstall

# Default target
help: ## Show this help message
	@echo "GopherTales - Interactive Adventure Game"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build configuration
BINARY_NAME=gophertales
BUILD_DIR=bin
CMD_DIR=cmd/server
MAIN_FILE=$(CMD_DIR)/main.go

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet

# Build flags
LDFLAGS=-ldflags "-s -w"
BUILD_FLAGS=-a -installsuffix cgo

# Development
dev: ## Run the application in development mode with auto-reload
	@echo "Starting GopherTales in development mode..."
	@air -c .air.toml || $(GOCMD) run $(MAIN_FILE)

run: ## Run the application
	@echo "Starting GopherTales..."
	@$(GOCMD) run $(MAIN_FILE)

# Building
build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)

build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_FILE)
	@GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_FILE)
	@GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_FILE)
	@GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_FILE)
	@echo "Built binaries:"
	@ls -la $(BUILD_DIR)/

build-prod: ## Build optimized production binary
	@echo "Building production binary..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)

# Installation
install: build ## Build and install the binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	@$(GOCMD) install $(MAIN_FILE)

uninstall: ## Remove the installed binary
	@echo "Removing $(BINARY_NAME)..."
	@rm -f $(GOPATH)/bin/$(BINARY_NAME)

# Testing
test: ## Run tests
	@echo "Running tests..."
	@$(GOTEST) -v ./...

test-cover: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@$(GOTEST) -cover ./...

test-race: ## Run tests with race detection
	@echo "Running tests with race detection..."
	@$(GOTEST) -race ./...

test-bench: ## Run benchmark tests
	@echo "Running benchmark tests..."
	@$(GOTEST) -bench=. ./...

test-all: test test-race test-cover ## Run all tests

# Code quality
fmt: ## Format Go code
	@echo "Formatting code..."
	@$(GOFMT) -s -w .

vet: ## Run go vet
	@echo "Running go vet..."
	@$(GOVET) ./...

lint: ## Run golangci-lint (requires golangci-lint to be installed)
	@echo "Running linter..."
	@golangci-lint run

check: fmt vet lint ## Run all code quality checks

# Dependencies
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@$(GOMOD) download

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	@$(GOMOD) tidy

deps-vendor: ## Vendor dependencies
	@echo "Vendoring dependencies..."
	@$(GOMOD) vendor

# Docker
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(BINARY_NAME):latest .

docker-run: docker-build ## Build and run Docker container
	@echo "Running Docker container..."
	@docker run -p 8000:8000 --rm $(BINARY_NAME):latest

docker-compose-dev: ## Run with docker-compose for development
	@echo "Starting development environment with docker-compose..."
	@docker-compose up --build

docker-compose-prod: ## Run with docker-compose for production
	@echo "Starting production environment with docker-compose..."
	@docker-compose --profile production up -d

docker-stop: ## Stop docker-compose services
	@echo "Stopping docker-compose services..."
	@docker-compose down

# Cleanup
clean: ## Clean build artifacts
	@echo "Cleaning up..."
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -rf vendor/

clean-docker: ## Clean Docker images and containers
	@echo "Cleaning Docker artifacts..."
	@docker-compose down --rmi all --volumes --remove-orphans
	@docker system prune -f

# Development tools
setup-dev: ## Setup development environment
	@echo "Setting up development environment..."
	@$(GOGET) -u github.com/cosmtrek/air@latest
	@$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Development tools installed!"

# Release
release: clean test build-all ## Create a release build
	@echo "Creating release..."
	@mkdir -p releases
	@tar -czf releases/$(BINARY_NAME)-linux-amd64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-linux-amd64
	@tar -czf releases/$(BINARY_NAME)-darwin-amd64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-darwin-amd64
	@tar -czf releases/$(BINARY_NAME)-darwin-arm64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-darwin-arm64
	@zip -j releases/$(BINARY_NAME)-windows-amd64.zip $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe
	@echo "Release packages created in releases/"

# Monitoring and health
health: ## Check application health
	@echo "Checking application health..."
	@curl -f http://localhost:8000/api/health || echo "Service is not running"

stats: ## Get application statistics
	@echo "Getting application statistics..."
	@curl -s http://localhost:8000/api/stats | python -m json.tool || echo "Service is not running"

# Database/Story management
validate-story: ## Validate story JSON format
	@echo "Validating story format..."
	@python -m json.tool gopher.json > /dev/null && echo "Story JSON is valid" || echo "Story JSON is invalid"

# Load testing
load-test: ## Run load test (requires apache bench)
	@echo "Running load test..."
	@ab -n 1000 -c 10 -k http://localhost:8000/

# Git hooks
git-hooks: ## Install git hooks
	@echo "Installing git hooks..."
	@cp scripts/pre-commit .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Git hooks installed!"

# Help
version: ## Show Go version
	@$(GOCMD) version

info: ## Show project information
	@echo "Project: GopherTales"
	@echo "Description: Interactive Adventure Game"
	@echo "Go Version: $$(go version)"
	@echo "Build Directory: $(BUILD_DIR)"
	@echo "Binary Name: $(BINARY_NAME)"
