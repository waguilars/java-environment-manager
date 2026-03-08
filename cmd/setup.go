package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/user/jem/internal/config"
	"github.com/user/jem/internal/platform"
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
	if _, err := os.Stat(configPath); err == nil {
		fmt.Println("jem is already configured.")
		fmt.Printf("Configuration file: %s\n", configPath)
		return nil
	}

	// Create directory structure
	fmt.Println("Creating jem directory structure...")
	if err := os.MkdirAll(jemDir, 0755); err != nil {
		return fmt.Errorf("failed to create %s: %w", jemDir, err)
	}

	dirs := []string{"bin", "jdks", "gradles"}
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
		InstalledJDKs:    make(map[string]config.JDKInfo),
		DetectedJDKs:     make(map[string]config.JDKInfo),
		InstalledGradles: make(map[string]config.GradleInfo),
		DetectedGradles:  make(map[string]config.GradleInfo),
	}

	if err := c.configRepo.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Configure shell
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
	}

	// Print success message
	fmt.Println("\n✓ jem setup completed successfully!")
	fmt.Printf("\nDirectory structure created at: %s\n", jemDir)
	fmt.Printf("Configuration file: %s\n", configPath)

	if shell == config.Shell("powershell") {
		fmt.Println("\nTo apply changes, restart PowerShell.")
	} else {
		fmt.Printf("\nTo apply changes, run: source %s\n", shellConfigPath)
		fmt.Println("Or restart your terminal.")
	}

	return nil
}

// isShellConfigured checks if the shell is already configured for jem
func (c *SetupCommand) isShellConfigured(shellConfigPath string) bool {
	if shellConfigPath == "$PROFILE" {
		// For PowerShell, check if jem is in PATH
		// We can't easily check PowerShell profile without parsing it
		// So we'll skip this check for now
		return false
	}

	content, err := os.ReadFile(shellConfigPath)
	if err != nil {
		return false
	}

	return strings.Contains(string(content), ".jem/bin")
}

// configureShell adds jem to the shell's PATH
func (c *SetupCommand) configureShell(shell config.Shell, shellConfigPath string) error {
	// Create backup if file exists
	if _, err := os.Stat(shellConfigPath); err == nil {
		backupPath := shellConfigPath + ".jem.backup"
		if err := os.Rename(shellConfigPath, backupPath); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(shellConfigPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Open file in append mode
	file, err := os.OpenFile(shellConfigPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Write configuration
	var lines []string
	switch shell {
	case config.Shell("bash"), config.Shell("zsh"):
		lines = []string{
			"",
			"# jem configuration",
			`export PATH="$HOME/.jem/bin:$PATH"`,
			`export JAVA_HOME="$HOME/.jem/jdks/current"`,
		}
	case config.Shell("powershell"):
		lines = []string{
			"",
			"# jem configuration",
			`$env:PATH = "$HOME\.jem\bin;$env:PATH"`,
			`$env:JAVA_HOME = "$HOME\.jem\jdks\current"`,
		}
	}

	if _, err := file.WriteString(strings.Join(lines, "\n") + "\n"); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
