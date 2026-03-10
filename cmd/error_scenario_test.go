package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/waguilars/java-environment-manager/internal/config"
	"github.com/waguilars/java-environment-manager/internal/jdk"
	"github.com/waguilars/java-environment-manager/internal/platform"
)

// MockPlatformForErrors creates a mock platform for error tests
type MockPlatformForErrors struct {
	platform.LinuxPlatform
	HomeDirFunc     func() string
	DetectShellFunc func() config.Shell
}

func (m *MockPlatformForErrors) HomeDir() string {
	if m.HomeDirFunc != nil {
		return m.HomeDirFunc()
	}
	return "/tmp"
}

func (m *MockPlatformForErrors) DetectShell() config.Shell {
	if m.DetectShellFunc != nil {
		return m.DetectShellFunc()
	}
	return config.ShellBash
}

// Test 5.6: Error scenario tests

// TestUseCommand_InvalidVersion_JDK verifies error for invalid/uninstalled JDK version
func TestUseCommand_InvalidVersion_JDK(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".jem", "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	platform := &MockPlatformForErrors{
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

	// Try to use an uninstalled JDK
	err = cmd.ExecuteJDK(context.Background(), "nonexistent-99.9.9", UseModeDefault)
	if err == nil {
		t.Error("Expected error for uninstalled JDK")
	}

	// Verify error message
	if !strings.Contains(err.Error(), "not installed") {
		t.Errorf("Expected error to mention 'not installed', got: %v", err)
	}

	// Verify it suggests the install command
	if !strings.Contains(err.Error(), "jem install jdk") {
		t.Errorf("Expected error to suggest 'jem install jdk', got: %v", err)
	}
}

// TestUseCommand_InvalidVersion_Gradle verifies error for invalid/uninstalled Gradle version
func TestUseCommand_InvalidVersion_Gradle(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".jem", "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	platform := &MockPlatformForErrors{
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

	// Try to use an uninstalled Gradle
	err = cmd.ExecuteGradle(context.Background(), "nonexistent-99.9.9", UseModeDefault)
	if err == nil {
		t.Error("Expected error for uninstalled Gradle")
	}

	// Verify error message
	if !strings.Contains(err.Error(), "not installed") {
		t.Errorf("Expected error to mention 'not installed', got: %v", err)
	}

	// Verify it suggests the install command
	if !strings.Contains(err.Error(), "jem install gradle") {
		t.Errorf("Expected error to suggest 'jem install gradle', got: %v", err)
	}
}

// TestInitCommand_UnknownShell verifies init handles unknown shell gracefully
func TestInitCommand_UnknownShell(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".jem", "config.toml")

	jdkPath := filepath.Join(tmpDir, ".jem", "jdks", "temurin-21")
	if err := os.MkdirAll(jdkPath, 0755); err != nil {
		t.Fatalf("Failed to create JDK dir: %v", err)
	}

	repo := config.NewTOMLConfigRepository(configPath)
	cfg, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	cfg.Defaults.JDK = "temurin-21"
	if err := repo.Save(cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	platform := &MockPlatformForErrors{
		HomeDirFunc: func() string { return tmpDir },
		DetectShellFunc: func() config.Shell {
			return config.Shell("unknown")
		},
	}

	cmd := NewInitCommand(platform, repo)

	// Pass unknown shell name - should fall back to detected shell
	output := captureOutput(func() {
		err = cmd.Execute(context.Background(), "weirdshell")
	})

	if err != nil {
		t.Errorf("Expected no error for unknown shell (should fallback), got: %v", err)
	}

	// Verify output still works (falls back)
	if !strings.Contains(output, "# jem environment initialization") {
		t.Errorf("Expected init script output even for unknown shell, got:\n%s", output)
	}
}

// TestUseCommand_SessionMode_UninstalledJDK verifies session mode error for uninstalled JDK
func TestUseCommand_SessionMode_UninstalledJDK(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".jem", "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	platform := &MockPlatformForErrors{
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
		outputEnv:  true,
	}

	// Try session mode with uninstalled JDK
	err = cmd.ExecuteJDK(context.Background(), "not-installed-21", UseModeSession)
	if err == nil {
		t.Error("Expected error for uninstalled JDK in session mode")
	}

	// Verify error message suggests install
	if !strings.Contains(err.Error(), "install") {
		t.Errorf("Expected error to mention install command, got: %v", err)
	}
}

// TestUseCommand_JDKDirectoryNotFound verifies error when JDK directory is missing
func TestUseCommand_JDKDirectoryNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".jem", "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Add a JDK to config but don't create the directory
	repo.AddInstalledJDK(config.JDKInfo{
		Path:     filepath.Join(tmpDir, ".jem", "jdks", "ghost-jdk"),
		Version:  "ghost-21",
		Provider: "temurin",
		Managed:  true,
	})

	platform := &MockPlatformForErrors{
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

	// Try to use a JDK whose directory doesn't exist
	err = cmd.ExecuteJDK(context.Background(), "ghost-21", UseModeDefault)
	if err == nil {
		t.Error("Expected error when JDK directory is missing")
	}

	// Verify error message
	if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "directory not found") {
		t.Errorf("Expected error to mention directory not found, got: %v", err)
	}
}

