# Makefile for jem - Java Environment Manager
# This Makefile handles versioning, building, testing, and releasing

# ============================================
# VERSIONING (SemVer 2.0.0)
# ============================================

# Version from git tag, fallback to dev
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# LDFLAGS for embedding version in binary
LDFLAGS := -ldflags "-X main.Version=$(VERSION)"

# Entry point
ENTRY := ./cmd/jem

# ============================================
# BUILD
# ============================================

# Binary name
BINARY := jem

# Build the binary
build:
	@echo "Building jem $(VERSION)..."
	go build $(LDFLAGS) -o $(BINARY) $(ENTRY)

# Build for specific platform
build-linux:
	@echo "Building jem $(VERSION) for Linux..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-linux-amd64 $(ENTRY)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY)-linux-arm64 $(ENTRY)

build-darwin:
	@echo "Building jem $(VERSION) for macOS..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-darwin-amd64 $(ENTRY)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY)-darwin-arm64 $(ENTRY)

build-windows:
	@echo "Building jem $(VERSION) for Windows..."
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-windows-amd64.exe $(ENTRY)
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY)-windows-arm64.exe $(ENTRY)

# ============================================
# INSTALL
# ============================================

install:
	@echo "Installing jem $(VERSION)..."
	go install $(LDFLAGS) $(ENTRY)

# ============================================
# TEST
# ============================================

# Run all tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-cover:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# Run tests and show coverage by package
test-cover-summary:
	@echo "Running tests with coverage summary..."
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | grep "total:" | awk '{print "Overall coverage: " $$3}'

# ============================================
# LINTING
# ============================================

lint:
	@echo "Running linter..."
	go vet ./...
	go fmt ./...

# ============================================
# VERSIONING COMMANDS
# ============================================

# Show current version
version:
	@echo "Version: $(VERSION)"

# ============================================
# CLEANUP
# ============================================

clean:
	@echo "Cleaning..."
	go clean -cache -testcache
	rm -f $(BINARY) $(BINARY)-*
	rm -f coverage.out
	rm -f *.exe

# ============================================
# HELP
# ============================================

help:
	@echo "jem Makefile - Java Environment Manager"
	@echo ""
	@echo "Available commands:"
	@echo "  make build              Build the binary"
	@echo "  make install            Install via go install"
	@echo "  make test               Run all tests"
	@echo "  make test-cover         Run tests with coverage"
	@echo "  make lint               Run linter"
	@echo "  make version            Show current version"
	@echo "  make clean              Clean build artifacts"
	@echo "  make help               Show this help"
	@echo ""
	@echo "Cross-compilation:"
	@echo "  make build-linux        Build for Linux (amd64, arm64)"
	@echo "  make build-darwin       Build for macOS (amd64, arm64)"
	@echo "  make build-windows      Build for Windows (amd64, arm64)"
	@echo ""
	@echo "Current version: $(VERSION)"

# ============================================
# DEFAULT TARGET
# ============================================

default: build
