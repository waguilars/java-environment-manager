package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/user/jem/internal/config"
	"github.com/user/jem/internal/jdk"
	"github.com/user/jem/internal/platform"
)

// UseCommand handles the 'jem use' command
type UseCommand struct {
	platform   platform.Platform
	configRepo config.ConfigRepository
	jdkService *jdk.JDKService
	prompter   Prompter
	force      bool
}

// NewUseCommand creates a new UseCommand
func NewUseCommand(platform platform.Platform, configRepo config.ConfigRepository, jdkService *jdk.JDKService) *UseCommand {
	return &UseCommand{
		platform:   platform,
		configRepo: configRepo,
		jdkService: jdkService,
		prompter:   &SurveyPrompter{},
		force:      false,
	}
}

// SetForce sets the force flag
func (c *UseCommand) SetForce(force bool) {
	c.force = force
}

// ExecuteJDK switches to a different JDK version
func (c *UseCommand) ExecuteJDK(ctx context.Context, jdkName string) error {
	// Validate JDK exists
	installedJDKs := c.configRepo.ListInstalledJDKs()
	detectedJDKs := c.configRepo.ListDetectedJDKs()

	var targetJDK *config.JDKInfo

	// Check installed JDKs
	for i := range installedJDKs {
		if installedJDKs[i].Version == jdkName {
			targetJDK = &installedJDKs[i]
			break
		}
	}

	// Check detected JDKs if not found in installed
	if targetJDK == nil {
		for i := range detectedJDKs {
			if detectedJDKs[i].Version == jdkName {
				targetJDK = &detectedJDKs[i]
				break
			}
		}
	}

	if targetJDK == nil {
		return fmt.Errorf("JDK '%s' not found", jdkName)
	}

	// If JDK is detected (not installed), prompt to import
	if !targetJDK.Managed {
		shouldImport := c.force
		if !c.force {
			shouldImport = c.prompter.Confirm(fmt.Sprintf("JDK '%s' is not managed by jem. Import it?", jdkName), false)
		}

		if !shouldImport {
			return fmt.Errorf("import cancelled by user")
		}
		if err := c.importJDK(ctx, targetJDK); err != nil {
			return fmt.Errorf("failed to import JDK: %w", err)
		}
		// Update targetJDK path to the imported location
		targetJDK.Path = filepath.Join(c.platform.HomeDir(), ".jem", "jdks", filepath.Base(targetJDK.Path))
	}

	// Update symlinks
	jemDir := filepath.Join(c.platform.HomeDir(), ".jem")
	jdksDir := filepath.Join(jemDir, "jdks")
	jdkPath := filepath.Join(jdksDir, targetJDK.Version)

	// Ensure jdks directory exists
	if err := os.MkdirAll(jdksDir, 0755); err != nil {
		return fmt.Errorf("failed to create jdks directory: %w", err)
	}

	// Create symlink to the JDK (remove existing first)
	currentLink := filepath.Join(jdksDir, "current")
	if _, err := os.Lstat(currentLink); err == nil {
		if err := os.Remove(currentLink); err != nil {
			return fmt.Errorf("failed to remove existing current symlink: %w", err)
		}
	}
	if err := c.platform.CreateLink(jdkPath, currentLink); err != nil {
		return fmt.Errorf("failed to create current symlink: %w", err)
	}

	// Update bin symlinks
	if err := c.updateJDKBinSymlinks(jdkPath); err != nil {
		return fmt.Errorf("failed to update bin symlinks: %w", err)
	}

	// Update config
	if err := c.configRepo.SetJDKCurrent(targetJDK.Version); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	fmt.Printf("✓ Now using JDK %s (%s)\n", targetJDK.Version, targetJDK.Path)

	return nil
}

// ExecuteGradle switches to a different Gradle version
func (c *UseCommand) ExecuteGradle(ctx context.Context, gradleName string) error {
	// Validate Gradle exists
	installedGradles := c.configRepo.ListInstalledGradles()
	detectedGradles := c.configRepo.ListDetectedGradles()

	var targetGradle *config.GradleInfo

	// Check installed Gradles
	for i := range installedGradles {
		if installedGradles[i].Version == gradleName {
			targetGradle = &installedGradles[i]
			break
		}
	}

	// Check detected Gradles if not found in installed
	if targetGradle == nil {
		for i := range detectedGradles {
			if detectedGradles[i].Version == gradleName {
				targetGradle = &detectedGradles[i]
				break
			}
		}
	}

	if targetGradle == nil {
		return fmt.Errorf("Gradle '%s' not found", gradleName)
	}

	// If Gradle is detected (not installed), prompt to import
	if !targetGradle.Managed {
		shouldImport := c.force
		if !c.force {
			shouldImport = c.prompter.Confirm(fmt.Sprintf("Gradle '%s' is not managed by jem. Import it?", gradleName), false)
		}

		if !shouldImport {
			return fmt.Errorf("import cancelled by user")
		}
		if err := c.importGradle(ctx, targetGradle); err != nil {
			return fmt.Errorf("failed to import Gradle: %w", err)
		}
		// Update targetGradle path to the imported location
		targetGradle.Path = filepath.Join(c.platform.HomeDir(), ".jem", "gradles", filepath.Base(targetGradle.Path))
	}

	// Update symlinks
	jemDir := filepath.Join(c.platform.HomeDir(), ".jem")
	gradlesDir := filepath.Join(jemDir, "gradles")
	gradlePath := filepath.Join(gradlesDir, targetGradle.Version)

	// Ensure gradles directory exists
	if err := os.MkdirAll(gradlesDir, 0755); err != nil {
		return fmt.Errorf("failed to create gradles directory: %w", err)
	}

	// Create symlink to the Gradle (remove existing first)
	currentLink := filepath.Join(gradlesDir, "current")
	if _, err := os.Lstat(currentLink); err == nil {
		if err := os.Remove(currentLink); err != nil {
			return fmt.Errorf("failed to remove existing current symlink: %w", err)
		}
	}
	if err := c.platform.CreateLink(gradlePath, currentLink); err != nil {
		return fmt.Errorf("failed to create current symlink: %w", err)
	}

	// Update config
	if err := c.configRepo.SetGradleCurrent(targetGradle.Version); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	fmt.Printf("✓ Now using Gradle %s (%s)\n", targetGradle.Version, targetGradle.Path)

	return nil
}

