package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/waguilars/java-environment-manager/internal/config"
	"github.com/waguilars/java-environment-manager/internal/platform"
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

	// Verify shell config was created and contains jem init configuration
	content, err := os.ReadFile(shellConfigPath)
	if err != nil {
		t.Errorf("Expected shell config to exist, got error: %v", err)
	}
	if !strings.Contains(string(content), `eval "$(jem init)"`) {
		t.Errorf("Expected shell config to contain 'eval \"$(jem init)\"', got:\n%s", string(content))
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

	// Verify shell config was created and contains jem init configuration
	content, err := os.ReadFile(shellConfigPath)
	if err != nil {
		t.Errorf("Expected shell config to exist, got error: %v", err)
	}
	if !strings.Contains(string(content), `eval "$(jem init)"`) {
		t.Errorf("Expected shell config to contain 'eval \"$(jem init)\"', got:\n%s", string(content))
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

	// Create shell config with jem already configured using new init pattern WITH wrapper
	originalContent := `# Existing config
jem() {
    case "$1" in
        use)
            shift
            eval "$(command jem use "$@" --output-env)"
            ;;
        *)
            command jem "$@"
            ;;
    esac
}
eval "$(jem init)"
`
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
	if !strings.Contains(string(content), `eval "$(jem init)"`) {
		t.Errorf("Expected actual config to contain 'eval \"$(jem init)\"', got:\n%s", string(content))
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

	// Verify backup was created with original content
	backupPath := shellConfigPath + ".jem.backup"
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Errorf("Expected backup to exist, got error: %v", err)
	}
	if string(backupContent) != originalContent {
		t.Error("Expected backup to contain original content")
	}

	// NEW: Verify shell config contains original content + jem config
	shellConfigContent, err := os.ReadFile(shellConfigPath)
	if err != nil {
		t.Errorf("Expected shell config to exist, got error: %v", err)
	}

	// Verify original content is preserved
	if !strings.Contains(string(shellConfigContent), originalContent) {
		t.Error("Expected shell config to contain original content")
	}

	// Verify jem config is present
	if !strings.Contains(string(shellConfigContent), `eval "$(jem init)"`) {
		t.Errorf("Expected shell config to contain 'eval \"$(jem init)\"', got:\n%s", string(shellConfigContent))
	}

	// Verify jem config appears AFTER original content
	originalIdx := strings.Index(string(shellConfigContent), "# Existing bashrc")
	jemIdx := strings.Index(string(shellConfigContent), "# jem initialization")
	if originalIdx >= jemIdx {
		t.Error("Expected jem configuration to appear after original content")
	}
}

// TestSetup_ShellConfigBackup_NoDataLoss verifies no data loss with comprehensive content
func TestSetup_ShellConfigBackup_NoDataLoss(t *testing.T) {
	tmpDir := t.TempDir()
	jemDir := filepath.Join(tmpDir, ".jem")
	os.MkdirAll(jemDir, 0755)

	configPath := filepath.Join(jemDir, "config.toml")
	shellConfigPath := filepath.Join(tmpDir, ".bashrc")

	// Create existing config
	os.WriteFile(configPath, []byte("[general]\ndefault_provider = 'temurin'\n"), 0644)

	// Create shell config with 50+ lines of diverse content
	originalLines := []string{
		"# ~/.bashrc: executed by bash(1) for non-login shells.",
		"# see /usr/share/doc/bash/examples/startup-files for examples.",
		"",
		"# If not running interactively, don't do anything",
		"case $- in",
		"    *i*) ;;",
		"      *) return;;",
		"esac",
		"",
		"# don't put duplicate lines or lines starting with space in the history.",
		"HISTCONTROL=ignoreboth",
		"",
		"# append to the history file, don't overwrite it",
		"shopt -s histappend",
		"",
		"# for setting history length see HISTSIZE and HISTFILESIZE in bash(1)",
		"HISTSIZE=1000",
		"HISTFILESIZE=2000",
		"",
		"# aliases",
		"alias ll='ls -alF'",
		"alias la='ls -A'",
		"alias l='ls -CF'",
		"alias ..='cd ..'",
		"alias ...='cd ../..'",
		"",
		"# custom functions",
		"mkcd() {",
		"    mkdir -p \"$1\"",
		"    cd \"$1\"",
		"}",
		"",
		"# custom exports",
		"export EDITOR=vim",
		"export VISUAL=vim",
		"export PATH=\"$HOME/.local/bin:$PATH\"",
		"",
		"# Git prompt",
		"parse_git_branch() {",
		"    git branch 2> /dev/null | sed -e '/^[^*]/d' -e 's/* \\(.*\\)/ (\\1)/'",
		"}",
		"export PS1=\"\\u@\\h \\[\\033[32m\\]\\w\\[\\033[33m\\]\\$(parse_git_branch)\\[\\033[00m\\] $ \"",
		"",
		"# NVM configuration",
		"export NVM_DIR=\"$HOME/.nvm\"",
		"[ -s \"$NVM_DIR/nvm.sh\" ] && \\. \"$NVM_DIR/nvm.sh\"",
		"",
		"# Pyenv configuration",
		"export PYENV_ROOT=\"$HOME/.pyenv\"",
		"export PATH=\"$PYENV_ROOT/bin:$PATH\"",
		"eval \"$(pyenv init -)\"",
	}
	originalContent := strings.Join(originalLines, "\n") + "\n"
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

	// Count original lines
	originalLineCount := len(strings.Split(strings.TrimSuffix(originalContent, "\n"), "\n"))

	// Read final shell config
	finalContent, err := os.ReadFile(shellConfigPath)
	if err != nil {
		t.Errorf("Expected shell config to exist, got error: %v", err)
	}

	// Count final lines
	finalLineCount := len(strings.Split(strings.TrimSuffix(string(finalContent), "\n"), "\n"))

	// Verify line count increased (original + jem lines)
	expectedAdditionalLines := 14 // jem adds 14 lines (empty + comment + wrapper function + init)
	if finalLineCount != originalLineCount+expectedAdditionalLines {
		t.Errorf("Expected %d lines, got %d", originalLineCount+expectedAdditionalLines, finalLineCount)
	}

	// Verify each original line is present
	for _, line := range originalLines {
		if !strings.Contains(string(finalContent), line) {
			t.Errorf("Expected shell config to contain line: %s", line)
		}
	}

	// Verify backup is exact copy
	backupPath := shellConfigPath + ".jem.backup"
	backupContent, _ := os.ReadFile(backupPath)
	if string(backupContent) != originalContent {
		t.Error("Expected backup to be exact copy of original")
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

	// Create shell config with jem already configured using new pattern WITH wrapper
	os.WriteFile(shellConfigPath, []byte(`jem() {
    case "$1" in
        use)
            shift
            eval "$(command jem use "$@" --output-env)"
            ;;
        *)
            command jem "$@"
            ;;
    esac
}
eval "$(jem init)"`+"\n"), 0644)

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

// TestIsShellConfigured_True verifies returns true when file contains wrapper AND jem init pattern
func TestIsShellConfigured_True(t *testing.T) {
	tmpDir := t.TempDir()
	shellConfigPath := filepath.Join(tmpDir, ".bashrc")
	os.WriteFile(shellConfigPath, []byte(`jem() {
    case "$1" in
        use)
            shift
            eval "$(command jem use "$@" --output-env)"
            ;;
        *)
            command jem "$@"
            ;;
    esac
}
eval "$(jem init)"
`), 0644)

	cmd := &SetupCommand{
		platform:   &MockPlatformForSetup{},
		configRepo: nil,
	}

	result := cmd.isShellConfigured(shellConfigPath)
	if !result {
		t.Error("Expected isShellConfigured to return true")
	}
}

// TestIsShellConfigured_True_InvokeExpression verifies returns true when PowerShell config contains wrapper AND jem init
func TestIsShellConfigured_True_InvokeExpression(t *testing.T) {
	tmpDir := t.TempDir()
	shellConfigPath := filepath.Join(tmpDir, "Microsoft.PowerShell_profile.ps1")
	os.WriteFile(shellConfigPath, []byte(`function jem {
    param([Parameter(ValueFromRemainingArguments)]$Args)
    if ($Args[0] -eq 'use') {
        $output = & jem $Args --output-env 2>&1
        if ($LASTEXITCODE -eq 0) {
            Invoke-Expression $output
        } else {
            Write-Host $output
        }
    } else {
        & jem @Args
    }
}
jem init | Invoke-Expression
`), 0644)

	cmd := &SetupCommand{
		platform:   &MockPlatformForSetup{},
		configRepo: nil,
	}

	result := cmd.isShellConfigured(shellConfigPath)
	if !result {
		t.Error("Expected isShellConfigured to return true for PowerShell")
	}
}

// TestIsShellConfigured_False verifies returns false when file does not contain jem init
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

// TestConfigureShell_Bash verifies correct jem init for bash
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

	contentStr := string(content)

	// Should contain wrapper function
	if !strings.Contains(contentStr, "jem() {") {
		t.Errorf("Expected shell config to contain wrapper function, got:\n%s", contentStr)
	}

	// Should contain jem init pattern
	expected := `eval "$(jem init)"`
	if !strings.Contains(contentStr, expected) {
		t.Errorf("Expected shell config to contain '%s', got:\n%s", expected, contentStr)
	}

	// Verify wrapper comes BEFORE init
	wrapperIdx := strings.Index(contentStr, "jem() {")
	initIdx := strings.Index(contentStr, expected)
	if wrapperIdx >= initIdx {
		t.Error("Expected wrapper function to appear BEFORE init line")
	}
}

// TestConfigureShell_Zsh verifies correct jem init for zsh
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

	contentStr := string(content)

	// Should contain wrapper function
	if !strings.Contains(contentStr, "jem() {") {
		t.Errorf("Expected shell config to contain wrapper function, got:\n%s", contentStr)
	}

	// Should contain jem init pattern
	expected := `eval "$(jem init)"`
	if !strings.Contains(contentStr, expected) {
		t.Errorf("Expected shell config to contain '%s', got:\n%s", expected, contentStr)
	}

	// Verify wrapper comes BEFORE init
	wrapperIdx := strings.Index(contentStr, "jem() {")
	initIdx := strings.Index(contentStr, expected)
	if wrapperIdx >= initIdx {
		t.Error("Expected wrapper function to appear BEFORE init line")
	}
}

// TestConfigureShell_PowerShell verifies correct jem init for PowerShell
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

	contentStr := string(content)

	// Should contain wrapper function
	if !strings.Contains(contentStr, "function jem {") {
		t.Errorf("Expected shell config to contain wrapper function, got:\n%s", contentStr)
	}

	// Should contain jem init | Invoke-Expression pattern
	expected := "jem init | Invoke-Expression"
	if !strings.Contains(contentStr, expected) {
		t.Errorf("Expected shell config to contain '%s', got:\n%s", expected, contentStr)
	}

	// Verify wrapper comes BEFORE init
	wrapperIdx := strings.Index(contentStr, "function jem {")
	initIdx := strings.Index(contentStr, expected)
	if wrapperIdx >= initIdx {
		t.Error("Expected wrapper function to appear BEFORE init line")
	}
}

