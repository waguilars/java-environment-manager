package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/user/jem/internal/config"
	"github.com/user/jem/internal/jdk"
	"github.com/user/jem/internal/platform"
)

// ScanCommand handles the 'jem scan' command
type ScanCommand struct {
	platform   platform.Platform
	configRepo config.ConfigRepository
	jdkService *jdk.JDKService
}

// Execute runs the scan command
func (c *ScanCommand) Execute(ctx context.Context) error {
	fmt.Println("Scanning for JDKs and Gradles...")

	// Scan for JDKs
	fmt.Println("\nScanning for JDKs...")
	detectedJDKs, err := c.jdkService.Detect()
	if err != nil {
		return fmt.Errorf("failed to detect JDKs: %w", err)
	}

	// Clear existing detected JDKs and add new ones
	if err := c.configRepo.ClearDetectedJDKs(); err != nil {
		return fmt.Errorf("failed to clear detected JDKs: %w", err)
	}

	for _, jdkInfo := range detectedJDKs {
		if err := c.configRepo.AddDetectedJDK(jdkInfo); err != nil {
			return fmt.Errorf("failed to add detected JDK: %w", err)
		}
	}

	// Scan for Gradles
	fmt.Println("Scanning for Gradles...")
	detectedGradles := scanGradles(c.platform)

	// Clear existing detected Gradles and add new ones
	if err := c.configRepo.ClearDetectedGradles(); err != nil {
		return fmt.Errorf("failed to clear detected Gradles: %w", err)
	}

	for _, gradleInfo := range detectedGradles {
		if err := c.configRepo.AddDetectedGradle(gradleInfo); err != nil {
			return fmt.Errorf("failed to add detected Gradle: %w", err)
		}
	}

	// Print results
	fmt.Printf("\n✓ Detection complete!\n")
	fmt.Printf("  Found %d JDK(s)\n", len(detectedJDKs))
	fmt.Printf("  Found %d Gradle(s)\n", len(detectedGradles))

	// List detected JDKs
	if len(detectedJDKs) > 0 {
		fmt.Println("\nDetected JDKs:")
		for _, jdkInfo := range detectedJDKs {
			fmt.Printf("  - %s (%s)\n", jdkInfo.Version, jdkInfo.Path)
		}
	}

	// List detected Gradles
	if len(detectedGradles) > 0 {
		fmt.Println("\nDetected Gradles:")
		for _, gradleInfo := range detectedGradles {
			fmt.Printf("  - %s (%s)\n", gradleInfo.Version, gradleInfo.Path)
		}
	}

	return nil
}

// scanGradles scans for Gradle installations in standard paths
func scanGradles(plat platform.Platform) []config.GradleInfo {
	paths := plat.GradleDetectionPaths()
	var detected []config.GradleInfo

	for _, path := range paths {
		gradles := scanGradlePath(path)
		detected = append(detected, gradles...)
	}

	return detected
}

// scanGradlePath scans a directory for Gradle installations
func scanGradlePath(path string) []config.GradleInfo {
	var gradles []config.GradleInfo

	entries, err := os.ReadDir(path)
	if err != nil {
		return gradles
	}

	for _, entry := range entries {
		if entry.IsDir() {
			fullPath := filepath.Join(path, entry.Name())
			version := detectGradleVersionFromPath(fullPath)
			if version != "" {
				gradles = append(gradles, config.GradleInfo{
					Path:    fullPath,
					Version: version,
					Managed: false,
				})
			}
		}
	}

	return gradles
}

// detectGradleVersionFromPath tries to detect Gradle version from a path
func detectGradleVersionFromPath(gradlePath string) string {
	// Try running gradle --version
	gradleBin := filepath.Join(gradlePath, "bin", "gradle")
	if _, err := os.Stat(gradleBin); err == nil {
		cmd := exec.Command(gradleBin, "--version")
		output, err := cmd.Output()
		if err == nil {
			return parseGradleVersionOutput(string(output))
		}
	}
	return ""
}

// parseGradleVersionOutput parses gradle --version output
func parseGradleVersionOutput(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// Look for version line like: Gradle 6.9.4
		if strings.HasPrefix(line, "Gradle ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}
	return ""
}
