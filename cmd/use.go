package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/waguilars/java-environment-manager/internal/config"
	"github.com/waguilars/java-environment-manager/internal/jdk"
	"github.com/waguilars/java-environment-manager/internal/platform"
	"github.com/waguilars/java-environment-manager/internal/symlink"
)

// UseMode represents the mode of the use command
type UseMode int

const (
	// UseModeSession outputs environment variables for the current shell session
	UseModeSession UseMode = iota
	// UseModeDefault updates the default version in config and updates symlinks
	UseModeDefault
)

// UseCommand handles the 'jem use' command
type UseCommand struct {
	platform       platform.Platform
	configRepo     config.ConfigRepository
	jdkService     *jdk.JDKService
	prompter       Prompter
	force          bool
	outputEnv      bool
	symlinkManager *symlink.SymlinkManager
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

// SetOutputEnv sets the outputEnv flag
func (c *UseCommand) SetOutputEnv(outputEnv bool) {
	c.outputEnv = outputEnv
}

// SetSymlinkManager sets the symlink manager
func (c *UseCommand) SetSymlinkManager(manager *symlink.SymlinkManager) {
	c.symlinkManager = manager
}

// ExecuteJDK switches to a different JDK version
func (c *UseCommand) ExecuteJDK(ctx context.Context, jdkName string, mode UseMode) error {
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
		return fmt.Errorf("✗ JDK '%s' is not installed.\nRun 'jem install jdk %s' first, or use 'jem list jdk' to see available versions.", jdkName, jdkName)
	}

	// Use targetJDK.Path if available, otherwise construct from version
	jdkPath := targetJDK.Path
	if jdkPath == "" {
		jdkPath = filepath.Join(c.platform.HomeDir(), ".jem", "jdks", targetJDK.Version)
	}

	// Session mode: output environment exports
	if mode == UseModeSession || c.outputEnv {
		// Validate JDK directory exists
		if _, err := os.Stat(jdkPath); os.IsNotExist(err) {
			return fmt.Errorf("✗ JDK directory not found: %s\nRun 'jem doctor' for diagnostics", jdkPath)
		}

		// Output export statements for shell eval
		fmt.Printf("export JAVA_HOME=\"%s\"\n", jdkPath)
		fmt.Printf("export PATH=\"%s/bin:$PATH\"\n", jdkPath)
		return nil
	}

	// Default mode: update config and symlinks
	if mode == UseModeDefault {
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
			jdkPath = targetJDK.Path
		}

		// Validate JDK directory exists before updating
		if _, err := os.Stat(jdkPath); os.IsNotExist(err) {
			return fmt.Errorf("✗ JDK directory not found: %s\nRun 'jem doctor' for diagnostics", jdkPath)
		}

		// Update config default
		if err := c.configRepo.SetDefaultJDK(targetJDK.Version); err != nil {
			return fmt.Errorf("failed to update config: %w", err)
		}

		// Update current JDK in config
		if err := c.configRepo.SetJDKCurrent(targetJDK.Version); err != nil {
			return fmt.Errorf("failed to update config: %w", err)
		}

		// Use SymlinkManager if available, otherwise fall back to manual symlink management
		if c.symlinkManager != nil {
			if err := c.symlinkManager.UpdateCurrentJava(targetJDK.Version); err != nil {
				return fmt.Errorf("failed to update Java symlink: %w", err)
			}
		} else {
			// Legacy symlink management
			jemDir := filepath.Join(c.platform.HomeDir(), ".jem")
			jdksDir := filepath.Join(jemDir, "jdks")

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
		}

		fmt.Printf("✓ Default JDK set to %s (%s)\n", targetJDK.Version, jdkPath)
		return nil
	}

	return fmt.Errorf("unknown use mode")
}

// ExecuteGradle switches to a different Gradle version
func (c *UseCommand) ExecuteGradle(ctx context.Context, gradleName string, mode UseMode) error {
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
		return fmt.Errorf("✗ Gradle '%s' is not installed.\nRun 'jem install gradle %s' first, or use 'jem list gradle' to see available versions.", gradleName, gradleName)
	}

	// Use targetGradle.Path if available, otherwise construct from version
	gradlePath := targetGradle.Path
	if gradlePath == "" {
		gradlePath = filepath.Join(c.platform.HomeDir(), ".jem", "gradles", targetGradle.Version)
	}

	// Session mode: output environment exports
	if mode == UseModeSession || c.outputEnv {
		// Validate Gradle directory exists
		if _, err := os.Stat(gradlePath); os.IsNotExist(err) {
			return fmt.Errorf("✗ Gradle directory not found: %s\nRun 'jem doctor' for diagnostics", gradlePath)
		}

		// Output export statements for shell eval
		fmt.Printf("export GRADLE_HOME=\"%s\"\n", gradlePath)
		fmt.Printf("export PATH=\"%s/bin:$PATH\"\n", gradlePath)
		return nil
	}

	// Default mode: update config and symlinks
	if mode == UseModeDefault {
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
			gradlePath = targetGradle.Path
		}

		// Validate Gradle directory exists before updating
		if _, err := os.Stat(gradlePath); os.IsNotExist(err) {
			return fmt.Errorf("✗ Gradle directory not found: %s\nRun 'jem doctor' for diagnostics", gradlePath)
		}

		// Update config default
		if err := c.configRepo.SetDefaultGradle(targetGradle.Version); err != nil {
			return fmt.Errorf("failed to update config: %w", err)
		}

		// Update current Gradle in config
		if err := c.configRepo.SetGradleCurrent(targetGradle.Version); err != nil {
			return fmt.Errorf("failed to update config: %w", err)
		}

		// Use SymlinkManager if available, otherwise fall back to manual symlink management
		if c.symlinkManager != nil {
			if err := c.symlinkManager.UpdateCurrentGradle(targetGradle.Version); err != nil {
				return fmt.Errorf("failed to update Gradle symlink: %w", err)
			}
		} else {
			// Legacy symlink management
			jemDir := filepath.Join(c.platform.HomeDir(), ".jem")
			gradlesDir := filepath.Join(jemDir, "gradles")

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
		}

		fmt.Printf("✓ Default Gradle set to %s (%s)\n", targetGradle.Version, gradlePath)
		return nil
	}

	return fmt.Errorf("unknown use mode")
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

	// Validate JDK bin directory exists before creating symlink
	if _, err := os.Stat(jdkBinDir); os.IsNotExist(err) {
		return fmt.Errorf("JDK bin directory not found: %s", jdkBinDir)
	}

	// Remove existing bin (could be broken symlink, directory, or file)
	if _, err := os.Lstat(binDir); err == nil {
		// It exists - remove it regardless of type
		if err := os.RemoveAll(binDir); err != nil {
			return fmt.Errorf("failed to remove existing bin: %w", err)
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
