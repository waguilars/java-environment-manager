package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/waguilars/java-environment-manager/internal/config"
	"github.com/waguilars/java-environment-manager/internal/platform"
)

// SetupCommand handles the 'jem setup' command
type SetupCommand struct {
	platform   platform.Platform
	configRepo config.ConfigRepository
}

// Execute runs the setup command
func (c *SetupCommand) Execute(ctx context.Context) error {
	homeDir := c.platform.HomeDir()
	jemDir := filepath.Join(homeDir, ".jem")

	// Check if already configured
	configPath := filepath.Join(jemDir, "config.toml")
	configExists := false
	if _, err := os.Stat(configPath); err == nil {
		configExists = true
	}

	// State tracking for summary messages
	configCreated := false
	shellConfigured := false

	// Create directory structure and config if needed
	if !configExists {
		fmt.Println("Creating jem directory structure...")
		if err := os.MkdirAll(jemDir, 0755); err != nil {
			return fmt.Errorf("failed to create %s: %w", jemDir, err)
		}

		dirs := []string{"jdks", "gradles", "current"}
		for _, dir := range dirs {
			if err := os.MkdirAll(filepath.Join(jemDir, dir), 0755); err != nil {
				return fmt.Errorf("failed to create %s: %w", filepath.Join(jemDir, dir), err)
			}
		}

		// Create default config
		cfg := &config.Config{
			General: config.GeneralConfig{
				DefaultProvider: "temurin",
			},
			JDK: config.JDKConfig{
				Current: "",
			},
			Gradle: config.GradleConfig{
				Current: "",
			},
			Defaults: config.DefaultsConfig{
				JDK:    "",
				Gradle: "",
			},
			InstalledJDKs:    make(map[string]config.JDKInfo),
			DetectedJDKs:     make(map[string]config.JDKInfo),
			InstalledGradles: make(map[string]config.GradleInfo),
			DetectedGradles:  make(map[string]config.GradleInfo),
		}

		if err := c.configRepo.Save(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		configCreated = true
	}

	// Configure shell - ALWAYS runs regardless of config state
	shell := c.platform.DetectShell()
	shellConfigPath := c.platform.ShellConfigPath(shell)

	fmt.Printf("\nConfiguring shell (%s)...\n", shell)

	// Check if already configured
	if c.isShellConfigured(shellConfigPath) {
		fmt.Printf("Shell already configured in %s\n", shellConfigPath)
	} else {
		if err := c.configureShell(shell, shellConfigPath); err != nil {
			return fmt.Errorf("failed to configure shell: %w", err)
		}
		shellConfigured = true
	}

	// Print success message based on what was done
	fmt.Println("\n✓ jem setup completed successfully!")

	if configCreated {
		fmt.Printf("\nDirectory structure created at: %s\n", jemDir)
		fmt.Printf("Configuration file: %s\n", configPath)
	} else {
		fmt.Println("\nConfiguration file already exists.")
	}

	if shellConfigured {
		if shell == config.Shell("powershell") {
			fmt.Println("\nTo apply changes, restart PowerShell.")
		} else {
			fmt.Printf("\nTo apply changes, run: source %s\n", shellConfigPath)
			fmt.Println("Or restart your terminal.")
		}
	} else {
		fmt.Println("\nShell already configured.")
	}

	return nil
}

// isShellConfigured checks if the shell is already configured for jem
func (c *SetupCommand) isShellConfigured(shellConfigPath string) bool {
	if shellConfigPath == "$PROFILE" {
		// For PowerShell, check if jem init pattern exists
		// We can't easily check PowerShell profile without parsing it
		// So we'll skip this check for now
		return false
	}

	// Resolve symlink if the path is a symlink
	resolvedPath := shellConfigPath
	if info, err := os.Lstat(shellConfigPath); err == nil && info.Mode()&os.ModeSymlink != 0 {
		resolvedPath, err = os.Readlink(shellConfigPath)
		if err != nil {
			return false
		}
		// Handle relative symlinks
		if !filepath.IsAbs(resolvedPath) {
			resolvedPath = filepath.Join(filepath.Dir(shellConfigPath), resolvedPath)
		}
	}

	content, err := os.ReadFile(resolvedPath)
	if err != nil {
		return false
	}

	// Check for new jem init pattern
	contentStr := string(content)
	return strings.Contains(contentStr, `eval "$(jem init)"`) ||
		strings.Contains(contentStr, `Invoke-Expression`) && strings.Contains(contentStr, "jem init")
}

// configureShell adds jem init to the shell's configuration
func (c *SetupCommand) configureShell(shell config.Shell, shellConfigPath string) error {
	// Handle Fish shell specially - it's not supported for session-only mode
	if shell == config.ShellFish {
		fmt.Println()
		fmt.Println("⚠ Fish shell is not supported for session-only mode.")
		fmt.Println()
		fmt.Println("You can still use jem with Fish by:")
		fmt.Println("1. Using 'jem use default jdk <version>' for persistent changes")
		fmt.Println("2. Manually setting environment variables:")
		fmt.Println("   set -x JAVA_HOME ~/.jem/current/java")
		fmt.Println("   set -x PATH $JAVA_HOME/bin $PATH")
		fmt.Println()
		fmt.Println("Contributions for Fish support are welcome!")
		return nil
	}

	// Resolve symlink if the path is a symlink
	resolvedPath := shellConfigPath
	if info, err := os.Lstat(shellConfigPath); err == nil && info.Mode()&os.ModeSymlink != 0 {
		resolvedPath, err = os.Readlink(shellConfigPath)
		if err != nil {
			return fmt.Errorf("failed to resolve symlink: %w", err)
		}
		// Handle relative symlinks
		if !filepath.IsAbs(resolvedPath) {
			resolvedPath = filepath.Join(filepath.Dir(shellConfigPath), resolvedPath)
		}
	}

	// Create backup if file exists
	if _, err := os.Stat(resolvedPath); err == nil {
		backupPath := resolvedPath + ".jem.backup"

		// Open source file for reading
		src, err := os.Open(resolvedPath)
		if err != nil {
			return fmt.Errorf("failed to open shell config for backup: %w", err)
		}

		// Create backup file
		dst, err := os.Create(backupPath)
		if err != nil {
			src.Close()
			return fmt.Errorf("failed to create backup file: %w", err)
		}

		// Copy content
		if _, err := io.Copy(dst, src); err != nil {
			src.Close()
			dst.Close()
			return fmt.Errorf("failed to copy to backup: %w", err)
		}

		// Close both files
		src.Close()
		dst.Close()
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(resolvedPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Open file in append mode
	file, err := os.OpenFile(resolvedPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Write configuration using jem init pattern
	var lines []string
	switch shell {
	case config.ShellBash, config.ShellZsh:
		lines = []string{
			"",
			"# jem initialization",
			`eval "$(jem init)"`,
		}
	case config.ShellPowerShell:
		lines = []string{
			"",
			"# jem initialization",
			`jem init | Invoke-Expression`,
		}
	}

	if _, err := file.WriteString(strings.Join(lines, "\n") + "\n"); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
