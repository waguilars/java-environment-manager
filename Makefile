# Makefile for jem - Java Environment Manager
# This Makefile handles versioning, building, testing, and releasing

# ============================================
# VERSIONING (SemVer 2.0.0)
# ============================================

# Version source: Use VERSION file, fallback to Makefile variable, then git tag
VERSION_FILE := $(shell cat VERSION 2>/dev/null || echo "")
VERSION ?= 0.2.0-beta

# Use VERSION file if available, otherwise use Makefile version
ifeq ($(VERSION_FILE),)
  VERSION_DISPLAY := $(VERSION)
else
  VERSION_DISPLAY := $(VERSION_FILE)
endif

# LDFLAGS for embedding version in binary
LDFLAGS := -ldflags "-X main.Version=$(VERSION_DISPLAY)"

# ============================================
# BUILD
# ============================================

# Binary name
BINARY := jem

# Build the binary
build: clean
	@echo "Building jem v$(VERSION_DISPLAY)..."
	go build $(LDFLAGS) -o $(BINARY) .

# Build for specific platform
build-linux: clean
	@echo "Building jem v$(VERSION_DISPLAY) for Linux..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY)-linux-arm64 .

build-darwin: clean
	@echo "Building jem v$(VERSION_DISPLAY) for macOS..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY)-darwin-arm64 .

build-windows: clean
	@echo "Building jem v$(VERSION_DISPLAY) for Windows..."
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY)-windows-arm64.exe .

# ============================================
# INSTALL
# ============================================

install: build
	@echo "Installing jem v$(VERSION_DISPLAY)..."
	go install $(LDFLAGS) .

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
	@echo "Current version: $(VERSION_DISPLAY)"
	@echo "Git version: $(GIT_VERSION)"
	@echo "Makefile version: $(VERSION)"

# Bump version (manual)
# Usage: make bump-major, make bump-minor, make bump-patch
bump-major:
	@echo "Bumping MAJOR version..."
	@awk -F. '{printf "%d.%d.%d", $$1+1, 0, 0}' VERSION > VERSION.new && mv VERSION.new VERSION
	@make version

bump-minor:
	@echo "Bumping MINOR version..."
	@awk -F. '{printf "%d.%d.%d", $$1, $$2+1, 0}' VERSION > VERSION.new && mv VERSION.new VERSION
	@make version

bump-patch:
	@echo "Bumping PATCH version..."
	@awk -F. '{printf "%d.%d.%d", $$1, $$2, $$3+1}' VERSION > VERSION.new && mv VERSION.new VERSION
	@make version

# ============================================
# RELEASE
# ============================================

# Create a new release
# Usage: make release version=1.0.0
release:
	@if [ -z "$(version)" ]; then \
		echo "ERROR: version is required"; \
		echo "Usage: make release version=1.0.0"; \
		exit 1; \
	fi
	@echo "Creating release v$(version)..."
	@# Update version file
	echo "$(version)" > VERSION
	@# Create git tag
	git tag -a "v$(version)" -m "Release v$(version)"
	@# Build binaries
	$(MAKE) build
	@# Create release assets (optional)
	@# ./create-release-assets.sh
	@echo "Release v$(version) created!"
	@echo "To publish, run: git push origin v$(version)"

# Pre-release (for beta/alpha versions)
# Usage: make prerelease version=1.0.0-beta
prerelease:
	@if [ -z "$(version)" ]; then \
		echo "ERROR: version is required"; \
		echo "Usage: make prerelease version=1.0.0-beta"; \
		exit 1; \
	fi
	@echo "Creating pre-release v$(version)..."
	@echo "$(version)" > VERSION
	git tag -a "v$(version)" -m "Pre-release v$(version)"
	$(MAKE) build
	@echo "Pre-release v$(version) created!"
	@echo "To publish, run: git push origin v$(version)"

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
	@echo "  make install            Install to GOPATH"
	@echo "  make test               Run all tests"
	@echo "  make test-cover         Run tests with coverage"
	@echo "  make lint               Run linter"
	@echo "  make version            Show current version"
	@echo "  make bump-major         Bump MAJOR version"
	@echo "  make bump-minor         Bump MINOR version"
	@echo "  make bump-patch         Bump PATCH version"
	@echo "  make release version=X  Create a release"
	@echo "  make prerelease version=X  Create a pre-release"
	@echo "  make clean              Clean build artifacts"
	@echo "  make help               Show this help"
	@echo ""
	@echo "Current version: $(VERSION_DISPLAY)"
	@echo "Git version: $(GIT_VERSION)"

# ============================================
# DEFAULT TARGET
# ============================================

default: build
