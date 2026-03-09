#!/bin/bash
# release.sh - Cross-platform release script for jem
# Works on Linux, macOS, and Windows (with Git Bash or WSL)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if version is provided
if [ -z "$1" ]; then
    echo -e "${RED}ERROR: Version is required${NC}"
    echo ""
    echo "Usage: $0 <version> [type]"
    echo ""
    echo "Examples:"
    echo "  $0 0.2.0-beta      # Pre-release"
    echo "  $0 0.2.0           # Stable release"
    echo "  $0 0.2.1           # Patch release"
    echo ""
    echo "Types:"
    echo "  major    - Bump MAJOR version (X.0.0)"
    echo "  minor    - Bump MINOR version (X.Y.0)"
    echo "  patch    - Bump PATCH version (X.Y.Z)"
    echo "  beta     - Create beta pre-release"
    echo "  rc       - Create release candidate"
    echo "  alpha    - Create alpha pre-release"
    echo ""
    echo "Version format: X.Y.Z[-pre-release][+build]"
    echo "Examples: 0.2.0, 0.2.0-beta, 0.2.1-rc.1"
    exit 1
fi

VERSION=$1
TYPE=${2:-}

# Validate version format (basic SemVer check)
if ! echo "$VERSION" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)?(\+[a-zA-Z0-9.]+)?$'; then
    echo -e "${RED}ERROR: Invalid version format${NC}"
    echo "Expected: X.Y.Z[-pre-release][+build]"
    echo "Examples: 0.2.0, 0.2.0-beta, 0.2.1-rc.1"
    exit 1
