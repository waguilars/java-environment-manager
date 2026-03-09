package provider

import (
	"context"
)

// JDKProvider interface defines the contract for JDK providers
type JDKProvider interface {
	// Name returns the provider's identifier
	Name() string

	// DisplayName returns the human-readable provider name
	DisplayName() string

	// ListAvailable returns available JDK releases matching the criteria
	ListAvailable(ctx context.Context, opts ListOptions) ([]JDKRelease, error)

	// GetLatestLTS returns the latest LTS release
	GetLatestLTS(ctx context.Context) (*JDKRelease, error)

	// GetLatest returns the latest release for a specific major version
	GetLatest(ctx context.Context, majorVersion int) (*JDKRelease, error)

	// Download downloads a JDK release to the specified path
	Download(ctx context.Context, release JDKRelease, dest string, progress ProgressFunc) error

	// GetChecksum returns the expected checksum for a release
	GetChecksum(release JDKRelease) string
}

// ListOptions defines criteria for listing available JDKs
type ListOptions struct {
	MajorVersion int
	OnlyLTS      bool
	Architecture string
	OS           string
}

// JDKRelease represents a specific JDK release
type JDKRelease struct {
	Version      string
	Major        int
	URL          string
	Checksum     string
	Architecture string
	ArchiveType  string // "zip" | "tar.gz"
	ReleaseType  string // "ga" | "ea"
}

// GradleProvider interface defines the contract for Gradle providers
type GradleProvider interface {
	// Name returns the provider's identifier
	Name() string

	// DisplayName returns the human-readable provider name
	DisplayName() string

	// ListAvailable returns all available Gradle releases
	ListAvailable(ctx context.Context) ([]GradleRelease, error)

	// GetLatest returns the latest stable Gradle release
	GetLatest(ctx context.Context) (*GradleRelease, error)

	// GetVersion returns a specific Gradle version
	GetVersion(ctx context.Context, version string) (*GradleRelease, error)

	// Download downloads a Gradle release to the specified path
	Download(ctx context.Context, release GradleRelease, dest string, progress ProgressFunc) error

	// GetChecksum returns the expected checksum for a release
	GetChecksum(ctx context.Context, release GradleRelease) (string, error)
}

// GradleRelease represents a specific Gradle release
type GradleRelease struct {
	Version     string // e.g., "8.5", "7.6.3"
	URL         string // Download URL for the distribution
	Checksum    string // SHA256 hash (may be fetched separately)
	ArchiveType string // Always "zip" for Gradle
}

// ProgressFunc is a callback for download progress
type ProgressFunc func(downloaded, total int64)

// ProgressInfo contains progress details
type ProgressInfo struct {
	Downloaded int64
	Total      int64
	Percent    float64
	Speed      float64 // bytes per second
	ETA        int64   // seconds
}
