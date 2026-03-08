# jem - Java Environment Manager

[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/user/jem)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

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

### From Source

```bash
git clone https://github.com/user/jem.git
cd jem
go build -o jem
sudo mv jem /usr/local/bin/
```

## Usage

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

```bash
# Switch JDK (imports automatically if detected but not managed)
jem use jdk 17.0.7

# Switch Gradle
jem use gradle 6.9.4

# Skip confirmation prompt
jem use jdk 21.0.1 --force
```

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
├── go.mod
└── main.go
```

## Configuration

jem stores configuration in `~/.jem/config.toml`:

```toml
[general]
  default_provider = "temurin"

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
  ["gradles.installed"."/home/user/.jem/gradles/7.6.1"]
    path = "/home/user/.jem/gradles/7.6.1"
    version = "7.6.1"
    managed = true
```

## Directory Structure

```
~/.jem/
├── bin/                    # Symlinks to current JDK/Gradle executables
│   └── java -> ../jdks/current/bin/java
├── jdks/                   # JDK installations (symlinks to imported)
│   ├── 21-amzn/
│   ├── 17.0.7-tem/
│   └── current -> 21-amzn/
├── gradles/                # Gradle installations
│   ├── 7.6.1/
│   └── current -> 7.6.1/
└── config.toml             # Configuration file
```

## Pending Issues

### `install` Command

The `install` command has known issues:

1. **JDK download returns 404**: The Temurin API integration needs to be fixed. The download URL generation is incorrect.

2. **No Gradle install support**: `jem install gradle <version>` is not implemented. Currently, you can only use Gradle versions detected by `jem scan`.

**Workaround**: Use SDKMAN or manually install JDKs/Gradles, then run `jem scan` to detect and `jem use` to switch.

```bash
# Workaround: Use SDKMAN to install, then import with jem
sdk install java 21.0.1-tem
jem scan
jem use jdk 21.0.1 --force
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) for the interactive UI
- Uses [Cobra](https://github.com/spf13/cobra) for CLI command structure
- Styling with [lipgloss](https://github.com/charmbracelet/lipgloss)