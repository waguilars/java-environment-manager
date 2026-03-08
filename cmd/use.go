package cmd

import (
	"context"
	"fmt"
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
}

// NewUseCommand creates a new UseCommand
func NewUseCommand(platform platform.Platform, configRepo config.ConfigRepository, jdkService *jdk.JDKService) *UseCommand {
	return &UseCommand{
		platform:   platform,
		configRepo: configRepo,
		jdkService: jdkService,
		prompter:   &SurveyPrompter{},
	}
}

// Execute runs the use command
func (c *UseCommand) Execute(ctx context.Context, jdkName string) error {
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
		if !c.prompter.Confirm(fmt.Sprintf("JDK '%s' is not managed by jem. Import it?", jdkName), false) {
			return fmt.Errorf("import cancelled by user")
		}
		if err := c.importJDK(ctx, targetJDK); err != nil {
			return fmt.Errorf("failed to import JDK: %w", err)
		}
	}

	// Update symlinks
	jemDir := filepath.Join(c.platform.HomeDir(), ".jem")
	jdkPath := filepath.Join(jemDir, "jdks", targetJDK.Version)

	// Create symlink to the JDK
	currentLink := filepath.Join(jemDir, "jdks", "current")
	if err := c.platform.CreateLink(jdkPath, currentLink); err != nil {
		return fmt.Errorf("failed to create current symlink: %w", err)
	}

	// Update bin symlinks
	if err := c.updateBinSymlinks(jdkPath); err != nil {
		return fmt.Errorf("failed to update bin symlinks: %w", err)
	}

	// Update config
	if err := c.configRepo.SetJDKCurrent(targetJDK.Version); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	fmt.Printf("✓ Now using JDK %s (%s)\n", targetJDK.Version, targetJDK.Path)

	return nil
}

// importJDK imports an external JDK into jem management
func (c *UseCommand) importJDK(ctx context.Context, jdkInfo *config.JDKInfo) error {
	homeDir := c.platform.HomeDir()
	jemDir := filepath.Join(homeDir, ".jem")
	importPath := filepath.Join(jemDir, "jdks", filepath.Base(jdkInfo.Path))

	// Create symlink to the external JDK
	if err := c.platform.CreateLink(jdkInfo.Path, importPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	// Update config
	newJDKInfo := *jdkInfo
	newJDKInfo.Managed = true
	newJDKInfo.Path = importPath

	if err := c.configRepo.AddInstalledJDK(newJDKInfo); err != nil {
		return fmt.Errorf("failed to add installed JDK: %w", err)
	}

	// Remove from detected JDKs
	if err := c.configRepo.RemoveInstalledJDK(jdkInfo.Version); err != nil {
		return fmt.Errorf("failed to remove detected JDK: %w", err)
	}

	return nil
}

// updateBinSymlinks creates symlinks for java, javac, etc. in ~/.jem/bin
func (c *UseCommand) updateBinSymlinks(jdkPath string) error {
	binDir := filepath.Join(c.platform.HomeDir(), ".jem", "bin")
	jdkBinDir := filepath.Join(jdkPath, "bin")

	// Create bin directory if it doesn't exist
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
