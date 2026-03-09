package jdk

import (
	"context"
	"os"
	"path/filepath"

	"github.com/waguilars/java-environment-manager/internal/platform"
)

// JDKInfo represents information about a JDK installation
type JDKInfo struct {
	Path     string
	Version  string
	Provider string
	Managed  bool
}

// JDKDetector interface for discovering JDKs in the system
type JDKDetector interface {
	Scan(ctx context.Context) ([]JDKInfo, error)
	DetectVersion(jdkPath string) (string, error)
}

// PlatformJDKDetector implements JDKDetector using Platform abstraction
type PlatformJDKDetector struct {
	platform platform.Platform
}

// NewPlatformJDKDetector creates a new JDK detector using the platform abstraction
func NewPlatformJDKDetector(platform platform.Platform) *PlatformJDKDetector {
	return &PlatformJDKDetector{
		platform: platform,
	}
}

// Scan scans for JDKs in all detection paths from the platform
func (d *PlatformJDKDetector) Scan(ctx context.Context) ([]JDKInfo, error) {
	detectionPaths := d.platform.JDKDetectionPaths()
	var jdkList []JDKInfo

	for _, path := range detectionPaths {
		jdks, err := d.scanPath(path)
		if err != nil {
			// Log error but continue scanning other paths
			continue
		}
		jdkList = append(jdkList, jdks...)
	}

	return jdkList, nil
}

// scanPath recursively scans a directory for JDK installations
func (d *PlatformJDKDetector) scanPath(path string) ([]JDKInfo, error) {
	var jdks []JDKInfo

	info, err := os.Stat(path)
	if err != nil {
		return jdks, nil // Path doesn't exist, skip
	}

	if !info.IsDir() {
		return jdks, nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return jdks, nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			fullPath := filepath.Join(path, entry.Name())
			version, err := d.DetectVersion(fullPath)
			if err == nil && version != "" {
				jdks = append(jdks, JDKInfo{
					Path:     fullPath,
					Version:  version,
					Provider: "detected",
					Managed:  false,
				})
			}
		}
	}

	return jdks, nil
}

// DetectVersion detects the JDK version from a given path
func (d *PlatformJDKDetector) DetectVersion(jdkPath string) (string, error) {
	// Try to read release file first (standard in JDK 9+)
	releasePath := filepath.Join(jdkPath, "release")
	if content, err := os.ReadFile(releasePath); err == nil {
		version := parseReleaseFile(string(content))
		if version != "" {
			return version, nil
		}
	}

	// Fallback: try java -version
	return d.detectVersionFromCommand(jdkPath)
}

// detectVersionFromCommand runs java -version to detect the version
func (d *PlatformJDKDetector) detectVersionFromCommand(jdkPath string) (string, error) {
	javaBin := filepath.Join(jdkPath, "bin", "java")
	if _, err := os.Stat(javaBin); err != nil {
		return "", os.ErrNotExist
	}

	// For now, return a placeholder - actual implementation would execute java -version
	// and parse the output
	return "", nil
}

// parseReleaseFile parses the release file format
func parseReleaseFile(content string) string {
	for _, line := range splitLines(content) {
		if len(line) == 0 {
			continue
		}
		// Format: KEY="value" or KEY=value
		parts := splitAtFirst(line, "=")
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := trimQuotes(parts[1])
		if key == "JAVA_VERSION" {
			return value
		}
	}
	return ""
}

// splitLines splits content by newlines
func splitLines(content string) []string {
	var lines []string
	var current string
	for _, r := range content {
		if r == '\n' {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

// splitAtFirst splits string at first occurrence of separator
func splitAtFirst(s, sep string) []string {
	idx := -1
	for i := 0; i < len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			idx = i
			break
		}
	}
	if idx == -1 {
		return []string{s}
	}
	return []string{s[:idx], s[idx+len(sep):]}
}

// trimQuotes removes surrounding quotes from a string
func trimQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
