//go:build !windows

package platform

import (
	"os"
	"path/filepath"

	"github.com/waguilars/java-environment-manager/internal/config"
)

// LinuxPlatform implements Platform interface for Linux
type LinuxPlatform struct{}

// Name returns "linux"
func (p *LinuxPlatform) Name() string {
	return "linux"
}

// HomeDir returns the user's home directory
func (p *LinuxPlatform) HomeDir() string {
	home, _ := os.UserHomeDir()
	return home
}

// DetectShell returns the detected shell
func (p *LinuxPlatform) DetectShell() config.Shell {
	return DetectShellLinux()
}

// CreateLink creates a symbolic link
func (p *LinuxPlatform) CreateLink(target, link string) error {
	return os.Symlink(target, link)
}

// RemoveLink removes a symbolic link
func (p *LinuxPlatform) RemoveLink(link string) error {
	return os.Remove(link)
}

// IsLink checks if path is a symbolic link
func (p *LinuxPlatform) IsLink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

// CanCreateSymlinks returns true for Linux (always can create symlinks)
func (p *LinuxPlatform) CanCreateSymlinks() bool {
	return true
}

// ShellConfigPath returns the path to the shell config file
func (p *LinuxPlatform) ShellConfigPath(shell config.Shell) string {
	switch shell {
	case config.ShellBash:
		return filepath.Join(p.HomeDir(), ".bashrc")
	case config.ShellZsh:
		return filepath.Join(p.HomeDir(), ".zshrc")
	default:
		return filepath.Join(p.HomeDir(), ".bashrc")
	}
}

// JDKDetectionPaths returns standard JDK detection paths on Linux
func (p *LinuxPlatform) JDKDetectionPaths() []string {
	return []string{
		"/usr/lib/jvm",
		filepath.Join(p.HomeDir(), ".jdks"),
		filepath.Join(p.HomeDir(), ".sdkman", "candidates", "java"),
	}
}

// GradleDetectionPaths returns standard Gradle detection paths on Linux
func (p *LinuxPlatform) GradleDetectionPaths() []string {
	return []string{
		filepath.Join(p.HomeDir(), ".sdkman", "candidates", "gradle"),
		"/opt/gradle",
	}
}
