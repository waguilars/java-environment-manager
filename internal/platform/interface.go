package platform

import "github.com/waguilars/java-environment-manager/internal/config"

// Platform interface defines the OS abstraction layer
type Platform interface {
	Name() string
	HomeDir() string
	DetectShell() config.Shell
	CreateLink(target, link string) error
	RemoveLink(link string) error
	IsLink(path string) bool
	CanCreateSymlinks() bool
	ShellConfigPath(shell config.Shell) string
	JDKDetectionPaths() []string
	GradleDetectionPaths() []string
}

// Detect returns the appropriate Platform implementation based on the OS
func Detect() Platform {
	return NewPlatform()
}
