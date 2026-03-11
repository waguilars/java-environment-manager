package shell

import (
	"os"
	"strings"
	"testing"

	"github.com/waguilars/java-environment-manager/internal/config"
)

func TestDetectShell_Bash(t *testing.T) {
	// Save original SHELL
	origShell := os.Getenv("SHELL")
	defer os.Setenv("SHELL", origShell)

	// Set SHELL to bash
	os.Setenv("SHELL", "/bin/bash")

	shell := DetectShell()
	if shell != config.ShellBash {
		t.Errorf("Expected ShellBash, got %s", shell)
	}
}

func TestDetectShell_Zsh(t *testing.T) {
	// Save original SHELL
	origShell := os.Getenv("SHELL")
	defer os.Setenv("SHELL", origShell)

	// Set SHELL to zsh
	os.Setenv("SHELL", "/bin/zsh")

	shell := DetectShell()
	if shell != config.ShellZsh {
		t.Errorf("Expected ShellZsh, got %s", shell)
	}
}

func TestDetectShell_PowerShell(t *testing.T) {
	// Save original env vars
	origShell := os.Getenv("SHELL")
	origPSModulePath := os.Getenv("PSModulePath")
	defer func() {
		os.Setenv("SHELL", origShell)
		os.Setenv("PSModulePath", origPSModulePath)
	}()

	// Clear SHELL and set PSModulePath
	os.Unsetenv("SHELL")
	os.Setenv("PSModulePath", "C:\\Windows\\System32\\WindowsPowerShell\\v1.0")

	shell := DetectShell()
	if shell != config.ShellPowerShell {
		t.Errorf("Expected ShellPowerShell, got %s", shell)
	}
}

func TestDetectShell_BashVersion(t *testing.T) {
	// Save original env vars
	origShell := os.Getenv("SHELL")
	origBashVersion := os.Getenv("BASH_VERSION")
	defer func() {
		os.Setenv("SHELL", origShell)
		if origBashVersion == "" {
			os.Unsetenv("BASH_VERSION")
		} else {
			os.Setenv("BASH_VERSION", origBashVersion)
		}
	}()

	// Clear SHELL and set BASH_VERSION
	os.Unsetenv("SHELL")
	os.Setenv("BASH_VERSION", "5.1.16")

	shell := DetectShell()
	if shell != config.ShellBash {
		t.Errorf("Expected ShellBash, got %s", shell)
	}
}

func TestDetectShell_Default(t *testing.T) {
	// Save original env vars
	origShell := os.Getenv("SHELL")
	origBashVersion := os.Getenv("BASH_VERSION")
	origZshVersion := os.Getenv("ZSH_VERSION")
	origFishVersion := os.Getenv("FISH_VERSION")
	origPSModulePath := os.Getenv("PSModulePath")
	defer func() {
		os.Setenv("SHELL", origShell)
		if origBashVersion == "" {
			os.Unsetenv("BASH_VERSION")
		} else {
			os.Setenv("BASH_VERSION", origBashVersion)
		}
		if origZshVersion == "" {
			os.Unsetenv("ZSH_VERSION")
		} else {
			os.Setenv("ZSH_VERSION", origZshVersion)
		}
		if origFishVersion == "" {
			os.Unsetenv("FISH_VERSION")
		} else {
			os.Setenv("FISH_VERSION", origFishVersion)
		}
		os.Setenv("PSModulePath", origPSModulePath)
	}()

	// Clear all shell-related env vars
	os.Unsetenv("SHELL")
	os.Unsetenv("BASH_VERSION")
	os.Unsetenv("ZSH_VERSION")
	os.Unsetenv("FISH_VERSION")
	os.Unsetenv("PSModulePath")

	shell := DetectShell()
	if shell != config.ShellBash {
		t.Errorf("Expected ShellBash as default, got %s", shell)
	}
}

func TestGetGenerator_Bash(t *testing.T) {
	generator := GetGenerator(config.ShellBash)
	if generator == nil {
		t.Error("Expected non-nil generator")
	}
	if generator.Name() != "bash" {
		t.Errorf("Expected name 'bash', got %s", generator.Name())
	}
}

func TestGetGenerator_Zsh(t *testing.T) {
	generator := GetGenerator(config.ShellZsh)
	if generator == nil {
		t.Error("Expected non-nil generator")
	}
	if generator.Name() != "zsh" {
		t.Errorf("Expected name 'zsh', got %s", generator.Name())
	}
}

func TestGetGenerator_PowerShell(t *testing.T) {
	generator := GetGenerator(config.ShellPowerShell)
	if generator == nil {
		t.Error("Expected non-nil generator")
	}
	if generator.Name() != "powershell" {
		t.Errorf("Expected name 'powershell', got %s", generator.Name())
	}
}