// importJDK imports an external JDK into jem management
func (c *UseCommand) importJDK(ctx context.Context, jdkInfo *config.JDKInfo) error {
	homeDir := c.platform.HomeDir()
	jemDir := filepath.Join(homeDir, ".jem")
	jdksDir := filepath.Join(jemDir, "jdks")
	importPath := filepath.Join(jdksDir, filepath.Base(jdkInfo.Path))

	// Create jdks directory if it doesn't exist
	if err := os.MkdirAll(jdksDir, 0755); err != nil {
		return fmt.Errorf("failed to create jdks directory: %w", err)
	}

	// Check if symlink already exists
	if _, err := os.Lstat(importPath); err == nil {
		// Symlink exists, remove it first
		if err := os.Remove(importPath); err != nil {
			return fmt.Errorf("failed to remove existing symlink: %w", err)
		}
	}

	// Create symlink to the external JDK
	if err := c.platform.CreateLink(jdkInfo.Path, importPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	// Update config - add to installed JDKs
	newJDKInfo := config.JDKInfo{
		Path:     importPath,
		Version:  jdkInfo.Version,
		Provider: "imported",
		Managed:  true,
	}

	if err := c.configRepo.AddInstalledJDK(newJDKInfo); err != nil {
		return fmt.Errorf("failed to add installed JDK: %w", err)
	}

	fmt.Printf("✓ Imported JDK %s\n", jdkInfo.Version)

	return nil
}

// importGradle imports an external Gradle into jem management
func (c *UseCommand) importGradle(ctx context.Context, gradleInfo *config.GradleInfo) error {
	homeDir := c.platform.HomeDir()
	jemDir := filepath.Join(homeDir, ".jem")
	gradlesDir := filepath.Join(jemDir, "gradles")
	importPath := filepath.Join(gradlesDir, filepath.Base(gradleInfo.Path))

	// Create gradles directory if it doesn't exist
	if err := os.MkdirAll(gradlesDir, 0755); err != nil {
		return fmt.Errorf("failed to create gradles directory: %w", err)
	}

	// Check if symlink already exists
	if _, err := os.Lstat(importPath); err == nil {
		// Symlink exists, remove it first
		if err := os.Remove(importPath); err != nil {
			return fmt.Errorf("failed to remove existing symlink: %w", err)
		}
	}

	// Create symlink to the external Gradle
	if err := c.platform.CreateLink(gradleInfo.Path, importPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	// Update config - add to installed Gradles
	newGradleInfo := config.GradleInfo{
		Path:    importPath,
		Version: gradleInfo.Version,
		Managed: true,
	}

	if err := c.configRepo.AddInstalledGradle(newGradleInfo); err != nil {
		return fmt.Errorf("failed to add installed Gradle: %w", err)
	}

	fmt.Printf("✓ Imported Gradle %s\n", gradleInfo.Version)

	return nil
}

// updateJDKBinSymlinks creates symlinks for java, javac, etc. in ~/.jem/bin
func (c *UseCommand) updateJDKBinSymlinks(jdkPath string) error {
	binDir := filepath.Join(c.platform.HomeDir(), ".jem", "bin")
	jdkBinDir := filepath.Join(jdkPath, "bin")

	// Remove existing bin symlink if it exists
	if _, err := os.Lstat(binDir); err == nil {
		if err := os.Remove(binDir); err != nil {
			return fmt.Errorf("failed to remove existing bin symlink: %w", err)
		}
	}

	// Create bin symlink
	if err := c.platform.CreateLink(jdkBinDir, binDir); err != nil {
		return fmt.Errorf("failed to create bin symlink: %w", err)
	}

	return nil
}

// Prompter interface for interactive prompts
type Prompter interface {
	Confirm(message string, defaultValue bool) bool
	Select(message string, options []string, defaultValue string) string
	Input(message string, defaultValue string) string
}

// SurveyPrompter implements Prompter using survey
type SurveyPrompter struct{}

// Confirm shows a confirmation prompt
func (p *SurveyPrompter) Confirm(message string, defaultValue bool) bool {
	var result bool
	prompt := &survey.Confirm{
		Message: message,
		Default: defaultValue,
	}
	_ = survey.AskOne(prompt, &result)
	return result
}

// Select shows a selection prompt
func (p *SurveyPrompter) Select(message string, options []string, defaultValue string) string {
	var result string
	prompt := &survey.Select{
		Message: message,
		Options: options,
		Default: defaultValue,
	}
	_ = survey.AskOne(prompt, &result)
	return result
}

// Input shows an input prompt
func (p *SurveyPrompter) Input(message string, defaultValue string) string {
	var result string
	prompt := &survey.Input{
		Message: message,
		Default: defaultValue,
	}
	_ = survey.AskOne(prompt, &result)
	return result
}
