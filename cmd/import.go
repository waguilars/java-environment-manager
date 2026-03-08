package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/jem/internal/config"
	"github.com/user/jem/internal/jdk"
	"github.com/user/jem/internal/platform"
)

// ImportCommand handles the 'jem import' command
type ImportCommand struct {
	platform   platform.Platform
	configRepo config.ConfigRepository
	jdkService *jdk.JDKService
}

// Execute runs the import command
func (c *ImportCommand) Execute(ctx context.Context, path string, name string) error {
	// Validate path exists
	if _, err := filepath.Abs(path); err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}

	// Validate it's a JDK directory
	javaBin := filepath.Join(path, "bin", "java")
	if _, err := os.Stat(javaBin); os.IsNotExist(err) {
		return fmt.Errorf("path is not a valid JDK directory (missing bin/java): %s", path)
	}

	homeDir := c.platform.HomeDir()
	jemDir := filepath.Join(homeDir, ".jem")
	importPath := filepath.Join(jemDir, "jdks", name)

	// Create symlink to the external JDK
	if err := c.platform.CreateLink(path, importPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	// Detect version
	version, err := c.jdkService.DetectVersion(path)
	if err != nil {
		return fmt.Errorf("failed to detect JDK version: %w", err)
	}

	// Register in config
	jdkInfo := config.JDKInfo{
		Path:     importPath,
		Version:  version,
		Provider: "external",
		Managed:  true,
	}

	if err := c.configRepo.AddInstalledJDK(jdkInfo); err != nil {
		return fmt.Errorf("failed to add installed JDK: %w", err)
	}

	fmt.Printf("✓ Imported JDK %s from %s\n", version, path)
	fmt.Printf("  Registered as: %s\n", name)

	return nil
}