func TestGetGenerator_Fish(t *testing.T) {
	generator := GetGenerator(config.ShellFish)
	if generator == nil {
		t.Error("Expected non-nil generator")
	}
	if generator.Name() != "fish" {
		t.Errorf("Expected name 'fish', got %s", generator.Name())
	}
}

func TestGetGenerator_Default(t *testing.T) {
	generator := GetGenerator(config.Shell("unknown"))
	if generator == nil {
		t.Error("Expected non-nil generator")
	}
	if generator.Name() != "bash" {
		t.Errorf("Expected default generator 'bash', got %s", generator.Name())
	}
}

// Bash Generator Tests
func TestBashGenerator_Name(t *testing.T) {
	gen := NewBashGenerator()
	if gen.Name() != "bash" {
		t.Errorf("Expected name 'bash', got %s", gen.Name())
	}
}

func TestBashGenerator_GenerateInitScript(t *testing.T) {
	gen := NewBashGenerator()
	envVars := map[string]string{
		"JAVA_HOME":   "/home/user/.jem/current/java",
		"GRADLE_HOME": "/home/user/.jem/current/gradle",
	}

	script := gen.GenerateInitScript(envVars)

	// Check for header comment
	if !strings.Contains(script, "# jem environment initialization") {
		t.Error("Expected script to contain header comment")
	}

	// Check for JAVA_HOME export
	if !strings.Contains(script, `export JAVA_HOME="/home/user/.jem/current/java"`) {
		t.Error("Expected script to contain JAVA_HOME export")
	}

	// Check for GRADLE_HOME export
	if !strings.Contains(script, `export GRADLE_HOME="/home/user/.jem/current/gradle"`) {
		t.Error("Expected script to contain GRADLE_HOME export")
	}

	// Check for PATH update
	if !strings.Contains(script, `export PATH="$JAVA_HOME/bin:$PATH"`) {
		t.Error("Expected script to contain PATH update")
	}
}

func TestBashGenerator_GenerateInitScript_NoJava(t *testing.T) {
	gen := NewBashGenerator()
	envVars := map[string]string{
		"GRADLE_HOME": "/home/user/.jem/current/gradle",
	}

	script := gen.GenerateInitScript(envVars)

	// Check that PATH is NOT updated without JAVA_HOME
	if strings.Contains(script, `export PATH=`) {
		t.Error("Expected script NOT to contain PATH update without JAVA_HOME")
	}
}

// Zsh Generator Tests
func TestZshGenerator_Name(t *testing.T) {
	gen := NewZshGenerator()
	if gen.Name() != "zsh" {
		t.Errorf("Expected name 'zsh', got %s", gen.Name())
	}
}

func TestZshGenerator_GenerateInitScript(t *testing.T) {
	gen := NewZshGenerator()
	envVars := map[string]string{
		"JAVA_HOME":   "/home/user/.jem/current/java",
		"GRADLE_HOME": "/home/user/.jem/current/gradle",
	}

	script := gen.GenerateInitScript(envVars)

	// Check for header comment
	if !strings.Contains(script, "# jem environment initialization") {
		t.Error("Expected script to contain header comment")
	}

	// Check for JAVA_HOME export
	if !strings.Contains(script, `export JAVA_HOME="/home/user/.jem/current/java"`) {
		t.Error("Expected script to contain JAVA_HOME export")
	}

	// Check for GRADLE_HOME export
	if !strings.Contains(script, `export GRADLE_HOME="/home/user/.jem/current/gradle"`) {
		t.Error("Expected script to contain GRADLE_HOME export")
	}

	// Check for PATH update
	if !strings.Contains(script, `export PATH="$JAVA_HOME/bin:$PATH"`) {
		t.Error("Expected script to contain PATH update")
	}
}

// PowerShell Generator Tests
func TestPowerShellGenerator_Name(t *testing.T) {
	gen := NewPowerShellGenerator()
	if gen.Name() != "powershell" {
		t.Errorf("Expected name 'powershell', got %s", gen.Name())
	}
}

