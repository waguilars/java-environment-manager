package jdk

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/waguilars/java-environment-manager/internal/config"
	"github.com/waguilars/java-environment-manager/internal/platform"
)

// JDKService handles JDK management operations
type JDKService struct {
	platform   platform.Platform
	configRepo config.ConfigRepository
	detector   *PlatformJDKDetector
	symlinker  *JDKSymlinker
}

// NewJDKService creates a new JDKService with dependencies
func NewJDKService(platform platform.Platform, configRepo config.ConfigRepository) *JDKService {
	return &JDKService{
		platform:   platform,
		configRepo: configRepo,
		detector:   NewPlatformJDKDetector(platform),
		symlinker:  NewJDKSymlinker(platform),
	}
}

// Install installs a JDK version by downloading and extracting it
func (s *JDKService) Install(version string) error {
	// For now, return not implemented
	return fmt.Errorf("install not yet implemented")
}

// Use switches to a different JDK version
func (s *JDKService) Use(name string) error {
	// For now, return not implemented
	return fmt.Errorf("use not yet implemented")
}

// List returns all installed JDKs
func (s *JDKService) List() ([]config.JDKInfo, error) {
	return s.configRepo.ListInstalledJDKs(), nil
}

// Detect scans for JDKs in the system
func (s *JDKService) Detect() ([]config.JDKInfo, error) {
	// Use Scan to properly scan subdirectories
	jdkInfos, err := s.detector.Scan(context.Background())
	if err != nil {
		return nil, err
	}

	// Convert to config.JDKInfo
	var detected []config.JDKInfo
	for _, info := range jdkInfos {
		detected = append(detected, config.JDKInfo{
			Path:     info.Path,
			Version:  info.Version,
			Provider: info.Provider,
			Managed:  info.Managed,
		})
	}

	return detected, nil
}

// GetCurrent returns the currently active JDK
func (s *JDKService) GetCurrent() (*config.JDKInfo, error) {
	currentName := s.configRepo.GetJDKCurrent()
	if currentName == "" {
		return nil, fmt.Errorf("no JDK currently active")
	}

	// Find the JDK in installed or detected
	allJDKs := append(
		s.configRepo.ListInstalledJDKs(),
		s.configRepo.ListDetectedJDKs()...,
	)

	for _, jdk := range allJDKs {
		if jdk.Version == currentName {
			return &jdk, nil
		}
	}

	return nil, fmt.Errorf("current JDK '%s' not found", currentName)
}

// DetectVersion detects the JDK version from a given path
func (s *JDKService) DetectVersion(jdkPath string) (string, error) {
	return s.detector.DetectVersion(jdkPath)
}

// GetJDKSymlinker returns the symlinker instance
func (s *JDKService) GetJDKSymlinker() *JDKSymlinker {
	return s.symlinker
}

// DetectSystemJava detects the currently active Java from JAVA_HOME or PATH
func DetectSystemJava() *config.JDKInfo {
	// Try JAVA_HOME first
	javaHome := os.Getenv("JAVA_HOME")
	if javaHome != "" {
		version := detectJavaVersionFromPath(javaHome)
		if version != "" {
			return &config.JDKInfo{
				Path:     javaHome,
				Version:  version,
				Provider: "system",
				Managed:  false,
			}
		}
	}

	// Try java from PATH
	javaPath, err := exec.LookPath("java")
	if err == nil {
		// Resolve symlink to find the actual JAVA_HOME
		realPath, err := filepath.EvalSymlinks(javaPath)
		if err == nil {
			// java is typically in bin/ directory, go up to find JAVA_HOME
			javaHome := filepath.Dir(filepath.Dir(realPath))
			version := detectJavaVersionFromPath(javaHome)
			if version != "" {
				return &config.JDKInfo{
					Path:     javaHome,
					Version:  version,
					Provider: "system",
					Managed:  false,
				}
			}
		}
	}

	return nil
}

// detectJavaVersionFromPath tries to detect Java version from a JDK path
func detectJavaVersionFromPath(jdkPath string) string {
	// Try release file first
	releasePath := filepath.Join(jdkPath, "release")
	if content, err := os.ReadFile(releasePath); err == nil {
		version := parseReleaseFile(string(content))
		if version != "" {
			return version
		}
	}

	// Try running java -version
	javaBin := filepath.Join(jdkPath, "bin", "java")
	if _, err := os.Stat(javaBin); err == nil {
		cmd := exec.Command(javaBin, "-version")
		output, err := cmd.CombinedOutput()
		if err == nil {
			return parseJavaVersionOutput(string(output))
		}
	}

	return ""
}

// parseJavaVersionOutput parses java -version output
func parseJavaVersionOutput(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// Look for version line like: openjdk version "21.0.1" or java version "1.8.0_342"
		if strings.Contains(line, "version") {
			// Extract version between quotes
			start := strings.Index(line, "\"")
			if start != -1 {
				end := strings.Index(line[start+1:], "\"")
				if end != -1 {
					version := line[start+1 : start+1+end]
					// Normalize version (e.g., 1.8.0_342 -> 8)
					return normalizeJavaVersion(version)
				}
			}
		}
	}
	return ""
}

// normalizeJavaVersion normalizes Java version string
func normalizeJavaVersion(version string) string {
	// Handle old-style versions like 1.8.0_342
	if strings.HasPrefix(version, "1.") {
		parts := strings.Split(version, ".")
		if len(parts) >= 2 {
			return parts[1]
		}
	}
	// Handle new-style versions like 21.0.1
	parts := strings.Split(version, ".")
	if len(parts) >= 1 {
		return parts[0]
	}
	return version
}
