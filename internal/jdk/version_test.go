package jdk

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected *Version
		hasError bool
	}{
		{
			name:     "simple version",
			version:  "21.0.2",
			expected: &Version{Major: 21, Minor: 0, Patch: 2, Raw: "21.0.2"},
			hasError: false,
		},
		{
			name:     "version with build",
			version:  "17.0.10+13",
			expected: &Version{Major: 17, Minor: 0, Patch: 10, Build: 13, Raw: "17.0.10+13"},
			hasError: false,
		},
		{
			name:     "major only",
			version:  "11",
			expected: &Version{Major: 11, Minor: 0, Patch: 0, Build: 0, Raw: "11"},
			hasError: false,
		},
		{
			name:     "major.minor",
			version:  "8.0",
			expected: &Version{Major: 8, Minor: 0, Patch: 0, Build: 0, Raw: "8.0"},
			hasError: false,
		},
		{
			name:     "invalid version",
			version:  "invalid",
			expected: nil,
			hasError: true,
		},
		{
			name:     "empty version",
			version:  "",
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseVersion(tt.version)

			if tt.hasError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result.Major != tt.expected.Major {
				t.Errorf("Expected Major %d, got %d", tt.expected.Major, result.Major)
			}
			if result.Minor != tt.expected.Minor {
				t.Errorf("Expected Minor %d, got %d", tt.expected.Minor, result.Minor)
			}
			if result.Patch != tt.expected.Patch {
				t.Errorf("Expected Patch %d, got %d", tt.expected.Patch, result.Patch)
			}
			if result.Build != tt.expected.Build {
				t.Errorf("Expected Build %d, got %d", tt.expected.Build, result.Build)
			}
			if result.Raw != tt.expected.Raw {
				t.Errorf("Expected Raw '%s', got '%s'", tt.expected.Raw, result.Raw)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int // -1, 0, 1
		hasError bool
	}{
		{
			name:     "v1 greater",
			v1:       "21.0.2",
			v2:       "17.0.10",
			expected: 1,
			hasError: false,
		},
		{
			name:     "v1 less",
			v1:       "17.0.10",
			v2:       "21.0.2",
			expected: -1,
			hasError: false,
		},
		{
			name:     "equal",
			v1:       "21.0.2",
			v2:       "21.0.2",
			expected: 0,
			hasError: false,
		},
		{
			name:     "same major, different minor",
			v1:       "21.0.1",
			v2:       "21.1.0",
			expected: -1,
			hasError: false,
		},
		{
			name:     "same major.minor, different patch",
			v1:       "21.0.2",
			v2:       "21.0.1",
			expected: 1,
			hasError: false,
		},
		{
			name:     "same version with different build",
			v1:       "21.0.2+13",
			v2:       "21.0.2+12",
			expected: 1,
			hasError: false,
		},
		{
			name:     "invalid v1",
			v1:       "invalid",
			v2:       "21.0.2",
			expected: 0,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CompareVersions(tt.v1, tt.v2)

			if tt.hasError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestVersion_IsLTS(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{
			name:     "JDK 8 is LTS",
			version:  "8.0.0",
			expected: true,
		},
		{
			name:     "JDK 11 is LTS",
			version:  "11.0.0",
			expected: true,
		},
		{
			name:     "JDK 17 is LTS",
			version:  "17.0.0",
			expected: true,
		},
		{
			name:     "JDK 21 is LTS",
			version:  "21.0.0",
			expected: true,
		},
		{
			name:     "JDK 19 is not LTS",
			version:  "19.0.0",
			expected: false,
		},
		{
			name:     "JDK 23 is not LTS",
			version:  "23.0.0",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := ParseVersion(tt.version)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			result := version.IsLTS()
			if result != tt.expected {
				t.Errorf("Expected IsLTS() = %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestVersion_IsGreaterThan(t *testing.T) {
	v1, _ := ParseVersion("21.0.2")
	v2, _ := ParseVersion("17.0.10")
	v3, _ := ParseVersion("21.0.2")

	if !v1.IsGreaterThan(v2) {
		t.Error("21.0.2 should be greater than 17.0.10")
	}

	if v1.IsGreaterThan(v3) {
		t.Error("21.0.2 should not be greater than 21.0.2")
	}
}

func TestVersion_IsLessThan(t *testing.T) {
	v1, _ := ParseVersion("17.0.10")
	v2, _ := ParseVersion("21.0.2")
	v3, _ := ParseVersion("17.0.10")

	if !v1.IsLessThan(v2) {
		t.Error("17.0.10 should be less than 21.0.2")
	}

	if v1.IsLessThan(v3) {
		t.Error("17.0.10 should not be less than 17.0.10")
	}
}

func TestVersion_Equals(t *testing.T) {
	v1, _ := ParseVersion("21.0.2")
	v2, _ := ParseVersion("21.0.2")
	v3, _ := ParseVersion("17.0.10")

	if !v1.Equals(v2) {
		t.Error("21.0.2 should equal 21.0.2")
	}

	if v1.Equals(v3) {
		t.Error("21.0.2 should not equal 17.0.10")
	}
}

func TestParseVersionError(t *testing.T) {
	err := &ParseVersionError{
		Version: "invalid",
		Message: "invalid version format",
	}

	expected := "invalid version format: invalid"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}
