package symlink

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/waguilars/java-environment-manager/internal/platform"
)

// SymlinkManager manages symlinks for current Java and Gradle versions
type SymlinkManager struct {
	platform platform.Platform
	homeDir  string
}

// NewSymlinkManager creates a new symlink manager
func NewSymlinkManager(p platform.Platform) *SymlinkManager {
	return &SymlinkManager{
		platform: p,
		homeDir:  p.HomeDir(),
	}
}

// UpdateCurrentJava updates the current Java symlink to point to the specified version
func (m *SymlinkManager) UpdateCurrentJava(version string, jdkPath string) error {
	jemDir := filepath.Join(m.homeDir, ".jem")
	currentDir := filepath.Join(jemDir, "current")
	javaLink := filepath.Join(currentDir, "java")

	// Ensure current directory exists
	if err := os.MkdirAll(currentDir, 0755); err != nil {
		return fmt.Errorf("failed to create current directory: %w", err)
	}

	// Validate the JDK path exists
	if _, err := os.Stat(jdkPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("java version '%s' not found at %s", version, jdkPath)
		}
		return fmt.Errorf("failed to check java version: %w", err)
	}

	// Remove existing symlink if it exists
	if m.platform.IsLink(javaLink) {
		if err := m.platform.RemoveLink(javaLink); err != nil {
			return fmt.Errorf("failed to remove existing java symlink: %w", err)
		}
	}

	// Create new symlink
	if err := m.platform.CreateLink(jdkPath, javaLink); err != nil {
		return fmt.Errorf("failed to create java symlink: %w", err)
	}

	return nil
}

// UpdateCurrentGradle updates the current Gradle symlink to point to the specified version
func (m *SymlinkManager) UpdateCurrentGradle(version string, gradlePath string) error {
	jemDir := filepath.Join(m.homeDir, ".jem")
	currentDir := filepath.Join(jemDir, "current")
	gradleLink := filepath.Join(currentDir, "gradle")

	// Ensure current directory exists
	if err := os.MkdirAll(currentDir, 0755); err != nil {
		return fmt.Errorf("failed to create current directory: %w", err)
	}

	// Validate the Gradle path exists
	if _, err := os.Stat(gradlePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("gradle version '%s' not found at %s", version, gradlePath)
		}
		return fmt.Errorf("failed to check gradle version: %w", err)
	}

	// Remove existing symlink if it exists
	if m.platform.IsLink(gradleLink) {
		if err := m.platform.RemoveLink(gradleLink); err != nil {
			return fmt.Errorf("failed to remove existing gradle symlink: %w", err)
		}
	}

	// Create new symlink
	if err := m.platform.CreateLink(gradlePath, gradleLink); err != nil {
		return fmt.Errorf("failed to create gradle symlink: %w", err)
	}

	return nil
}

// GetCurrentJava returns the target path of the current Java symlink
func (m *SymlinkManager) GetCurrentJava() (string, error) {
	javaLink := filepath.Join(m.homeDir, ".jem", "current", "java")

	if !m.platform.IsLink(javaLink) {
		return "", fmt.Errorf("no current Java version set")
	}

	target, err := os.Readlink(javaLink)
	if err != nil {
		return "", fmt.Errorf("failed to read java symlink: %w", err)
	}

	return target, nil
}

// GetCurrentGradle returns the target path of the current Gradle symlink
func (m *SymlinkManager) GetCurrentGradle() (string, error) {
	gradleLink := filepath.Join(m.homeDir, ".jem", "current", "gradle")

	if !m.platform.IsLink(gradleLink) {
		return "", fmt.Errorf("no current Gradle version set")
	}

	target, err := os.Readlink(gradleLink)
	if err != nil {
		return "", fmt.Errorf("failed to read gradle symlink: %w", err)
	}

	return target, nil
}

// ValidateVersionExists checks if the specified version directory exists
func (m *SymlinkManager) ValidateVersionExists(tool, version string) error {
	var versionPath string

	switch tool {
	case "java":
		versionPath = filepath.Join(m.homeDir, ".jem", "jdks", version)
	case "gradle":
		versionPath = filepath.Join(m.homeDir, ".jem", "gradles", version)
	default:
		return fmt.Errorf("unknown tool: %s", tool)
	}

	info, err := os.Stat(versionPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s version '%s' not found at %s", tool, version, versionPath)
		}
		return fmt.Errorf("failed to check %s version: %w", tool, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("%s version '%s' exists but is not a directory", tool, version)
	}

	return nil
}

// RemoveCurrentJava removes the current Java symlink
func (m *SymlinkManager) RemoveCurrentJava() error {
	javaLink := filepath.Join(m.homeDir, ".jem", "current", "java")

	if m.platform.IsLink(javaLink) {
		if err := m.platform.RemoveLink(javaLink); err != nil {
			return fmt.Errorf("failed to remove java symlink: %w", err)
		}
	}

	return nil
}

// RemoveCurrentGradle removes the current Gradle symlink
func (m *SymlinkManager) RemoveCurrentGradle() error {
	gradleLink := filepath.Join(m.homeDir, ".jem", "current", "gradle")

	if m.platform.IsLink(gradleLink) {
		if err := m.platform.RemoveLink(gradleLink); err != nil {
			return fmt.Errorf("failed to remove gradle symlink: %w", err)
		}
	}

	return nil
}

// HasCurrentJava checks if a current Java symlink exists
func (m *SymlinkManager) HasCurrentJava() bool {
	javaLink := filepath.Join(m.homeDir, ".jem", "current", "java")
	return m.platform.IsLink(javaLink)
}

// HasCurrentGradle checks if a current Gradle symlink exists
func (m *SymlinkManager) HasCurrentGradle() bool {
	gradleLink := filepath.Join(m.homeDir, ".jem", "current", "gradle")
	return m.platform.IsLink(gradleLink)
}
