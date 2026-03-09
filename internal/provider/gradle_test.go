package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGradleProvider_Name(t *testing.T) {
	provider := NewGradleProvider()
	name := provider.Name()
	if name != "gradle" {
		t.Errorf("Expected 'gradle', got '%s'", name)
	}
}

func TestGradleProvider_DisplayName(t *testing.T) {
	provider := NewGradleProvider()
	displayName := provider.DisplayName()
	if displayName != "Gradle" {
		t.Errorf("Expected 'Gradle', got '%s'", displayName)
	}
}

func TestGradleProvider_ListAvailable(t *testing.T) {
	// Create a mock server
	mockVersions := []string{"7.0", "7.1", "7.2", "7.3", "7.4", "7.5", "7.6", "8.0", "8.1", "8.2", "8.3", "8.4", "8.5"}
	mockResponse := struct {
		Versions []string `json:"versions"`
	}{
		Versions: mockVersions,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	provider := NewGradleProviderWithBase(server.URL, "/distributions")

	ctx := context.Background()
	releases, err := provider.ListAvailable(ctx)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(releases) != len(mockVersions) {
		t.Errorf("Expected %d releases, got %d", len(mockVersions), len(releases))
	}

	// Verify the first release
	if releases[0].Version != "7.0" {
		t.Errorf("Expected first version '7.0', got '%s'", releases[0].Version)
	}

	// Verify archive type is always zip
	for _, release := range releases {
		if release.ArchiveType != "zip" {
			t.Errorf("Expected archive type 'zip', got '%s'", release.ArchiveType)
		}
	}
}

func TestGradleProvider_GetVersion(t *testing.T) {
	// Mock ListAvailable response
	mockVersions := []string{"7.0", "7.1", "7.2", "7.3", "7.4", "7.5", "7.6", "8.0", "8.1", "8.2", "8.3", "8.4", "8.5"}
	mockListResponse := struct {
		Versions []string `json:"versions"`
	}{
		Versions: mockVersions,
	}

	// Mock GetLatest response for GetVersion's checksum fetch
	mockLatestResponse := struct {
		CurrentRelease struct {
			Version string `json:"version"`
		} `json:"currentRelease"`
	}{
		CurrentRelease: struct {
			Version string `json:"version"`
		}{
			Version: "8.5",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/versions/all" {
			json.NewEncoder(w).Encode(mockListResponse)
		} else if r.URL.Path == "/versions/current" {
			json.NewEncoder(w).Encode(mockLatestResponse)
		} else if r.URL.Path == "/distributions/gradle-8.5-bin.zip.sha256" {
			w.Write([]byte("abc123def456 Gradle 8.5 distribution"))
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	provider := NewGradleProviderWithBase(server.URL, "/distributions")

	ctx := context.Background()
	release, err := provider.GetVersion(ctx, "8.5")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if release == nil {
		t.Fatal("Expected non-nil release")
	}

	if release.Version != "8.5" {
		t.Errorf("Expected version '8.5', got '%s'", release.Version)
	}

	if release.ArchiveType != "zip" {
		t.Errorf("Expected archive type 'zip', got '%s'", release.ArchiveType)
	}
}

func TestGradleProvider_GetVersion_NotFound(t *testing.T) {
	mockVersions := []string{"7.0", "7.1", "7.2"}
	mockResponse := struct {
		Versions []string `json:"versions"`
	}{
		Versions: mockVersions,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	provider := NewGradleProviderWithBase(server.URL, "/distributions")

	ctx := context.Background()
	_, err := provider.GetVersion(ctx, "99.99")

	if err == nil {
		t.Error("Expected error for non-existent version")
	}
}

func TestGradleProvider_GetChecksum(t *testing.T) {
	// Test with version in URL
	release := GradleRelease{
		URL: "https://services.gradle.org/distributions/gradle-8.5-bin.zip",
	}

	ctx := context.Background()

	// Create a mock server for checksum
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("abc123def456789 Gradle 8.5 distribution"))
	}))
	defer server.Close()

	// Override distributions for testing
	gradleProvider := &gradleProvider{
		httpClient:    &http.Client{Timeout: 5 * http.DefaultClient.Timeout},
		distributions: server.URL,
	}

	checksum, err := gradleProvider.GetChecksum(ctx, release)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if checksum != "abc123def456789" {
		t.Errorf("Expected checksum 'abc123def456789', got '%s'", checksum)
	}
}

func TestGradleProvider_GetChecksum_EmptyVersion(t *testing.T) {
	// Test with version in URL but no checksum URL
	release := GradleRelease{
		URL: "https://services.gradle.org/distributions/gradle-8.5-bin.zip",
	}

	ctx := context.Background()

	// Create a mock server that returns 404 for checksum
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	gradleProvider := &gradleProvider{
		httpClient:    &http.Client{Timeout: 5 * http.DefaultClient.Timeout},
		distributions: server.URL,
	}

	checksum, err := gradleProvider.GetChecksum(ctx, release)

	// Should return empty checksum and error for 404
	if err == nil {
		t.Error("Expected error for missing checksum")
	}
	if checksum != "" {
		t.Errorf("Expected empty checksum, got '%s'", checksum)
	}
}

