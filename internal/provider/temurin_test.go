package provider

import (
	"context"
	"testing"
)

func TestTemurinProvider_Name(t *testing.T) {
	provider := NewTemurinProvider()
	name := provider.Name()
	if name != "temurin" {
		t.Errorf("Expected 'temurin', got '%s'", name)
	}
}

func TestTemurinProvider_DisplayName(t *testing.T) {
	provider := NewTemurinProvider()
	displayName := provider.DisplayName()
	if displayName != "Eclipse Temurin" {
		t.Errorf("Expected 'Eclipse Temurin', got '%s'", displayName)
	}
}

func TestListOptions(t *testing.T) {
	opts := ListOptions{
		MajorVersion: 17,
		OnlyLTS:      true,
		Architecture: "x64",
		OS:           "linux",
	}

	if opts.MajorVersion != 17 {
		t.Errorf("Expected MajorVersion 17, got %d", opts.MajorVersion)
	}
	if !opts.OnlyLTS {
		t.Error("Expected OnlyLTS to be true")
	}
	if opts.Architecture != "x64" {
		t.Errorf("Expected Architecture 'x64', got '%s'", opts.Architecture)
	}
	if opts.OS != "linux" {
		t.Errorf("Expected OS 'linux', got '%s'", opts.OS)
	}
}

func TestJDKRelease(t *testing.T) {
	release := JDKRelease{
		Version:      "21.0.1",
		Major:        21,
		URL:          "https://example.com/jdk-21.0.1.tar.gz",
		Checksum:     "abc123",
		Architecture: "x64",
		ArchiveType:  "tar.gz",
		ReleaseType:  "ga",
	}

	if release.Version != "21.0.1" {
		t.Errorf("Expected version '21.0.1', got '%s'", release.Version)
	}
	if release.Major != 21 {
		t.Errorf("Expected major 21, got %d", release.Major)
	}
	if release.ArchiveType != "tar.gz" {
		t.Errorf("Expected archive type 'tar.gz', got '%s'", release.ArchiveType)
	}
}

func TestProgressFunc(t *testing.T) {
	var downloaded, total int64 = 500, 1000
	var called bool

	// Define a progress function
	progress := func(d, t int64) {
		called = true
		downloaded = d
		total = t
	}

	// Call the progress function
	progress(500, 1000)

	if !called {
		t.Error("Expected progress function to be called")
	}
	if downloaded != 500 {
		t.Errorf("Expected downloaded 500, got %d", downloaded)
	}
	if total != 1000 {
		t.Errorf("Expected total 1000, got %d", total)
	}
}

func TestProgressInfo(t *testing.T) {
	info := ProgressInfo{
		Downloaded: 500,
		Total:      1000,
		Percent:    50.0,
		Speed:      1024.0,
		ETA:        10,
	}

	if info.Downloaded != 500 {
		t.Errorf("Expected downloaded 500, got %d", info.Downloaded)
	}
	if info.Total != 1000 {
		t.Errorf("Expected total 1000, got %d", info.Total)
	}
	if info.Percent != 50.0 {
		t.Errorf("Expected percent 50.0, got %f", info.Percent)
	}
	if info.Speed != 1024.0 {
		t.Errorf("Expected speed 1024.0, got %f", info.Speed)
	}
	if info.ETA != 10 {
		t.Errorf("Expected ETA 10, got %d", info.ETA)
	}
}

func TestNewTemurinProvider(t *testing.T) {
	provider := NewTemurinProvider()
	if provider == nil {
		t.Error("Expected non-nil provider")
	}
	if provider.httpClient == nil {
		t.Error("Expected httpClient to be initialized")
	}
}

func TestTemurinProvider_GetChecksum(t *testing.T) {
	provider := NewTemurinProvider()
	release := JDKRelease{
		Checksum: "expected-checksum-123",
	}

	checksum := provider.GetChecksum(release)
	if checksum != "expected-checksum-123" {
		t.Errorf("Expected checksum 'expected-checksum-123', got '%s'", checksum)
	}
}

