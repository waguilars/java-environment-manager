package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/waguilars/java-environment-manager/internal/config"
	"github.com/waguilars/java-environment-manager/internal/jdk"
	"github.com/waguilars/java-environment-manager/internal/platform"
	"github.com/waguilars/java-environment-manager/internal/symlink"
)

// MockPrompter implements Prompter for testing
type MockPrompter struct {
	ConfirmFunc  func(message string, defaultValue bool) bool
	SelectFunc   func(message string, options []string, defaultValue string) string
	InputFunc    func(message string, defaultValue string) string
	ConfirmCalls int
}

func (m *MockPrompter) Confirm(message string, defaultValue bool) bool {
	m.ConfirmCalls++
	if m.ConfirmFunc != nil {
		return m.ConfirmFunc(message, defaultValue)
	}
	return true
}

func (m *MockPrompter) Select(message string, options []string, defaultValue string) string {
	if m.SelectFunc != nil {
		return m.SelectFunc(message, options, defaultValue)
	}
	return options[0]
}

func (m *MockPrompter) Input(message string, defaultValue string) string {
	if m.InputFunc != nil {
		return m.InputFunc(message, defaultValue)
	}
	return defaultValue
}

// MockPlatformForUse creates a mock platform with custom paths
type MockPlatformForUse struct {
	platform.LinuxPlatform
	HomeDirFunc     func() string
	CreateLinkFunc  func(target, link string) error
	RemoveLinkFunc  func(link string) error
	IsLinkFunc      func(path string) bool
	JDKPathsFunc    func() []string
	GradlePathsFunc func() []string
}

func (m *MockPlatformForUse) HomeDir() string {
	if m.HomeDirFunc != nil {
		return m.HomeDirFunc()
	}
	return "/tmp"
}

func (m *MockPlatformForUse) CreateLink(target, link string) error {
	if m.CreateLinkFunc != nil {
		return m.CreateLinkFunc(target, link)
	}
	return nil
}

func (m *MockPlatformForUse) RemoveLink(link string) error {
	if m.RemoveLinkFunc != nil {
		return m.RemoveLinkFunc(link)
	}
	return nil
}

func (m *MockPlatformForUse) IsLink(path string) bool {
	if m.IsLinkFunc != nil {
		return m.IsLinkFunc(path)
	}
	return false
}

func (m *MockPlatformForUse) JDKDetectionPaths() []string {
	if m.JDKPathsFunc != nil {
		return m.JDKPathsFunc()
	}
	return []string{}
}

func (m *MockPlatformForUse) GradleDetectionPaths() []string {
	if m.GradlePathsFunc != nil {
		return m.GradlePathsFunc()
	}
	return []string{}
}

func TestUseCommand_ExecuteJDK_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	platform := &MockPlatformForUse{}
	prompter := &MockPrompter{}
	jdkService := &jdk.JDKService{}

	cmd := &UseCommand{
		platform:   platform,
		configRepo: repo,
		jdkService: jdkService,
		prompter:   prompter,
		force:      false,
	}

	err = cmd.ExecuteJDK(context.Background(), "non-existent-version", UseModeDefault)
	if err == nil {
		t.Error("Expected error for non-existent JDK")
	}
}