// TestConfigureShell_Fish verifies Fish shell warning is shown
func TestConfigureShell_Fish(t *testing.T) {
	tmpDir := t.TempDir()
	shellConfigPath := filepath.Join(tmpDir, ".config", "fish", "config.fish")

	cmd := &SetupCommand{
		platform:   &MockPlatformForSetup{},
		configRepo: nil,
	}

	// Fish shell should not create a config file, just show warning
	err := cmd.configureShell(config.ShellFish, shellConfigPath)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify no shell config was created
	if _, err := os.Stat(shellConfigPath); !os.IsNotExist(err) {
		t.Error("Expected no shell config to be created for Fish")
	}
}

// Phase 4: New wrapper detection tests

// TestIsShellConfigured_WithWrapper verifies returns true when wrapper AND init present
func TestIsShellConfigured_WithWrapper(t *testing.T) {
	tmpDir := t.TempDir()
	shellConfigPath := filepath.Join(tmpDir, ".bashrc")
	// Shell config with BOTH wrapper and init
	content := `jem() {
    case "$1" in
        use)
            shift
            eval "$(command jem use "$@" --output-env)"
            ;;
        *)
            command jem "$@"
            ;;
    esac
}
eval "$(jem init)"
`
	os.WriteFile(shellConfigPath, []byte(content), 0644)

	cmd := &SetupCommand{
		platform:   &MockPlatformForSetup{},
		configRepo: nil,
	}

	result := cmd.isShellConfigured(shellConfigPath)
	if !result {
		t.Error("Expected isShellConfigured to return true when wrapper AND init present")
	}
}

