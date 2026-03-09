package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/waguilars/java-environment-manager/internal/config"
	"github.com/waguilars/java-environment-manager/internal/jdk"
	"github.com/waguilars/java-environment-manager/internal/platform"
)

// ImportCommand handles the 'jem import' command
type ImportCommand struct {
	platform   platform.Platform
	configRepo config.ConfigRepository
	jdkService *jdk.JDKService
}

// ExecuteJDK imports an external JDK into jem management
func (c *ImportCommand) ExecuteJDK(ctx context.Context, path string, name string) error {
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

// ExecuteGradle imports an external Gradle into jem management
func (c *ImportCommand) ExecuteGradle(ctx context.Context, path string, name string) error {
	// Validate path exists
	if _, err := filepath.Abs(path); err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}

	// Validate it's a Gradle directory
	gradleBin := filepath.Join(path, "bin", "gradle")
	gradleBat := filepath.Join(path, "bin", "gradle.bat")
	gradleBinExists := false
	gradleBatExists := false
	if _, err := os.Stat(gradleBin); err == nil {
		gradleBinExists = true
	}
	if _, err := os.Stat(gradleBat); err == nil {
		gradleBatExists = true
	}
	if !gradleBinExists && !gradleBatExists {
		return fmt.Errorf("path is not a valid Gradle directory (missing bin/gradle or bin/gradle.bat): %s", path)
	}

	// Detect version
	version, err := c.detectGradleVersion(path)
	if err != nil {
		return fmt.Errorf("failed to detect Gradle version: %w", err)
	}

	homeDir := c.platform.HomeDir()
	jemDir := filepath.Join(homeDir, ".jem")
	importPath := filepath.Join(jemDir, "gradles", name)

	// Create symlink to the external Gradle
	if err := c.platform.CreateLink(path, importPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	// Register in config
	gradleInfo := config.GradleInfo{
		Path:    importPath,
		Version: version,
		Managed: true,
	}

	if err := c.configRepo.AddInstalledGradle(gradleInfo); err != nil {
		return fmt.Errorf("failed to add installed Gradle: %w", err)
	}

	fmt.Printf("✓ Imported Gradle %s from %s\n", version, path)
	fmt.Printf("  Registered as: %s\n", name)

	return nil
}

// detectGradleVersion detects the version of a Gradle installation
func (c *ImportCommand) detectGradleVersion(path string) (string, error) {
	// Try to read from gradle-installation-beam.properties first
	propertiesFile := filepath.Join(path, "lib", "gradle-installation-beam.properties")
	if data, err := os.ReadFile(propertiesFile); err == nil {
		// Parse version from properties file
		version := c.parseGradleVersionFromProperties(string(data))
		if version != "" {
			return version, nil
		}
	}

	// Try to read from gradle-core-*.jar filename
	libDir := filepath.Join(path, "lib")
	entries, err := os.ReadDir(libDir)
	if err == nil {
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), "gradle-core-") && strings.HasSuffix(entry.Name(), ".jar") {
				// Extract version from filename: gradle-core-8.5.jar
				re := regexp.MustCompile(`gradle-core-([0-9][^-]*)`)
				matches := re.FindStringSubmatch(entry.Name())
				if len(matches) > 1 {
					return matches[1], nil
				}
			}
		}
	}

	// Try to run gradle --version
	gradleBin := filepath.Join(path, "bin", "gradle")
	if _, err := os.Stat(gradleBin); err == nil {
		// This would require executing the command, which is complex
		// For now, return a placeholder
		return "", fmt.Errorf("could not detect Gradle version from installation")
	}

	return "", fmt.Errorf("could not detect Gradle version from installation")
}

// parseGradleVersionFromProperties extracts version from gradle-installation-beam.properties
func (c *ImportCommand) parseGradleVersionFromProperties(data string) string {
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "gradle.version=") {
			return strings.TrimPrefix(line, "gradle.version=")
		}
	}
	return ""
}
