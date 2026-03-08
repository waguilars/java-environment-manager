package jdk

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/user/jem/internal/config"
)

func TestPlatformJDKDetector_Scan(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()

	// Create a fake JDK directory structure
	jdkPath := filepath.Join(tmpDir, "temurin-21")
	binPath := filepath.Join(jdkPath, "bin")

	// Create directories
	if err := os.MkdirAll(binPath, 0755); err != nil {
		t.Fatalf("Failed to create JDK directory: %v", err)
	}

	// Create a fake release file
	releaseContent := `JAVA_VERSION="21.0.2"
OPENJDK_RUNTIME_ENVIRONMENT=21.0.2+13
`
	if err := os.WriteFile(filepath.Join(jdkPath, "release"), []byte(releaseContent), 0644); err != nil {
		t.Fatalf("Failed to write release file: %v", err)
	}

	// Create a mock platform that returns our temp dir as detection path
	platform := &MockPlatform{
		JDKDetectionPathsFunc: func() []string {
			return []string{tmpDir}
		},
	}

	detector := NewPlatformJDKDetector(platform)

	ctx := context.Background()
	jdks, err := detector.Scan(ctx)

	if err != nil {
		t.Fatalf("Scan() should not error: %v", err)
	}

	if len(jdks) == 0 {
		t.Error("Expected at least one JDK to be detected")
	}
}

func TestPlatformJDKDetector_DetectVersion(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a fake JDK directory with release file
	jdkPath := filepath.Join(tmpDir, "temurin-21")
	if err := os.MkdirAll(jdkPath, 0755); err != nil {
		t.Fatalf("Failed to create JDK directory: %v", err)
	}

	// Create a fake release file
	releaseContent := `JAVA_VERSION="21.0.2"
OPENJDK_RUNTIME_ENVIRONMENT=21.0.2+13
`
	if err := os.WriteFile(filepath.Join(jdkPath, "release"), []byte(releaseContent), 0644); err != nil {
		t.Fatalf("Failed to write release file: %v", err)
	}

	platform := &MockPlatform{}
	detector := NewPlatformJDKDetector(platform)

	version, err := detector.DetectVersion(jdkPath)

	if err != nil {
		t.Fatalf("DetectVersion() should not error: %v", err)
	}

	if version != "21.0.2" {
		t.Errorf("Expected version '21.0.2', got '%s'", version)
	}
}

func TestParseReleaseFile(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "valid version",
			content:  `JAVA_VERSION="21.0.2"`,
			expected: "21.0.2",
		},
		{
			name:     "version with build info",
			content:  `JAVA_VERSION="17.0.10"`,
			expected: "17.0.10",
		},
		{
			name: "multiple lines",
			content: `JAVA_VERSION="21.0.2"
OTHER_VAR=value`,
			expected: "21.0.2",
		},
		{
			name:     "single quotes",
			content:  `JAVA_VERSION='21.0.2'`,
			expected: "21.0.2",
		},
		{
			name:     "no version",
			content:  `OTHER_VAR=value`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseReleaseFile(tt.content)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "single line",
			content:  "hello",
			expected: []string{"hello"},
		},
		{
			name:     "multiple lines",
			content:  "line1\nline2\nline3",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "empty string",
			content:  "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitLines(tt.content)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d lines, got %d", len(tt.expected), len(result))
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("Line %d: expected '%s', got '%s'", i, tt.expected[i], result[i])
				}
			}
		})
	}
}

func TestSplitAtFirst(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		sep      string
		expected []string
	}{
		{
			name:     "simple split",
			s:        "key=value",
			sep:      "=",
			expected: []string{"key", "value"},
		},
		{
			name:     "no separator",
			s:        "noequals",
			sep:      "=",
			expected: []string{"noequals"},
		},
		{
			name:     "multiple separators",
			s:        "a=b=c",
			sep:      "=",
			expected: []string{"a", "b=c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitAtFirst(tt.s, tt.sep)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d parts, got %d", len(tt.expected), len(result))
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("Part %d: expected '%s', got '%s'", i, tt.expected[i], result[i])
				}
			}
		})
	}
}

func TestTrimQuotes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "double quotes",
			input:    `"hello"`,
			expected: "hello",
		},
		{
			name:     "single quotes",
			input:    `'hello'`,
			expected: "hello",
		},
		{
			name:     "no quotes",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := trimQuotes(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// MockPlatform is a mock implementation of Platform interface for testing
type MockPlatform struct {
	HomeDirFunc           func() string
	CreateLinkFunc        func(target, link string) error
	IsLinkFunc            func(path string) bool
	RemoveLinkFunc        func(link string) error
	JDKDetectionPathsFunc func() []string
}

func (m *MockPlatform) Name() string {
	return "mock"
}

func (m *MockPlatform) HomeDir() string {
	if m.HomeDirFunc != nil {
		return m.HomeDirFunc()
	}
	return "/tmp"
}

func (m *MockPlatform) DetectShell() config.Shell {
	return config.ShellBash
}

func (m *MockPlatform) CreateLink(target, link string) error {
	if m.CreateLinkFunc != nil {
		return m.CreateLinkFunc(target, link)
	}
	return nil
}

func (m *MockPlatform) RemoveLink(link string) error {
	if m.RemoveLinkFunc != nil {
		return m.RemoveLinkFunc(link)
	}
	return nil
}

func (m *MockPlatform) IsLink(path string) bool {
	if m.IsLinkFunc != nil {
		return m.IsLinkFunc(path)
	}
	return false
}

func (m *MockPlatform) CanCreateSymlinks() bool {
	return true
}

func (m *MockPlatform) ShellConfigPath(shell config.Shell) string {
	return filepath.Join(m.HomeDir(), ".bashrc")
}

func (m *MockPlatform) JDKDetectionPaths() []string {
	if m.JDKDetectionPathsFunc != nil {
		return m.JDKDetectionPathsFunc()
	}
	return []string{}
}

func (m *MockPlatform) GradleDetectionPaths() []string {
	return []string{}
}