// TestIsShellConfigured_Legacy verifies returns false when only init present (no wrapper)
func TestIsShellConfigured_Legacy(t *testing.T) {
	tmpDir := t.TempDir()
	shellConfigPath := filepath.Join(tmpDir, ".bashrc")
	// Legacy shell config with ONLY init, no wrapper
	os.WriteFile(shellConfigPath, []byte(`eval "$(jem init)"`+"\n"), 0644)

	cmd := &SetupCommand{
		platform:   &MockPlatformForSetup{},
		configRepo: nil,
	}

	result := cmd.isShellConfigured(shellConfigPath)
	if result {
		t.Error("Expected isShellConfigured to return false for legacy config (init only, no wrapper)")
	}
}

// TestIsShellConfigured_OnlyWrapper verifies returns false when only wrapper present
func TestIsShellConfigured_OnlyWrapper(t *testing.T) {
	tmpDir := t.TempDir()
	shellConfigPath := filepath.Join(tmpDir, ".bashrc")
	// Shell config with ONLY wrapper, no init
	content := `jem() {
    case "$1" in
        use)
            shift
            eval "$(command jem use "$@" --output-env)"
            ;;
        *)
            command jem "$@"
            ;;
    esac
}
`
	os.WriteFile(shellConfigPath, []byte(content), 0644)

	cmd := &SetupCommand{
		platform:   &MockPlatformForSetup{},
		configRepo: nil,
	}

	result := cmd.isShellConfigured(shellConfigPath)
	if result {
		t.Error("Expected isShellConfigured to return false when only wrapper present (no init)")
	}
}