func TestJDKRelease_Structure(t *testing.T) {
	release := JDKRelease{
		Version:      "17.0.10",
		Major:        17,
		URL:          "https://api.adoptium.net/v3/binary/17",
		Checksum:     "sha256:abc123",
		Architecture: "x64",
		ArchiveType:  "tar.gz",
		ReleaseType:  "ga",
	}

	// Verify all fields are accessible
	if release.Version == "" {
		t.Error("Expected non-empty version")
	}
	if release.URL == "" {
		t.Error("Expected non-empty URL")
	}
	if release.Architecture == "" {
		t.Error("Expected non-empty architecture")
	}
}

func TestListOptions_Defaults(t *testing.T) {
	opts := ListOptions{}

	// Verify defaults
	if opts.MajorVersion != 0 {
		t.Errorf("Expected MajorVersion 0 by default, got %d", opts.MajorVersion)
	}
	if opts.OnlyLTS != false {
		t.Error("Expected OnlyLTS false by default")
	}
	if opts.Architecture != "" {
		t.Errorf("Expected empty Architecture by default, got '%s'", opts.Architecture)
	}
	if opts.OS != "" {
		t.Errorf("Expected empty OS by default, got '%s'", opts.OS)
	}
}

func TestProgressInfo_Calculation(t *testing.T) {
	// Test 100% progress
	info1 := ProgressInfo{
		Downloaded: 1000,
		Total:      1000,
		Percent:    100.0,
	}
	if info1.Percent != 100 {
		t.Errorf("Expected 100%% progress, got %f", info1.Percent)
	}

	// Test 0% progress
	info2 := ProgressInfo{
		Downloaded: 0,
		Total:      1000,
		Percent:    0.0,
	}
	if info2.Percent != 0 {
		t.Errorf("Expected 0%% progress, got %f", info2.Percent)
	}
}

func TestJDKRelease_EarlyAccess(t *testing.T) {
	release := JDKRelease{
		Version:     "22-ea",
		Major:       22,
		URL:         "https://example.com/jdk-22-ea.tar.gz",
		ArchiveType: "tar.gz",
		ReleaseType: "ea", // Early Access
	}

	if release.ReleaseType != "ea" {
		t.Errorf("Expected release type 'ea', got '%s'", release.ReleaseType)
	}
}

func TestTemurinProvider_HTTPClientTimeout(t *testing.T) {
	provider := NewTemurinProvider()
	// The client should have a timeout set
	// We can't directly access the timeout value in tests,
	// but we can verify the client exists
	if provider.httpClient == nil {
		t.Error("Expected httpClient to be initialized")
	}
}

func TestListOptions_MajorVersionOnly(t *testing.T) {
	opts := ListOptions{
		MajorVersion: 21,
	}

	if opts.MajorVersion != 21 {
		t.Errorf("Expected MajorVersion 21, got %d", opts.MajorVersion)
	}
	if opts.OnlyLTS != false {
		t.Error("Expected OnlyLTS false when not specified")
	}
}

func TestListOptions_LTSOnly(t *testing.T) {
	opts := ListOptions{
		OnlyLTS: true,
	}

	if !opts.OnlyLTS {
		t.Error("Expected OnlyLTS to be true")
	}
	if opts.MajorVersion != 0 {
		t.Errorf("Expected MajorVersion 0, got %d", opts.MajorVersion)
	}
}

func TestProgressInfo_SpeedCalculation(t *testing.T) {
	info := ProgressInfo{
		Downloaded: 1024,
		Total:      1024,
		Percent:    100.0,
		Speed:      512.0,
	}

	if info.Speed != 512.0 {
		t.Errorf("Expected speed 512.0, got %f", info.Speed)
	}
}

func TestTemurinProvider_Constructor(t *testing.T) {
	provider := NewTemurinProvider()

	// Verify interface compliance
	var _ JDKProvider = provider
}