func TestPowerShellGenerator_GenerateInitScript(t *testing.T) {
	gen := NewPowerShellGenerator()
	envVars := map[string]string{
		"JAVA_HOME":   "C:\\Users\\user\\.jem\\current\\java",
		"GRADLE_HOME": "C:\\Users\\user\\.jem\\current\\gradle",
	}

	script := gen.GenerateInitScript(envVars)

	// Check for header comment
	if !strings.Contains(script, "# jem environment initialization") {
		t.Error("Expected script to contain header comment")
	}

	// Check for JAVA_HOME assignment
	if !strings.Contains(script, `$env:JAVA_HOME = "C:\Users\user\.jem\current\java"`) {
		t.Error("Expected script to contain JAVA_HOME assignment")
	}

	// Check for GRADLE_HOME assignment
	if !strings.Contains(script, `$env:GRADLE_HOME = "C:\Users\user\.jem\current\gradle"`) {
		t.Error("Expected script to contain GRADLE_HOME assignment")
	}

	// Check for PATH update
	if !strings.Contains(script, `$env:PATH = "$env:JAVA_HOME\bin;$env:PATH"`) {
		t.Error("Expected script to contain PATH update")
	}
}

func TestPowerShellGenerator_GenerateInitScript_NoJava(t *testing.T) {
	gen := NewPowerShellGenerator()
	envVars := map[string]string{
		"GRADLE_HOME": "C:\\Users\\user\\.jem\\current\\gradle",
	}

	script := gen.GenerateInitScript(envVars)

	// Check that PATH is NOT updated without JAVA_HOME
	if strings.Contains(script, `$env:PATH =`) {
		t.Error("Expected script NOT to contain PATH update without JAVA_HOME")
	}
}

// Fish Generator Tests
func TestFishGenerator_Name(t *testing.T) {
	gen := NewFishGenerator()
	if gen.Name() != "fish" {
		t.Errorf("Expected name 'fish', got %s", gen.Name())
	}
}

func TestFishGenerator_GenerateInitScript(t *testing.T) {
	gen := NewFishGenerator()
	envVars := map[string]string{
		"JAVA_HOME":   "/home/user/.jem/current/java",
		"GRADLE_HOME": "/home/user/.jem/current/gradle",
	}

	script := gen.GenerateInitScript(envVars)

	// Check for header comment
	if !strings.Contains(script, "# jem environment initialization") {
		t.Error("Expected script to contain header comment")
	}

	// Check for JAVA_HOME export (Fish uses set -x)
	if !strings.Contains(script, `set -x JAVA_HOME "/home/user/.jem/current/java"`) {
		t.Error("Expected script to contain JAVA_HOME set -x")
	}

	// Check for GRADLE_HOME export (Fish uses set -x)
	if !strings.Contains(script, `set -x GRADLE_HOME "/home/user/.jem/current/gradle"`) {
		t.Error("Expected script to contain GRADLE_HOME set -x")
	}

	// Check for PATH update (Fish has different syntax)
	if !strings.Contains(script, `set -x PATH "$JAVA_HOME/bin" $PATH`) {
		t.Error("Expected script to contain PATH update")
	}
}

func TestFishGenerator_GenerateInitScript_NoJava(t *testing.T) {
	gen := NewFishGenerator()
	envVars := map[string]string{
		"GRADLE_HOME": "/home/user/.jem/current/gradle",
	}

	script := gen.GenerateInitScript(envVars)

	// Check that PATH is NOT updated without JAVA_HOME
	if strings.Contains(script, `set -x PATH`) {
		t.Error("Expected script NOT to contain PATH update without JAVA_HOME")
	}
}

func TestDetectShell_FishVersion(t *testing.T) {
	// Save original env vars
	origShell := os.Getenv("SHELL")
	origFishVersion := os.Getenv("FISH_VERSION")
	defer func() {
		os.Setenv("SHELL", origShell)
		if origFishVersion == "" {
			os.Unsetenv("FISH_VERSION")
		} else {
			os.Setenv("FISH_VERSION", origFishVersion)
		}
	}()

	// Clear SHELL and set FISH_VERSION
	os.Unsetenv("SHELL")
	os.Setenv("FISH_VERSION", "3.5.1")

	shell := DetectShell()
	if shell != config.ShellFish {
		t.Errorf("Expected ShellFish, got %s", shell)
	}
}

// Bash Wrapper Function Tests
func TestBashGenerator_GenerateWrapperFunction(t *testing.T) {
	gen := NewBashGenerator()
	wrapper := gen.GenerateWrapperFunction()

	// Check for function definition
	if !strings.Contains(wrapper, "jem() {") {
		t.Error("Expected wrapper to contain 'jem() {' function definition")
	}

	// Check for "use" subcommand interception
	if !strings.Contains(wrapper, `case "$1" in`) {
		t.Error("Expected wrapper to contain case statement for $1")
	}

	// Check for use pattern
	if !strings.Contains(wrapper, "use)") {
		t.Error("Expected wrapper to contain 'use)' pattern")
	}

	// Check for command jem use "$@" --output-env
	if !strings.Contains(wrapper, `command jem use "$@" --output-env`) {
		t.Error("Expected wrapper to contain 'command jem use \"$@\" --output-env'")
	}

	// Check for eval
	if !strings.Contains(wrapper, `eval "$(`) {
		t.Error("Expected wrapper to contain eval statement")
	}

	// Check for pass-through to non-use commands
	if !strings.Contains(wrapper, "*)") {
		t.Error("Expected wrapper to contain '*' case for pass-through")
	}

	// Check for command jem "$@" for pass-through
	if !strings.Contains(wrapper, `command jem "$@"`) {
		t.Error("Expected wrapper to contain 'command jem \"$@\"' for pass-through")
	}
}

