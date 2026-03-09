# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.1-beta] - 2026-03-09

### Fixed
- **CRITICAL**: `jem setup` no longer destroys shell config files (was using `os.Rename` instead of copy)
- Shell config backups now properly preserve original content using `io.Copy`
- Added comprehensive tests for shell config backup with no data loss (50+ line configs)
- Shell config now retains all user content AND adds jem configuration

## [0.3.0-beta] - 2026-03-09

### Added
- `jem install jdk <version>` - Download and install JDKs from Temurin (Eclipse Adoptium)
- `jem install gradle <version>` - Download and install Gradle distributions
- `jem import gradle <path>` - Import existing Gradle installations
- GradleProvider interface for Gradle distribution downloads
- SHA256 checksum validation for all downloads
- Download progress display with percentage and size
- `--lts` flag for `jem install jdk` to install latest LTS version
- `latest` keyword support for `jem install gradle latest`
- Comprehensive test suite with 81.5% coverage for providers

### Fixed
- Temurin API URL construction (was returning 404 errors)
- Temurin API response parsing with correct JSON structure
- Platform-aware download URLs (Windows: zip, Linux: tar.gz)

### Changed
- Refactored provider architecture for better testability
- Improved error handling for network failures
- Enhanced download infrastructure with checksum verification

## [0.2.0-beta] - 2026-03-08

### Added
- Support for Gradle version management
- System detection for Gradle (GRADLE_HOME/PATH)
- `jem list gradle` and `jem use gradle` commands
- `jem scan` now detects Gradle installations
- `jem use --force` flag for non-interactive mode
- Import functionality for external JDK/Gradle installations

### Fixed
- Fixed `install` command 404 error (Temurin API URL generation)
- Fixed `current` command to detect system Java/Gradle
- Improved symlink handling for JDK and Gradle

### Changed
- Updated architecture to support Gradle alongside JDK management
- Improved CLI command structure with better separation of concerns

## [0.1.0-beta] - 2026-03-08

### Added
- Initial release with JDK management
- Interactive TUI menu with Bubble Tea v2
- Cross-platform support (Windows/Linux)
- System detection for JDK (JAVA_HOME/PATH)
- `jem setup` command for shell configuration
- `jem scan` command for detecting JDKs
- `jem list` command for listing installed/detected JDKs
- `jem current` command for showing current environment
- `jem use` command for switching JDK versions
- Configuration management with TOML
