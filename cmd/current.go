package cmd

import (
	"fmt"

	"github.com/user/jem/internal/config"
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
		// Find the JDK info
		installedJDKs := c.configRepo.ListInstalledJDKs()
		detectedJDKs := c.configRepo.ListDetectedJDKs()
		allJDKs := append(installedJDKs, detectedJDKs...)

		found := false
		for _, jdkInfo := range allJDKs {
			if jdkInfo.Version == currentJDK {
				fmt.Printf("JDK:     %s (%s)\n", jdkInfo.Version, jdkInfo.Path)
				found = true
				break
			}
		}

		if !found {
			fmt.Printf("JDK:     %s (not found)\n", currentJDK)
		}
	} else {
		fmt.Println("JDK:     (not configured)")
	}

	// Print current Gradle
	if currentGradle != "" {
		// Find the Gradle info
		installedGradles := c.configRepo.ListInstalledGradles()
		detectedGradles := c.configRepo.ListDetectedGradles()
		allGradles := append(installedGradles, detectedGradles...)

		found := false
		for _, gradleInfo := range allGradles {
			if gradleInfo.Version == currentGradle {
				fmt.Printf("Gradle:  %s (%s)\n", gradleInfo.Version, gradleInfo.Path)
				found = true
				break
			}
		}

		if !found {
			fmt.Printf("Gradle:  %s (not found)\n", currentGradle)
		}
	} else {
		fmt.Println("Gradle:  (not configured)")
	}

	return nil
}
