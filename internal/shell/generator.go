package shell

import (
	"os"
	"strings"

	"github.com/waguilars/java-environment-manager/internal/config"
)

// ShellGenerator generates shell initialization scripts
type ShellGenerator interface {
	// GenerateInitScript generates a shell script that sets up the environment
	GenerateInitScript(envVars map[string]string) string
	// Name returns the shell name
	Name() string
}

// DetectShell detects the current shell from the environment
func DetectShell() config.Shell {
	// Check SHELL environment variable
	shellEnv := os.Getenv("SHELL")
	if shellEnv != "" {
		shell := strings.ToLower(shellEnv)
		if strings.Contains(shell, "bash") {
			return config.ShellBash
		}
		if strings.Contains(shell, "zsh") {
			return config.ShellZsh
		}
		if strings.Contains(shell, "powershell") || strings.Contains(shell, "pwsh") {
			return config.ShellPowerShell
		}
		if strings.Contains(shell, "fish") {
			return config.ShellFish
		}
	}

	// Check for PowerShell on Windows
	if os.Getenv("PSModulePath") != "" {
		return config.ShellPowerShell
	}

	// Check version-specific environment variables
	if os.Getenv("BASH_VERSION") != "" {
		return config.ShellBash
	}
	if os.Getenv("ZSH_VERSION") != "" {
		return config.ShellZsh
	}
	if os.Getenv("FISH_VERSION") != "" {
		return config.ShellFish
	}

	// Default to bash
	return config.ShellBash
}

// GetGenerator returns the appropriate generator for the given shell type
func GetGenerator(shell config.Shell) ShellGenerator {
	switch shell {
	case config.ShellBash:
		return NewBashGenerator()
	case config.ShellZsh:
		return NewZshGenerator()
	case config.ShellPowerShell:
		return NewPowerShellGenerator()
	case config.ShellFish:
		return NewFishGenerator()
	default:
		return NewBashGenerator()
	}
}
