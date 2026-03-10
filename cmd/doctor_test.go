package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/waguilars/java-environment-manager/internal/config"
	"github.com/waguilars/java-environment-manager/internal/platform"
)

// MockPlatformForDoctor creates a mock platform for doctor tests
type MockPlatformForDoctor struct {
	platform.LinuxPlatform
	HomeDirFunc func() string
}

func (m *MockPlatformForDoctor) HomeDir() string {
	if m.HomeDirFunc != nil {
		return m.HomeDirFunc()
	}
	return "/tmp"
}

func TestDoctorCommand_Execute_AllPass(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".jem", "config.toml")

	// Create directory structure
	jdksDir := filepath.Join(tmpDir, ".jem", "jdks")
	jdkPath := filepath.Join(jdksDir, "temurin-21")

	if err := os.MkdirAll(jdkPath, 0755); err != nil {
		t.Fatalf("Failed to create JDK dir: %v", err)
	}

	// Create current symlink
	currentLink := filepath.Join(jdksDir, "current")
	if err := os.Symlink(jdkPath, currentLink); err != nil {
		t.Fatalf("Failed to create current symlink: %v", err)
	}

	// Setup config
	repo := config.NewTOMLConfigRepository(configPath)
	repo.SetJDKCurrent("temurin-21")

	cmd := &DoctorCommand{
		platform: &MockPlatformForDoctor{
			HomeDirFunc: func() string { return tmpDir },
		},
		configRepo: repo,
	}

	exitCode := cmd.Execute()
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestDoctorCommand_Execute_BrokenSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".jem", "config.toml")

	// Create directory structure
	jdksDir := filepath.Join(tmpDir, ".jem", "jdks")
	if err := os.MkdirAll(jdksDir, 0755); err != nil {
		t.Fatalf("Failed to create jdks dir: %v", err)
	}

	// Create broken current symlink
	currentLink := filepath.Join(jdksDir, "current")
	if err := os.Symlink("/nonexistent/path", currentLink); err != nil {
		t.Fatalf("Failed to create broken symlink: %v", err)
	}

	repo := config.NewTOMLConfigRepository(configPath)
	repo.SetJDKCurrent("temurin-21")

	cmd := &DoctorCommand{
		platform: &MockPlatformForDoctor{
			HomeDirFunc: func() string { return tmpDir },
		},
		configRepo: repo,
	}

	exitCode := cmd.Execute()
	if exitCode == 0 {
		t.Error("Expected non-zero exit code for broken symlink")
	}
}

func TestDoctorCommand_Execute_NoConfiguredJDK(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".jem", "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	// No JDK configured

	cmd := &DoctorCommand{
		platform: &MockPlatformForDoctor{
			HomeDirFunc: func() string { return tmpDir },
		},
		configRepo: repo,
	}

	// Should complete without panic
	_ = cmd.Execute()
	// Don't check exit code - just verify it doesn't panic
}

func TestParseJavaVersionFromDoctor_OpenJDK(t *testing.T) {
	output := `openjdk version "21.0.2" 2024-01-16 LTS
OpenJDK Runtime Environment Temurin-21.0.2+13 (build 21.0.2+13-LTS)
OpenJDK 64-Bit Server VM Temurin-21.0.2+13 (build 21.0.2+13-LTS, mixed mode, sharing)`

	version := parseJavaVersionFromDoctor(output)
	if version != "21.0.2" {
		t.Errorf("Expected version '21.0.2', got '%s'", version)
	}
}

func TestParseJavaVersionFromDoctor_OracleJDK(t *testing.T) {
	output := `java version "17.0.8" 2023-07-18 LTS
Java(TM) SE Runtime Environment (build 17.0.8+9-LTS-211)
Java HotSpot(TM) 64-Bit Server VM (build 17.0.8+9-LTS-211, mixed mode, sharing)`

	version := parseJavaVersionFromDoctor(output)
	if version != "17.0.8" {
		t.Errorf("Expected version '17.0.8', got '%s'", version)
	}
}

func TestExtractMajorVersionFromDoctor(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"21.0.2", "21"},
		{"17.0.8", "17"},
		{"11", "11"},
		{"temurin-21.0.2", "21"},
		{"openjdk-17.0.5", "17"},
	}

	for _, tc := range tests {
		result := extractMajorVersionFromDoctor(tc.input)
		if result != tc.expected {
			t.Errorf("extractMajorVersionFromDoctor(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}
