package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/jem/internal/config"
	"github.com/user/jem/internal/downloader"
	"github.com/user/jem/internal/jdk"
	"github.com/user/jem/internal/platform"
	"github.com/user/jem/internal/provider"
)

// InstallCommand handles the 'jem install' command
type InstallCommand struct {
	platform       platform.Platform
	configRepo     config.ConfigRepository
	jdkService     *jdk.JDKService
	jdkProvider    provider.JDKProvider
	gradleProvider provider.GradleProvider
	downloader     *downloader.Downloader
	extractor      *downloader.Extractor
	force          bool
	onlyLTS        bool
	majorVersion   int
}

// NewInstallCommand creates a new InstallCommand
func NewInstallCommand(platform platform.Platform, configRepo config.ConfigRepository, jdkService *jdk.JDKService) *InstallCommand {
	return &InstallCommand{
		platform:   platform,
		configRepo: configRepo,
		jdkService: jdkService,
		downloader: downloader.NewDownloader(),
		extractor:  &downloader.Extractor{},
	}
}

// ExecuteJDK runs the JDK install command
func (c *InstallCommand) ExecuteJDK(ctx context.Context, version string) error {
	// Determine which provider to use
	providerName := c.configRepo.GetDefaultProvider()
	if providerName == "" {
		providerName = "temurin"
	}

	// Create the appropriate provider
	switch providerName {
	case "temurin":
		c.jdkProvider = provider.NewTemurinProvider()
	default:
		return fmt.Errorf("unknown provider: %s", providerName)
	}

	// Determine the version to install
	var release *provider.JDKRelease
	var err error

	if c.onlyLTS {
		release, err = c.jdkProvider.GetLatestLTS(ctx)
		if err != nil {
			return fmt.Errorf("failed to get latest LTS: %w", err)
		}
	} else if c.majorVersion > 0 {
		release, err = c.jdkProvider.GetLatest(ctx, c.majorVersion)
		if err != nil {
			return fmt.Errorf("failed to get latest version: %w", err)
		}
	} else {
		// Parse version string to determine major version
		// For now, try to get latest for the specified version
		release, err = c.findRelease(ctx, version)
		if err != nil {
			return err
		}
	}

	// Download the JDK
	if err := c.downloadJDK(ctx, release); err != nil {
		return err
	}

	// Extract the JDK
	if err := c.extractJDK(ctx, release); err != nil {
		return err
	}

	// Register the JDK in config
	if err := c.registerJDK(ctx, release); err != nil {
		return err
	}

	fmt.Printf("✓ Successfully installed JDK %s\n", release.Version)

	// Optionally set as current
	if !c.force {
		// Prompt user to set as current (TODO: implement interactive prompt)
		// For now, auto-set
		if err := c.setAsCurrent(ctx, release.Version); err != nil {
			return fmt.Errorf("failed to set as current: %w", err)
		}
	}

	return nil
}

// ExecuteGradle runs the Gradle install command
func (c *InstallCommand) ExecuteGradle(ctx context.Context, version string) error {
	// Use the default Gradle provider (currently only "gradle" is supported)
	providerName := "gradle"

	// Create the appropriate provider
	switch providerName {
	case "gradle":
		c.gradleProvider = provider.NewGradleProvider()
	default:
		return fmt.Errorf("unknown Gradle provider: %s", providerName)
	}

	// Determine the version to install
	var release *provider.GradleRelease
	var err error

	if version == "latest" {
		release, err = c.gradleProvider.GetLatest(ctx)
		if err != nil {
			return fmt.Errorf("failed to get latest Gradle: %w", err)
		}
	} else {
		release, err = c.gradleProvider.GetVersion(ctx, version)
		if err != nil {
			return err
		}
	}

	// Download the Gradle
	if err := c.downloadGradle(ctx, release); err != nil {
		return err
	}

	// Extract the Gradle
	if err := c.extractGradle(ctx, release); err != nil {
		return err
	}

	// Register the Gradle in config
	if err := c.registerGradle(ctx, release); err != nil {
		return err
	}

	fmt.Printf("✓ Successfully installed Gradle %s\n", release.Version)

	// Optionally set as current
	if !c.force {
		if err := c.setGradleAsCurrent(ctx, release.Version); err != nil {
			return fmt.Errorf("failed to set Gradle as current: %w", err)
		}
	}

	return nil
}

