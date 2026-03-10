package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/waguilars/java-environment-manager/internal/config"
	"github.com/waguilars/java-environment-manager/internal/jdk"
)

// CurrentCommand handles the 'jem current' command
type CurrentCommand struct {
	configRepo config.ConfigRepository
}

// Execute runs the current command
func (c *CurrentCommand) Execute() error {
	currentJDK := c.configRepo.GetJDKCurrent()
	currentGradle := c.configRepo.GetGradleCurrent()

	fmt.Println("=== Current Environment ===")

	// Print current JDK
	if currentJDK != "" {
		// Find the JDK info from jem config
		installedJDKs := c.configRepo.ListInstalledJDKs()
		detectedJDKs := c.configRepo.ListDetectedJDKs()
		allJDKs := append(installedJDKs, detectedJDKs...)

		found := false
		for _, jdkInfo := range allJDKs {
			if jdkInfo.Version == currentJDK {
				fmt.Printf("JDK:     %s (%s) [jem]\n", jdkInfo.Version, jdkInfo.Path)
				found = true
				break
			}
		}

		if !found {
			fmt.Printf("JDK:     %s (not found)\n", currentJDK)
		}

		// Version consistency check: compare configured version with actual java -version
		actualVersion := detectActualJavaVersion()
		if actualVersion != "" {
			configuredMajor := extractMajorVersion(currentJDK)
			actualMajor := extractMajorVersion(actualVersion)
			if configuredMajor != "" && actualMajor != "" && configuredMajor != actualMajor {
				fmt.Printf("⚠ Configured version (%s) does not match active Java (%s)\n", configuredMajor, actualMajor)
				fmt.Println("  Check PATH priority or run 'jem doctor'")
			}
		} else {
			fmt.Println("⚠ java executable not found in PATH")
		}
	} else {
		// No JDK configured in jem, try to detect from system
		systemJava := jdk.DetectSystemJava()
		if systemJava != nil {
			fmt.Printf("JDK:     %s (%s) [system]\n", systemJava.Version, systemJava.Path)
		} else {
			fmt.Println("JDK:     (not configured)")
		}
	}

	// Print current Gradle
	if currentGradle != "" {
		// Find the Gradle info from jem config
		installedGradles := c.configRepo.ListInstalledGradles()
		detectedGradles := c.configRepo.ListDetectedGradles()
		allGradles := append(installedGradles, detectedGradles...)

		found := false
		for _, gradleInfo := range allGradles {
			if gradleInfo.Version == currentGradle {
				fmt.Printf("Gradle:  %s (%s) [jem]\n", gradleInfo.Version, gradleInfo.Path)
				found = true
				break
			}
		}

		if !found {
			fmt.Printf("Gradle:  %s (not found)\n", currentGradle)
		}
	} else {
		// No Gradle configured in jem, try to detect from system
		systemGradle := detectSystemGradle()
		if systemGradle != nil {
			fmt.Printf("Gradle:  %s (%s) [system]\n", systemGradle.Version, systemGradle.Path)
		} else {
			fmt.Println("Gradle:  (not configured)")
		}
	}

	return nil
}

// detectSystemGradle detects the currently active Gradle from GRADLE_HOME or PATH
func detectSystemGradle() *config.GradleInfo {
	// Try GRADLE_HOME first
	gradleHome := os.Getenv("GRADLE_HOME")
	if gradleHome != "" {
		version := detectGradleVersionFromPath(gradleHome)
		if version != "" {
			return &config.GradleInfo{
				Path:    gradleHome,
				Version: version,
				Managed: false,
			}
		}
	}

	// Try gradle from PATH
	gradlePath, err := exec.LookPath("gradle")
	if err == nil {
		// Resolve symlink to find the actual GRADLE_HOME
		realPath, err := filepath.EvalSymlinks(gradlePath)
		if err == nil {
			// gradle is typically in bin/ directory, go up to find GRADLE_HOME
			gradleHome := filepath.Dir(filepath.Dir(realPath))
			version := detectGradleVersionFromPath(gradleHome)
			if version != "" {
				return &config.GradleInfo{
					Path:    gradleHome,
					Version: version,
					Managed: false,
				}
			}
		}
	}

	return nil
}

// detectActualJavaVersion executes java -version and parses the output
func detectActualJavaVersion() string {
	cmd := exec.Command("java", "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	return parseJavaVersion(string(output))
}

// parseJavaVersion parses the version string from java -version output
// Handles both OpenJDK and Oracle JDK formats
func parseJavaVersion(output string) string {
	// Parse OpenJDK format: openjdk version "21.0.2" 2024-01-16 LTS
	// Parse Oracle format: java version "21.0.2" 2024-01-16 LTS
	lines := strings.Split(output, "\n")
	if len(lines) == 0 {
		return ""
	}

	firstLine := lines[0]

	// Extract version between quotes
	start := strings.Index(firstLine, "\"")
	if start == -1 {
		return ""
	}
	end := strings.Index(firstLine[start+1:], "\"")
	if end == -1 {
		return ""
	}

	return firstLine[start+1 : start+1+end]
}

// extractMajorVersion extracts the major version number from a version string
// Handles formats like "temurin-21.0.2", "21.0.2", "17"
func extractMajorVersion(version string) string {
	// Remove any prefix before the number (e.g., "temurin-" from "temurin-21.0.2")
	re := regexp.MustCompile(`(\d+)`)
	match := re.FindStringSubmatch(version)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}
