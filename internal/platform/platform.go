//go:build !windows

package platform

import (
	"os"
	"strings"

	"github.com/waguilars/java-environment-manager/internal/config"
)

// NewPlatformLinux creates the Platform implementation for Linux/Unix
func NewPlatformLinux() Platform {
	return &LinuxPlatform{}
}

// DetectShell detects the shell based on the OS
func DetectShell() config.Shell {
	return DetectShellLinux()
}

// DetectShellLinux detects the shell on Linux/Unix systems
func DetectShellLinux() config.Shell {
	// Check parent process
	if parent, err := os.LookupEnv("SHELL"); err {
		shell := strings.ToLower(parent)
		if strings.Contains(shell, "bash") {
			return config.ShellBash
		}
		if strings.Contains(shell, "zsh") {
			return config.ShellZsh
		}
		if strings.Contains(shell, "powershell") {
			return config.ShellPowerShell
		}
	}

	// Check environment variables
	if shell := os.Getenv("BASH_VERSION"); shell != "" {
		return config.ShellBash
	}
	if shell := os.Getenv("ZSH_VERSION"); shell != "" {
		return config.ShellZsh
	}

	// Default to bash
	return config.ShellBash
}

// NewPlatform creates the appropriate Platform implementation based on the OS
func NewPlatform() Platform {
	return NewPlatformLinux()
}