fi

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Starting release process for v${VERSION}...${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# Step 1: Update VERSION file
echo -e "${YELLOW}Step 1: Updating VERSION file...${NC}"
echo "$VERSION" > VERSION
echo -e "${GREEN}✓ VERSION file updated to $VERSION${NC}"
echo ""

# Step 2: Update main.go version
echo -e "${YELLOW}Step 2: Updating main.go version...${NC}"
perl -i -pe 's/var Version = ".*"/var Version = "'"$VERSION"'"/' main.go
echo -e "${GREEN}✓ main.go updated${NC}"
echo ""

# Step 3: Update cmd/root.go version
echo -e "${YELLOW}Step 3: Updating cmd/root.go version...${NC}"
perl -i -pe 's/Version: "[^"]*"/Version: "'"$VERSION"'"/' cmd/root.go
echo -e "${GREEN}✓ cmd/root.go updated${NC}"
echo ""

# Step 3.5: Clean up any backup files
echo -e "${YELLOW}Step 3.5: Cleaning up backup files...${NC}"
rm -f *.bak cmd/*.bak
echo -e "${GREEN}✓ Backup files cleaned${NC}"
echo ""

# Step 4: Run tests
echo -e "${YELLOW}Step 4: Running tests...${NC}"
if ! go test -v ./...; then
    echo -e "${RED}ERROR: Tests failed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ All tests passed${NC}"
echo ""

# Step 5: Build binaries
echo -e "${YELLOW}Step 5: Building binaries...${NC}"
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
    go build -ldflags "-X main.Version=$VERSION" -o "jem.exe" .
else
    go build -ldflags "-X main.Version=$VERSION" -o "jem" .
fi
echo -e "${GREEN}✓ Binary built successfully${NC}"
echo ""

# Step 6: Create git tag
echo -e "${YELLOW}Step 6: Creating git tag...${NC}"
git add VERSION main.go cmd/root.go
git commit -m "Release v${VERSION}"
git tag -a "v${VERSION}" -m "Release v${VERSION}"
echo -e "${GREEN}✓ Git tag v${VERSION} created${NC}"
echo ""

# Step 7: Build release assets
echo -e "${YELLOW}Step 7: Building release assets...${NC}"
echo ""

# Windows
echo -e "${BLUE}Building Windows binaries...${NC}"
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
    go build -ldflags "-X main.Version=$VERSION" -o "jem-windows-amd64.exe" .
    go build -ldflags "-X main.Version=$VERSION" -o "jem-windows-arm64.exe" .
else
    GOOS=windows GOARCH=amd64 go build -ldflags "-X main.Version=$VERSION" -o "jem-windows-amd64.exe" .
    GOOS=windows GOARCH=arm64 go build -ldflags "-X main.Version=$VERSION" -o "jem-windows-arm64.exe" .
fi
echo -e "${GREEN}✓ Windows binaries built${NC}"
echo ""

# Linux
echo -e "${BLUE}Building Linux binaries...${NC}"
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=$VERSION" -o "jem-linux-amd64" .
GOOS=linux GOARCH=arm64 go build -ldflags "-X main.Version=$VERSION" -o "jem-linux-arm64" .
echo -e "${GREEN}✓ Linux binaries built${NC}"
echo ""

# macOS
echo -e "${BLUE}Building macOS binaries...${NC}"
GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.Version=$VERSION" -o "jem-darwin-amd64" .
GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.Version=$VERSION" -o "jem-darwin-arm64" .
echo -e "${GREEN}✓ macOS binaries built${NC}"
echo ""

# Step 8: Verify builds
echo -e "${YELLOW}Step 8: Verifying builds...${NC}"
echo -e "${BLUE}Verifying Windows binaries...${NC}"
if [[ -f "jem-windows-amd64.exe" ]]; then
    echo -e "${GREEN}✓ jem-windows-amd64.exe exists${NC}"
else
    echo -e "${RED}✗ jem-windows-amd64.exe not found${NC}"
    exit 1
fi

if [[ -f "jem-windows-arm64.exe" ]]; then
    echo -e "${GREEN}✓ jem-windows-arm64.exe exists${NC}"
else
    echo -e "${RED}✗ jem-windows-arm64.exe not found${NC}"
    exit 1
fi

echo -e "${BLUE}Verifying Linux binaries...${NC}"
if [[ -f "jem-linux-amd64" ]]; then
    echo -e "${GREEN}✓ jem-linux-amd64 exists${NC}"
else
    echo -e "${RED}✗ jem-linux-amd64 not found${NC}"
    exit 1
fi

if [[ -f "jem-linux-arm64" ]]; then
    echo -e "${GREEN}✓ jem-linux-arm64 exists${NC}"
else
    echo -e "${RED}✗ jem-linux-arm64 not found${NC}"
    exit 1
fi

echo -e "${BLUE}Verifying macOS binaries...${NC}"
if [[ -f "jem-darwin-amd64" ]]; then
    echo -e "${GREEN}✓ jem-darwin-amd64 exists${NC}"
else
    echo -e "${RED}✗ jem-darwin-amd64 not found${NC}"
    exit 1
fi

if [[ -f "jem-darwin-arm64" ]]; then
    echo -e "${GREEN}✓ jem-darwin-arm64 exists${NC}"
else
    echo -e "${RED}✗ jem-darwin-arm64 not found${NC}"
    exit 1
fi

echo ""

# Summary
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Release v${VERSION} preparation complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${YELLOW}Generated assets:${NC}"
echo -e "  ${BLUE}Windows:${NC} jem-windows-amd64.exe, jem-windows-arm64.exe"
echo -e "  ${BLUE}Linux:${NC}   jem-linux-amd64, jem-linux-arm64"
echo -e "  ${BLUE}macOS:${NC}   jem-darwin-amd64, jem-darwin-arm64"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo -e "  1. Review changes: ${BLUE}git diff HEAD~1${NC}"
echo -e "  2. Push to remote: ${BLUE}git push origin main --tags${NC}"
echo -e "  3. Create GitHub release with assets:"
echo -e "     ${BLUE}- jem-windows-amd64.exe${NC}"
echo -e "     ${BLUE}- jem-windows-arm64.exe${NC}"
echo -e "     ${BLUE}- jem-linux-amd64${NC}"
echo -e "     ${BLUE}- jem-linux-arm64${NC}"
echo -e "     ${BLUE}- jem-darwin-amd64${NC}"
echo -e "     ${BLUE}- jem-darwin-arm64${NC}"
echo ""
echo -e "${YELLOW}Or use GitHub CLI:${NC}"
echo -e "  ${BLUE}gh release create v${VERSION} --title \"Release v${VERSION}\" ./jem-*${NC}"
echo ""
echo -e "${YELLOW}Or use Makefile:${NC}"
echo -e "  ${BLUE}make release version=${VERSION}${NC}"
echo ""
