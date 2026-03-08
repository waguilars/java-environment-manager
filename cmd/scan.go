package cmd

import (
	"context"
	"fmt"

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
	// TODO: Implement Gradle detection

	// Print results
	fmt.Printf("\n✓ Detection complete!\n")
	fmt.Printf("  Found %d JDK(s)\n", len(detectedJDKs))

	// List detected JDKs
	if len(detectedJDKs) > 0 {
		fmt.Println("\nDetected JDKs:")
		for _, jdkInfo := range detectedJDKs {
			fmt.Printf("  - %s (%s)\n", jdkInfo.Version, jdkInfo.Path)
		}
	}

	return nil
}