func TestUseCommand_ExecuteJDK_Installed(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Add an installed JDK - create the actual directory structure
	jdkPath := filepath.Join(tmpDir, ".jem", "jdks", "21.0.1")
	jdkBinPath := filepath.Join(jdkPath, "bin")
	if err := os.MkdirAll(jdkBinPath, 0755); err != nil {
		t.Fatalf("Failed to create JDK dir: %v", err)
	}

	repo.AddInstalledJDK(config.JDKInfo{
		Path:     jdkPath,
		Version:  "21.0.1",
		Provider: "temurin",
		Managed:  true,
	})

	platform := &MockPlatformForUse{
		HomeDirFunc: func() string { return tmpDir },
		CreateLinkFunc: func(target, link string) error {
			return os.Symlink(target, link)
		},
		IsLinkFunc: func(path string) bool {
			_, err := os.Lstat(path)
			return err == nil && (os.FileMode(0)&os.ModeSymlink != 0)
		},
	}

	prompter := &MockPrompter{}
	jdkService := &jdk.JDKService{}

	cmd := &UseCommand{
		platform:       platform,
		configRepo:     repo,
		jdkService:     jdkService,
		prompter:       prompter,
		force:          false,
		symlinkManager: symlink.NewSymlinkManager(platform),
	}

	err = cmd.ExecuteJDK(context.Background(), "21.0.1", UseModeDefault)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestUseCommand_ExecuteJDK_Detected_Import(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Add a detected JDK (not managed) - create the actual directory structure with bin
	detectedPath := filepath.Join(tmpDir, "detected-jdk")
	detectedBinPath := filepath.Join(detectedPath, "bin")
	if err := os.MkdirAll(detectedBinPath, 0755); err != nil {
		t.Fatalf("Failed to create detected JDK dir: %v", err)
	}

	repo.AddDetectedJDK(config.JDKInfo{
		Path:     detectedPath,
		Version:  "17.0.10",
		Provider: "corretto",
		Managed:  false,
	})

	platform := &MockPlatformForUse{
		HomeDirFunc: func() string { return tmpDir },
		CreateLinkFunc: func(target, link string) error {
			// Actually create the symlink so validation passes
			return os.Symlink(target, link)
		},
		IsLinkFunc: func(path string) bool {
			_, err := os.Lstat(path)
			return err == nil
		},
	}

	prompter := &MockPrompter{
		ConfirmFunc: func(message string, defaultValue bool) bool {
			return true // Simulate user confirming import
		},
	}

	jdkService := &jdk.JDKService{}

	cmd := &UseCommand{
		platform:       platform,
		configRepo:     repo,
		jdkService:     jdkService,
		prompter:       prompter,
		force:          false,
		symlinkManager: symlink.NewSymlinkManager(platform),
	}

	err = cmd.ExecuteJDK(context.Background(), "17.0.10", UseModeDefault)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify the JDK was imported
	installedJDKs := repo.ListInstalledJDKs()
	if len(installedJDKs) != 1 {
		t.Errorf("Expected 1 installed JDK, got %d", len(installedJDKs))
	}

	if installedJDKs[0].Managed != true {
		t.Error("Expected imported JDK to be managed")
	}
}

func TestUseCommand_ExecuteJDK_Detected_SkipImport(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Add a detected JDK (not managed)
	detectedPath := filepath.Join(tmpDir, "detected-jdk")
	if err := os.MkdirAll(detectedPath, 0755); err != nil {
		t.Fatalf("Failed to create detected JDK dir: %v", err)
	}

	repo.AddDetectedJDK(config.JDKInfo{
		Path:     detectedPath,
		Version:  "17.0.10",
		Provider: "corretto",
		Managed:  false,
	})

	platform := &MockPlatformForUse{
		HomeDirFunc: func() string { return tmpDir },
		CreateLinkFunc: func(target, link string) error {
			return nil
		},
	}

	prompter := &MockPrompter{
		ConfirmFunc: func(message string, defaultValue bool) bool {
			return false // Simulate user declining import
		},
	}

	jdkService := &jdk.JDKService{}

	cmd := &UseCommand{
		platform:   platform,
		configRepo: repo,
		jdkService: jdkService,
		prompter:   prompter,
		force:      false,
	}

	err = cmd.ExecuteJDK(context.Background(), "17.0.10", UseModeDefault)
	if err == nil {
		t.Error("Expected error when import is cancelled")
	}
}

func TestUseCommand_ExecuteJDK_Force(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Add an installed JDK - create the actual directory structure
	jdkPath := filepath.Join(tmpDir, ".jem", "jdks", "21.0.1")
	jdkBinPath := filepath.Join(jdkPath, "bin")
	if err := os.MkdirAll(jdkBinPath, 0755); err != nil {
		t.Fatalf("Failed to create JDK dir: %v", err)
	}

	repo.AddInstalledJDK(config.JDKInfo{
		Path:     jdkPath,
		Version:  "21.0.1",
		Provider: "temurin",
		Managed:  true,
	})

	platform := &MockPlatformForUse{
		HomeDirFunc: func() string { return tmpDir },
		CreateLinkFunc: func(target, link string) error {
			return os.Symlink(target, link)
		},
		IsLinkFunc: func(path string) bool {
			_, err := os.Lstat(path)
			return err == nil
		},
	}

	prompter := &MockPrompter{}
	jdkService := &jdk.JDKService{}

	cmd := &UseCommand{
		platform:       platform,
		configRepo:     repo,
		jdkService:     jdkService,
		prompter:       prompter,
		force:          true, // Force mode
		symlinkManager: symlink.NewSymlinkManager(platform),
	}

	err = cmd.ExecuteJDK(context.Background(), "21.0.1", UseModeDefault)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestUseCommand_ExecuteGradle_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	platform := &MockPlatformForUse{}
	prompter := &MockPrompter{}
	jdkService := &jdk.JDKService{}

	cmd := &UseCommand{
		platform:   platform,
		configRepo: repo,
		jdkService: jdkService,
		prompter:   prompter,
		force:      false,
	}

	err = cmd.ExecuteGradle(context.Background(), "non-existent-version", UseModeDefault)
	if err == nil {
		t.Error("Expected error for non-existent Gradle")
	}
}

func TestUseCommand_ExecuteGradle_Installed(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Add an installed Gradle - create the actual directory structure
	gradlePath := filepath.Join(tmpDir, ".jem", "gradles", "7.6.1")
	if err := os.MkdirAll(gradlePath, 0755); err != nil {
		t.Fatalf("Failed to create Gradle dir: %v", err)
	}

	repo.AddInstalledGradle(config.GradleInfo{
		Path:    gradlePath,
		Version: "7.6.1",
		Managed: true,
	})

	platform := &MockPlatformForUse{
		HomeDirFunc: func() string { return tmpDir },
		CreateLinkFunc: func(target, link string) error {
			return os.Symlink(target, link)
		},
		IsLinkFunc: func(path string) bool {
			_, err := os.Lstat(path)
			return err == nil
		},
	}

	prompter := &MockPrompter{}
	jdkService := &jdk.JDKService{}

	cmd := &UseCommand{
		platform:       platform,
		configRepo:     repo,
		jdkService:     jdkService,
		prompter:       prompter,
		force:          false,
		symlinkManager: symlink.NewSymlinkManager(platform),
	}

	err = cmd.ExecuteGradle(context.Background(), "7.6.1", UseModeDefault)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestUseCommand_ImportJDK(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create an external JDK directory
	externalJDK := filepath.Join(tmpDir, "external-jdk")
	if err := os.MkdirAll(externalJDK, 0755); err != nil {
		t.Fatalf("Failed to create external JDK: %v", err)
	}

	platform := &MockPlatformForUse{
		HomeDirFunc: func() string { return tmpDir },
		CreateLinkFunc: func(target, link string) error {
			return nil
		},
	}

	prompter := &MockPrompter{}
	jdkService := &jdk.JDKService{}

	cmd := &UseCommand{
		platform:   platform,
		configRepo: repo,
		jdkService: jdkService,
		prompter:   prompter,
		force:      false,
	}

	jdkInfo := &config.JDKInfo{
		Path:     externalJDK,
		Version:  "21.0.1",
		Provider: "system",
		Managed:  false,
	}

	err = cmd.importJDK(context.Background(), jdkInfo)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify the JDK was added to installed
	installedJDKs := repo.ListInstalledJDKs()
	if len(installedJDKs) != 1 {
		t.Errorf("Expected 1 installed JDK, got %d", len(installedJDKs))
	}

	if installedJDKs[0].Managed != true {
		t.Error("Expected imported JDK to be managed")
	}

	if installedJDKs[0].Provider != "imported" {
		t.Errorf("Expected provider 'imported', got '%s'", installedJDKs[0].Provider)
	}
}

func TestUseCommand_ImportGradle(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create an external Gradle directory
	externalGradle := filepath.Join(tmpDir, "external-gradle")
	if err := os.MkdirAll(externalGradle, 0755); err != nil {
		t.Fatalf("Failed to create external Gradle: %v", err)
	}

	platform := &MockPlatformForUse{
		HomeDirFunc: func() string { return tmpDir },
		CreateLinkFunc: func(target, link string) error {
			return nil
		},
	}

	prompter := &MockPrompter{}
	jdkService := &jdk.JDKService{}

	cmd := &UseCommand{
		platform:   platform,
		configRepo: repo,
		jdkService: jdkService,
		prompter:   prompter,
		force:      false,
	}

	gradleInfo := &config.GradleInfo{
		Path:    externalGradle,
		Version: "7.6.1",
		Managed: false,
	}

	err = cmd.importGradle(context.Background(), gradleInfo)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify the Gradle was added to installed
	installedGradles := repo.ListInstalledGradles()
	if len(installedGradles) != 1 {
		t.Errorf("Expected 1 installed Gradle, got %d", len(installedGradles))
	}

	if installedGradles[0].Managed != true {
		t.Error("Expected imported Gradle to be managed")
	}
}

func TestUseCommand_ExecuteJDK_InvalidPath(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Add an installed JDK but don't create the directory (simulates corrupted config)
	jdkPath := filepath.Join(tmpDir, ".jem", "jdks", "nonexistent-17")
	// Note: we don't create the directory

	repo.AddInstalledJDK(config.JDKInfo{
		Path:     jdkPath,
		Version:  "nonexistent-17",
		Provider: "temurin",
		Managed:  true,
	})

	platform := &MockPlatformForUse{
		HomeDirFunc: func() string { return tmpDir },
	}

	prompter := &MockPrompter{}
	jdkService := &jdk.JDKService{}

	cmd := &UseCommand{
		platform:   platform,
		configRepo: repo,
		jdkService: jdkService,
		prompter:   prompter,
		force:      false,
	}

	err = cmd.ExecuteJDK(context.Background(), "nonexistent-17", UseModeDefault)
	if err == nil {
		t.Error("Expected error for non-existent JDK path")
	}
	if !contains(err.Error(), "JDK directory not found") {
		t.Errorf("Expected error to contain 'JDK directory not found', got: %v", err)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Tests for Session Mode (Phase 3)

func TestUseCommand_ExecuteJDK_SessionMode(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Add an installed JDK - create the actual directory structure
	jdkPath := filepath.Join(tmpDir, ".jem", "jdks", "21.0.1")
	jdkBinPath := filepath.Join(jdkPath, "bin")
	if err := os.MkdirAll(jdkBinPath, 0755); err != nil {
		t.Fatalf("Failed to create JDK dir: %v", err)
	}

	repo.AddInstalledJDK(config.JDKInfo{
		Path:     jdkPath,
		Version:  "21.0.1",
		Provider: "temurin",
		Managed:  true,
	})

	platform := &MockPlatformForUse{
		HomeDirFunc: func() string { return tmpDir },
	}

	prompter := &MockPrompter{}
	jdkService := &jdk.JDKService{}

	cmd := &UseCommand{
		platform:   platform,
		configRepo: repo,
		jdkService: jdkService,
		prompter:   prompter,
		force:      false,
	}

	// Use session mode - should output env vars, not update symlinks
	err = cmd.ExecuteJDK(context.Background(), "21.0.1", UseModeSession)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// In session mode, config should NOT be updated
	if repo.GetJDKCurrent() == "21.0.1" {
		t.Error("Session mode should not update current JDK in config")
	}
}

func TestUseCommand_ExecuteJDK_SessionMode_Uninstalled(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	platform := &MockPlatformForUse{
		HomeDirFunc: func() string { return tmpDir },
	}

	prompter := &MockPrompter{}
	jdkService := &jdk.JDKService{}

	cmd := &UseCommand{
		platform:   platform,
		configRepo: repo,
		jdkService: jdkService,
		prompter:   prompter,
		force:      false,
	}

	// Use session mode with uninstalled JDK
	err = cmd.ExecuteJDK(context.Background(), "uninstalled-21", UseModeSession)
	if err == nil {
		t.Error("Expected error for uninstalled JDK in session mode")
	}

	// Error should mention install command
	if !contains(err.Error(), "install") {
		t.Errorf("Expected error to mention install command, got: %v", err)
	}
}

func TestUseCommand_ExecuteGradle_SessionMode(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Add an installed Gradle - create the actual directory structure
	gradlePath := filepath.Join(tmpDir, ".jem", "gradles", "7.6.1")
	if err := os.MkdirAll(gradlePath, 0755); err != nil {
		t.Fatalf("Failed to create Gradle dir: %v", err)
	}

	repo.AddInstalledGradle(config.GradleInfo{
		Path:    gradlePath,
		Version: "7.6.1",
		Managed: true,
	})

	platform := &MockPlatformForUse{
		HomeDirFunc: func() string { return tmpDir },
	}

	prompter := &MockPrompter{}
	jdkService := &jdk.JDKService{}

	cmd := &UseCommand{
		platform:   platform,
		configRepo: repo,
		jdkService: jdkService,
		prompter:   prompter,
		force:      false,
	}

	// Use session mode - should output env vars
	err = cmd.ExecuteGradle(context.Background(), "7.6.1", UseModeSession)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// In session mode, config should NOT be updated
	if repo.GetGradleCurrent() == "7.6.1" {
		t.Error("Session mode should not update current Gradle in config")
	}
}

func TestUseCommand_ExecuteJDK_DefaultMode_SetsDefault(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Add an installed JDK - create the actual directory structure
	jdkPath := filepath.Join(tmpDir, ".jem", "jdks", "21.0.1")
	jdkBinPath := filepath.Join(jdkPath, "bin")
	if err := os.MkdirAll(jdkBinPath, 0755); err != nil {
		t.Fatalf("Failed to create JDK dir: %v", err)
	}

	repo.AddInstalledJDK(config.JDKInfo{
		Path:     jdkPath,
		Version:  "21.0.1",
		Provider: "temurin",
		Managed:  true,
	})

	platform := &MockPlatformForUse{
		HomeDirFunc: func() string { return tmpDir },
		CreateLinkFunc: func(target, link string) error {
			return os.Symlink(target, link)
		},
		IsLinkFunc: func(path string) bool {
			_, err := os.Lstat(path)
			return err == nil
		},
	}

	prompter := &MockPrompter{}
	jdkService := &jdk.JDKService{}

	cmd := &UseCommand{
		platform:       platform,
		configRepo:     repo,
		jdkService:     jdkService,
		prompter:       prompter,
		force:          false,
		symlinkManager: symlink.NewSymlinkManager(platform),
	}

	// Use default mode - should set both current and default
	err = cmd.ExecuteJDK(context.Background(), "21.0.1", UseModeDefault)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// In default mode, both current and default should be set
	if repo.GetJDKCurrent() != "21.0.1" {
		t.Errorf("Expected current JDK to be set to 21.0.1, got %s", repo.GetJDKCurrent())
	}

	if repo.GetDefaultJDK() != "21.0.1" {
		t.Errorf("Expected default JDK to be set to 21.0.1, got %s", repo.GetDefaultJDK())
	}
}

func TestUseCommand_ExecuteJDK_OutputEnvFlag(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Add an installed JDK - create the actual directory structure
	jdkPath := filepath.Join(tmpDir, ".jem", "jdks", "21.0.1")
	jdkBinPath := filepath.Join(jdkPath, "bin")
	if err := os.MkdirAll(jdkBinPath, 0755); err != nil {
		t.Fatalf("Failed to create JDK dir: %v", err)
	}

	repo.AddInstalledJDK(config.JDKInfo{
		Path:     jdkPath,
		Version:  "21.0.1",
		Provider: "temurin",
		Managed:  true,
	})

	platform := &MockPlatformForUse{
		HomeDirFunc: func() string { return tmpDir },
	}

	prompter := &MockPrompter{}
	jdkService := &jdk.JDKService{}

	cmd := &UseCommand{
		platform:   platform,
		configRepo: repo,
		jdkService: jdkService,
		prompter:   prompter,
		force:      false,
		outputEnv:  true, // Set outputEnv flag
	}

	// With outputEnv flag, should behave like session mode
	err = cmd.ExecuteJDK(context.Background(), "21.0.1", UseModeSession)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Should not update current
	if repo.GetJDKCurrent() == "21.0.1" {
		t.Error("With outputEnv flag, should not update current JDK")
	}
}
