package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/jem/internal/config"
	"github.com/user/jem/internal/jdk"
	"github.com/user/jem/internal/platform"
)

// CommandFactory creates and wires up command instances with dependencies
type CommandFactory struct {
	ctx        context.Context
	platform   platform.Platform
	configRepo config.ConfigRepository
	jdkService *jdk.JDKService
}

// NewCommandFactory creates a new factory instance
func NewCommandFactory() (*CommandFactory, error) {
	// Detect platform
	plat := platform.Detect()

	// Determine config path
	homeDir := plat.HomeDir()
	configPath := filepath.Join(homeDir, ".jem", "config.toml")

	// Create config repository and load existing config
	configRepo := config.NewTOMLConfigRepository(configPath)
	if _, err := configRepo.Load(); err != nil {
		// Non-fatal: continue with empty config
		fmt.Fprintf(os.Stderr, "Warning: could not load config: %v\n", err)
	}

	// Create JDK service with dependencies
	jdkService := jdk.NewJDKService(plat, configRepo)

	return &CommandFactory{
		ctx:        context.Background(),
		platform:   plat,
		configRepo: configRepo,
		jdkService: jdkService,
	}, nil
}

// CreateSetupCommand creates a setup command instance
func (f *CommandFactory) CreateSetupCommand() *SetupCommand {
	return &SetupCommand{
		platform:   f.platform,
		configRepo: f.configRepo,
	}
}

// CreateScanCommand creates a scan command instance
func (f *CommandFactory) CreateScanCommand() *ScanCommand {
	return &ScanCommand{
		platform:   f.platform,
		configRepo: f.configRepo,
		jdkService: f.jdkService,
	}
}

// CreateListCommand creates a list command instance
func (f *CommandFactory) CreateListCommand() *ListCommand {
	return &ListCommand{
		configRepo: f.configRepo,
	}
}

// CreateCurrentCommand creates a current command instance
func (f *CommandFactory) CreateCurrentCommand() *CurrentCommand {
	return &CurrentCommand{
		configRepo: f.configRepo,
	}
}

// CreateUseCommand creates a use command instance
func (f *CommandFactory) CreateUseCommand() *UseCommand {
	return &UseCommand{
		platform:   f.platform,
		configRepo: f.configRepo,
		jdkService: f.jdkService,
		prompter:   &SurveyPrompter{},
	}
}

// CreateInstallCommand creates an install command instance
func (f *CommandFactory) CreateInstallCommand() *InstallCommand {
	return &InstallCommand{
		platform:   f.platform,
		configRepo: f.configRepo,
		jdkService: f.jdkService,
	}
}

// CreateImportCommand creates an import command instance
func (f *CommandFactory) CreateImportCommand() *ImportCommand {
	return &ImportCommand{
		platform:   f.platform,
		configRepo: f.configRepo,
		jdkService: f.jdkService,
	}
}

// Context returns the command context
func (f *CommandFactory) Context() context.Context {
	return f.ctx
}

// Platform returns the platform instance
func (f *CommandFactory) Platform() platform.Platform {
	return f.platform
}

// ConfigRepo returns the config repository
func (f *CommandFactory) ConfigRepo() config.ConfigRepository {
	return f.configRepo
}

// JDKService returns the JDK service
func (f *CommandFactory) JDKService() *jdk.JDKService {
	return f.jdkService
}

// Error handling utilities for CLI commands
func PrintSuccess(msg string) {
	fmt.Printf("✓ %s\n", msg)
}

func PrintError(msg string) {
	fmt.Fprintf(os.Stderr, "✗ %s\n", msg)
}

func PrintWarning(msg string) {
	fmt.Fprintf(os.Stderr, "⚠ %s\n", msg)
}

func PrintInfo(msg string) {
	fmt.Printf("ℹ %s\n", msg)
}

// FatalError exits with error message
func FatalError(msg string) {
	PrintError(msg)
	os.Exit(1)
}

// FatalErrorWithHint exits with error message and hint
func FatalErrorWithHint(msg, hint string) {
	PrintError(msg)
	if hint != "" {
		fmt.Fprintf(os.Stderr, "\nHint: %s\n", hint)
	}
	os.Exit(1)
}