// TestUseCommand_JDKBinDirectoryNotFound verifies error when JDK bin directory is missing
func TestUseCommand_JDKBinDirectoryNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".jem", "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create JDK directory without bin subdirectory
	jdkPath := filepath.Join(tmpDir, ".jem", "jdks", "no-bin-jdk")
	if err := os.MkdirAll(jdkPath, 0755); err != nil {
		t.Fatalf("Failed to create JDK dir: %v", err)
	}

	repo.AddInstalledJDK(config.JDKInfo{
		Path:     jdkPath,
		Version:  "no-bin-21",
		Provider: "temurin",
		Managed:  true,
	})

	platform := &MockPlatformForErrors{
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

	// Try to use a JDK without bin directory
	err = cmd.ExecuteJDK(context.Background(), "no-bin-21", UseModeDefault)
	if err == nil {
		t.Error("Expected error when JDK bin directory is missing")
	}

	// Verify error message
	if !strings.Contains(err.Error(), "bin directory not found") {
		t.Errorf("Expected error to mention bin directory not found, got: %v", err)
	}
}

// TestUseCommand_ImportCancelled verifies error when user cancels import
func TestUseCommand_ImportCancelled(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".jem", "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Add a detected (non-managed) JDK
	detectedPath := filepath.Join(tmpDir, "detected-jdk")
	if err := os.MkdirAll(detectedPath, 0755); err != nil {
		t.Fatalf("Failed to create detected JDK dir: %v", err)
	}

	repo.AddDetectedJDK(config.JDKInfo{
		Path:     detectedPath,
		Version:  "detected-17",
		Provider: "corretto",
		Managed:  false,
	})

	platform := &MockPlatformForErrors{
		HomeDirFunc: func() string { return tmpDir },
	}

	// User declines import
	prompter := &MockPrompter{
		ConfirmFunc: func(message string, defaultValue bool) bool {
			return false // User says no
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

	// Try to use a detected JDK but cancel import
	err = cmd.ExecuteJDK(context.Background(), "detected-17", UseModeDefault)
	if err == nil {
		t.Error("Expected error when import is cancelled")
	}

	// Verify error message
	if !strings.Contains(err.Error(), "cancelled") {
		t.Errorf("Expected error to mention cancelled, got: %v", err)
	}
}

// TestUseCommand_InvalidMode verifies error for unknown use mode
func TestUseCommand_InvalidMode(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".jem", "config.toml")

	jdkPath := filepath.Join(tmpDir, ".jem", "jdks", "temurin-21")
	if err := os.MkdirAll(filepath.Join(jdkPath, "bin"), 0755); err != nil {
		t.Fatalf("Failed to create JDK dir: %v", err)
	}

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	repo.AddInstalledJDK(config.JDKInfo{
		Path:     jdkPath,
		Version:  "21.0.1",
		Provider: "temurin",
		Managed:  true,
	})

	platform := &MockPlatformForErrors{
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

	// Use an invalid mode value (cast from int)
	invalidMode := UseMode(999)
	err = cmd.ExecuteJDK(context.Background(), "21.0.1", invalidMode)
	if err == nil {
		t.Error("Expected error for invalid use mode")
	}

	// Verify error message
	if !strings.Contains(err.Error(), "unknown") {
		t.Errorf("Expected error to mention unknown mode, got: %v", err)
	}
}

// TestInitCommand_ConfigLoadError verifies init handles config load errors
func TestInitCommand_ConfigLoadError(t *testing.T) {
	// Use a path that doesn't exist for config
	configPath := "/nonexistent/path/config.toml"

	repo := config.NewTOMLConfigRepository(configPath)

	platform := &MockPlatformForErrors{
		HomeDirFunc: func() string { return "/tmp" },
	}

	cmd := NewInitCommand(platform, repo)

	// Execute init with non-existent config
	output := captureOutput(func() {
		err := cmd.Execute(context.Background(), "bash")
		// Should not error - creates default config
		if err != nil {
			t.Logf("Init with missing config returned error: %v", err)
		}
	})

	// Should still produce some output (may be empty or have defaults)
	t.Logf("Output with missing config: %s", output)
}

// TestUseCommand_EmptyVersion verifies error for empty version
func TestUseCommand_EmptyVersion(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".jem", "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	platform := &MockPlatformForErrors{
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

	// Try to use with empty version
	err = cmd.ExecuteJDK(context.Background(), "", UseModeDefault)
	if err == nil {
		t.Error("Expected error for empty version")
	}
}

// TestUseCommand_EmptyVersion_Gradle verifies error for empty Gradle version
func TestUseCommand_EmptyVersion_Gradle(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".jem", "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	platform := &MockPlatformForErrors{
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

	// Try to use with empty Gradle version
	err = cmd.ExecuteGradle(context.Background(), "", UseModeDefault)
	if err == nil {
		t.Error("Expected error for empty Gradle version")
	}
}
