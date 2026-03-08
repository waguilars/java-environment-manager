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
}

// NewTemurinProvider creates a new TemurinProvider
func NewTemurinProvider() *TemurinProvider {
	return &TemurinProvider{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
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
	url := fmt.Sprintf("%s/releases?architecture=x64&os=linux&image_type=jdk&sortMethod=VERSION&sortOrder=DESC", temurinAPIBase)

	if opts.MajorVersion > 0 {
		url = fmt.Sprintf("%s&version=%d", url, opts.MajorVersion)
	}

	if opts.OnlyLTS {
		url = fmt.Sprintf("%s&lts=true", url)
	}

	return p.fetchReleases(ctx, url)
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
func (p *TemurinProvider) fetchReleases(ctx context.Context, url string) ([]JDKRelease, error) {
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

	// Parse the response
	// Note: This is a simplified parser - actual API response format may vary
	var releases []JDKRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	return releases, nil
}

// parseJDKRelease parses a single release from the API response
func parseJDKRelease(data map[string]interface{}) (*JDKRelease, error) {
	versionStr, ok := data["version"].(string)
	if !ok {
		return nil, fmt.Errorf("missing version field")
	}

	// Parse version components
	var major int
	parts := strings.Split(versionStr, ".")
	if len(parts) > 0 {
		fmt.Sscanf(parts[0], "%d", &major)
	}

	// Extract release metadata
	release := &JDKRelease{
		Version:      versionStr,
		Major:        major,
		Architecture: "x64",
		ArchiveType:  "tar.gz",
		ReleaseType:  "ga",
	}

	// Parse additional fields if present
	if v, ok := data["binary"].(map[string]interface{}); ok {
		if url, ok := v["link"].(string); ok {
			release.URL = url
		}
		if checksum, ok := v["checksum"].(string); ok {
			release.Checksum = checksum
		}
		if archiveType, ok := v["package"].(map[string]interface{}); ok {
			if atype, ok := archiveType["link"].(string); ok {
				if strings.HasSuffix(atype, ".zip") {
					release.ArchiveType = "zip"
				}
			}
		}
	}

	return release, nil
}
