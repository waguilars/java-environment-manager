package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/waguilars/java-environment-manager/internal/provider"
)

// Downloader handles HTTP downloads with progress tracking
type Downloader struct {
	httpClient *http.Client
	maxRetries int
	retryDelay time.Duration
	proxyURL   *url.URL
	timeout    time.Duration
	cacheDir   string
	statistics *DownloadStatistics
}

// NewDownloader creates a new Downloader with default settings
func NewDownloader() *Downloader {
	return &Downloader{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		maxRetries: 3,
		retryDelay: 1 * time.Second,
		timeout:    30 * time.Second,
		cacheDir:   "",
		statistics: NewDownloadStatistics(),
	}
}

// Download downloads a file from the given URL to the destination path
func (d *Downloader) Download(ctx context.Context, url, dest string, progress provider.ProgressFunc) error {
	return d.DownloadWithRetries(ctx, url, dest, progress, d.maxRetries)
}

// DownloadWithRetries downloads a file with retry logic
func (d *Downloader) DownloadWithRetries(ctx context.Context, url, dest string, progress provider.ProgressFunc, maxRetries int) error {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(d.retryDelay * time.Duration(attempt-1)):
			}
		}

		lastErr = d.downloadSingleAttempt(ctx, url, dest, progress)
		if lastErr == nil {
			return nil
		}

		// Only retry for certain error types
		if !d.shouldRetry(lastErr) {
			break
		}
	}

	return fmt.Errorf("download failed after %d attempts: %w", maxRetries, lastErr)
}

// downloadSingleAttempt performs a single download attempt
func (d *Downloader) downloadSingleAttempt(ctx context.Context, url, dest string, progress provider.ProgressFunc) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set proxy if configured
	if d.proxyURL != nil {
		transport := &http.Transport{
			Proxy: http.ProxyURL(d.proxyURL),
		}
		d.httpClient.Transport = transport
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to initiate download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	// Create destination directory if needed
	destDir := filepath.Dir(dest)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Create the output file
	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	// Track download progress
	var downloaded int64
	total := resp.ContentLength

	// Create progress wrapper
	wrapper := &progressWriter{
		writer: out,
		onProgress: func(n int64) {
			downloaded = n
			if progress != nil {
				progress(downloaded, total)
			}
			d.statistics.RecordDownload(n)
		},
	}

	// Copy with progress tracking
	if _, err := io.Copy(wrapper, resp.Body); err != nil {
		// Clean up partial download on error
		out.Close()
		os.Remove(dest)
		return fmt.Errorf("download failed: %w", err)
	}

	return nil
}

// DownloadWithChecksum downloads and verifies checksum
func (d *Downloader) DownloadWithChecksum(ctx context.Context, url, dest, expectedChecksum string, progress provider.ProgressFunc) error {
	// Download to temporary file first
	tmpDest := dest + ".part"

	if err := d.DownloadWithRetries(ctx, url, tmpDest, progress, d.maxRetries); err != nil {
		return err
	}

	// Verify checksum
	if err := VerifySHA256(tmpDest, expectedChecksum); err != nil {
		os.Remove(tmpDest)
		return fmt.Errorf("checksum verification failed: %w", err)
	}

	// Move to final destination
	if err := os.Rename(tmpDest, dest); err != nil {
		return fmt.Errorf("failed to move downloaded file: %w", err)
	}

	return nil
}

// shouldRetry determines if an error is retryable
func (d *Downloader) shouldRetry(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	// Retryable errors
	retryable := []string{
		"timeout",
		"connection reset",
		"connection refused",
		"network is unreachable",
		"no route to host",
		"broken pipe",
		"unexpected eof",
		"context deadline exceeded",
		"context canceled",
	}

	for _, msg := range retryable {
		if strings.Contains(errStr, msg) {
			return true
		}
	}

	return false
}

// SetProxy configures the proxy for downloads
func (d *Downloader) SetProxy(proxyURL string) error {
	if proxyURL == "" {
		d.proxyURL = nil
		return nil
	}

	parsed, err := url.Parse(proxyURL)
	if err != nil {
		return fmt.Errorf("invalid proxy URL: %w", err)
	}

	d.proxyURL = parsed
	return nil
}

// SetTimeout configures the download timeout
func (d *Downloader) SetTimeout(timeout time.Duration) {
	d.httpClient.Timeout = timeout
	d.timeout = timeout
}

// SetCacheDir configures the download cache directory
func (d *Downloader) SetCacheDir(dir string) {
	d.cacheDir = dir
}

// GetStatistics returns download statistics
func (d *Downloader) GetStatistics() *DownloadStatistics {
	return d.statistics
}

// progressWriter wraps an io.Writer to track progress
type progressWriter struct {
	writer     io.Writer
	onProgress func(int64)
	written    int64
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	pw.written += int64(n)
	pw.onProgress(pw.written)
	return n, err
}