// TestSetup_WrapperInstalled_Bash verifies fresh setup installs wrapper + init for Bash
func TestSetup_WrapperInstalled_Bash(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".jem", "config.toml")
	shellConfigPath := filepath.Join(tmpDir, ".bashrc")

	// Create config directory
	os.MkdirAll(filepath.Dir(configPath), 0755)

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

	content, err := os.ReadFile(shellConfigPath)
	if err != nil {
		t.Fatalf("Expected shell config to exist, got error: %v", err)
	}

	contentStr := string(content)

	// Should contain wrapper function
	if !strings.Contains(contentStr, "jem() {") {
		t.Error("Expected shell config to contain wrapper function")
	}

	// Should contain jem init
	if !strings.Contains(contentStr, `eval "$(jem init)"`) {
		t.Error("Expected shell config to contain jem init")
	}

	// Verify wrapper comes BEFORE init
	wrapperIdx := strings.Index(contentStr, "jem() {")
	initIdx := strings.Index(contentStr, `eval "$(jem init)"`)
	if wrapperIdx >= initIdx {
		t.Error("Expected wrapper function to appear BEFORE init line")
	}
}

// TestSetup_WrapperInstalled_PowerShell verifies fresh setup installs wrapper + init for PowerShell
func TestSetup_WrapperInstalled_PowerShell(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".jem", "config.toml")
	shellConfigPath := filepath.Join(tmpDir, "Microsoft.PowerShell_profile.ps1")

	// Create config directory
	os.MkdirAll(filepath.Dir(configPath), 0755)

	repo := config.NewTOMLConfigRepository(configPath)

	platform := &MockPlatformForSetup{
		HomeDirFunc:         func() string { return tmpDir },
		DetectShellFunc:     func() config.Shell { return config.ShellPowerShell },
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

	content, err := os.ReadFile(shellConfigPath)
	if err != nil {
		t.Fatalf("Expected shell config to exist, got error: %v", err)
	}

	contentStr := string(content)

	// Should contain wrapper function
	if !strings.Contains(contentStr, "function jem {") {
		t.Error("Expected shell config to contain wrapper function")
	}

	// Should contain jem init
	if !strings.Contains(contentStr, "jem init | Invoke-Expression") {
		t.Error("Expected shell config to contain jem init | Invoke-Expression")
	}

	// Verify wrapper comes BEFORE init
	wrapperIdx := strings.Index(contentStr, "function jem {")
	initIdx := strings.Index(contentStr, "jem init | Invoke-Expression")
	if wrapperIdx >= initIdx {
		t.Error("Expected wrapper function to appear BEFORE init line")
	}
}