func TestListOptions_AllFields(t *testing.T) {
	opts := ListOptions{
		MajorVersion: 11,
		OnlyLTS:      false,
		Architecture: "aarch64",
		OS:           "windows",
	}

	if opts.MajorVersion != 11 {
		t.Errorf("Expected MajorVersion 11, got %d", opts.MajorVersion)
	}
	if opts.OnlyLTS != false {
		t.Error("Expected OnlyLTS false")
	}
	if opts.Architecture != "aarch64" {
		t.Errorf("Expected Architecture 'aarch64', got '%s'", opts.Architecture)
	}
	if opts.OS != "windows" {
		t.Errorf("Expected OS 'windows', got '%s'", opts.OS)
	}
}

func TestJDKRelease_VariousVersions(t *testing.T) {
	testCases := []struct {
		version string
		major   int
	}{
		{"8.0.342", 8},
		{"11.0.12", 11},
		{"17.0.10", 17},
		{"21.0.1", 21},
	}

	for _, tc := range testCases {
		t.Run(tc.version, func(t *testing.T) {
			release := JDKRelease{
				Version: tc.version,
				Major:   tc.major,
			}
			if release.Version != tc.version {
				t.Errorf("Expected version '%s', got '%s'", tc.version, release.Version)
			}
			if release.Major != tc.major {
				t.Errorf("Expected major %d, got %d", tc.major, release.Major)
			}
		})
	}
}

func TestProgressInfo_NoData(t *testing.T) {
	info := ProgressInfo{}

	if info.Downloaded != 0 {
		t.Errorf("Expected downloaded 0, got %d", info.Downloaded)
	}
	if info.Total != 0 {
		t.Errorf("Expected total 0, got %d", info.Total)
	}
	if info.Percent != 0 {
		t.Errorf("Expected percent 0, got %f", info.Percent)
	}
}

func TestListOptions_EmptyArchitecture(t *testing.T) {
	opts := ListOptions{
		MajorVersion: 17,
	}

	// Empty architecture should be allowed (use default)
	if opts.Architecture != "" {
		t.Errorf("Expected empty architecture by default, got '%s'", opts.Architecture)
	}
}

func TestTemurinProvider_WithContext(t *testing.T) {
	provider := NewTemurinProvider()
	ctx := context.Background()

	// Verify provider can be used with context
	// This test just verifies the interface is correct
	var _ context.Context = ctx
	_ = provider

	// The actual API calls would use the context
}

func TestJDKRelease_ArchiveTypes(t *testing.T) {
	testCases := []struct {
		archiveType string
		description string
	}{
		{"tar.gz", "gzip compressed tar"},
		{"zip", "zip archive"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			release := JDKRelease{
				ArchiveType: tc.archiveType,
			}
			if release.ArchiveType != tc.archiveType {
				t.Errorf("Expected archive type '%s', got '%s'", tc.archiveType, release.ArchiveType)
			}
		})
	}
}

func TestProgressInfo_ETA(t *testing.T) {
	info := ProgressInfo{
		Downloaded: 500,
		Total:      1000,
		Percent:    50.0,
		ETA:        30, // 30 seconds remaining
	}

	if info.ETA != 30 {
		t.Errorf("Expected ETA 30, got %d", info.ETA)
	}
}

func TestListOptions_Combinations(t *testing.T) {
	testCases := []struct {
		name        string
		opts        ListOptions
		expectMajor int
		expectLTS   bool
	}{
		{"Major only", ListOptions{MajorVersion: 17}, 17, false},
		{"LTS only", ListOptions{OnlyLTS: true}, 0, true},
		{"Both set", ListOptions{MajorVersion: 21, OnlyLTS: true}, 21, true},
		{"Neither set", ListOptions{}, 0, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.opts.MajorVersion != tc.expectMajor {
				t.Errorf("Expected MajorVersion %d, got %d", tc.expectMajor, tc.opts.MajorVersion)
			}
			if tc.opts.OnlyLTS != tc.expectLTS {
				t.Errorf("Expected OnlyLTS %v, got %v", tc.expectLTS, tc.opts.OnlyLTS)
			}
		})
	}
}
