package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/user/jem/internal/config"
	"github.com/user/jem/internal/platform"
)

// MockPlatformForSetup creates a mock platform for setup tests
type MockPlatformForSetup struct {
	platform.LinuxPlatform
	HomeDirFunc         func() string
	DetectShellFunc     func() config.Shell
	ShellConfigPathFunc func(shell config.Shell) string
}

func (m *MockPlatformForSetup) HomeDir() string {
	if m.HomeDirFunc != nil {
		return m.HomeDirFunc()
	}
	return "/tmp"
}

func (m *MockPlatformForSetup) DetectShell() config.Shell {
	if m.DetectShellFunc != nil {
		return m.DetectShellFunc()
	}
	return config.ShellBash
}

func (m *MockPlatformForSetup) ShellConfigPath(shell config.Shell) string {
	if m.ShellConfigPathFunc != nil {
		return m.ShellConfigPathFunc(shell)
	}
	return filepath.Join(m.HomeDir(), ".bashrc")
}

// TestSetup_ConfigNotExists verifies fresh install creates config AND configures shell
func TestSetup_ConfigNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".jem", "config.toml")
	shellConfigPath := filepath.Join(tmpDir, ".bashrc")

	repo := config.NewTOMLConfigRepository(configPath)

	platform := &MockPlatformForSetup{
		HomeDirFunc:         func() string { return tmpDir },
		DetectShellFunc:     func() config.Shell { return config.ShellBash },
		ShellConfigPathFunc: func(shell config.Shell) string { return shellConfigPath },
	}

	cmd := &SetupCommand{
		platform:   platform,
		configRepo: repo,
	}

	err := cmd.Execute(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify config was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Expected config.toml to be created")
	}

	// Verify shell config was created and contains jem configuration
	content, err := os.ReadFile(shellConfigPath)
	if err != nil {
		t.Errorf("Expected shell config to exist, got error: %v", err)
	}
	if !strings.Contains(string(content), ".jem/bin") {
		t.Error("Expected shell config to contain '.jem/bin'")
	}
}