// TestSetup_LegacyUpgrade verifies existing init-only config gets wrapper added
func TestSetup_LegacyUpgrade(t *testing.T) {
	tmpDir := t.TempDir()
	jemDir := filepath.Join(tmpDir, ".jem")
	os.MkdirAll(jemDir, 0755)

	configPath := filepath.Join(jemDir, "config.toml")
	shellConfigPath := filepath.Join(tmpDir, ".bashrc")

	// Create existing config
	os.WriteFile(configPath, []byte("[general]\ndefault_provider = 'temurin'\n"), 0644)

	// Create shell config with LEGACY setup (init only, no wrapper)
	originalContent := "# Legacy setup\neval \"$(jem init)\"\n"
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

	content, err := os.ReadFile(shellConfigPath)
	if err != nil {
		t.Fatalf("Expected shell config to exist, got error: %v", err)
	}

	contentStr := string(content)

	// Should still contain original content
	if !strings.Contains(contentStr, "# Legacy setup") {
		t.Error("Expected shell config to preserve original content")
	}

	// Should now contain wrapper function
	if !strings.Contains(contentStr, "jem() {") {
		t.Error("Expected shell config to have wrapper added")
	}

	// Should still contain jem init
	if !strings.Contains(contentStr, `eval "$(jem init)"`) {
		t.Error("Expected shell config to still contain jem init")
	}

	// Verify wrapper appears BEFORE init
	wrapperIdx := strings.Index(contentStr, "jem() {")
	initIdx := strings.Index(contentStr, `eval "$(jem init)"`)
	if wrapperIdx >= initIdx {
		t.Error("Expected wrapper function to appear BEFORE init line")
	}
}

// TestSetup_NoDuplicateWrapper verifies running setup twice doesn't duplicate wrapper
func TestSetup_NoDuplicateWrapper(t *testing.T) {
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

	// First setup
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Fatalf("First setup failed: %v", err)
	}

	content1, err := os.ReadFile(shellConfigPath)
	if err != nil {
		t.Fatalf("Expected shell config to exist after first setup, got error: %v", err)
	}

	// Second setup - should be idempotent
	err = cmd.Execute(context.Background())
	if err != nil {
		t.Errorf("Second setup failed: %v", err)
	}

	content2, err := os.ReadFile(shellConfigPath)
	if err != nil {
		t.Fatalf("Expected shell config to exist after second setup, got error: %v", err)
	}

	// Content should be identical after second setup (idempotent)
	if string(content1) != string(content2) {
		t.Errorf("Expected shell config to be unchanged after second setup. First:\n%s\n\nSecond:\n%s", content1, content2)
	}

	// Count wrapper occurrences - should be exactly 1
	wrapperCount := strings.Count(string(content2), "jem() {")
	if wrapperCount != 1 {
		t.Errorf("Expected exactly 1 wrapper function, found %d", wrapperCount)
	}
}