// findRelease finds a release matching the version string
func (c *InstallCommand) findRelease(ctx context.Context, version string) (*provider.JDKRelease, error) {
	// Parse major version from version string
	major := c.parseMajorVersion(version)

	if major > 0 {
		return c.jdkProvider.GetLatest(ctx, major)
	}

	// Try to find exact version
	releases, err := c.jdkProvider.ListAvailable(ctx, provider.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, r := range releases {
		if r.Version == version {
			return &r, nil
		}
	}

	return nil, fmt.Errorf("version '%s' not found", version)
}

// parseMajorVersion extracts the major version number from a version string
func (c *InstallCommand) parseMajorVersion(version string) int {
	// Simple parsing - extract first number
	for i := 0; i < len(version); i++ {
		if version[i] >= '0' && version[i] <= '9' {
			start := i
			for i < len(version) && (version[i] >= '0' && version[i] <= '9') {
				i++
			}
			// Parse the number
			var major int
			fmt.Sscanf(version[start:i], "%d", &major)
			return major
		}
	}
	return 0
}

// downloadJDK downloads the JDK
func (c *InstallCommand) downloadJDK(ctx context.Context, release *provider.JDKRelease) error {
	homeDir := c.platform.HomeDir()
	jemDir := filepath.Join(homeDir, ".jem")
	jdkDir := filepath.Join(jemDir, "jdks")

	// Create JDK directory if needed
	if err := c.platform.CreateLink(jdkDir, jdkDir); err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create JDK directory: %w", err)
	}

	// Determine archive filename
	archiveName := filepath.Base(release.URL)
	archivePath := filepath.Join(jdkDir, archiveName)

	// Download with progress
	fmt.Printf("Downloading JDK %s...\n", release.Version)
	fmt.Printf("URL: %s\n", release.URL)

	if err := c.downloader.DownloadWithChecksum(ctx, release.URL, archivePath, release.Checksum, c.progressCallback); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	return nil
}

// downloadGradle downloads the Gradle
func (c *InstallCommand) downloadGradle(ctx context.Context, release *provider.GradleRelease) error {
	homeDir := c.platform.HomeDir()
	jemDir := filepath.Join(homeDir, ".jem")
	gradleDir := filepath.Join(jemDir, "gradles")

	// Create Gradle directory if needed
	if err := c.platform.CreateLink(gradleDir, gradleDir); err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create Gradle directory: %w", err)
	}

	// Determine archive filename
	archiveName := filepath.Base(release.URL)
	archivePath := filepath.Join(gradleDir, archiveName)

	// Download with progress
	fmt.Printf("Downloading Gradle %s...\n", release.Version)
	fmt.Printf("URL: %s\n", release.URL)

	if err := c.downloader.DownloadWithChecksum(ctx, release.URL, archivePath, release.Checksum, c.progressCallback); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	return nil
}

// extractJDK extracts the downloaded archive
func (c *InstallCommand) extractJDK(ctx context.Context, release *provider.JDKRelease) error {
	homeDir := c.platform.HomeDir()
	jemDir := filepath.Join(homeDir, ".jem")
	jdkDir := filepath.Join(jemDir, "jdks")

	// Determine archive filename
	archiveName := filepath.Base(release.URL)
	archivePath := filepath.Join(jdkDir, archiveName)

	// Extract directory name
	extractDir := filepath.Join(jdkDir, release.Version)

	fmt.Printf("Extracting JDK to %s...\n", extractDir)

	if err := c.extractor.Extract(archivePath, extractDir); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	// Remove archive after extraction
	os.Remove(archivePath)

	return nil
}

// extractGradle extracts the downloaded archive
func (c *InstallCommand) extractGradle(ctx context.Context, release *provider.GradleRelease) error {
	homeDir := c.platform.HomeDir()
	jemDir := filepath.Join(homeDir, ".jem")
	gradleDir := filepath.Join(jemDir, "gradles")

	// Determine archive filename
	archiveName := filepath.Base(release.URL)
	archivePath := filepath.Join(gradleDir, archiveName)

	// Extract directory name
	extractDir := filepath.Join(gradleDir, release.Version)

	fmt.Printf("Extracting Gradle to %s...\n", extractDir)

	if err := c.extractor.Extract(archivePath, extractDir); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	// Remove archive after extraction
	os.Remove(archivePath)

	return nil
}

