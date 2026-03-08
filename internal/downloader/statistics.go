package downloader

import (
	"sync"
	"time"

	"github.com/user/jem/internal/provider"
)

// DownloadStatistics tracks download statistics
type DownloadStatistics struct {
	mu           sync.Mutex
	downloads    int64
	bytesTotal   int64
	bytesFailed  int64
	durations    []time.Duration
	startTime    time.Time
	lastActivity time.Time
}

// NewDownloadStatistics creates a new DownloadStatistics instance
func NewDownloadStatistics() *DownloadStatistics {
	return &DownloadStatistics{
		startTime: time.Now(),
	}
}

// RecordDownload records a successful download
func (s *DownloadStatistics) RecordDownload(bytes int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.downloads++
	s.bytesTotal += bytes
	s.lastActivity = time.Now()
}

// RecordFailure records a failed download attempt
func (s *DownloadStatistics) RecordFailure(bytes int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.bytesFailed += bytes
	s.lastActivity = time.Now()
}

// RecordDuration records a download duration
func (s *DownloadStatistics) RecordDuration(duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.durations = append(s.durations, duration)
}

// GetStatistics returns current statistics
func (s *DownloadStatistics) GetStatistics() Stats {
	s.mu.Lock()
	defer s.mu.Unlock()

	var avgDuration time.Duration
	if len(s.durations) > 0 {
		var total time.Duration
		for _, d := range s.durations {
			total += d
		}
		avgDuration = total / time.Duration(len(s.durations))
	}

	var avgSpeed float64
	if avgDuration > 0 {
		if s.bytesTotal > 0 {
			avgSpeed = float64(s.bytesTotal) / avgDuration.Seconds()
		}
	}

	return Stats{
		TotalDownloads: s.downloads,
		TotalBytes:     s.bytesTotal,
		FailedBytes:    s.bytesFailed,
		AverageSpeed:   avgSpeed,
		AvgDuration:    avgDuration,
		Uptime:         time.Since(s.startTime),
		LastActivity:   s.lastActivity,
	}
}

// Stats contains download statistics
type Stats struct {
	TotalDownloads int64
	TotalBytes     int64
	FailedBytes    int64
	AverageSpeed   float64
	AvgDuration    time.Duration
	Uptime         time.Duration
	LastActivity   time.Time
}

// GetDownloadProgress calculates current progress
func GetDownloadProgress(downloaded, total int64) provider.ProgressInfo {
	var percent float64
	if total > 0 {
		percent = float64(downloaded) / float64(total) * 100
	}

	return provider.ProgressInfo{
		Downloaded: downloaded,
		Total:      total,
		Percent:    percent,
	}
}
