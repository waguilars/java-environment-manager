package jdk

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/jem/internal/config"
	"github.com/user/jem/internal/platform"
)

// JDKSymlinker manages symlinks for JDK installations
type JDKSymlinker struct {
	platform platform.Platform
	config   *config.Config
}

// NewJDKSymlinker creates a new JDK symlinker
func NewJDKSymlinker(platform platform.Platform) *JDKSymlinker {
	return &JDKSymlinker{
		platform: platform,
		config:   &config.Config{},
	}
}

// SetConfig sets the configuration for the symlinker
func (s *JDKSymlinker) SetConfig(config *config.Config) {
	s.config = config
}

// UpdateCurrentLink updates the current JDK symlink
func (s *JDKSymlinker) UpdateCurrentLink(jdkName, jdkPath string) error {
	// Determine the current symlink path
	currentLink := filepath.Join(s.platform.HomeDir(), ".jem", "jdks", "current")

	// Remove existing symlink if it exists
	if s.platform.IsLink(currentLink) {
		if err := s.platform.RemoveLink(currentLink); err != nil {
			return err
		}
	}

	// Create the new symlink
	if err := s.platform.CreateLink(jdkPath, currentLink); err != nil {
		return err
	}

	return nil
}

// UpdateBinLinks creates symlinks for Java binaries in the bin directory
func (s *JDKSymlinker) UpdateBinLinks(jdkPath string) error {
	binDir := filepath.Join(s.platform.HomeDir(), ".jem", "bin")

	// Create bin directory if it doesn't exist
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}

	// Java binaries to symlink
	binaries := []string{"java", "javac", "jar", "javadoc", "javap"}

	for _, binary := range binaries {
		source := filepath.Join(jdkPath, "bin", binary)
		link := filepath.Join(binDir, binary)

		// Remove existing symlink if it exists
		if s.platform.IsLink(link) {
			if err := s.platform.RemoveLink(link); err != nil {
				return err
			}
		}

		// Create the new symlink
		if err := s.platform.CreateLink(source, link); err != nil {
			return err
		}
	}

	return nil
}

// RemoveCurrentLink removes the current JDK symlink
func (s *JDKSymlinker) RemoveCurrentLink() error {
	currentLink := filepath.Join(s.platform.HomeDir(), ".jem", "jdks", "current")

	if s.platform.IsLink(currentLink) {
		return s.platform.RemoveLink(currentLink)
	}

	return nil
}

// RemoveBinLinks removes all Java binary symlinks
func (s *JDKSymlinker) RemoveBinLinks() error {
	binDir := filepath.Join(s.platform.HomeDir(), ".jem", "bin")

	// Java binaries to remove
	binaries := []string{"java", "javac", "jar", "javadoc", "javap"}

	for _, binary := range binaries {
		link := filepath.Join(binDir, binary)

		if s.platform.IsLink(link) {
			if err := s.platform.RemoveLink(link); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetCurrentJDK returns the path of the currently active JDK
func (s *JDKSymlinker) GetCurrentJDK() (string, error) {
	currentLink := filepath.Join(s.platform.HomeDir(), ".jem", "jdks", "current")

	if !s.platform.IsLink(currentLink) {
		return "", os.ErrNotExist
	}

	// Read the symlink target
	target, err := os.Readlink(currentLink)
	if err != nil {
		return "", err
	}

	return target, nil
}

// UpdateLinks updates both current and bin symlinks
func (s *JDKSymlinker) UpdateLinks(jdkPath string) error {
	// Extract JDK name from path
	jdkName := filepath.Base(jdkPath)

	// Update current link
	if err := s.UpdateCurrentLink(jdkName, jdkPath); err != nil {
		return fmt.Errorf("failed to update current link: %w", err)
	}

	// Update bin links
	if err := s.UpdateBinLinks(jdkPath); err != nil {
		return fmt.Errorf("failed to update bin links: %w", err)
	}

	return nil
}
