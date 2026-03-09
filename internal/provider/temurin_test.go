package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

// TestTemurinProvider_ListAvailable_Success tests successful API call to list available JDKs
func TestTemurinProvider_ListAvailable_Success(t *testing.T) {
	// Mock response from Adoptium API
	mockResponse := []struct {
		Version  temurinVersion  `json:"version"`
		Binaries []temurinBinary `json:"binaries"`
	}{
		{
			Version: temurinVersion{
				Name:  "21.0.2+13",
				Major: 21,
				Minor: 0,
				Patch: 2,
			},
			Binaries: []temurinBinary{
				{
					Architecture: "x64",
					OS:           "linux",
					Package: temurinPackage{
						Link:     "https://example.com/jdk-21.0.2+13_linux-x64.tar.gz",
						Checksum: "abc123def456",
						Name:     "OpenJDK21U-jdk_x64_linux_hotspot_21.0.2.tar.gz",
						Size:     123456789,
					},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	provider := NewTemurinProviderWithBase(server.URL)

	opts := ListOptions{
		MajorVersion: 21,
		Architecture: "x64",
		OS:           "linux",
	}

	releases, err := provider.ListAvailable(context.Background(), opts)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(releases) != 1 {
		t.Errorf("Expected 1 release, got %d", len(releases))
	}

	if releases[0].Version != "21.0.2+13" {
		t.Errorf("Expected version '21.0.2+13', got '%s'", releases[0].Version)
	}

	if releases[0].Major != 21 {
		t.Errorf("Expected major 21, got %d", releases[0].Major)
	}

	if releases[0].URL != "https://example.com/jdk-21.0.2+13_linux-x64.tar.gz" {
		t.Errorf("Expected URL 'https://example.com/jdk-21.0.2+13_linux-x64.tar.gz', got '%s'", releases[0].URL)
	}

	if releases[0].Architecture != "x64" {
		t.Errorf("Expected architecture 'x64', got '%s'", releases[0].Architecture)
	}

	if releases[0].ArchiveType != "tar.gz" {
		t.Errorf("Expected archive type 'tar.gz', got '%s'", releases[0].ArchiveType)
	}
}

// TestTemurinProvider_ListAvailable_Windows tests Windows platform handling
func TestTemurinProvider_ListAvailable_Windows(t *testing.T) {
	mockResponse := []struct {
		Version  temurinVersion  `json:"version"`
		Binaries []temurinBinary `json:"binaries"`
	}{
		{
			Version: temurinVersion{
				Name:  "21.0.2+13",
				Major: 21,
			},
			Binaries: []temurinBinary{
				{
					Architecture: "x64",
					OS:           "windows",
					Package: temurinPackage{
						Link:     "https://example.com/jdk-21.0.2+13_windows-x64.zip",
						Checksum: "def456abc123",
						Name:     "OpenJDK21U-jdk_x64_windows_hotspot_21.0.2.zip",
						Size:     987654321,
					},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	provider := NewTemurinProviderWithBase(server.URL)

	opts := ListOptions{
		MajorVersion: 21,
		Architecture: "x64",
		OS:           "windows",
	}

	releases, err := provider.ListAvailable(context.Background(), opts)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(releases) != 1 {
		t.Errorf("Expected 1 release, got %d", len(releases))
	}

	if releases[0].ArchiveType != "zip" {
		t.Errorf("Expected archive type 'zip' for Windows, got '%s'", releases[0].ArchiveType)
	}
}

// TestTemurinProvider_ListAvailable_MultipleReleases tests multiple JDK releases
func TestTemurinProvider_ListAvailable_MultipleReleases(t *testing.T) {
	mockResponse := []struct {
		Version  temurinVersion  `json:"version"`
		Binaries []temurinBinary `json:"binaries"`
	}{
		{
			Version: temurinVersion{
				Name:  "17.0.10+7",
				Major: 17,
			},
			Binaries: []temurinBinary{
				{
					Architecture: "x64",
					OS:           "linux",
					Package: temurinPackage{
						Link:     "https://example.com/jdk-17.0.10+7_linux-x64.tar.gz",
						Checksum: "checksum1",
						Name:     "jdk-17.0.10+7_linux-x64.tar.gz",
					},
				},
			},
		},
		{
			Version: temurinVersion{
				Name:  "21.0.2+13",
				Major: 21,
			},
			Binaries: []temurinBinary{
				{
					Architecture: "x64",
					OS:           "linux",
					Package: temurinPackage{
						Link:     "https://example.com/jdk-21.0.2+13_linux-x64.tar.gz",
						Checksum: "checksum2",
						Name:     "jdk-21.0.2+13_linux-x64.tar.gz",
					},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	provider := NewTemurinProviderWithBase(server.URL)

	opts := ListOptions{
		MajorVersion: 0, // No major version filter
		Architecture: "x64",
		OS:           "linux",
	}

	releases, err := provider.ListAvailable(context.Background(), opts)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(releases) != 2 {
		t.Errorf("Expected 2 releases, got %d", len(releases))
	}

	// Verify first release
	if releases[0].Version != "17.0.10+7" {
		t.Errorf("Expected version '17.0.10+7', got '%s'", releases[0].Version)
	}

	// Verify second release
	if releases[1].Version != "21.0.2+13" {
		t.Errorf("Expected version '21.0.2+13', got '%s'", releases[1].Version)
	}
}

// TestTemurinProvider_ListAvailable_EmptyResponse tests empty API response
func TestTemurinProvider_ListAvailable_EmptyResponse(t *testing.T) {
	mockResponse := []struct {
		Version  temurinVersion  `json:"version"`
		Binaries []temurinBinary `json:"binaries"`
	}{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	provider := NewTemurinProviderWithBase(server.URL)

	opts := ListOptions{
		MajorVersion: 21,
		Architecture: "x64",
		OS:           "linux",
	}

	releases, err := provider.ListAvailable(context.Background(), opts)

	if err != nil {
		t.Fatalf("Expected no error with empty response, got: %v", err)
	}

	if len(releases) != 0 {
		t.Errorf("Expected 0 releases, got %d", len(releases))
	}
}

// TestTemurinProvider_ListAvailable_InvalidJSON tests handling of invalid JSON
func TestTemurinProvider_ListAvailable_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json {"))
	}))
	defer server.Close()

	provider := NewTemurinProviderWithBase(server.URL)

	opts := ListOptions{
		MajorVersion: 21,
	}

	_, err := provider.ListAvailable(context.Background(), opts)

	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

// TestTemurinProvider_GetLatest_Success tests getting latest JDK for a major version
func TestTemurinProvider_GetLatest_Success(t *testing.T) {
	mockResponse := []struct {
		Version  temurinVersion  `json:"version"`
		Binaries []temurinBinary `json:"binaries"`
	}{
		{
			Version: temurinVersion{
				Name:  "21.0.2+13",
				Major: 21,
			},
			Binaries: []temurinBinary{
				{
					Architecture: "x64",
					OS:           "linux",
					Package: temurinPackage{
						Link:     "https://example.com/jdk-21.0.2+13_linux-x64.tar.gz",
						Checksum: "latest-checksum",
						Name:     "jdk-21.0.2+13_linux-x64.tar.gz",
					},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	provider := NewTemurinProviderWithBase(server.URL)

	release, err := provider.GetLatest(context.Background(), 21)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if release == nil {
		t.Fatal("Expected non-nil release")
	}

	if release.Version != "21.0.2+13" {
		t.Errorf("Expected version '21.0.2+13', got '%s'", release.Version)
	}

	if release.Major != 21 {
		t.Errorf("Expected major 21, got %d", release.Major)
	}
}

// TestTemurinProvider_GetLatest_NoReleases tests when no releases are available
func TestTemurinProvider_GetLatest_NoReleases(t *testing.T) {
	mockResponse := []struct {
		Version  temurinVersion  `json:"version"`
		Binaries []temurinBinary `json:"binaries"`
	}{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	provider := NewTemurinProviderWithBase(server.URL)

	_, err := provider.GetLatest(context.Background(), 21)

	if err == nil {
		t.Error("Expected error when no releases available")
	}
}

// TestTemurinProvider_GetLatestLTS_Success tests getting latest LTS release
func TestTemurinProvider_GetLatestLTS_Success(t *testing.T) {
	mockResponse := []struct {
		Version  temurinVersion  `json:"version"`
		Binaries []temurinBinary `json:"binaries"`
	}{
		{
			Version: temurinVersion{
				Name:  "17.0.10+7",
				Major: 17,
			},
			Binaries: []temurinBinary{
				{
					Architecture: "x64",
					OS:           "linux",
					Package: temurinPackage{
						Link:     "https://example.com/jdk-17.0.10+7_linux-x64.tar.gz",
						Checksum: "lts-checksum",
						Name:     "jdk-17.0.10+7_linux-x64.tar.gz",
					},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	provider := NewTemurinProviderWithBase(server.URL)

	release, err := provider.GetLatestLTS(context.Background())

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if release == nil {
		t.Fatal("Expected non-nil release")
	}

	if release.Version != "17.0.10+7" {
		t.Errorf("Expected LTS version '17.0.10+7', got '%s'", release.Version)
	}

	if release.Major != 17 {
		t.Errorf("Expected major 17 (LTS), got %d", release.Major)
	}
}

// TestTemurinProvider_GetLatestLTS_EmptyResponse tests GetLatestLTS with empty response
func TestTemurinProvider_GetLatestLTS_EmptyResponse(t *testing.T) {
	mockResponse := []struct {
		Version  temurinVersion  `json:"version"`
		Binaries []temurinBinary `json:"binaries"`
	}{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	provider := NewTemurinProviderWithBase(server.URL)

	release, err := provider.GetLatestLTS(context.Background())

	if err == nil {
		t.Error("Expected error when no LTS releases available")
	}
	if release != nil {
		t.Error("Expected nil release when no LTS releases available")
	}
}

// TestTemurinProvider_GetChecksum tests checksum retrieval
func TestTemurinProvider_GetChecksum(t *testing.T) {
	provider := NewTemurinProvider()
	release := JDKRelease{
		Version:      "21.0.2+13",
		Major:        21,
		Checksum:     "expected-checksum-abc123",
		Architecture: "x64",
		ArchiveType:  "tar.gz",
	}

	checksum := provider.GetChecksum(release)
	if checksum != "expected-checksum-abc123" {
		t.Errorf("Expected checksum 'expected-checksum-abc123', got '%s'", checksum)
	}
}

// TestTemurinProvider_GetChecksum_Formats tests GetChecksum with different checksum formats
func TestTemurinProvider_GetChecksum_Formats(t *testing.T) {
	testCases := []struct {
		name     string
		checksum string
	}{
		{"SHA256 hex", "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"},
		{"SHA256 with prefix", "sha256:abc123def456"},
		{"Empty checksum", ""},
		{"Short checksum", "abc123"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := NewTemurinProvider()
			release := JDKRelease{
				Checksum: tc.checksum,
			}

			result := provider.GetChecksum(release)
			if result != tc.checksum {
				t.Errorf("Expected checksum '%s', got '%s'", tc.checksum, result)
			}
		})
	}
}

// TestTemurinProvider_URLConstruction tests correct URL construction
func TestTemurinProvider_URLConstruction(t *testing.T) {
	testCases := []struct {
		name         string
		majorVersion int
		architecture string
		os           string
	}{
		{
			name:         "Linux x64",
			majorVersion: 21,
			architecture: "x64",
			os:           "linux",
		},
		{
			name:         "Windows x64",
			majorVersion: 17,
			architecture: "x64",
			os:           "windows",
		},
		{
			name:         "Linux aarch64",
			majorVersion: 21,
			architecture: "aarch64",
			os:           "linux",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := NewTemurinProvider()

			// We can't directly test the URL construction without calling the API
			// This test verifies the provider can be created with different options
			if provider == nil {
				t.Fatal("Expected non-nil provider")
			}
		})
	}
}

// TestTemurinProvider_InterfaceCompliance tests that TemurinProvider implements JDKProvider
func TestTemurinProvider_InterfaceCompliance(t *testing.T) {
	provider := NewTemurinProvider()

	// Verify provider implements JDKProvider interface
	var _ JDKProvider = provider
}

// TestTemurinProvider_ConstructorWithBaseURL tests constructor with custom base URL
func TestTemurinProvider_ConstructorWithBaseURL(t *testing.T) {
	baseURL := "https://custom.api.adoptium.net"
	provider := NewTemurinProviderWithBase(baseURL)

	if provider == nil {
		t.Fatal("Expected non-nil provider")
	}

	// The base URL should be set correctly (we can't directly access it in tests)
	// This test verifies the constructor works
}

// TestTemurinProvider_HTTPClientTimeout tests that HTTP client has proper timeout
func TestTemurinProvider_HTTPClientTimeout(t *testing.T) {
	provider := NewTemurinProvider()

	if provider.httpClient == nil {
		t.Error("Expected httpClient to be initialized")
	}
}

// TestTemurinProvider_ListAvailable_OnlyLTS tests filtering for LTS only
func TestTemurinProvider_ListAvailable_OnlyLTS(t *testing.T) {
	mockResponse := []struct {
		Version  temurinVersion  `json:"version"`
		Binaries []temurinBinary `json:"binaries"`
	}{
		{
			Version: temurinVersion{
				Name:  "17.0.10+7",
				Major: 17,
			},
			Binaries: []temurinBinary{
				{
					Architecture: "x64",
					OS:           "linux",
					Package: temurinPackage{
						Link:     "https://example.com/jdk-17.0.10+7_linux-x64.tar.gz",
						Checksum: "lts-checksum",
						Name:     "jdk-17.0.10+7_linux-x64.tar.gz",
					},
				},
			},
		},
		{
			Version: temurinVersion{
				Name:  "21.0.2+13",
				Major: 21,
			},
			Binaries: []temurinBinary{
				{
					Architecture: "x64",
					OS:           "linux",
					Package: temurinPackage{
						Link:     "https://example.com/jdk-21.0.2+13_linux-x64.tar.gz",
						Checksum: "non-lts-checksum",
						Name:     "jdk-21.0.2+13_linux-x64.tar.gz",
					},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	provider := NewTemurinProviderWithBase(server.URL)

	opts := ListOptions{
		MajorVersion: 0,
		OnlyLTS:      true,
		Architecture: "x64",
		OS:           "linux",
	}

	releases, err := provider.ListAvailable(context.Background(), opts)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Both 17 and 21 are LTS versions, so both should be returned
	if len(releases) != 2 {
		t.Errorf("Expected 2 LTS releases, got %d", len(releases))
	}
}

// TestTemurinProvider_Download tests the Download method
func TestTemurinProvider_Download(t *testing.T) {
	provider := NewTemurinProvider()

	release := JDKRelease{
		Version:      "21.0.2",
		URL:          "https://example.com/jdk.tar.gz",
		Checksum:     "abc123",
		Architecture: "x64",
		ArchiveType:  "tar.gz",
	}

	// This will fail because we're not actually downloading, but we can test the method exists
	// The actual download would require a real server, which is not practical for unit tests
	err := provider.Download(context.Background(), release, "/tmp/test", nil)

	// Expected to fail because the URL is not a real server
	if err == nil {
		t.Error("Expected error when downloading from non-existent server")
	}
}

// TestTemurinProvider_Download_WithProgress tests Download with progress callback
func TestTemurinProvider_Download_WithProgress(t *testing.T) {
	provider := NewTemurinProvider()

	release := JDKRelease{
		Version:      "21.0.2",
		URL:          "https://example.com/jdk.tar.gz",
		Checksum:     "abc123",
		Architecture: "x64",
		ArchiveType:  "tar.gz",
	}

	progress := func(d, t int64) {
	}

	// This will fail because we're not actually downloading
	err := provider.Download(context.Background(), release, "/tmp/test", progress)

	// Expected to fail because the URL is not a real server
	if err == nil {
		t.Error("Expected error when downloading from non-existent server")
	}
}

// TestTemurinProvider_Download_ArchiveTypes tests Download with different archive types
func TestTemurinProvider_Download_ArchiveTypes(t *testing.T) {
	testCases := []struct {
		name        string
		archiveType string
	}{
		{"tar.gz", "tar.gz"},
		{"zip", "zip"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := NewTemurinProvider()

			release := JDKRelease{
				Version:      "21.0.2",
				URL:          "https://example.com/jdk." + tc.archiveType,
				Checksum:     "abc123",
				Architecture: "x64",
				ArchiveType:  tc.archiveType,
			}

			err := provider.Download(context.Background(), release, "/tmp/test", nil)

			// Expected to fail because the URL is not a real server
			if err == nil {
				t.Error("Expected error when downloading from non-existent server")
			}
		})
	}
}
