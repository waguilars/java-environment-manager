package jdk

import (
	"fmt"

	"github.com/user/jem/internal/config"
	"github.com/user/jem/internal/platform"
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
	paths := s.platform.JDKDetectionPaths()
	var detected []config.JDKInfo

	for _, path := range paths {
		info, err := s.detector.DetectVersion(path)
		if err == nil && info != "" {
			detected = append(detected, config.JDKInfo{
				Path:     path,
				Version:  info,
				Provider: "detected",
				Managed:  false,
			})
		}
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
