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
