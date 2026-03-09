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

// gradleProvider implements GradleProvider for official Gradle distributions
type gradleProvider struct {
	httpClient    *http.Client
	servicesBase  string
	distributions string
}

// NewGradleProvider creates a new GradleProvider
func NewGradleProvider() GradleProvider {
	return &gradleProvider{
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		servicesBase:  "https://services.gradle.org",
		distributions: "/distributions",
	}
}

// NewGradleProviderWithBase creates a new GradleProvider with custom base URLs (for testing)
func NewGradleProviderWithBase(servicesBase, distributions string) GradleProvider {
	return &gradleProvider{
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		servicesBase:  servicesBase,
		distributions: distributions,
	}
}

// Name returns the provider identifier
func (p *gradleProvider) Name() string {
	return "gradle"
}

// DisplayName returns the human-readable name
func (p *gradleProvider) DisplayName() string {
	return "Gradle"
}

// ListAvailable returns all available Gradle releases
func (p *gradleProvider) ListAvailable(ctx context.Context) ([]GradleRelease, error) {
	url := p.servicesBase + "/versions/all"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch versions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// The API returns a JSON object with a "versions" array
	var response struct {
		Versions []string `json:"versions"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Convert version strings to GradleRelease objects
	var releases []GradleRelease
	for _, version := range response.Versions {
		release := GradleRelease{
			Version:     version,
			ArchiveType: "zip",
		}
		releases = append(releases, release)
	}

	return releases, nil
}

// GetLatest returns the latest stable Gradle release
func (p *gradleProvider) GetLatest(ctx context.Context) (*GradleRelease, error) {
	url := p.servicesBase + "/versions/current"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch version: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// The API returns a JSON object with version info
	var response struct {
		CurrentRelease struct {
			Version string `json:"version"`
		} `json:"currentRelease"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	version := response.CurrentRelease.Version
	release := &GradleRelease{
		Version:     version,
		URL:         fmt.Sprintf("%s%s/gradle-%s-bin.zip", p.servicesBase, p.distributions, version),
		ArchiveType: "zip",
	}

	// Fetch checksum
	checksum, err := p.GetChecksum(ctx, *release)
	if err != nil {
		// Log warning but continue
		fmt.Fprintf(io.Discard, "Warning: failed to fetch checksum for Gradle %s: %v\n", version, err)
	}
	release.Checksum = checksum

	return release, nil
}

// GetVersion returns a specific Gradle version
func (p *gradleProvider) GetVersion(ctx context.Context, version string) (*GradleRelease, error) {
	releases, err := p.ListAvailable(ctx)
	if err != nil {
		return nil, err
	}

	for _, release := range releases {
		if release.Version == version {
			release.URL = fmt.Sprintf("%s%s/gradle-%s-bin.zip", p.servicesBase, p.distributions, version)
			release.ArchiveType = "zip"

			// Fetch checksum
			checksum, err := p.GetChecksum(ctx, release)
			if err != nil {
				// Log warning but continue
				fmt.Fprintf(io.Discard, "Warning: failed to fetch checksum for Gradle %s: %v\n", version, err)
			}
			release.Checksum = checksum

			return &release, nil
		}
	}

	return nil, fmt.Errorf("version '%s' not found", version)
}

// Download downloads a Gradle release to the specified path
func (p *gradleProvider) Download(ctx context.Context, release GradleRelease, dest string, progress ProgressFunc) error {
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
func (p *gradleProvider) GetChecksum(ctx context.Context, release GradleRelease) (string, error) {
	// Extract version from URL or use the version field
	version := release.Version
	if version == "" {
		// Try to extract from URL
		parts := strings.Split(release.URL, "/")
		for i, part := range parts {
			if strings.HasPrefix(part, "gradle-") && strings.HasSuffix(part, "-bin.zip") {
				version = strings.TrimSuffix(strings.TrimPrefix(part, "gradle-"), "-bin.zip")
				break
			} else if i == len(parts)-1 && strings.HasPrefix(part, "gradle-") {
				// Last part is the filename
				version = strings.TrimSuffix(strings.TrimPrefix(part, "gradle-"), "-bin.zip")
				break
			}
		}
	}

	// Construct checksum URL
	checksumURL := fmt.Sprintf("%s%s/gradle-%s-bin.zip.sha256", p.servicesBase, p.distributions, version)

	req, err := http.NewRequestWithContext(ctx, "GET", checksumURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch checksum: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("checksum request failed: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read checksum: %w", err)
	}

	// The checksum file contains just the hash and filename
	checksum := strings.TrimSpace(string(body))
	// Extract just the hash (first 64 chars)
	parts := strings.Fields(checksum)
	if len(parts) > 0 {
		return parts[0], nil
	}

	return checksum, nil
}
