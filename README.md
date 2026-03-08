# jem - Java Environment Manager

[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/user/jem)
[![Tests](https://img.shields.io/badge/tests-56%2F56%20passing-brightgreen.svg)](https://github.com/user/jem)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

`jem` (Java Environment Manager) is a CLI tool for managing multiple JDK and Gradle versions on your local development machine. It provides a simple, interactive interface for installing, switching, and managing Java development environments.

## Features

- **Multi-version management**: Install and switch between multiple JDK versions
- **Interactive menu**: User-friendly terminal interface with keyboard navigation
- **Automatic detection**: Scan your system for existing JDK installations
- **Cross-platform**: Works on Windows and Linux with automatic platform detection
- **Persistent configuration**: Settings persist across shell sessions
- **Symlink-based**: Uses symlinks for fast version switching (< 3 commands)
- **Multiple providers**: Support for Temurin, Corretto, Zulu, and more
- **Gradle support**: Manage Gradle versions alongside JDKs

## Prerequisites

- **Go 1.26.1+** (for building from source)
- **Windows 10+** or **Linux** (with bash/zsh)
- **Internet connection** (for downloading JDKs)
- **Developer Mode** (Windows, for symlinks - or run as administrator)

## Installation

### From Binary (Recommended)

Download the latest release from the [Releases page](https://github.com/user/jem/releases).

```bash
# Extract and move to PATH
tar -xzf jem-linux-amd64.tar.gz
sudo mv jem /usr/local/bin/
```

### From Source

```bash
git clone https://github.com/user/jem.git
cd jem
go build -o jem
sudo mv jem /usr/local/bin/
```

### Build from Source (with version)

```bash
go build -ldflags "-X main.Version=1.0.0" -o jem
```

## Basic Usage

### First Time Setup

```bash
# Initialize jem configuration
jem setup

# This will:
# - Create ~/.jem directory structure
# - Add ~/.jem/bin to your PATH
# - Configure JAVA_HOME and GRADLE_HOME
```

### Interactive Mode

Run `jem` without arguments to open the interactive menu:

```bash
jem
```

**Menu Options:**
- `Setup` - Initialize jem configuration
- `Scan` - Detect JDKs on your system
- `List` - List installed JDKs
- `Current` - Show active JDK version
- `Use` - Switch JDK version
- `Install` - Download and install JDK
- `Exit` - Close menu

**Navigation:**
- `↑↓` Arrow keys - Navigate menu
- `Enter` - Select option
- `q` or `Ctrl+C` - Quit

### CLI Commands

```bash
# List all installed JDKs
jem list jdk

# Show current JDK version
jem current

# Switch to a specific JDK version
jem use 17

# Install a specific JDK version
jem install jdk 21

# Install latest LTS version
jem install jdk --lts

# Scan system for existing JDKs
jem scan

# Generate shell completion
jem completion bash  # or zsh, fish, powershell
```

### Interactive Menus

#### Use JDK Menu

When running `jem use`, you'll see a menu of installed JDKs:

```
Select JDK to use:
  ✓ temurin-17 (managed)
  ✓ temurin-21 (managed)
  ✓ corretto-17 (external)
  ✓ zulu-11 (external)
```

#### Install Menu

When running `jem install`, you'll see available versions:

```
Select JDK version:
  ● 21.0.2 (LTS) [temurin]
  ● 20.0.2         [temurin]
  ○ 17.0.10 (LTS)  [temurin]
  ○ 11.0.22        [temurin]

Press 'l' to toggle LTS filter
```

## Supported JDK Providers

| Provider | Distribution | Default |
|----------|-------------|---------|
| Eclipse Temurin | Adoptium | ✅ |
| Amazon Corretto | Amazon | |
| Azul Zulu | Azul Systems | |
| Microsoft Build of OpenJDK | Microsoft | |
| Oracle GraalVM | Oracle | |
| BellSoft Liberica | BellSoft | |
| IBM Semeru | IBM | |
| IBM SDK (WebSphere) | IBM | |

## Project Structure

```
jem/
├── cmd/              # CLI command definitions
├── internal/
│   ├── config/       # Configuration management
│   ├── downloader/   # Download and extraction logic
│   ├── jdk/          # JDK management logic
│   ├── menu/         # Interactive menu system
│   ├── platform/     # OS-specific functionality
│   ├── provider/     # JDK provider integrations
│   └── ui/           # UI components (spinner, progress)
├── pkg/
│   └── interactive/  # Interactive utilities
├── go.mod
└── main.go
```

## Configuration

jem stores configuration in `~/.jem/config.toml`:

```toml
[jdk]
current = "temurin-21"

[jdks]
  [jdks.temurin-21]
    path = "/home/user/.jem/jdks/temurin-21"
    provider = "temurin"
    version = "21.0.2"
    managed = true

[gradle]
current = "gradle-8.6"
```

## Directory Structure

```
~/.jem/
├── bin/                    # Symlinks to executables (add to PATH)
│   ├── java -> ../jdks/current/bin/java
│   ├── javac -> ../jdks/current/bin/javac
│   └── gradle -> ../gradles/current/bin/gradle
├── jdks/                   # JDK installations
│   ├── temurin-17/
│   ├── temurin-21/
│   └── current -> temurin-21/
├── gradles/                # Gradle installations
│   ├── gradle-8.5/
│   ├── gradle-8.6/
│   └── current -> gradle-8.6/
└── config.toml             # Configuration file
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) for the interactive UI
- Uses [Cobra](https://github.com/spf13/cobra) for CLI command structure
- Icons from [lipgloss](https://github.com/charmbracelet/lipgloss)

## Status

✅ **Production Ready** - All 84 tasks complete, 56 tests passing, 13MB binary
