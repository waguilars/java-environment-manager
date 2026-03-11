# jem - Java Environment Manager

[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/waguilars/java-environment-manager)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Version](https://img.shields.io/badge/version-0.3.1--beta-orange.svg)](https://github.com/waguilars/java-environment-manager)

`jem` (Java Environment Manager) is a CLI tool for managing multiple JDK and Gradle versions on your local development machine. It provides a simple interface for detecting, importing, and switching between Java development environments.

## Features

- **Multi-version management**: Switch between multiple JDK and Gradle versions
- **Automatic detection**: Scan your system for existing JDK/Gradle installations (SDKMAN, /usr/lib/jvm, etc.)
- **Import existing installations**: Use JDKs/Gradles already on your system without reinstalling
- **Cross-platform**: Works on Windows and Linux with automatic platform detection
- **Persistent configuration**: Settings persist across shell sessions
- **Symlink-based**: Uses symlinks for fast version switching

## Prerequisites

- **Go 1.21+** (for building from source)
- **Windows 10+** or **Linux** (with bash/zsh)
- **Developer Mode** (Windows, for symlinks - or run as administrator)

## Installation

### Using `go install` (Recommended)

The easiest way to install `jem` is using Go's built-in install command:

```bash
go install github.com/waguilars/java-environment-manager/cmd/jem@latest
```

This will download, build, and install the latest version to your `$GOPATH/bin` (make sure `$GOPATH/bin` is in your `$PATH`).

To install a specific version:

```bash
go install github.com/waguilars/java-environment-manager/cmd/jem@v0.3.1
```

### From Source

```bash
git clone https://github.com/waguilars/java-environment-manager.git
cd java-environment-manager

# Build with Makefile
make build

# Or build directly
go build -o jem ./cmd/jem

# Install to $GOPATH/bin
go install ./cmd/jem
```

### Pre-built Binaries

Download the latest release for your platform:

| Platform | Architecture | Binary |
|----------|--------------|--------|
| Windows | amd64 | `jem-windows-amd64.exe` |
| Windows | arm64 | `jem-windows-arm64.exe` |
| Linux | amd64 | `jem-linux-amd64` |
| Linux | arm64 | `jem-linux-arm64` |
| macOS | amd64 | `jem-darwin-amd64` |
| macOS | arm64 | `jem-darwin-arm64` |

## Usage

### Quick Start

After installation, run the setup command to initialize jem:

```bash
# Initialize jem configuration and shell integration
jem setup

# Restart your shell or source the config
source ~/.zshrc  # or ~/.bashrc on Linux

# jem use now works immediately (no source needed!)
jem use jdk 21.0.1

# Verify
java -version
```

### Automatic `jem use` (Shell Wrapper)

Starting with v0.4.0, `jem setup` installs a shell function wrapper that makes `jem use` work immediately without requiring `source` or `eval`:

```bash
# This now updates the environment immediately
jem use jdk 21.0.1

# No need for:
#   source ~/.zshrc
#   eval "$(jem use jdk 21.0.1 --output-env)"
```

**How it works:**

The wrapper intercepts `jem use` commands and auto-evaluates the `--output-env` output, updating your shell's environment variables in place.

| Shell | Supported | Notes |
|-------|-----------|-------|
| Bash | ✅ | Linux |
| Zsh | ✅ | Linux/macOS |
| PowerShell | ✅ | Windows |
| Fish | ❌ | Use `jem use jdk 21 --default` for persistent changes |

**Validation Status:**

| Platform | Status |
|----------|--------|
| Linux (Bash/Zsh) | ✅ Validated |
| Windows (PowerShell) | ⏳ Pending validation |

**For existing users:** Run `jem setup` again to install the wrapper, then restart your shell.

### First Time Setup

```bash
# Initialize jem configuration
jem setup

# Scan for existing JDKs and Gradles on your system
jem scan
```

### Interactive Mode

```bash
# Launch interactive TUI menu
jem tui
```

**Navigation:**
- `↑↓` Arrow keys - Navigate menu
- `Enter` - Select option
- `q` or `Ctrl+C` - Quit

### CLI Commands

#### Initialize Shell (`jem init`)

The `jem init` command generates shell initialization scripts and updates symlinks based on your default versions:

```bash
# Auto-detect shell and output initialization script
eval "$(jem init)"

# Specify shell explicitly
eval "$(jem init bash)"
eval "$(jem init zsh)"
eval "$(jem init powershell)"

# For PowerShell
jem init powershell | Invoke-Expression
```

This command:
- Creates/updates symlinks in `~/.jem/current/` based on your default versions
- Outputs shell-specific environment variable exports
- Should be added to your shell profile via `jem setup`

#### List Available Versions

```bash
# List JDKs
jem list jdk

# List Gradles
jem list gradle

# List both
jem list
```

#### Show Current Versions

```bash
jem current
```

Output:
```
=== Current Environment ===
JDK:     21.0.1 (/home/user/.jem/jdks/21-amzn) [jem]
Gradle:  7.6.1 (/home/user/.jem/gradles/7.6.1) [jem]
```

#### Switch Versions

**Session Mode** (temporary, outputs env exports):
```bash
# Temporarily use a JDK for current shell session
jem use jdk 17.0.7 --session

# Temporarily use a Gradle for current shell session
jem use gradle 8.5 --session

# Or use the --output-env flag
jem use jdk 21.0.1 --output-env
```

**Default Mode** (persistent, updates symlinks and config):
```bash
# Set default JDK (updates symlinks and config)
jem use jdk 21.0.1 --default

# Set default Gradle
jem use gradle 8.5 --default

# These are equivalent (default mode is the default)
jem use jdk 21.0.1
```

**Force Mode** (skip confirmation prompts):
```bash
# Skip confirmation prompt for imports
jem use jdk 21.0.1 --force
```

#### Set Default Versions

You can also set the default version separately:

```bash
# Set default JDK version
jem use default jdk 21.0.1

# Set default Gradle version
jem use default gradle 8.5
```

This updates:
- The `[defaults]` section in `~/.jem/config.toml`
- The symlinks in `~/.jem/current/`

#### Scan System

```bash
# Detect JDKs and Gradles installed on your system
jem scan
```

Detects installations from:
- `~/.sdkman/candidates/java/` and `~/.sdkman/candidates/gradle/` (SDKMAN)
- `/usr/lib/jvm/` (Linux)
- `~/.jdks/` (IntelliJ IDEA)
- `/opt/gradle/` (Linux)

#### Install JDKs and Gradle

```bash
# Install latest LTS JDK
jem install jdk --lts

# Install specific JDK version
jem install jdk 21

# Install specific Gradle version
jem install gradle 8.5

# Install latest Gradle
jem install gradle latest
```

Downloads are verified with SHA256 checksums. JDKs are downloaded from Eclipse Temurin (Adoptium) and Gradle from the official Gradle distributions.

#### Import Existing Installations

```bash
# Import an existing JDK installation
jem import jdk /path/to/jdk-21

# Import an existing Gradle installation
jem import gradle /opt/gradle-8.5

# Import with custom name
jem import gradle /opt/gradle-8.5 --name work-gradle
```

## Project Structure

```
jem/
├── cmd/                    # CLI command definitions
│   ├── root.go            # Root command and subcommands
│   ├── current.go         # Current command
│   ├── list.go            # List command (jdk/gradle)
│   ├── scan.go            # Scan command
│   ├── use.go             # Use command (jdk/gradle)
│   ├── install.go         # Install command
│   └── factory.go         # Dependency injection
├── internal/
│   ├── config/            # Configuration management (TOML)
│   ├── downloader/        # Download and extraction logic
│   ├── jdk/               # JDK detection and management
│   ├── menu/              # Interactive TUI (Bubble Tea)
│   ├── platform/          # OS-specific functionality
│   ├── provider/          # JDK provider integrations (Temurin)
│   └── ui/                # UI components (spinner, progress)
├── pkg/
│   └── interactive/       # Interactive utilities
├── Makefile               # Build automation
├── VERSION                # Current version (SemVer 2.0.0)
├── CHANGELOG.md           # Version history
├── go.mod
└── cmd/
    └── jem/
        └── main.go        # Entry point for go install
```

## Configuration

jem stores configuration in `~/.jem/config.toml`:

```toml
[general]
  default_provider = "temurin"

[defaults]
  jdk = "21.0.1"
  gradle = "8.5"

[jdk]
  current = "21.0.1"

[gradle]
  current = "7.6.1"

["jdks.installed"]
  ["jdks.installed"."/home/user/.jem/jdks/21-amzn"]
    path = "/home/user/.jem/jdks/21-amzn"
    version = "21.0.1"
    provider = "imported"
    managed = true

["gradles.installed"]
  ["gradles.installed"."/home/user/.jem/gradles/8.5"]
    path = "/home/user/.jem/gradles/8.5"
    version = "8.5"
    managed = true
```

### Configuration Sections

- **`[general]`** - General settings like default provider
- **`[defaults]`** - Default JDK and Gradle versions (used by `jem init`)
- **`[jdk]`** - Current JDK (deprecated, use `[defaults]`)
- **`[gradle]`** - Current Gradle (deprecated, use `[defaults]`)
- **`["jdks.installed"]`** - Installed JDKs managed by jem
- **`["gradles.installed"]`** - Installed Gradles managed by jem

## Directory Structure

```
~/.jem/
├── current/                # Symlinks to default JDK/Gradle versions (updated by jem init)
│   ├── java -> ../jdks/temurin-21
│   └── gradle -> ../gradles/8.5
├── jdks/                   # JDK installations (symlinks to imported)
│   ├── 21-amzn/
│   ├── 17.0.7-tem/
│   └── current -> 21-amzn/  # (legacy, use ../current/java)
├── gradles/                # Gradle installations
│   ├── 7.6.1/
│   └── current -> 7.6.1/    # (legacy, use ../current/gradle)
└── config.toml             # Configuration file
```

**Important:** The `~/.jem/current/` directory contains symlinks that are updated by `jem init` based on your `[defaults]` configuration. These are used to set up your shell environment.

## Development

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-linux build-darwin build-windows

# Run tests
make test

# Run tests with coverage
make test-cover

# Show current version
make version
```

### Versioning

This project follows [Semantic Versioning 2.0.0](https://semver.org/).

| Component | When to increment |
|-----------|-------------------|
| **MAJOR** | Incompatible API changes |
| **MINOR** | New backward-compatible features |
| **PATCH** | Backward-compatible bug fixes |
| **Pre-release** | `-beta`, `-rc.1`, `-alpha` suffixes |

#### Creating a Release

This project uses [GoReleaser](https://goreleaser.com/) for automated releases. Releases are triggered by pushing a git tag.

```bash
# Tag a new version
git tag -a v0.3.0 -m "Release v0.3.0"
git push origin v0.3.0
```

The CI will automatically:
1. Run all tests
2. Build binaries for all platforms (Linux, macOS, Windows)
3. Create a GitHub release with assets
4. Generate checksums

For local testing (without creating a release):
```bash
# Test the release process locally
goreleaser release --snapshot --clean
```

See [CHANGELOG.md](CHANGELOG.md) for version history.

## Migration Guide

### Upgrading to v0.4.0+ (Automatic jem use)

Starting with v0.4.0, jem introduces **automatic `jem use`** via shell function wrappers:

#### Key Changes

- **Shell function wrapper**: `jem setup` now installs a wrapper that intercepts `jem use` commands
- **Immediate environment updates**: `jem use jdk 21` works immediately without `source` or `eval`
- **Cross-platform**: Supported on Bash, Zsh (Linux/macOS), and PowerShell (Windows)

#### Migration Steps

1. **Run `jem setup`** to install the shell wrapper:
   ```bash
   jem setup
   ```
   This will:
   - Install the shell function wrapper in your profile
   - Add `eval "$(jem init)"` (or PowerShell equivalent) to your shell profile
   - Migrate old config format to new format

2. **Restart your shell or source the config**:
   ```bash
   source ~/.zshrc  # or ~/.bashrc
   # On Windows: restart PowerShell
   ```

3. **Test the wrapper**:
   ```bash
   jem use jdk 21.0.1
   java -version  # Should show the new version immediately
   ```

#### Old vs New Behavior

| Old Behavior (pre-0.4.0) | New Behavior (0.4.0+) |
|-------------------------|----------------------|
| `jem use jdk 21` required `source ~/.zshrc` | `jem use jdk 21` works immediately |
| Manual `eval "$(jem use ... --output-env)"` | Automatic via wrapper |
| No session mode | `jem use jdk 21 --session` for temporary use |
| `jem setup` managed PATH | Shell wrapper + `jem init` |

#### Shell Support

| Shell | Wrapper Support | Notes |
|-------|-----------------|-------|
| Bash | ✅ | Linux |
| Zsh | ✅ | Linux/macOS |
| PowerShell | ✅ | Windows |
| Fish | ❌ | Use `jem use jdk 21 --default` |

#### Validation Status

| Platform | Status |
|----------|--------|
| Linux (Bash/Zsh) | ✅ Validated |
| Windows (PowerShell) | ⏳ Pending validation |

#### Backward Compatibility

- Existing `eval "$(jem use ... --output-env)"` still works
- Users can remove the wrapper from their profile if they prefer manual eval
- All existing installed JDKs and Gradles are preserved

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) for the interactive UI
- Uses [Cobra](https://github.com/spf13/cobra) for CLI command structure
- Styling with [Lipgloss](https://github.com/charmbracelet/lipgloss)
- Versioning follows [SemVer 2.0.0](https://semver.org/)
- Changelog follows [Keep a Changelog](https://keepachangelog.com/)