// Zsh Wrapper Function Tests
func TestZshGenerator_GenerateWrapperFunction(t *testing.T) {
	gen := NewZshGenerator()
	wrapper := gen.GenerateWrapperFunction()

	// Check for function definition
	if !strings.Contains(wrapper, "jem() {") {
		t.Error("Expected wrapper to contain 'jem() {' function definition")
	}

	// Check for "use" subcommand interception
	if !strings.Contains(wrapper, `case "$1" in`) {
		t.Error("Expected wrapper to contain case statement for $1")
	}

	// Check for use pattern
	if !strings.Contains(wrapper, "use)") {
		t.Error("Expected wrapper to contain 'use)' pattern")
	}

	// Check for command jem use "$@" --output-env
	if !strings.Contains(wrapper, `command jem use "$@" --output-env`) {
		t.Error("Expected wrapper to contain 'command jem use \"$@\" --output-env'")
	}

	// Check for eval
	if !strings.Contains(wrapper, `eval "$(`) {
		t.Error("Expected wrapper to contain eval statement")
	}

	// Check for pass-through to non-use commands
	if !strings.Contains(wrapper, "*)") {
		t.Error("Expected wrapper to contain '*' case for pass-through")
	}

	// Check for command jem "$@" for pass-through
	if !strings.Contains(wrapper, `command jem "$@"`) {
		t.Error("Expected wrapper to contain 'command jem \"$@\"' for pass-through")
	}
}

// PowerShell Wrapper Function Tests
func TestPowerShellGenerator_GenerateWrapperFunction(t *testing.T) {
	gen := NewPowerShellGenerator()
	wrapper := gen.GenerateWrapperFunction()

	// Check for function definition
	if !strings.Contains(wrapper, "function jem {") {
		t.Error("Expected wrapper to contain 'function jem {' definition")
	}

	// Check for param block with ValueFromRemainingArguments
	if !strings.Contains(wrapper, "[Parameter(ValueFromRemainingArguments)]$Args") {
		t.Error("Expected wrapper to contain param block with ValueFromRemainingArguments")
	}

	// Check for use subcommand interception
	if !strings.Contains(wrapper, "if ($Args[0] -eq 'use')") {
		t.Error("Expected wrapper to check for 'use' subcommand")
	}

	// Check for capturing output with 2>&1
	if !strings.Contains(wrapper, "--output-env 2>&1") {
		t.Error("Expected wrapper to capture output with '2>&1'")
	}

	// Check for LASTEXITCODE validation
	if !strings.Contains(wrapper, "$LASTEXITCODE -eq 0") {
		t.Error("Expected wrapper to check $LASTEXITCODE")
	}

	// Check for Invoke-Expression on success
	if !strings.Contains(wrapper, "Invoke-Expression $output") {
		t.Error("Expected wrapper to contain 'Invoke-Expression $output'")
	}

	// Check for error display via Write-Host
	if !strings.Contains(wrapper, "Write-Host $output") {
		t.Error("Expected wrapper to display errors via 'Write-Host $output'")
	}

	// Check for pass-through to non-use commands
	if !strings.Contains(wrapper, "} else {") {
		t.Error("Expected wrapper to contain else block for pass-through")
	}

	// Check for & jem @Args for pass-through
	if !strings.Contains(wrapper, "& jem @Args") {
		t.Error("Expected wrapper to contain '& jem @Args' for pass-through")
	}
}

// Fish Wrapper Function Tests
func TestFishGenerator_GenerateWrapperFunction(t *testing.T) {
	gen := NewFishGenerator()
	wrapper := gen.GenerateWrapperFunction()

	// Fish wrapper should return a comment explaining lack of support
	if !strings.Contains(wrapper, "Fish shell wrapper not supported") {
		t.Error("Expected wrapper to indicate Fish shell is not supported")
	}

	if !strings.Contains(wrapper, "jem use default") {
		t.Error("Expected wrapper to suggest using 'jem use default'")
	}
}
