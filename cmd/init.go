package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/waguilars/java-environment-manager/internal/config"
	"github.com/waguilars/java-environment-manager/internal/platform"
	"github.com/waguilars/java-environment-manager/internal/shell"
	"github.com/waguilars/java-environment-manager/internal/symlink"
)

// InitCommand handles the 'jem init' command
type InitCommand struct {
	platform       platform.Platform
	configRepo     config.ConfigRepository
	symlinkManager *symlink.SymlinkManager
}

// NewInitCommand creates a new InitCommand
func NewInitCommand(platform platform.Platform, configRepo config.ConfigRepository) *InitCommand {
	return &InitCommand{
		platform:       platform,
		configRepo:     configRepo,
		symlinkManager: symlink.NewSymlinkManager(platform),
	}
}

// Execute runs the init command
func (c *InitCommand) Execute(ctx context.Context, shellName string) error {
	// Auto-detect shell if not provided
	shellType := c.detectShell(shellName)

	// Load config to get defaults
	cfg, err := c.configRepo.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Update symlinks for default versions
	if err := c.updateSymlinks(cfg); err != nil {
		return err
	}

	// Generate and output shell init script
	initScript := c.generateInitScript(cfg, shellType)
	fmt.Println(initScript)

	return nil
}

// detectShell determines the shell to use
func (c *InitCommand) detectShell(shellName string) config.Shell {
	if shellName != "" {
		// Parse provided shell name
		switch shellName {
		case "bash":
			return config.ShellBash
		case "zsh":
			return config.ShellZsh
		case "powershell", "pwsh":
			return config.ShellPowerShell
		case "fish":
			return config.ShellFish
		default:
			// Fall back to auto-detection
		}
	}

	// Auto-detect from platform
	return c.platform.DetectShell()
}

// updateSymlinks updates the current symlinks based on config defaults
func (c *InitCommand) updateSymlinks(cfg *config.Config) error {
	homeDir := c.platform.HomeDir()

	// Update Java symlink if default is set
	if cfg.Defaults.JDK != "" {
		jdkPath := filepath.Join(homeDir, ".jem", "jdks", cfg.Defaults.JDK)
		if _, err := os.Stat(jdkPath); err == nil {
			if err := c.symlinkManager.UpdateCurrentJava(cfg.Defaults.JDK); err != nil {
				return fmt.Errorf("failed to update current Java symlink: %w", err)
			}
		}
	}

	// Update Gradle symlink if default is set
	if cfg.Defaults.Gradle != "" {
		gradlePath := filepath.Join(homeDir, ".jem", "gradles", cfg.Defaults.Gradle)
		if _, err := os.Stat(gradlePath); err == nil {
			if err := c.symlinkManager.UpdateCurrentGradle(cfg.Defaults.Gradle); err != nil {
				return fmt.Errorf("failed to update current Gradle symlink: %w", err)
			}
		}
	}

	return nil
}

// generateInitScript generates the shell initialization script
func (c *InitCommand) generateInitScript(cfg *config.Config, shellType config.Shell) string {
	homeDir := c.platform.HomeDir()
	envVars := make(map[string]string)

	// Set JAVA_HOME if current Java symlink exists
	javaLink := filepath.Join(homeDir, ".jem", "current", "java")
	if c.platform.IsLink(javaLink) {
		target, err := os.Readlink(javaLink)
		if err == nil {
			envVars["JAVA_HOME"] = target
		}
	}

	// Set GRADLE_HOME if current Gradle symlink exists
	gradleLink := filepath.Join(homeDir, ".jem", "current", "gradle")
	if c.platform.IsLink(gradleLink) {
		target, err := os.Readlink(gradleLink)
		if err == nil {
			envVars["GRADLE_HOME"] = target
		}
	}

	// Get the appropriate generator
	generator := shell.GetGenerator(shellType)

	// Generate the init script
	return generator.GenerateInitScript(envVars)
}

// GetSymlinkManager returns the symlink manager for testing
func (c *InitCommand) GetSymlinkManager() *symlink.SymlinkManager {
	return c.symlinkManager
}
