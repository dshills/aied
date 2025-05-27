# AIED Makefile

# Variables
BINARY_NAME=aied
MAIN_PACKAGE=.
GO=go
GOFLAGS=
LDFLAGS=-s -w

# Version info
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
BUILD_FLAGS=-ldflags "$(LDFLAGS) -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Platforms
PLATFORMS=darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64

.PHONY: all build clean test coverage lint fmt vet run install help

# Default target
all: test build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) $(BUILD_FLAGS) -o $(BINARY_NAME) $(MAIN_PACKAGE)

# Run the application
run: build
	./$(BINARY_NAME)

# Install the binary
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install $(GOFLAGS) $(BUILD_FLAGS) $(MAIN_PACKAGE)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -rf dist/
	@$(GO) clean -cache

# Run tests
test:
	@echo "Running tests..."
	$(GO) test $(GOFLAGS) -race -short ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	$(GO) test $(GOFLAGS) -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GO) test $(GOFLAGS) -bench=. -benchmem ./...

# Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

# Run linters
lint: fmt vet
	@echo "Running golangci-lint..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found, install from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...

# Update dependencies
deps:
	@echo "Updating dependencies..."
	$(GO) mod download
	$(GO) mod tidy
	$(GO) mod verify

# Generate example config
example-config:
	@echo "Generating example configuration..."
	@mkdir -p examples
	$(GO) run $(MAIN_PACKAGE) -generate-config > examples/config.yaml

# Build for all platforms
build-all: clean
	@echo "Building for all platforms..."
	@mkdir -p dist
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} \
		$(GO) build $(GOFLAGS) $(BUILD_FLAGS) \
		-o dist/$(BINARY_NAME)-$${platform%/*}-$${platform#*/}$(if $(findstring windows,$${platform}),.exe,) \
		$(MAIN_PACKAGE); \
		echo "Built: dist/$(BINARY_NAME)-$${platform%/*}-$${platform#*/}$(if $(findstring windows,$${platform}),.exe,)"; \
	done

# Create release archives
release: build-all
	@echo "Creating release archives..."
	@cd dist && for file in *; do \
		if [[ "$$file" == *.exe ]]; then \
			zip "$${file%.exe}.zip" "$$file" ../README.md ../LICENSE ../CHANGELOG.md ../.aied.yaml.example; \
		else \
			tar czf "$$file.tar.gz" "$$file" -C .. README.md LICENSE CHANGELOG.md .aied.yaml.example; \
		fi; \
		echo "Created: dist/$$file archive"; \
	done

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	$(GO) mod download
	@echo "Installing development tools..."
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Development setup complete!"

# Quick check before committing
pre-commit: fmt vet test
	@echo "Pre-commit checks passed!"

# Help target
help:
	@echo "AIED Makefile Commands:"
	@echo "  make build       - Build the binary"
	@echo "  make run         - Build and run the application"
	@echo "  make install     - Install the binary to GOPATH/bin"
	@echo "  make test        - Run tests"
	@echo "  make coverage    - Run tests with coverage report"
	@echo "  make bench       - Run benchmarks"
	@echo "  make clean       - Remove build artifacts"
	@echo "  make fmt         - Format code"
	@echo "  make vet         - Run go vet"
	@echo "  make lint        - Run linters"
	@echo "  make deps        - Update dependencies"
	@echo "  make build-all   - Build for all platforms"
	@echo "  make release     - Create release archives"
	@echo "  make dev-setup   - Set up development environment"
	@echo "  make pre-commit  - Run pre-commit checks"
	@echo "  make help        - Show this help message"