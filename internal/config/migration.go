package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// MigrateCurrentToDefaults migrates old config format (jdk.current, gradle.current)
// to new format (defaults.jdk, defaults.gradle) and creates symlinks for current versions.
// This is called during config loading to ensure backward compatibility.
func MigrateCurrentToDefaults(cfg *Config, jemDir string) error {
	// Migrate JDK current to defaults.jdk
	if cfg.Defaults.JDK == "" && cfg.JDK.Current != "" {
		cfg.Defaults.JDK = cfg.JDK.Current
		// Create symlink for current/java
		if err := createCurrentSymlink(jemDir, "java", cfg.JDK.Current); err != nil {
			return fmt.Errorf("failed to create java symlink: %w", err)
		}
	}

	// Migrate Gradle current to defaults.gradle
	if cfg.Defaults.Gradle == "" && cfg.Gradle.Current != "" {
		cfg.Defaults.Gradle = cfg.Gradle.Current
		// Create symlink for current/gradle
		if err := createCurrentSymlink(jemDir, "gradle", cfg.Gradle.Current); err != nil {
			return fmt.Errorf("failed to create gradle symlink: %w", err)
		}
	}

	return nil
}

// createCurrentSymlink creates or updates a symlink in the current directory
// pointing to the specified version
func createCurrentSymlink(jemDir string, tool string, version string) error {
	currentDir := filepath.Join(jemDir, "current")
	symlinkPath := filepath.Join(currentDir, tool)

	// Ensure the current directory exists
	if err := os.MkdirAll(currentDir, 0755); err != nil {
		return fmt.Errorf("failed to create current directory: %w", err)
	}

	// Remove existing symlink if it exists
	if _, err := os.Lstat(symlinkPath); err == nil {
		if err := os.Remove(symlinkPath); err != nil {
			return fmt.Errorf("failed to remove existing symlink: %w", err)
		}
	}

	// Determine the target path based on the tool
	var targetPath string
	switch tool {
	case "java":
		targetPath = filepath.Join(jemDir, "jdks", version)
	case "gradle":
		targetPath = filepath.Join(jemDir, "gradles", version)
	default:
		return fmt.Errorf("unknown tool: %s", tool)
	}

	// Check if target exists before creating symlink
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		// Target doesn't exist yet, skip symlink creation
		// The symlink will be created when the tool is installed
		return nil
	}

	// Create the symlink
	if err := os.Symlink(targetPath, symlinkPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

// UpdateCurrentSymlinks updates the current symlinks to match the defaults
// This should be called after setting a default version
func UpdateCurrentSymlinks(cfg *Config, jemDir string) error {
	// Update java symlink if default is set
	if cfg.Defaults.JDK != "" {
		if err := createCurrentSymlink(jemDir, "java", cfg.Defaults.JDK); err != nil {
			return fmt.Errorf("failed to update java symlink: %w", err)
		}
	}

	// Update gradle symlink if default is set
	if cfg.Defaults.Gradle != "" {
		if err := createCurrentSymlink(jemDir, "gradle", cfg.Defaults.Gradle); err != nil {
			return fmt.Errorf("failed to update gradle symlink: %w", err)
		}
	}

	return nil
}
