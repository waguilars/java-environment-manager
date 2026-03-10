package cmd

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/waguilars/java-environment-manager/internal/config"
	"github.com/waguilars/java-environment-manager/internal/jdk"
	"github.com/waguilars/java-environment-manager/internal/platform"
	"github.com/waguilars/java-environment-manager/internal/symlink"
)

// MockPlatformForUseIntegration creates a mock platform for use integration tests
type MockPlatformForUseIntegration struct {
	platform.LinuxPlatform
	HomeDirFunc    func() string
	CreateLinkFunc func(target, link string) error
	RemoveLinkFunc func(link string) error
	IsLinkFunc     func(path string) bool
}

func (m *MockPlatformForUseIntegration) HomeDir() string {
	if m.HomeDirFunc != nil {
		return m.HomeDirFunc()
	}
	return "/tmp"
}

func (m *MockPlatformForUseIntegration) CreateLink(target, link string) error {
	if m.CreateLinkFunc != nil {
		return m.CreateLinkFunc(target, link)
	}
	return os.Symlink(target, link)
}

func (m *MockPlatformForUseIntegration) RemoveLink(link string) error {
	if m.RemoveLinkFunc != nil {
		return m.RemoveLinkFunc(link)
	}
	return os.Remove(link)
}

func (m *MockPlatformForUseIntegration) IsLink(path string) bool {
	if m.IsLinkFunc != nil {
		return m.IsLinkFunc(path)
	}
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

// captureOutput captures stdout during test execution
func captureOutputUse(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// Test 5.2: jem use jdk --session (session mode)

// TestUseCommand_SessionMode_OutputsEnv verifies session mode outputs env exports without changing symlinks
func TestUseCommand_SessionMode_OutputsEnv(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".jem", "config.toml")

	// Create JDK directory with bin subdirectory
	jdkPath := filepath.Join(tmpDir, ".jem", "jdks", "21.0.1")
	jdkBinPath := filepath.Join(jdkPath, "bin")
	if err := os.MkdirAll(jdkBinPath, 0755); err != nil {
		t.Fatalf("Failed to create JDK dir: %v", err)
	}

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Add installed JDK
	repo.AddInstalledJDK(config.JDKInfo{
		Path:     jdkPath,
		Version:  "21.0.1",
		Provider: "temurin",
		Managed:  true,
	})

	// Set a default to verify it doesn't change
	repo.SetDefaultJDK("17.0.0")

	platform := &MockPlatformForUseIntegration{
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

	output := captureOutputUse(func() {
		err = cmd.ExecuteJDK(context.Background(), "21.0.1", UseModeSession)
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify output contains export statements
	if !strings.Contains(output, "export JAVA_HOME=") {
		t.Errorf("Expected output to contain JAVA_HOME export, got:\n%s", output)
	}
	if !strings.Contains(output, "export PATH=") {
		t.Errorf("Expected output to contain PATH export, got:\n%s", output)
	}

	// Verify default was NOT changed
	if repo.GetDefaultJDK() != "17.0.0" {
		t.Errorf("Expected default JDK to remain unchanged, got: %s", repo.GetDefaultJDK())
	}

	// Verify no symlinks were created
	javaLink := filepath.Join(tmpDir, ".jem", "current", "java")
	if _, err := os.Lstat(javaLink); !os.IsNotExist(err) {
		t.Error("Expected no Java symlink in session mode")
	}
}

// TestUseCommand_SessionMode_DefaultSymlinkUnchanged verifies default symlink not changed in session mode
func TestUseCommand_SessionMode_DefaultSymlinkUnchanged(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".jem", "config.toml")

	// Create two JDK directories
	jdk21Path := filepath.Join(tmpDir, ".jem", "jdks", "21.0.1")
	jdk17Path := filepath.Join(tmpDir, ".jem", "jdks", "17.0.0")
	if err := os.MkdirAll(filepath.Join(jdk21Path, "bin"), 0755); err != nil {
		t.Fatalf("Failed to create JDK 21 dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(jdk17Path, "bin"), 0755); err != nil {
		t.Fatalf("Failed to create JDK 17 dir: %v", err)
	}

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	repo.AddInstalledJDK(config.JDKInfo{
		Path:     jdk21Path,
		Version:  "21.0.1",
		Provider: "temurin",
		Managed:  true,
	})
	repo.AddInstalledJDK(config.JDKInfo{
		Path:     jdk17Path,
		Version:  "17.0.0",
		Provider: "temurin",
		Managed:  true,
	})

	// Set default to 17.0.0 and create its symlink
	repo.SetDefaultJDK("17.0.0")
	currentDir := filepath.Join(tmpDir, ".jem", "current")
	os.MkdirAll(currentDir, 0755)
	javaLink := filepath.Join(currentDir, "java")
	os.Symlink(jdk17Path, javaLink)

	platform := &MockPlatformForUseIntegration{
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

	_ = captureOutputUse(func() {
		err = cmd.ExecuteJDK(context.Background(), "21.0.1", UseModeSession)
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify symlink still points to 17.0.0
	target, err := os.Readlink(javaLink)
	if err != nil {
		t.Fatalf("Failed to read symlink: %v", err)
	}
	if target != jdk17Path {
		t.Errorf("Expected symlink to still point to JDK 17, got: %s", target)
	}

	// Verify default was NOT changed
	if repo.GetDefaultJDK() != "17.0.0" {
		t.Errorf("Expected default JDK to remain 17.0.0, got: %s", repo.GetDefaultJDK())
	}
}

// Test 5.3: jem use jdk --default (default mode)

// TestUseCommand_DefaultMode_UpdatesConfigDefault verifies default mode updates config default
func TestUseCommand_DefaultMode_UpdatesConfigDefault(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".jem", "config.toml")

	jdkPath := filepath.Join(tmpDir, ".jem", "jdks", "21.0.1")
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

	platform := &MockPlatformForUseIntegration{
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

	err = cmd.ExecuteJDK(context.Background(), "21.0.1", UseModeDefault)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify default was updated
	if repo.GetDefaultJDK() != "21.0.1" {
		t.Errorf("Expected default JDK to be 21.0.1, got: %s", repo.GetDefaultJDK())
	}
}

// TestUseCommand_DefaultMode_UpdatesSymlinks verifies default mode updates symlinks
func TestUseCommand_DefaultMode_UpdatesSymlinks(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".jem", "config.toml")

	jdkPath := filepath.Join(tmpDir, ".jem", "jdks", "21.0.1")
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

	platform := &MockPlatformForUseIntegration{
		HomeDirFunc: func() string { return tmpDir },
		CreateLinkFunc: func(target, link string) error {
			return os.Symlink(target, link)
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

	// Set up symlink manager
	symlinkManager := symlink.NewSymlinkManager(platform)
	cmd.SetSymlinkManager(symlinkManager)

	err = cmd.ExecuteJDK(context.Background(), "21.0.1", UseModeDefault)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify symlink was created in current directory
	javaLink := filepath.Join(tmpDir, ".jem", "current", "java")
	if _, err := os.Lstat(javaLink); os.IsNotExist(err) {
		t.Error("Expected Java symlink to be created")
	}

	// Verify symlink points to correct JDK
	target, err := os.Readlink(javaLink)
	if err != nil {
		t.Fatalf("Failed to read symlink: %v", err)
	}
	if target != jdkPath {
		t.Errorf("Expected symlink to point to %s, got: %s", jdkPath, target)
	}
}

// Test 5.4: jem use default jdk

// TestUseCommand_UseDefaultJDK_UpdatesConfigDefault verifies use default updates config default
func TestUseCommand_UseDefaultJDK_UpdatesConfigDefault(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".jem", "config.toml")

	jdk21Path := filepath.Join(tmpDir, ".jem", "jdks", "21.0.1")
	jdk17Path := filepath.Join(tmpDir, ".jem", "jdks", "17.0.0")
	if err := os.MkdirAll(filepath.Join(jdk21Path, "bin"), 0755); err != nil {
		t.Fatalf("Failed to create JDK 21 dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(jdk17Path, "bin"), 0755); err != nil {
		t.Fatalf("Failed to create JDK 17 dir: %v", err)
	}

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	repo.AddInstalledJDK(config.JDKInfo{
		Path:     jdk21Path,
		Version:  "21.0.1",
		Provider: "temurin",
		Managed:  true,
	})
	repo.AddInstalledJDK(config.JDKInfo{
		Path:     jdk17Path,
		Version:  "17.0.0",
		Provider: "temurin",
		Managed:  true,
	})

	// Set initial default
	repo.SetDefaultJDK("17.0.0")

	platform := &MockPlatformForUseIntegration{
		HomeDirFunc: func() string { return tmpDir },
		CreateLinkFunc: func(target, link string) error {
			return os.Symlink(target, link)
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

	// Use default mode to set JDK 21 as the new default
	err = cmd.ExecuteJDK(context.Background(), "21.0.1", UseModeDefault)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify default was updated
	if repo.GetDefaultJDK() != "21.0.1" {
		t.Errorf("Expected default JDK to be 21.0.1, got: %s", repo.GetDefaultJDK())
	}
}

// TestUseCommand_UseDefaultJDK_UpdatesSymlinks verifies use default updates symlinks
func TestUseCommand_UseDefaultJDK_UpdatesSymlinks(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".jem", "config.toml")

	jdk21Path := filepath.Join(tmpDir, ".jem", "jdks", "21.0.1")
	jdk17Path := filepath.Join(tmpDir, ".jem", "jdks", "17.0.0")
	if err := os.MkdirAll(filepath.Join(jdk21Path, "bin"), 0755); err != nil {
		t.Fatalf("Failed to create JDK 21 dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(jdk17Path, "bin"), 0755); err != nil {
		t.Fatalf("Failed to create JDK 17 dir: %v", err)
	}

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	repo.AddInstalledJDK(config.JDKInfo{
		Path:     jdk21Path,
		Version:  "21.0.1",
		Provider: "temurin",
		Managed:  true,
	})
	repo.AddInstalledJDK(config.JDKInfo{
		Path:     jdk17Path,
		Version:  "17.0.0",
		Provider: "temurin",
		Managed:  true,
	})

	// Set initial default and create symlink
	repo.SetDefaultJDK("17.0.0")
	currentDir := filepath.Join(tmpDir, ".jem", "current")
	os.MkdirAll(currentDir, 0755)
	javaLink := filepath.Join(currentDir, "java")
	os.Symlink(jdk17Path, javaLink)

	platform := &MockPlatformForUseIntegration{
		HomeDirFunc: func() string { return tmpDir },
		CreateLinkFunc: func(target, link string) error {
			// Remove existing link first
			os.Remove(link)
			return os.Symlink(target, link)
		},
		RemoveLinkFunc: func(link string) error {
			return os.Remove(link)
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

	// Set up symlink manager
	symlinkManager := symlink.NewSymlinkManager(platform)
	cmd.SetSymlinkManager(symlinkManager)

	// Use default mode to set JDK 21 as the new default
	err = cmd.ExecuteJDK(context.Background(), "21.0.1", UseModeDefault)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify symlink was updated
	target, err := os.Readlink(javaLink)
	if err != nil {
		t.Fatalf("Failed to read symlink: %v", err)
	}
	if target != jdk21Path {
		t.Errorf("Expected symlink to point to JDK 21, got: %s", target)
	}
}

// TestUseCommand_SessionMode_Gradle verifies session mode works for Gradle
func TestUseCommand_SessionMode_Gradle(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".jem", "config.toml")

	gradlePath := filepath.Join(tmpDir, ".jem", "gradles", "8.5")
	if err := os.MkdirAll(gradlePath, 0755); err != nil {
		t.Fatalf("Failed to create Gradle dir: %v", err)
	}

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	repo.AddInstalledGradle(config.GradleInfo{
		Path:    gradlePath,
		Version: "8.5",
		Managed: true,
	})

	// Set a default Gradle to verify it doesn't change
	repo.SetDefaultGradle("7.6")

	platform := &MockPlatformForUseIntegration{
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

	output := captureOutputUse(func() {
		err = cmd.ExecuteGradle(context.Background(), "8.5", UseModeSession)
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify output contains GRADLE_HOME export
	if !strings.Contains(output, "export GRADLE_HOME=") {
		t.Errorf("Expected output to contain GRADLE_HOME export, got:\n%s", output)
	}

	// Verify default was NOT changed
	if repo.GetDefaultGradle() != "7.6" {
		t.Errorf("Expected default Gradle to remain unchanged, got: %s", repo.GetDefaultGradle())
	}
}

// TestUseCommand_DefaultMode_Gradle verifies default mode works for Gradle
func TestUseCommand_DefaultMode_Gradle(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".jem", "config.toml")

	gradlePath := filepath.Join(tmpDir, ".jem", "gradles", "8.5")
	if err := os.MkdirAll(gradlePath, 0755); err != nil {
		t.Fatalf("Failed to create Gradle dir: %v", err)
	}

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	repo.AddInstalledGradle(config.GradleInfo{
		Path:    gradlePath,
		Version: "8.5",
		Managed: true,
	})

	platform := &MockPlatformForUseIntegration{
		HomeDirFunc: func() string { return tmpDir },
		CreateLinkFunc: func(target, link string) error {
			return os.Symlink(target, link)
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

	// Set up symlink manager
	symlinkManager := symlink.NewSymlinkManager(platform)
	cmd.SetSymlinkManager(symlinkManager)

	err = cmd.ExecuteGradle(context.Background(), "8.5", UseModeDefault)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify default was updated
	if repo.GetDefaultGradle() != "8.5" {
		t.Errorf("Expected default Gradle to be 8.5, got: %s", repo.GetDefaultGradle())
	}

	// Verify symlink was created in current directory
	gradleLink := filepath.Join(tmpDir, ".jem", "current", "gradle")
	if _, err := os.Lstat(gradleLink); os.IsNotExist(err) {
		t.Error("Expected Gradle symlink to be created")
	}
}
