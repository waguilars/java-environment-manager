package downloader

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestVerifySHA256_Success(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	expectedChecksum := "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"

	// Write test content that produces the expected checksum
	content := []byte("abc")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	err := VerifySHA256(filePath, expectedChecksum)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestVerifySHA256_Failure(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	expectedChecksum := "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"

	// Write different content
	content := []byte("xyz")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	err := VerifySHA256(filePath, expectedChecksum)
	if err == nil {
		t.Error("Expected checksum mismatch error")
	}
}

func TestVerifySHA256_FileNotFound(t *testing.T) {
	err := VerifySHA256("/non/existent/file.txt", "abc123")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestComputeSHA256(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")

	// Write test content
	content := []byte("abc")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Compute expected checksum
	hash := sha256.Sum256(content)
	expectedChecksum := hex.EncodeToString(hash[:])

	result, err := ComputeSHA256(filePath)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result != expectedChecksum {
		t.Errorf("Expected checksum '%s', got '%s'", expectedChecksum, result)
	}
}

func TestComputeSHA256_FileNotFound(t *testing.T) {
	_, err := ComputeSHA256("/non/existent/file.txt")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestVerifyMD5_Success(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	expectedChecksum := "900150983cd24fb0d6963f7d28e17f72" // MD5 of "abc"

	content := []byte("abc")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	err := VerifyMD5(filePath, expectedChecksum)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestVerifyMD5_Failure(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	expectedChecksum := "900150983cd24fb0d6963f7d28e17f72" // MD5 of "abc"

	content := []byte("xyz")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	err := VerifyMD5(filePath, expectedChecksum)
	if err == nil {
		t.Error("Expected checksum mismatch error")
	}
}

func TestVerifyChecksum_SHA256(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	expectedChecksum := "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"

	content := []byte("abc")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	err := VerifyChecksum(filePath, expectedChecksum, "sha256")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestVerifyChecksum_MD5(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	expectedChecksum := "900150983cd24fb0d6963f7d28e17f72"

	content := []byte("abc")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	err := VerifyChecksum(filePath, expectedChecksum, "md5")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestVerifyChecksum_UnsupportedAlgorithm(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")

	err := VerifyChecksum(filePath, "abc123", "unsupported")
	if err == nil {
		t.Error("Expected error for unsupported algorithm")
	}
}

func TestExtractor_GetFormat_ZIP(t *testing.T) {
	e := &Extractor{}
	format := e.GetFormat("archive.zip")
	if format != FormatZIP {
		t.Errorf("Expected FormatZIP, got %v", format)
	}
}

func TestExtractor_GetFormat_TarGZ(t *testing.T) {
	e := &Extractor{}
	format := e.GetFormat("archive.tar.gz")
	if format != FormatTarGZ {
		t.Errorf("Expected FormatTarGZ, got %v", format)
	}
}

func TestExtractor_GetFormat_GZ(t *testing.T) {
	e := &Extractor{}
	format := e.GetFormat("archive.gz")
	if format != FormatUnknown {
		t.Errorf("Expected FormatUnknown for .gz, got %v", format)
	}
}

func TestExtractor_GetFormat_Unknown(t *testing.T) {
	e := &Extractor{}
	format := e.GetFormat("archive.tar")
	if format != FormatUnknown {
		t.Errorf("Expected FormatUnknown, got %v", format)
	}
}

func TestExtractor_GetFormat_CaseInsensitive(t *testing.T) {
	e := &Extractor{}
	format := e.GetFormat("ARCHIVE.ZIP")
	if format != FormatZIP {
		t.Errorf("Expected FormatZIP, got %v", format)
	}
}

func TestArchiveFormat_String(t *testing.T) {
	tests := []struct {
		format   ArchiveFormat
		expected string
	}{
		{FormatZIP, "zip"},
		{FormatTarGZ, "tar.gz"},
		{FormatUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.format.String()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestNewDownloadStatistics(t *testing.T) {
	stats := NewDownloadStatistics()
	if stats == nil {
		t.Error("Expected non-nil stats")
	}
}

func TestDownloadStatistics_RecordDownload(t *testing.T) {
	stats := NewDownloadStatistics()
	stats.RecordDownload(1024)

	s := stats.GetStatistics()
	if s.TotalDownloads != 1 {
		t.Errorf("Expected 1 download, got %d", s.TotalDownloads)
	}
	if s.TotalBytes != 1024 {
		t.Errorf("Expected 1024 bytes, got %d", s.TotalBytes)
	}
}

func TestDownloadStatistics_RecordFailure(t *testing.T) {
	stats := NewDownloadStatistics()
	stats.RecordFailure(512)

	s := stats.GetStatistics()
	if s.FailedBytes != 512 {
		t.Errorf("Expected 512 failed bytes, got %d", s.FailedBytes)
	}
}

func TestDownloadStatistics_RecordDuration(t *testing.T) {
	stats := NewDownloadStatistics()
	stats.RecordDuration(100 * time.Millisecond)

	s := stats.GetStatistics()
	if s.AvgDuration == 0 {
		t.Error("Expected non-zero duration")
	}
}

func TestGetDownloadProgress(t *testing.T) {
	// Test 50% progress
	info := GetDownloadProgress(500, 1000)
	if info.Percent != 50 {
		t.Errorf("Expected 50%%, got %f", info.Percent)
	}
	if info.Downloaded != 500 {
		t.Errorf("Expected 500 downloaded, got %d", info.Downloaded)
	}
	if info.Total != 1000 {
		t.Errorf("Expected 1000 total, got %d", info.Total)
	}
}

func TestGetDownloadProgress_ZeroTotal(t *testing.T) {
	info := GetDownloadProgress(500, 0)
	if info.Percent != 0 {
		t.Errorf("Expected 0%% when total is 0, got %f", info.Percent)
	}
}

func TestGetDownloadProgress_Full(t *testing.T) {
	info := GetDownloadProgress(1000, 1000)
	if info.Percent != 100 {
		t.Errorf("Expected 100%%, got %f", info.Percent)
	}
}
