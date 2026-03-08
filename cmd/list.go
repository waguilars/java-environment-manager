package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/user/jem/internal/config"
)

// ListCommand handles the 'jem list' command
type ListCommand struct {
	configRepo config.ConfigRepository
}

// Execute runs the list command
func (c *ListCommand) Execute() error {
	installedJDKs := c.configRepo.ListInstalledJDKs()
	detectedJDKs := c.configRepo.ListDetectedJDKs()
	currentJDK := c.configRepo.GetJDKCurrent()

	// Create tab writer for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Print installed JDKs
	fmt.Println("\n=== Installed JDKs ===")
	if len(installedJDKs) == 0 {
		fmt.Println("No JDKs installed. Use 'jem install jdk <version>' to install.")
	} else {
		for _, jdkInfo := range installedJDKs {
			currentMarker := ""
			if jdkInfo.Version == currentJDK {
				currentMarker = " [active]"
			}
			fmt.Fprintf(w, "%s\t(%s)%s\n", jdkInfo.Version, jdkInfo.Path, currentMarker)
		}
		w.Flush()
	}

	// Print detected JDKs
	fmt.Println("\n=== Detected JDKs ===")
	if len(detectedJDKs) == 0 {
		fmt.Println("No JDKs detected. Use 'jem scan' to detect JDKs on your system.")
	} else {
		for _, jdkInfo := range detectedJDKs {
			currentMarker := ""
			if jdkInfo.Version == currentJDK {
				currentMarker = " [active]"
			}
			fmt.Fprintf(w, "%s\t(%s)%s\n", jdkInfo.Version, jdkInfo.Path, currentMarker)
		}
		w.Flush()
	}

	// Print current JDK
	fmt.Println("\n=== Current JDK ===")
	if currentJDK != "" {
		// Find the current JDK info
		allJDKs := append(installedJDKs, detectedJDKs...)
		for _, jdkInfo := range allJDKs {
			if jdkInfo.Version == currentJDK {
				fmt.Printf("Active: %s (%s)\n", jdkInfo.Version, jdkInfo.Path)
				break
			}
		}
	} else {
		fmt.Println("No JDK currently active. Use 'jem use jdk <version>' to set.")
	}

	return nil
}
