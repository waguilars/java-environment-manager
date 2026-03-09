package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	temurinAPIBase = "https://api.adoptium.net/v3"
)

// TemurinProvider implements JDKProvider for Eclipse Temurin (Adoptium)
type TemurinProvider struct {
	httpClient *http.Client
	apiBaseURL string
}

// NewTemurinProvider creates a new TemurinProvider with the default API base URL
func NewTemurinProvider() *TemurinProvider {
	return &TemurinProvider{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiBaseURL: temurinAPIBase,
	}
}

// NewTemurinProviderWithBase creates a new TemurinProvider with a custom API base URL (for testing)
func NewTemurinProviderWithBase(baseURL string) *TemurinProvider {
	return &TemurinProvider{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiBaseURL: baseURL,
	}
}

// Name returns the provider identifier
func (p *TemurinProvider) Name() string {
	return "temurin"
}

// DisplayName returns the human-readable name
func (p *TemurinProvider) DisplayName() string {
	return "Eclipse Temurin"
}

// ListAvailable returns available JDK releases matching the criteria
func (p *TemurinProvider) ListAvailable(ctx context.Context, opts ListOptions) ([]JDKRelease, error) {
	// Determine OS and architecture based on options
	os := "linux"
	architecture := "x64"
	archiveType := "tar.gz"

	if opts.OS != "" {
		os = opts.OS
	}
	if opts.Architecture != "" {
		architecture = opts.Architecture
	}

	// Determine archive type based on OS
	if os == "windows" {
		archiveType = "zip"
	}

	// Build the API URL
	url := fmt.Sprintf("%s/assets/latest/%d/hotspot?architecture=%s&os=%s&image_type=jdk",
		p.apiBaseURL, opts.MajorVersion, architecture, os)

	// Add LTS filter if requested
	if opts.OnlyLTS {
		url += "&lts=true"
	}

	return p.fetchReleases(ctx, url, archiveType)
}

// GetLatestLTS returns the latest LTS release
func (p *TemurinProvider) GetLatestLTS(ctx context.Context) (*JDKRelease, error) {
	releases, err := p.ListAvailable(ctx, ListOptions{OnlyLTS: true})
	if err != nil {
		return nil, err
	}

	if len(releases) == 0 {
		return nil, fmt.Errorf("no LTS releases found")
	}

	return &releases[0], nil
}

// GetLatest returns the latest release for a specific major version
func (p *TemurinProvider) GetLatest(ctx context.Context, majorVersion int) (*JDKRelease, error) {
	releases, err := p.ListAvailable(ctx, ListOptions{MajorVersion: majorVersion})
	if err != nil {
		return nil, err
	}

	if len(releases) == 0 {
		return nil, fmt.Errorf("no releases found for Java %d", majorVersion)
	}

	return &releases[0], nil
}

// Download downloads a JDK release to the specified path
func (p *TemurinProvider) Download(ctx context.Context, release JDKRelease, dest string, progress ProgressFunc) error {
	req, err := http.NewRequestWithContext(ctx, "GET", release.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	// TODO: Implement download with progress
	return nil
}

// GetChecksum returns the expected checksum for a release
func (p *TemurinProvider) GetChecksum(release JDKRelease) string {
	return release.Checksum
}

// fetchReleases fetches releases from the Adoptium API
func (p *TemurinProvider) fetchReleases(ctx context.Context, url string, archiveType string) ([]JDKRelease, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse the response - the API returns an array of release objects
	var releases []temurinAPIResponse
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Convert to JDKRelease format
	var result []JDKRelease
	for _, r := range releases {
		jdkRelease, err := parseTemurinResponse(r, archiveType)
		if err != nil {
			return nil, err
		}
		result = append(result, *jdkRelease)
	}

	return result, nil
}

// parseTemurinResponse parses a single release from the API response
func parseTemurinResponse(resp temurinAPIResponse, archiveType string) (*JDKRelease, error) {
	versionStr := resp.Version.Name
	major := resp.Version.Major

	// Find the correct binary for the requested architecture and OS
	var binary temurinBinary
	found := false

	for _, b := range resp.Binaries {
		if b.Architecture == "x64" || b.Architecture == "x86" || b.Architecture == "x32" {
			binary = b
			found = true
			break
		}
	}

	if !found && len(resp.Binaries) > 0 {
		binary = resp.Binaries[0]
		found = true
	}

	if !found {
		return nil, fmt.Errorf("no binary found for release %s", versionStr)
	}

	packageInfo := binary.Package
	if packageInfo.Link == "" {
		return nil, fmt.Errorf("no download link found for release %s", versionStr)
	}

	// Parse version components
	var majorVersion int
	parts := strings.Split(versionStr, "+")
	if len(parts) > 0 {
		// Try to extract major from the version string
		versionParts := strings.Split(parts[0], ".")
		if len(versionParts) > 0 {
			fmt.Sscanf(versionParts[0], "%d", &majorVersion)
		}
	}

	// Determine archive type from the download link
	archiveTypeFromURL := archiveType
	if strings.HasSuffix(packageInfo.Link, ".zip") {
		archiveTypeFromURL = "zip"
	}

	release := &JDKRelease{
		Version:      versionStr,
		Major:        major,
		URL:          packageInfo.Link,
		Checksum:     packageInfo.Checksum,
		Architecture: binary.Architecture,
		ArchiveType:  archiveTypeFromURL,
		ReleaseType:  "ga",
	}

	return release, nil
}

// temurinAPIResponse represents the Adoptium API response structure
type temurinAPIResponse struct {
	Version  temurinVersion  `json:"version"`
	Binaries []temurinBinary `json:"binaries"`
}

type temurinVersion struct {
	Name  string `json:"name"`  // e.g., "21.0.2+13"
	Major int    `json:"major"` // e.g., 21
	Minor int    `json:"minor"`
	Patch int    `json:"patch"`
}

type temurinBinary struct {
	Architecture string         `json:"architecture"` // e.g., "x64"
	OS           string         `json:"os"`           // e.g., "linux"
	Package      temurinPackage `json:"package"`
}

type temurinPackage struct {
	Link     string `json:"link"`     // Download URL
	Checksum string `json:"checksum"` // SHA256
	Name     string `json:"name"`     // Filename
	Size     int64  `json:"size"`     // Bytes
}
