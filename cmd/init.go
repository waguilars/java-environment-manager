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

	// Check if we have any defaults configured (for informational purposes only)
	hasDefaults := cfg.Defaults.JDK != "" || cfg.Defaults.Gradle != ""
	if !hasDefaults {
		// Check old format for migration
		if cfg.JDK.Current == "" && cfg.Gradle.Current == "" {
			// No defaults configured - this is fine, init should still work
			// Output a warning but continue to generate the basic init script
			fmt.Fprintf(os.Stderr, "# Note: No default JDK or Gradle configured. Run 'jem use default jdk <version>' after sourcing init.\n")
		} else {
			// Migrate old format to new
			fmt.Fprintf(os.Stderr, "Warning: using deprecated config format. Run 'jem use default jdk <version>' to update.\n")
		}
	}

	// Update symlinks for default versions (if any are configured)
	if err := c.updateSymlinks(cfg); err != nil {
		return err
	}

	// Generate and output shell init script
	initScript := c.generateInitScript(cfg, shellType)
	if initScript == "" {
		return fmt.Errorf("failed to generate init script")
	}
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
			if err := c.symlinkManager.UpdateCurrentJava(cfg.Defaults.JDK, jdkPath); err != nil {
				return fmt.Errorf("failed to update current Java symlink: %w", err)
			}
		}
	}

	// Update Gradle symlink if default is set
	if cfg.Defaults.Gradle != "" {
		gradlePath := filepath.Join(homeDir, ".jem", "gradles", cfg.Defaults.Gradle)
		if _, err := os.Stat(gradlePath); err == nil {
			if err := c.symlinkManager.UpdateCurrentGradle(cfg.Defaults.Gradle, gradlePath); err != nil {
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

	// Set JAVA_HOME - try current symlink first, then fallback to config
	javaHome := c.getCurrentJavaHome(homeDir, cfg)
	if javaHome != "" {
		envVars["JAVA_HOME"] = javaHome
	}

	// Set GRADLE_HOME - try current symlink first, then fallback to config
	gradleHome := c.getCurrentGradleHome(homeDir, cfg)
	if gradleHome != "" {
		envVars["GRADLE_HOME"] = gradleHome
	}

	// Get the appropriate generator
	generator := shell.GetGenerator(shellType)

	// Generate the init script
	return generator.GenerateInitScript(envVars)
}

// getCurrentJavaHome gets JAVA_HOME from symlink or config fallback
func (c *InitCommand) getCurrentJavaHome(homeDir string, cfg *config.Config) string {
	// Try to read the current symlink first (most reliable)
	javaLink := filepath.Join(homeDir, ".jem", "current", "java")
	if target, err := os.Readlink(javaLink); err == nil {
		// Verify the target exists
		if _, err := os.Stat(target); err == nil {
			return target
		}
	}

	// Fallback: try config defaults
	jdkVersion := cfg.Defaults.JDK
	if jdkVersion == "" {
		jdkVersion = cfg.JDK.Current
	}
	if jdkVersion != "" {
		jdkPath := filepath.Join(homeDir, ".jem", "jdks", jdkVersion)
		if _, err := os.Stat(jdkPath); err == nil {
			return jdkPath
		}
	}

	return ""
}

// getCurrentGradleHome gets GRADLE_HOME from symlink or config fallback
func (c *InitCommand) getCurrentGradleHome(homeDir string, cfg *config.Config) string {
	// Try to read the current symlink first (most reliable)
	gradleLink := filepath.Join(homeDir, ".jem", "current", "gradle")
	if target, err := os.Readlink(gradleLink); err == nil {
		// Verify the target exists
		if _, err := os.Stat(target); err == nil {
			return target
		}
	}

	// Fallback: try config defaults
	gradleVersion := cfg.Defaults.Gradle
	if gradleVersion == "" {
		gradleVersion = cfg.Gradle.Current
	}
	if gradleVersion != "" {
		gradlePath := filepath.Join(homeDir, ".jem", "gradles", gradleVersion)
		if _, err := os.Stat(gradlePath); err == nil {
			return gradlePath
		}
	}

	return ""
}

// GetSymlinkManager returns the symlink manager for testing
func (c *InitCommand) GetSymlinkManager() *symlink.SymlinkManager {
	return c.symlinkManager
}
