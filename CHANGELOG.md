# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
