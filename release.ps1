# Release script for Windows
# Usage: .\release.ps1 -Version "0.2.0-beta" [-Type "beta"]

param(
    [Parameter(Mandatory=$true)]
    [ValidatePattern("^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)?(\+[a-zA-Z0-9.]+)?$")]
    [string]$Version,
    
    [Parameter(Mandatory=$false)]
    [ValidateSet("major", "minor", "patch", "beta", "rc", "alpha")]
    [string]$Type = ""
)

# Colors for output
function Write-Color {
    param(
        [Parameter(Mandatory=$true)]
        [string]$Message,
        
        [Parameter(Mandatory=$false)]
        [ConsoleColor]$Color = [ConsoleColor]::White
    )
    Write-Host $Message -ForegroundColor $Color
}

function Write-Info {
    param([string]$Message)
    Write-Color $Message -Color Yellow
}

function Write-Success {
    param([string]$Message)
    Write-Color $Message -Color Green
}

function Write-Error {
    param([string]$Message)
    Write-Color $Message -Color Red
}

Write-Info "Starting release process for v$Version..."

# Step 1: Update VERSION file
Write-Info "Step 1: Updating VERSION file..."
$Version | Out-File -FilePath "VERSION" -Encoding utf8
Write-Success "✓ VERSION file updated to $Version"

# Step 2: Update main.go version
Write-Info "Step 2: Updating main.go version..."
$content = Get-Content "main.go" -Raw
$content = $content -replace 'var Version = ".*"', "var Version = `"$Version`""
$content | Out-File -FilePath "main.go" -Encoding utf8 -NoNewline
Write-Success "✓ main.go updated"

# Step 3: Update cmd/root.go version
Write-Info "Step 3: Updating cmd/root.go version..."
$content = Get-Content "cmd/root.go" -Raw
$content = $content -replace 'Version: "[^"]*"', "Version: `"$Version`""
$content | Out-File -FilePath "cmd/root.go" -Encoding utf8 -NoNewline
Write-Success "✓ cmd/root.go updated"

# Step 4: Run tests
Write-Info "Step 4: Running tests..."
$testResult = & go test -v ./...
if ($LASTEXITCODE -ne 0) {
    Write-Error "ERROR: Tests failed"
    exit 1
}
Write-Success "✓ All tests passed"

# Step 5: Build binaries
Write-Info "Step 5: Building binaries..."
& go build -ldflags "-X main.Version=$Version" -o "jem.exe" .
if ($LASTEXITCODE -ne 0) {
    Write-Error "ERROR: Build failed"
    exit 1
}
Write-Success "✓ Binary built successfully"

# Step 6: Create git tag
Write-Info "Step 6: Creating git tag..."
git add VERSION main.go cmd/root.go
git commit -Message "Release v$Version"
git tag -a "v$Version" -m "Release v$Version"
Write-Success "✓ Git tag v$Version created"

# Step 6.5: Clean up any backup files
Write-Info "Step 6.5: Cleaning up backup files..."
Remove-Item -Path "*.bak" -Force -ErrorAction SilentlyContinue
Remove-Item -Path "cmd\*.bak" -Force -ErrorAction SilentlyContinue
Write-Success "✓ Backup files cleaned"

# Step 7: Build release assets
Write-Info "Step 7: Building release assets..."

Write-Info "Building Windows binaries..."
& go build -ldflags "-X main.Version=$Version" -o "jem-windows-amd64.exe" .
& go build -ldflags "-X main.Version=$Version" -o "jem-windows-arm64.exe" .
Write-Success "✓ Windows binaries built"

Write-Info "Building Linux binaries..."
& GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=$Version" -o "jem-linux-amd64" .
& GOOS=linux GOARCH=arm64 go build -ldflags "-X main.Version=$Version" -o "jem-linux-arm64" .
Write-Success "✓ Linux binaries built"

Write-Info "Building macOS binaries..."
& GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.Version=$Version" -o "jem-darwin-amd64" .
& GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.Version=$Version" -o "jem-darwin-arm64" .
Write-Success "✓ macOS binaries built"

# Summary
Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "Release v$Version preparation complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "  1. Review changes: git diff HEAD~1" -ForegroundColor White
Write-Host "  2. Push to remote: git push origin main --tags" -ForegroundColor White
Write-Host "  3. Create GitHub release with assets:" -ForegroundColor White
Write-Host "     - jem-windows-amd64.exe" -ForegroundColor White
Write-Host "     - jem-windows-arm64.exe" -ForegroundColor White
Write-Host "     - jem-linux-amd64" -ForegroundColor White
Write-Host "     - jem-linux-arm64" -ForegroundColor White
Write-Host "     - jem-darwin-amd64" -ForegroundColor White
Write-Host "     - jem-darwin-arm64" -ForegroundColor White
Write-Host ""
Write-Host "Or use GitHub CLI:" -ForegroundColor Yellow
Write-Host "  gh release create v$Version --title `"Release v$Version`" ./jem-*" -ForegroundColor White