// registerJDK registers the installed JDK in the config
func (c *InstallCommand) registerJDK(ctx context.Context, release *provider.JDKRelease) error {
	homeDir := c.platform.HomeDir()
	jemDir := filepath.Join(homeDir, ".jem")
	jdkPath := filepath.Join(jemDir, "jdks", release.Version)

	// Create JDK info
	jdkInfo := config.JDKInfo{
		Path:     jdkPath,
		Version:  release.Version,
		Provider: c.jdkProvider.Name(),
		Managed:  true,
	}

	// Add to installed JDKs
	if err := c.configRepo.AddInstalledJDK(jdkInfo); err != nil {
		return fmt.Errorf("failed to register JDK: %w", err)
	}

	return nil
}

// registerGradle registers the installed Gradle in the config
func (c *InstallCommand) registerGradle(ctx context.Context, release *provider.GradleRelease) error {
	homeDir := c.platform.HomeDir()
	jemDir := filepath.Join(homeDir, ".jem")
	gradlePath := filepath.Join(jemDir, "gradles", release.Version)

	// Create Gradle info
	gradleInfo := config.GradleInfo{
		Path:    gradlePath,
		Version: release.Version,
		Managed: true,
	}

	// Add to installed Gradles
	if err := c.configRepo.AddInstalledGradle(gradleInfo); err != nil {
		return fmt.Errorf("failed to register Gradle: %w", err)
	}

	return nil
}

// setAsCurrent sets the installed JDK as the current version
func (c *InstallCommand) setAsCurrent(ctx context.Context, version string) error {
	if err := c.configRepo.SetJDKCurrent(version); err != nil {
		return fmt.Errorf("failed to set current JDK: %w", err)
	}

	// Update symlinks
	jdkPath := filepath.Join(c.platform.HomeDir(), ".jem", "jdks", version)
	if err := c.jdkService.GetJDKSymlinker().UpdateLinks(jdkPath); err != nil {
		return fmt.Errorf("failed to update symlinks: %w", err)
	}

	fmt.Printf("✓ Set JDK %s as current\n", version)
	return nil
}

// setGradleAsCurrent sets the installed Gradle as the current version
func (c *InstallCommand) setGradleAsCurrent(ctx context.Context, version string) error {
	if err := c.configRepo.SetGradleCurrent(version); err != nil {
		return fmt.Errorf("failed to set current Gradle: %w", err)
	}

	// Update symlink
	gradlePath := filepath.Join(c.platform.HomeDir(), ".jem", "gradles", version)
	currentLink := filepath.Join(filepath.Dir(gradlePath), "current")

	// Remove existing symlink if it exists
	if _, err := os.Lstat(currentLink); err == nil {
		if err := os.Remove(currentLink); err != nil {
			return fmt.Errorf("failed to remove existing current symlink: %w", err)
		}
	}

	if err := c.platform.CreateLink(gradlePath, currentLink); err != nil {
		return fmt.Errorf("failed to create current symlink: %w", err)
	}

	fmt.Printf("✓ Set Gradle %s as current\n", version)
	return nil
}

// progressCallback is the download progress callback
func (c *InstallCommand) progressCallback(downloaded, total int64) {
	// Calculate percentage
	var percent float64
	if total > 0 {
		percent = float64(downloaded) / float64(total) * 100
	}

	// Calculate speed and ETA
	// For now, just show progress
	fmt.Printf("\rDownloading: %.1f%% (%.1f MB / %.1f MB)",
		percent,
		float64(downloaded)/1024/1024,
		float64(total)/1024/1024)

	if downloaded >= total && total > 0 {
		fmt.Println() // New line when complete
	}
}

// SetForce sets the force flag
func (c *InstallCommand) SetForce(force bool) {
	c.force = force
}

// SetOnlyLTS sets the onlyLTS flag
func (c *InstallCommand) SetOnlyLTS(onlyLTS bool) {
	c.onlyLTS = onlyLTS
}

// SetMajorVersion sets the major version
func (c *InstallCommand) SetMajorVersion(major int) {
	c.majorVersion = major
}
