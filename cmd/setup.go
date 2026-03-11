package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/waguilars/java-environment-manager/internal/config"
	"github.com/waguilars/java-environment-manager/internal/platform"
	shellgen "github.com/waguilars/java-environment-manager/internal/shell"
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

	// Check for wrapper function presence and jem init pattern
	contentStr := string(content)
	hasWrapper := strings.Contains(contentStr, "jem()") || strings.Contains(contentStr, "function jem")
	hasInit := strings.Contains(contentStr, `eval "$(jem init)"`) ||
		(strings.Contains(contentStr, `Invoke-Expression`) && strings.Contains(contentStr, "jem init"))

	return hasWrapper && hasInit
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

	// Get the shell generator for wrapper function
	generator := shellgen.GetGenerator(shell)

	// Prepare the jem configuration block
	var jemLines []string
	jemLines = append(jemLines, "") // Empty line before jem section
	jemLines = append(jemLines, "# jem initialization")
	jemLines = append(jemLines, generator.GenerateWrapperFunction())

	switch shell {
	case config.ShellBash, config.ShellZsh:
		jemLines = append(jemLines, `eval "$(jem init)"`)
	case config.ShellPowerShell:
		jemLines = append(jemLines, `jem init | Invoke-Expression`)
	}

	jemBlock := strings.Join(jemLines, "\n") + "\n"

	// Check if file exists
	_, statErr := os.Stat(resolvedPath)
	if os.IsNotExist(statErr) {
		// File doesn't exist - create it with jem config
		dir := filepath.Dir(resolvedPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
		if err := os.WriteFile(resolvedPath, []byte(jemBlock), 0644); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}
		return nil
	}

	if statErr != nil {
		return fmt.Errorf("failed to stat shell config: %w", statErr)
	}

	// File exists - check if it's a legacy config (init but no wrapper)
	content, err := os.ReadFile(resolvedPath)
	if err != nil {
		return fmt.Errorf("failed to read shell config: %w", err)
	}

	contentStr := string(content)
	hasWrapper := strings.Contains(contentStr, "jem()") || strings.Contains(contentStr, "function jem")
	hasInit := strings.Contains(contentStr, `eval "$(jem init)"`) ||
		(strings.Contains(contentStr, "Invoke-Expression") && strings.Contains(contentStr, "jem init"))

	// Create backup before modification
	backupPath := resolvedPath + ".jem.backup"
	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	var newContent string
	if hasInit && !hasWrapper {
		// LEGACY CONFIG: Has init but no wrapper - insert wrapper BEFORE init
		newContent = insertWrapperBeforeInit(contentStr, generator.GenerateWrapperFunction(), shell)
	} else {
		// FRESH CONFIG: No jem config at all - append at the end
		newContent = contentStr + jemBlock
	}

	if err := os.WriteFile(resolvedPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// insertWrapperBeforeInit inserts the wrapper function before the jem init line
func insertWrapperBeforeInit(content string, wrapper string, shell config.Shell) string {
	lines := strings.Split(content, "\n")
	var result []string
	wrapperInserted := false

	for _, line := range lines {
		// Check if this is the init line
		isInitLine := false
		switch shell {
		case config.ShellBash, config.ShellZsh:
			if strings.Contains(line, `eval "$(jem init)"`) {
				isInitLine = true
			}
		case config.ShellPowerShell:
			if strings.Contains(line, "jem init") && strings.Contains(line, "Invoke-Expression") {
				isInitLine = true
			}
		}

		if isInitLine && !wrapperInserted {
			// Insert wrapper before init line
			result = append(result, "")
			result = append(result, "# jem initialization")
			result = append(result, wrapper)
			wrapperInserted = true
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}