// TestSetup_ConfigExists_ShellNotConfigured verifies shell is configured when config.toml exists but shell is not
func TestSetup_ConfigExists_ShellNotConfigured(t *testing.T) {
	tmpDir := t.TempDir()
	jemDir := filepath.Join(tmpDir, ".jem")
	os.MkdirAll(jemDir, 0755)

	configPath := filepath.Join(jemDir, "config.toml")
	shellConfigPath := filepath.Join(tmpDir, ".bashrc")

	// Create existing config
	os.WriteFile(configPath, []byte("[general]\ndefault_provider = 'temurin'\n"), 0644)

	repo := config.NewTOMLConfigRepository(configPath)

	platform := &MockPlatformForSetup{
		HomeDirFunc:         func() string { return tmpDir },
		DetectShellFunc:     func() config.Shell { return config.ShellBash },
		ShellConfigPathFunc: func(shell config.Shell) string { return shellConfigPath },
	}

	cmd := &SetupCommand{
		platform:   platform,
		configRepo: repo,
	}

	err := cmd.Execute(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify shell config was created and contains jem configuration
	content, err := os.ReadFile(shellConfigPath)
	if err != nil {
		t.Errorf("Expected shell config to exist, got error: %v", err)
	}
	if !strings.Contains(string(content), ".jem/bin") {
		t.Error("Expected shell config to contain '.jem/bin'")
	}
}

// TestSetup_ConfigExists_ShellConfigured verifies no modifications when both config and shell exist
func TestSetup_ConfigExists_ShellConfigured(t *testing.T) {
	tmpDir := t.TempDir()
	jemDir := filepath.Join(tmpDir, ".jem")
	os.MkdirAll(jemDir, 0755)

	configPath := filepath.Join(jemDir, "config.toml")
	shellConfigPath := filepath.Join(tmpDir, ".bashrc")

	// Create existing config
	os.WriteFile(configPath, []byte("[general]\ndefault_provider = 'temurin'\n"), 0644)

	// Create shell config with jem already configured
	originalContent := "# Existing config\nexport PATH=\"$HOME/.jem/bin:$PATH\"\n"
	os.WriteFile(shellConfigPath, []byte(originalContent), 0644)

	repo := config.NewTOMLConfigRepository(configPath)

	platform := &MockPlatformForSetup{
		HomeDirFunc:         func() string { return tmpDir },
		DetectShellFunc:     func() config.Shell { return config.ShellBash },
		ShellConfigPathFunc: func(shell config.Shell) string { return shellConfigPath },
	}

	cmd := &SetupCommand{
		platform:   platform,
		configRepo: repo,
	}

	err := cmd.Execute(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify shell config was NOT modified
	content, err := os.ReadFile(shellConfigPath)
	if err != nil {
		t.Errorf("Expected shell config to exist, got error: %v", err)
	}
	if string(content) != originalContent {
		t.Error("Expected shell config to NOT be modified")
	}

	// Verify no backup was created
	backupPath := shellConfigPath + ".jem.backup"
	if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
		t.Error("Expected no backup file to be created")
	}
}

// TestSetup_ShellConfigSymlink verifies symlinked shell configs are resolved and written to target
func TestSetup_ShellConfigSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	jemDir := filepath.Join(tmpDir, ".jem")
	os.MkdirAll(jemDir, 0755)

	configPath := filepath.Join(jemDir, "config.toml")
	shellConfigPath := filepath.Join(tmpDir, ".bashrc")
	actualConfigPath := filepath.Join(tmpDir, "actual_bashrc")

	// Create the actual file
	os.WriteFile(actualConfigPath, []byte("# Actual bashrc\n"), 0644)

	// Create symlink
	os.Symlink(actualConfigPath, shellConfigPath)

	// Create existing config
	os.WriteFile(configPath, []byte("[general]\ndefault_provider = 'temurin'\n"), 0644)

	repo := config.NewTOMLConfigRepository(configPath)

	platform := &MockPlatformForSetup{
		HomeDirFunc:         func() string { return tmpDir },
		DetectShellFunc:     func() config.Shell { return config.ShellBash },
		ShellConfigPathFunc: func(shell config.Shell) string { return shellConfigPath },
	}

	cmd := &SetupCommand{
		platform:   platform,
		configRepo: repo,
	}

	err := cmd.Execute(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify symlink still exists
	if _, err := os.Lstat(shellConfigPath); err != nil {
		t.Errorf("Expected symlink to still exist: %v", err)
	}

	// Verify the actual target file was modified
	content, err := os.ReadFile(actualConfigPath)
	if err != nil {
		t.Errorf("Expected actual config to exist, got error: %v", err)
	}
	if !strings.Contains(string(content), ".jem/bin") {
		t.Error("Expected actual config to contain '.jem/bin'")
	}
}

// TestSetup_ShellConfigBackup verifies backup is created before modification
func TestSetup_ShellConfigBackup(t *testing.T) {
	tmpDir := t.TempDir()
	jemDir := filepath.Join(tmpDir, ".jem")
	os.MkdirAll(jemDir, 0755)

	configPath := filepath.Join(jemDir, "config.toml")
	shellConfigPath := filepath.Join(tmpDir, ".bashrc")

	// Create existing config
	os.WriteFile(configPath, []byte("[general]\ndefault_provider = 'temurin'\n"), 0644)

	// Create existing shell config (without jem)
	originalContent := "# Existing bashrc\nexport PATH=/usr/local/bin:$PATH\n"
	os.WriteFile(shellConfigPath, []byte(originalContent), 0644)

	repo := config.NewTOMLConfigRepository(configPath)

	platform := &MockPlatformForSetup{
		HomeDirFunc:         func() string { return tmpDir },
		DetectShellFunc:     func() config.Shell { return config.ShellBash },
		ShellConfigPathFunc: func(shell config.Shell) string { return shellConfigPath },
	}

	cmd := &SetupCommand{
		platform:   platform,
		configRepo: repo,
	}

	err := cmd.Execute(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify backup was created
	backupPath := shellConfigPath + ".jem.backup"
	content, err := os.ReadFile(backupPath)
	if err != nil {
		t.Errorf("Expected backup to exist, got error: %v", err)
	}
	if string(content) != originalContent {
		t.Error("Expected backup to contain original content")
	}
}

// TestSetup_ShellConfigNoBackup verifies no backup when shell already configured
func TestSetup_ShellConfigNoBackup(t *testing.T) {
	tmpDir := t.TempDir()
	jemDir := filepath.Join(tmpDir, ".jem")
	os.MkdirAll(jemDir, 0755)

	configPath := filepath.Join(jemDir, "config.toml")
	shellConfigPath := filepath.Join(tmpDir, ".bashrc")

	// Create existing config
	os.WriteFile(configPath, []byte("[general]\ndefault_provider = 'temurin'\n"), 0644)

	// Create shell config with jem already configured
	os.WriteFile(shellConfigPath, []byte("export PATH=\"$HOME/.jem/bin:$PATH\"\n"), 0644)

	repo := config.NewTOMLConfigRepository(configPath)

	platform := &MockPlatformForSetup{
		HomeDirFunc:         func() string { return tmpDir },
		DetectShellFunc:     func() config.Shell { return config.ShellBash },
		ShellConfigPathFunc: func(shell config.Shell) string { return shellConfigPath },
	}

	cmd := &SetupCommand{
		platform:   platform,
		configRepo: repo,
	}

	err := cmd.Execute(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify no backup was created
	backupPath := shellConfigPath + ".jem.backup"
	if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
		t.Error("Expected no backup file to be created")
	}
}

// TestSetup_ShellConfigWriteError verifies error handling for permission denied
func TestSetup_ShellConfigWriteError(t *testing.T) {
	tmpDir := t.TempDir()
	jemDir := filepath.Join(tmpDir, ".jem")
	os.MkdirAll(jemDir, 0755)

	configPath := filepath.Join(jemDir, "config.toml")
	shellConfigDir := filepath.Join(tmpDir, "readonly")
	shellConfigPath := filepath.Join(shellConfigDir, ".bashrc")

	// Create existing config
	os.WriteFile(configPath, []byte("[general]\ndefault_provider = 'temurin'\n"), 0644)

	// Create the shell config directory and file
	os.MkdirAll(shellConfigDir, 0755)
	os.WriteFile(shellConfigPath, []byte("# Existing bashrc\n"), 0644)

	// Make the directory read-only (this will cause OpenFile to fail)
	os.Chmod(shellConfigDir, 0555)
	defer os.Chmod(shellConfigDir, 0755) // Restore permissions for cleanup

	repo := config.NewTOMLConfigRepository(configPath)

	platform := &MockPlatformForSetup{
		HomeDirFunc:         func() string { return tmpDir },
		DetectShellFunc:     func() config.Shell { return config.ShellBash },
		ShellConfigPathFunc: func(shell config.Shell) string { return shellConfigPath },
	}

	cmd := &SetupCommand{
		platform:   platform,
		configRepo: repo,
	}

	err := cmd.Execute(context.Background())
	if err == nil {
		t.Error("Expected error for write permission denied")
	}
	if err != nil && !strings.Contains(err.Error(), "failed to configure shell") {
		t.Errorf("Expected error to contain 'failed to configure shell', got: %v", err)
	}
}

// TestSetup_ShellConfigDirectoryError verifies error handling for directory creation failure
func TestSetup_ShellConfigDirectoryError(t *testing.T) {
	tmpDir := t.TempDir()
	jemDir := filepath.Join(tmpDir, ".jem")
	os.MkdirAll(jemDir, 0755)

	configPath := filepath.Join(jemDir, "config.toml")

	// Create existing config
	os.WriteFile(configPath, []byte("[general]\ndefault_provider = 'temurin'\n"), 0644)

	// Use a path in a non-existent directory that can't be created
	shellConfigPath := filepath.Join("/nonexistent", "readonly", ".bashrc")

	repo := config.NewTOMLConfigRepository(configPath)

	platform := &MockPlatformForSetup{
		HomeDirFunc:         func() string { return tmpDir },
		DetectShellFunc:     func() config.Shell { return config.ShellBash },
		ShellConfigPathFunc: func(shell config.Shell) string { return shellConfigPath },
	}

	cmd := &SetupCommand{
		platform:   platform,
		configRepo: repo,
	}

	err := cmd.Execute(context.Background())
	if err == nil {
		t.Error("Expected error for directory creation failure")
	}
	if err != nil && !strings.Contains(err.Error(), "failed to create directory") {
		t.Errorf("Expected error to contain 'failed to create directory', got: %v", err)
	}
}

// TestIsShellConfigured_True verifies returns true when file contains .jem/bin
func TestIsShellConfigured_True(t *testing.T) {
	tmpDir := t.TempDir()
	shellConfigPath := filepath.Join(tmpDir, ".bashrc")
	os.WriteFile(shellConfigPath, []byte("export PATH=\"$HOME/.jem/bin:$PATH\"\n"), 0644)

	cmd := &SetupCommand{
		platform:   &MockPlatformForSetup{},
		configRepo: nil,
	}

	result := cmd.isShellConfigured(shellConfigPath)
	if !result {
		t.Error("Expected isShellConfigured to return true")
	}
}

// TestIsShellConfigured_False verifies returns false when file does not contain .jem/bin
func TestIsShellConfigured_False(t *testing.T) {
	tmpDir := t.TempDir()
	shellConfigPath := filepath.Join(tmpDir, ".bashrc")
	os.WriteFile(shellConfigPath, []byte("export PATH=/usr/local/bin:$PATH\n"), 0644)

	cmd := &SetupCommand{
		platform:   &MockPlatformForSetup{},
		configRepo: nil,
	}

	result := cmd.isShellConfigured(shellConfigPath)
	if result {
		t.Error("Expected isShellConfigured to return false")
	}
}

// TestIsShellConfigured_FileNotExists verifies returns false when shell config file doesn't exist
func TestIsShellConfigured_FileNotExists(t *testing.T) {
	nonExistentPath := "/nonexistent/path/.bashrc"

	cmd := &SetupCommand{
		platform:   &MockPlatformForSetup{},
		configRepo: nil,
	}

	result := cmd.isShellConfigured(nonExistentPath)
	if result {
		t.Error("Expected isShellConfigured to return false for non-existent file")
	}
}

// TestConfigureShell_Bash verifies correct PATH export for bash
func TestConfigureShell_Bash(t *testing.T) {
	tmpDir := t.TempDir()
	shellConfigPath := filepath.Join(tmpDir, ".bashrc")

	cmd := &SetupCommand{
		platform:   &MockPlatformForSetup{},
		configRepo: nil,
	}

	err := cmd.configureShell(config.ShellBash, shellConfigPath)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	content, err := os.ReadFile(shellConfigPath)
	if err != nil {
		t.Errorf("Expected shell config to exist, got error: %v", err)
	}

	expected := `export PATH="$HOME/.jem/bin:$PATH"`
	if !strings.Contains(string(content), expected) {
		t.Errorf("Expected shell config to contain '%s', got:\n%s", expected, string(content))
	}
}

// TestConfigureShell_Zsh verifies correct PATH export for zsh
func TestConfigureShell_Zsh(t *testing.T) {
	tmpDir := t.TempDir()
	shellConfigPath := filepath.Join(tmpDir, ".zshrc")

	cmd := &SetupCommand{
		platform:   &MockPlatformForSetup{},
		configRepo: nil,
	}

	err := cmd.configureShell(config.ShellZsh, shellConfigPath)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	content, err := os.ReadFile(shellConfigPath)
	if err != nil {
		t.Errorf("Expected shell config to exist, got error: %v", err)
	}

	expected := `export PATH="$HOME/.jem/bin:$PATH"`
	if !strings.Contains(string(content), expected) {
		t.Errorf("Expected shell config to contain '%s', got:\n%s", expected, string(content))
	}
}

// TestConfigureShell_PowerShell verifies correct PATH export for PowerShell
func TestConfigureShell_PowerShell(t *testing.T) {
	tmpDir := t.TempDir()
	shellConfigPath := filepath.Join(tmpDir, "Microsoft.PowerShell_profile.ps1")

	cmd := &SetupCommand{
		platform:   &MockPlatformForSetup{},
		configRepo: nil,
	}

	err := cmd.configureShell(config.ShellPowerShell, shellConfigPath)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	content, err := os.ReadFile(shellConfigPath)
	if err != nil {
		t.Errorf("Expected shell config to exist, got error: %v", err)
	}

	expected := `$env:PATH = "$HOME\.jem\bin;$env:PATH"`
	if !strings.Contains(string(content), expected) {
		t.Errorf("Expected shell config to contain '%s', got:\n%s", expected, string(content))
	}
}
