package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectSystemGradle(t *testing.T) {
	// Save original environment
	originalGradleHome := os.Getenv("GRADLE_HOME")
	originalPath := os.Getenv("PATH")
	defer func() {
		os.Setenv("GRADLE_HOME", originalGradleHome)
		os.Setenv("PATH", originalPath)
	}()

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

	// Test 1: GRADLE_HOME set
	t.Run("GRADLE_HOME set", func(t *testing.T) {
		os.Setenv("GRADLE_HOME", gradlePath)
		os.Setenv("PATH", binPath)

		result := detectSystemGradle()
		if result == nil {
			t.Error("Expected Gradle info, got nil")
			return
		}
		if result.Version != "7.6.1" {
			t.Errorf("Expected version '7.6.1', got '%s'", result.Version)
		}
		if result.Path != gradlePath {
			t.Errorf("Expected path '%s', got '%s'", gradlePath, result.Path)
		}
		if result.Managed {
			t.Error("Expected Managed to be false")
		}
	})

	// Test 2: GRADLE_HOME not set, gradle in PATH
	t.Run("gradle in PATH", func(t *testing.T) {
		os.Unsetenv("GRADLE_HOME")
		os.Setenv("PATH", binPath)

		result := detectSystemGradle()
		if result == nil {
			t.Error("Expected Gradle info, got nil")
			return
		}
		if result.Version != "7.6.1" {
			t.Errorf("Expected version '7.6.1', got '%s'", result.Version)
		}
	})

	// Test 3: No Gradle available
	t.Run("no gradle available", func(t *testing.T) {
		os.Unsetenv("GRADLE_HOME")
		os.Setenv("PATH", "/usr/bin:/bin")

		result := detectSystemGradle()
		if result != nil {
			t.Errorf("Expected nil, got %+v", result)
		}
	})
}

func TestDetectSystemGradle_NonExistentPath(t *testing.T) {
	// Save original environment
	originalGradleHome := os.Getenv("GRADLE_HOME")
	defer func() {
		os.Setenv("GRADLE_HOME", originalGradleHome)
	}()

	os.Setenv("GRADLE_HOME", "/non/existent/gradle")
	os.Setenv("PATH", "/non/existent/gradle/bin")

	result := detectSystemGradle()
	if result != nil {
		t.Errorf("Expected nil for non-existent path, got %+v", result)
	}
}

func TestDetectActualJavaVersion(t *testing.T) {
	// This test may not work in CI if java is not installed
	// Just verify the function doesn't panic
	_ = detectActualJavaVersion()
}

func TestParseJavaVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "OpenJDK format",
			input:    "openjdk version \"21.0.2\" 2024-01-16 LTS",
			expected: "21.0.2",
		},
		{
			name:     "Oracle JDK format",
			input:    "java version \"17.0.8\" 2023-07-18 LTS",
			expected: "17.0.8",
		},
		{
			name:     "No quotes",
			input:    "some random output",
			expected: "",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := parseJavaVersion(tc.input)
			if result != tc.expected {
				t.Errorf("parseJavaVersion(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestExtractMajorVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"21.0.2", "21"},
		{"17.0.8", "17"},
		{"11", "11"},
		{"8", "8"},
		{"temurin-21.0.2", "21"},
		{"openjdk-17.0.5", "17"},
		{"", ""},
		{"abc", ""},
	}

	for _, tc := range tests {
		result := extractMajorVersion(tc.input)
		if result != tc.expected {
			t.Errorf("extractMajorVersion(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}
