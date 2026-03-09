package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/waguilars/java-environment-manager/internal/config"
)

// MockPlatformForScan creates a mock platform with custom Gradle detection paths
type MockPlatformForScan struct {
	GradleDetectionPathsFunc func() []string
}

func (m *MockPlatformForScan) Name() string {
	return "mock"
}

func (m *MockPlatformForScan) HomeDir() string {
	return "/tmp"
}

func (m *MockPlatformForScan) DetectShell() config.Shell {
	return config.ShellBash
}

func (m *MockPlatformForScan) CreateLink(target, link string) error {
	return nil
}

func (m *MockPlatformForScan) RemoveLink(link string) error {
	return nil
}

func (m *MockPlatformForScan) IsLink(path string) bool {
	return false
}

func (m *MockPlatformForScan) CanCreateSymlinks() bool {
	return true
}

func (m *MockPlatformForScan) ShellConfigPath(shell config.Shell) string {
	return filepath.Join(m.HomeDir(), ".bashrc")
}

func (m *MockPlatformForScan) JDKDetectionPaths() []string {
	return []string{}
}

func (m *MockPlatformForScan) GradleDetectionPaths() []string {
	if m.GradleDetectionPathsFunc != nil {
		return m.GradleDetectionPathsFunc()
	}
	return []string{}
}

func TestParseGradleVersionOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected string
	}{
		{
			name:     "valid version",
			output:   "Gradle 7.6.1\nBuild time: 2022-03-30 15:53:23 UTC",
			expected: "7.6.1",
		},
		{
			name:     "version with build info",
			output:   "Gradle 8.5.0\nBuild time: 2023-05-10 12:00:00 UTC",
			expected: "8.5.0",
		},
		{
			name:     "no gradle prefix",
			output:   "Build time: 2022-03-30 15:53:23 UTC",
			expected: "",
		},
		{
			name:     "empty output",
			output:   "",
			expected: "",
		},
		{
			name:     "multiple lines with version",
			output:   "Gradle 7.6.1\nSome other text\nGradle 8.5.0",
			expected: "7.6.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseGradleVersionOutput(tt.output)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestDetectGradleVersionFromPath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a fake Gradle directory structure
	gradlePath := filepath.Join(tmpDir, "gradle-7.6.1")
	binPath := filepath.Join(gradlePath, "bin")
	if err := os.MkdirAll(binPath, 0755); err != nil {
		t.Fatalf("Failed to create Gradle directory: %v", err)
	}

	// Create a fake gradle executable
	gradleBin := filepath.Join(binPath, "gradle")
	if err := os.WriteFile(gradleBin, []byte("#!/bin/sh\necho 'Gradle 7.6.1'"), 0755); err != nil {
		t.Fatalf("Failed to write gradle script: %v", err)
	}

	result := detectGradleVersionFromPath(gradlePath)
	if result != "7.6.1" {
		t.Errorf("Expected '7.6.1', got '%s'", result)
	}
}

func TestScanGradlePath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a fake Gradle directory structure
	gradlePath := filepath.Join(tmpDir, "gradle-7.6.1")
	binPath := filepath.Join(gradlePath, "bin")
	if err := os.MkdirAll(binPath, 0755); err != nil {
		t.Fatalf("Failed to create Gradle directory: %v", err)
	}

	// Create a fake gradle executable
	gradleBin := filepath.Join(binPath, "gradle")
	if err := os.WriteFile(gradleBin, []byte("#!/bin/sh\necho 'Gradle 7.6.1'"), 0755); err != nil {
		t.Fatalf("Failed to write gradle script: %v", err)
	}

	result := scanGradlePath(tmpDir)
	if len(result) != 1 {
		t.Errorf("Expected 1 Gradle, got %d", len(result))
	}

	if len(result) > 0 && result[0].Version != "7.6.1" {
		t.Errorf("Expected version '7.6.1', got '%s'", result[0].Version)
	}
}

func TestScanGradles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a fake Gradle directory structure
	gradlePath := filepath.Join(tmpDir, "gradle-7.6.1")
	binPath := filepath.Join(gradlePath, "bin")
	if err := os.MkdirAll(binPath, 0755); err != nil {
		t.Fatalf("Failed to create Gradle directory: %v", err)
	}

	// Create a fake gradle executable
	gradleBin := filepath.Join(binPath, "gradle")
	if err := os.WriteFile(gradleBin, []byte("#!/bin/sh\necho 'Gradle 7.6.1'"), 0755); err != nil {
		t.Fatalf("Failed to write gradle script: %v", err)
	}

	// Create mock platform
	platform := &MockPlatformForScan{
		GradleDetectionPathsFunc: func() []string {
			return []string{tmpDir}
		},
	}

	result := scanGradles(platform)
	if len(result) != 1 {
		t.Errorf("Expected 1 Gradle, got %d", len(result))
	}

	if len(result) > 0 && result[0].Version != "7.6.1" {
		t.Errorf("Expected version '7.6.1', got '%s'", result[0].Version)
	}
}

func TestScanGradlePath_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	result := scanGradlePath(tmpDir)
	if len(result) != 0 {
		t.Errorf("Expected 0 Gradles, got %d", len(result))
	}
}

func TestScanGradlePath_NonExistentDirectory(t *testing.T) {
	result := scanGradlePath("/non/existent/path")
	if len(result) != 0 {
		t.Errorf("Expected 0 Gradles, got %d", len(result))
	}
}

func TestScanGradlePath_WithoutGradleExecutable(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory without gradle executable
	gradlePath := filepath.Join(tmpDir, "gradle-7.6.1")
	if err := os.MkdirAll(gradlePath, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	result := scanGradlePath(gradlePath)
	if len(result) != 0 {
		t.Errorf("Expected 0 Gradles, got %d", len(result))
	}
}
