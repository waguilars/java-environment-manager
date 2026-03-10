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
	"github.com/waguilars/java-environment-manager/internal/platform"
)

// MockPlatformForInit creates a mock platform for init tests
type MockPlatformForInit struct {
	platform.LinuxPlatform
	HomeDirFunc     func() string
	DetectShellFunc func() config.Shell
	CreateLinkFunc  func(target, link string) error
	RemoveLinkFunc  func(link string) error
	IsLinkFunc      func(path string) bool
}

func (m *MockPlatformForInit) HomeDir() string {
	if m.HomeDirFunc != nil {
		return m.HomeDirFunc()
	}
	return "/tmp"
}

func (m *MockPlatformForInit) DetectShell() config.Shell {
	if m.DetectShellFunc != nil {
		return m.DetectShellFunc()
	}
	return config.ShellBash
}

func (m *MockPlatformForInit) CreateLink(target, link string) error {
	if m.CreateLinkFunc != nil {
		return m.CreateLinkFunc(target, link)
	}
	return os.Symlink(target, link)
}

func (m *MockPlatformForInit) RemoveLink(link string) error {
	if m.RemoveLinkFunc != nil {
		return m.RemoveLinkFunc(link)
	}
	return os.Remove(link)
}

func (m *MockPlatformForInit) IsLink(path string) bool {
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
func captureOutput(f func()) string {
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

// Test 5.1: jem init creates symlinks and outputs script

// TestInitCommand_CreatesJavaSymlink verifies init creates Java symlink when default is set
func TestInitCommand_CreatesJavaSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".jem", "config.toml")

	// Create JDK directory
	jdkPath := filepath.Join(tmpDir, ".jem", "jdks", "temurin-21")
	if err := os.MkdirAll(jdkPath, 0755); err != nil {
		t.Fatalf("Failed to create JDK dir: %v", err)
	}

	repo := config.NewTOMLConfigRepository(configPath)
	cfg, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Set default JDK
	cfg.Defaults.JDK = "temurin-21"
	if err := repo.Save(cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	platform := &MockPlatformForInit{
		HomeDirFunc: func() string { return tmpDir },
	}

	cmd := NewInitCommand(platform, repo)

	output := captureOutput(func() {
		err = cmd.Execute(context.Background(), "")
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify Java symlink was created
	javaLink := filepath.Join(tmpDir, ".jem", "current", "java")
	if _, err := os.Lstat(javaLink); os.IsNotExist(err) {
		t.Error("Expected Java symlink to be created")
	}

	// Verify output contains shell script
	if !strings.Contains(output, "# jem environment initialization") {
		t.Errorf("Expected output to contain init script header, got:\n%s", output)
	}
}

// TestInitCommand_CreatesGradleSymlink verifies init creates Gradle symlink when default is set
func TestInitCommand_CreatesGradleSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".jem", "config.toml")

	// Create Gradle directory
	gradlePath := filepath.Join(tmpDir, ".jem", "gradles", "8.5")
	if err := os.MkdirAll(gradlePath, 0755); err != nil {
		t.Fatalf("Failed to create Gradle dir: %v", err)
	}

	repo := config.NewTOMLConfigRepository(configPath)
	cfg, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Set default Gradle
	cfg.Defaults.Gradle = "8.5"
	if err := repo.Save(cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	platform := &MockPlatformForInit{
		HomeDirFunc: func() string { return tmpDir },
	}

	cmd := NewInitCommand(platform, repo)

	output := captureOutput(func() {
		err = cmd.Execute(context.Background(), "")
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify Gradle symlink was created
	gradleLink := filepath.Join(tmpDir, ".jem", "current", "gradle")
	if _, err := os.Lstat(gradleLink); os.IsNotExist(err) {
		t.Error("Expected Gradle symlink to be created")
	}

	// Verify output contains GRADLE_HOME
	if !strings.Contains(output, "GRADLE_HOME") {
		t.Errorf("Expected output to contain GRADLE_HOME, got:\n%s", output)
	}
}

// TestInitCommand_CreatesBothSymlinks verifies init creates both symlinks when both defaults are set
func TestInitCommand_CreatesBothSymlinks(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".jem", "config.toml")

	// Create directories
	jdkPath := filepath.Join(tmpDir, ".jem", "jdks", "temurin-21")
	gradlePath := filepath.Join(tmpDir, ".jem", "gradles", "8.5")
	if err := os.MkdirAll(jdkPath, 0755); err != nil {
		t.Fatalf("Failed to create JDK dir: %v", err)
	}
	if err := os.MkdirAll(gradlePath, 0755); err != nil {
		t.Fatalf("Failed to create Gradle dir: %v", err)
	}

	repo := config.NewTOMLConfigRepository(configPath)
	cfg, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	cfg.Defaults.JDK = "temurin-21"
	cfg.Defaults.Gradle = "8.5"
	if err := repo.Save(cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	platform := &MockPlatformForInit{
		HomeDirFunc: func() string { return tmpDir },
	}

	cmd := NewInitCommand(platform, repo)

	output := captureOutput(func() {
		err = cmd.Execute(context.Background(), "")
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify both symlinks were created
	javaLink := filepath.Join(tmpDir, ".jem", "current", "java")
	gradleLink := filepath.Join(tmpDir, ".jem", "current", "gradle")

	if _, err := os.Lstat(javaLink); os.IsNotExist(err) {
		t.Error("Expected Java symlink to be created")
	}
	if _, err := os.Lstat(gradleLink); os.IsNotExist(err) {
		t.Error("Expected Gradle symlink to be created")
	}

	// Verify output contains both env vars
	if !strings.Contains(output, "JAVA_HOME") {
		t.Errorf("Expected output to contain JAVA_HOME, got:\n%s", output)
	}
	if !strings.Contains(output, "GRADLE_HOME") {
		t.Errorf("Expected output to contain GRADLE_HOME, got:\n%s", output)
	}
}

// TestInitCommand_BashOutput verifies init outputs correct Bash script
func TestInitCommand_BashOutput(t *testing.T) {
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

	platform := &MockPlatformForInit{
		HomeDirFunc:     func() string { return tmpDir },
		DetectShellFunc: func() config.Shell { return config.ShellBash },
	}

	cmd := NewInitCommand(platform, repo)

	output := captureOutput(func() {
		err = cmd.Execute(context.Background(), "bash")
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify Bash-specific syntax
	if !strings.Contains(output, `export JAVA_HOME=`) {
		t.Errorf("Expected Bash export syntax for JAVA_HOME, got:\n%s", output)
	}
	if !strings.Contains(output, `export PATH="$JAVA_HOME/bin:$PATH"`) {
		t.Errorf("Expected Bash PATH update syntax, got:\n%s", output)
	}
}

// TestInitCommand_ZshOutput verifies init outputs correct Zsh script
func TestInitCommand_ZshOutput(t *testing.T) {
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

	platform := &MockPlatformForInit{
		HomeDirFunc:     func() string { return tmpDir },
		DetectShellFunc: func() config.Shell { return config.ShellZsh },
	}

	cmd := NewInitCommand(platform, repo)

	output := captureOutput(func() {
		err = cmd.Execute(context.Background(), "zsh")
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify Zsh uses same syntax as Bash
	if !strings.Contains(output, `export JAVA_HOME=`) {
		t.Errorf("Expected Zsh export syntax for JAVA_HOME, got:\n%s", output)
	}
}

// TestInitCommand_PowerShellOutput verifies init outputs correct PowerShell script
func TestInitCommand_PowerShellOutput(t *testing.T) {
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

	platform := &MockPlatformForInit{
		HomeDirFunc:     func() string { return tmpDir },
		DetectShellFunc: func() config.Shell { return config.ShellPowerShell },
	}

	cmd := NewInitCommand(platform, repo)

	output := captureOutput(func() {
		err = cmd.Execute(context.Background(), "powershell")
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify PowerShell-specific syntax
	if !strings.Contains(output, `$env:JAVA_HOME =`) {
		t.Errorf("Expected PowerShell syntax for JAVA_HOME, got:\n%s", output)
	}
	if !strings.Contains(output, `$env:PATH = "$env:JAVA_HOME\bin;$env:PATH"`) {
		t.Errorf("Expected PowerShell PATH update syntax, got:\n%s", output)
	}
}

// TestInitCommand_FishOutput verifies init outputs correct Fish script
func TestInitCommand_FishOutput(t *testing.T) {
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

	platform := &MockPlatformForInit{
		HomeDirFunc:     func() string { return tmpDir },
		DetectShellFunc: func() config.Shell { return config.ShellFish },
	}

	cmd := NewInitCommand(platform, repo)

	output := captureOutput(func() {
		err = cmd.Execute(context.Background(), "fish")
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify Fish-specific syntax
	if !strings.Contains(output, `set -x JAVA_HOME`) {
		t.Errorf("Expected Fish syntax for JAVA_HOME, got:\n%s", output)
	}
	if !strings.Contains(output, `set -x PATH "$JAVA_HOME/bin" $PATH`) {
		t.Errorf("Expected Fish PATH update syntax, got:\n%s", output)
	}
}

// TestInitCommand_AutoDetectShell verifies init auto-detects shell when not specified
func TestInitCommand_AutoDetectShell(t *testing.T) {
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

	platform := &MockPlatformForInit{
		HomeDirFunc:     func() string { return tmpDir },
		DetectShellFunc: func() config.Shell { return config.ShellZsh },
	}

	cmd := NewInitCommand(platform, repo)

	output := captureOutput(func() {
		// Pass empty string to trigger auto-detection
		err = cmd.Execute(context.Background(), "")
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify output (should use Zsh syntax since we mocked it)
	if !strings.Contains(output, "# jem environment initialization") {
		t.Errorf("Expected init script output, got:\n%s", output)
	}
}

// TestInitCommand_NoDefaultsSet verifies init works when no defaults are set
func TestInitCommand_NoDefaultsSet(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".jem", "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	platform := &MockPlatformForInit{
		HomeDirFunc: func() string { return tmpDir },
	}

	cmd := NewInitCommand(platform, repo)

	output := captureOutput(func() {
		err = cmd.Execute(context.Background(), "bash")
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify no symlinks were created
	javaLink := filepath.Join(tmpDir, ".jem", "current", "java")
	gradleLink := filepath.Join(tmpDir, ".jem", "current", "gradle")

	if _, err := os.Lstat(javaLink); !os.IsNotExist(err) {
		t.Error("Expected no Java symlink when no default set")
	}
	if _, err := os.Lstat(gradleLink); !os.IsNotExist(err) {
		t.Error("Expected no Gradle symlink when no default set")
	}

	// Verify output still contains header
	if !strings.Contains(output, "# jem environment initialization") {
		t.Errorf("Expected init script header even with no defaults, got:\n%s", output)
	}
}

// TestInitCommand_InvalidShell falls back to bash for unknown shell
func TestInitCommand_InvalidShell(t *testing.T) {
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

	platform := &MockPlatformForInit{
		HomeDirFunc: func() string { return tmpDir },
	}

	cmd := NewInitCommand(platform, repo)

	output := captureOutput(func() {
		// Pass invalid shell name - should fall back to auto-detection
		err = cmd.Execute(context.Background(), "unknownshell")
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify output still works (falls back to detected shell)
	if !strings.Contains(output, "# jem environment initialization") {
		t.Errorf("Expected init script output, got:\n%s", output)
	}
}
