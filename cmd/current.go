package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/user/jem/internal/config"
	"github.com/user/jem/internal/jdk"
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
