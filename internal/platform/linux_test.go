package platform

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/waguilars/java-environment-manager/internal/config"
)

func TestLinuxPlatform_Name(t *testing.T) {
	p := &LinuxPlatform{}
	if p.Name() != "linux" {
		t.Errorf("Expected name 'linux', got '%s'", p.Name())
	}
}

func TestLinuxPlatform_HomeDir(t *testing.T) {
	p := &LinuxPlatform{}
	home := p.HomeDir()

	if home == "" {
		t.Error("Expected non-empty home directory")
	}

	// Check that home directory exists
	if _, err := os.Stat(home); err != nil {
		t.Errorf("Home directory does not exist: %v", err)
	}
}

func TestLinuxPlatform_CreateLink(t *testing.T) {
	p := &LinuxPlatform{}

	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "target")
	link := filepath.Join(tmpDir, "link")

	// Create target directory
	if err := os.Mkdir(target, 0755); err != nil {
		t.Fatalf("Failed to create target: %v", err)
	}

	// Create symlink
	if err := p.CreateLink(target, link); err != nil {
		t.Errorf("Failed to create symlink: %v", err)
	}

	// Verify symlink was created
	if _, err := os.Stat(link); err != nil {
		t.Errorf("Symlink was not created: %v", err)
	}

	// Clean up
	os.Remove(link)
}

func TestLinuxPlatform_RemoveLink(t *testing.T) {
	p := &LinuxPlatform{}

	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "target")
	link := filepath.Join(tmpDir, "link")

	// Create target and symlink
	os.Mkdir(target, 0755)
	p.CreateLink(target, link)

	// Remove symlink
	if err := p.RemoveLink(link); err != nil {
		t.Errorf("Failed to remove symlink: %v", err)
	}

	// Verify symlink was removed
	if _, err := os.Stat(link); err == nil {
		t.Error("Symlink still exists after removal")
	}
}

func TestLinuxPlatform_IsLink(t *testing.T) {
	p := &LinuxPlatform{}

	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "target")
	link := filepath.Join(tmpDir, "link")

	// Create target and symlink
	os.Mkdir(target, 0755)
	p.CreateLink(target, link)

	// Test IsLink
	if !p.IsLink(link) {
		t.Error("IsLink should return true for symlink")
	}

	// Test with non-link
	if p.IsLink(target) {
		t.Error("IsLink should return false for regular directory")
	}

	// Clean up
	os.Remove(link)
}

func TestLinuxPlatform_CanCreateSymlinks(t *testing.T) {
	p := &LinuxPlatform{}

	// Linux should always be able to create symlinks
	if !p.CanCreateSymlinks() {
		t.Error("LinuxPlatform should always return true for CanCreateSymlinks")
	}
}

func TestLinuxPlatform_DetectShell(t *testing.T) {
	p := &LinuxPlatform{}
	shell := p.DetectShell()

	// Should return one of the valid shells
	validShells := map[config.Shell]bool{
		config.ShellBash:       true,
		config.ShellZsh:        true,
		config.ShellPowerShell: false, // Unlikely on Linux
	}

	if !validShells[shell] {
		t.Errorf("Unexpected shell detected: %s", shell)
	}
}

func TestLinuxPlatform_ShellConfigPath(t *testing.T) {
	p := &LinuxPlatform{}

	paths := map[config.Shell]string{
		config.ShellBash:       filepath.Join(p.HomeDir(), ".bashrc"),
		config.ShellZsh:        filepath.Join(p.HomeDir(), ".zshrc"),
		config.ShellPowerShell: filepath.Join(p.HomeDir(), ".bashrc"), // Default on Linux
	}

	for shell, expected := range paths {
		result := p.ShellConfigPath(shell)
		if result != expected {
			t.Errorf("For shell %s, expected %s, got %s", shell, expected, result)
		}
	}
}

func TestLinuxPlatform_JDKDetectionPaths(t *testing.T) {
	p := &LinuxPlatform{}
	paths := p.JDKDetectionPaths()

	if len(paths) == 0 {
		t.Error("Expected at least one JDK detection path")
	}

	// Check for expected paths
	expectedPaths := []string{
		"/usr/lib/jvm",
	}

	for _, expected := range expectedPaths {
		found := false
		for _, path := range paths {
			if path == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected JDK path %s not found", expected)
		}
	}
}

func TestLinuxPlatform_GradleDetectionPaths(t *testing.T) {
	p := &LinuxPlatform{}
	paths := p.GradleDetectionPaths()

	if len(paths) == 0 {
		t.Error("Expected at least one Gradle detection path")
	}
}