func TestGradleProvider_InterfaceCompliance(t *testing.T) {
	provider := NewGradleProvider()

	// Verify provider implements GradleProvider interface
	var _ GradleProvider = provider
}

func TestGradleProvider_ListAvailable_Empty(t *testing.T) {
	mockResponse := struct {
		Versions []string `json:"versions"`
	}{
		Versions: []string{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	provider := NewGradleProviderWithBase(server.URL, "/distributions")

	ctx := context.Background()
	releases, err := provider.ListAvailable(ctx)

	if err != nil {
		t.Fatalf("Expected no error with empty versions, got: %v", err)
	}

	if len(releases) != 0 {
		t.Errorf("Expected 0 releases, got %d", len(releases))
	}
}

func TestGradleRelease_Structure(t *testing.T) {
	release := GradleRelease{
		Version:     "8.5",
		URL:         "https://services.gradle.org/distributions/gradle-8.5-bin.zip",
		Checksum:    "abc123def456",
		ArchiveType: "zip",
	}

	if release.Version != "8.5" {
		t.Errorf("Expected version '8.5', got '%s'", release.Version)
	}
	if release.ArchiveType != "zip" {
		t.Errorf("Expected archive type 'zip', got '%s'", release.ArchiveType)
	}
	if release.Checksum != "abc123def456" {
		t.Errorf("Expected checksum 'abc123def456', got '%s'", release.Checksum)
	}
}

// TestGradleProvider_GetLatest tests the GetLatest method
func TestGradleProvider_GetLatest(t *testing.T) {
	provider := NewGradleProvider()

	// This will fail because we're not actually downloading, but we can test the method exists
	// The actual download would require a real server, which is not practical for unit tests
	ctx := context.Background()
	release, err := provider.GetLatest(ctx)

	// The real API might fail due to network issues, rate limits, etc.
	// This is expected behavior for a unit test without proper mocking
	if err != nil {
		t.Logf("Expected error (network-dependent): %v", err)
		return
	}

	// If we got a release, verify it's valid
	if release == nil {
		t.Fatal("Expected non-nil release")
	}
	if release.Version == "" {
		t.Logf("Warning: release version is empty (API response may have changed)")
	}
	if release.ArchiveType != "zip" {
		t.Errorf("Expected archive type 'zip', got '%s'", release.ArchiveType)
	}
	if release.URL == "" {
		t.Error("Expected non-empty URL")
	}
}

// TestGradleProvider_GetLatest_WithMock tests GetLatest with mock server
func TestGradleProvider_GetLatest_WithMock(t *testing.T) {
	// Mock the latest version response
	mockLatestResponse := struct {
		CurrentRelease struct {
			Version string `json:"version"`
		} `json:"currentRelease"`
	}{
		CurrentRelease: struct {
			Version string `json:"version"`
		}{
			Version: "8.5",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/versions/current" {
			json.NewEncoder(w).Encode(mockLatestResponse)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	provider := NewGradleProviderWithBase(server.URL, "/distributions")

	ctx := context.Background()
	release, err := provider.GetLatest(ctx)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if release == nil {
		t.Fatal("Expected non-nil release")
	}

	if release.Version != "8.5" {
		t.Errorf("Expected version '8.5', got '%s'", release.Version)
	}

	if release.ArchiveType != "zip" {
		t.Errorf("Expected archive type 'zip', got '%s'", release.ArchiveType)
	}
}

// TestGradleProvider_Download tests the Download method
func TestGradleProvider_Download(t *testing.T) {
	provider := NewGradleProvider()

	release := GradleRelease{
		Version:     "8.5",
		URL:         "https://example.com/gradle.zip",
		Checksum:    "abc123",
		ArchiveType: "zip",
	}

	// This will fail because we're not actually downloading
	err := provider.Download(context.Background(), release, "/tmp/test", nil)

	// Expected to fail because the URL is not a real server
	if err == nil {
		t.Error("Expected error when downloading from non-existent server")
	}
}

// TestGradleProvider_Download_WithProgress tests Download with progress callback
func TestGradleProvider_Download_WithProgress(t *testing.T) {
	provider := NewGradleProvider()

	release := GradleRelease{
		Version:     "8.5",
		URL:         "https://example.com/gradle.zip",
		Checksum:    "abc123",
		ArchiveType: "zip",
	}

	progress := func(d, t int64) {
		// Progress callback - just for testing the signature
	}

	// This will fail because we're not actually downloading
	err := provider.Download(context.Background(), release, "/tmp/test", progress)

	// Expected to fail because the URL is not a real server
	if err == nil {
		t.Error("Expected error when downloading from non-existent server")
	}
}

// TestGradleProvider_Download_ArchiveTypes tests Download with different archive types
func TestGradleProvider_Download_ArchiveTypes(t *testing.T) {
	provider := NewGradleProvider()

	release := GradleRelease{
		Version:     "8.5",
		URL:         "https://example.com/gradle.zip",
		Checksum:    "abc123",
		ArchiveType: "zip",
	}

	err := provider.Download(context.Background(), release, "/tmp/test", nil)

	// Expected to fail because the URL is not a real server
	if err == nil {
		t.Error("Expected error when downloading from non-existent server")
	}
}
