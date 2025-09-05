# Vulkan-Go Makefile

.PHONY: all build test clean install generate update shaders example docs lint fmt vet

# Variables
PROJECT_NAME := vulkan-go
BINARY_NAME := vulkan-go
BUILD_DIR := build
GO_FILES := $(shell find . -name '*.go' -type f)

# Build configuration
GO_BUILD_FLAGS := -ldflags="-s -w"
GO_TEST_FLAGS := -race -coverprofile=coverage.out

# Default target
all: fmt lint vet test build

# Build the project
build:
	@echo "Building $(PROJECT_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/triangle
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run tests
test:
	@echo "Running tests..."
	@go test $(GO_TEST_FLAGS) ./...
	@echo "Tests complete"

# Run tests with coverage
test-coverage: test
	@echo "Generating coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -rf .temp
	@rm -rf .backup
	@rm -f coverage.out coverage.html
	@rm -rf assets/shaders/compiled
	@echo "Clean complete"

# Install dependencies
install:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies installed"

# Generate Vulkan bindings
generate:
	@echo "Generating Vulkan bindings..."
	@go run scripts/generate.go
	@echo "Generation complete"

# Update Vulkan API definitions
update:
	@echo "Updating Vulkan API definitions..."
	@chmod +x scripts/update.sh
	@./scripts/update.sh
	@echo "Update complete"

# Compile shaders
shaders:
	@echo "Compiling shaders..."
	@chmod +x assets/shaders/compile.sh
	@cd assets/shaders && ./compile.sh
	@echo "Shader compilation complete"

# Build and run triangle example
example: shaders build
	@echo "Running triangle example..."
	@$(BUILD_DIR)/$(BINARY_NAME)

# Build triangle example specifically
triangle: shaders
	@echo "Building triangle example..."
	@mkdir -p $(BUILD_DIR)
	@go build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/triangle ./cmd/triangle
	@echo "Triangle example built: $(BUILD_DIR)/triangle"

# Build compute example
compute:
	@echo "Building compute example..."
	@mkdir -p $(BUILD_DIR)
	@go build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/compute ./cmd/compute
	@echo "Compute example built: $(BUILD_DIR)/compute"

# Generate documentation
docs:
	@echo "Generating documentation..."
	@go doc -all ./pkg/vk > docs/api.md
	@go doc -all ./pkg/vulkan > docs/bindings.md
	@echo "Documentation generated"

# Lint the code
lint:
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping..."; \
	fi

# Format the code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Code formatted"

# Vet the code
vet:
	@echo "Vetting code..."
	@go vet ./...
	@echo "Code vetted"

# Development setup
dev-setup: install
	@echo "Setting up development environment..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Development setup complete"

# Full development build
dev: clean fmt lint vet test shaders build
	@echo "Development build complete"

# Release build
release: clean update generate fmt lint vet test shaders
	@echo "Building release binaries..."
	@mkdir -p $(BUILD_DIR)/release
	
	# Build for multiple platforms
	@GOOS=windows GOARCH=amd64 go build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-windows-amd64.exe ./cmd/triangle
	@GOOS=linux GOARCH=amd64 go build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-amd64 ./cmd/triangle
	@GOOS=darwin GOARCH=amd64 go build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-amd64 ./cmd/triangle
	@GOOS=darwin GOARCH=arm64 go build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-arm64 ./cmd/triangle
	
	@echo "Release build complete"
	@ls -la $(BUILD_DIR)/release/

# Benchmark tests
bench:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

# Memory profiling
profile:
	@echo "Running memory profile..."
	@go test -memprofile=mem.prof -bench=. ./...
	@go tool pprof mem.prof

# Check for vulnerabilities
vuln:
	@echo "Checking for vulnerabilities..."
	@go run golang.org/x/vuln/cmd/govulncheck@latest ./...

# Tidy modules
tidy:
	@echo "Tidying Go modules..."
	@go mod tidy
	@echo "Modules tidied"

# Show project statistics
stats:
	@echo "Project Statistics:"
	@echo "=================="
	@echo "Lines of Go code:"
	@find . -name '*.go' -not -path './vendor/*' | xargs wc -l | tail -n1
	@echo ""
	@echo "Number of Go files:"
	@find . -name '*.go' -not -path './vendor/*' | wc -l
	@echo ""
	@echo "Number of packages:"
	@go list ./... | wc -l
	@echo ""
	@echo "Dependencies:"
	@go list -m all | wc -l

# Help target
help:
	@echo "Available targets:"
	@echo "  all          - Format, lint, vet, test, and build"
	@echo "  build        - Build the triangle example"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage report"
	@echo "  clean        - Clean build artifacts"
	@echo "  install      - Install dependencies"
	@echo "  generate     - Generate Vulkan bindings"
	@echo "  update       - Update Vulkan API definitions"
	@echo "  shaders      - Compile SPIR-V shaders"
	@echo "  example      - Build and run triangle example"
	@echo "  triangle     - Build triangle example"
	@echo "  compute      - Build compute example"
	@echo "  docs         - Generate documentation"
	@echo "  lint         - Run linters"
	@echo "  fmt          - Format code"
	@echo "  vet          - Vet code"
	@echo "  dev-setup    - Setup development environment"
	@echo "  dev          - Full development build"
	@echo "  release      - Build release binaries"
	@echo "  bench        - Run benchmarks"
	@echo "  profile      - Run memory profiling"
	@echo "  vuln         - Check for vulnerabilities"
	@echo "  tidy         - Tidy Go modules"
	@echo "  stats        - Show project statistics"
	@echo "  help         - Show this help message"