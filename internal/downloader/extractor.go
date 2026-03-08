package downloader

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Extractor handles archive extraction
type Extractor struct{}

// Extract extracts an archive to the destination directory
func (e *Extractor) Extract(archivePath, destDir string) error {
	format := e.GetFormat(archivePath)

	switch format {
	case FormatZIP:
		return e.ExtractZIP(archivePath, destDir)
	case FormatTarGZ:
		return e.ExtractTarGZ(archivePath, destDir)
	default:
		return fmt.Errorf("unsupported archive format: %s", format)
	}
}

// ExtractZIP extracts a ZIP archive
func (e *Extractor) ExtractZIP(archivePath, destDir string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer reader.Close()

	// Ensure destination directory exists
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Extract each file
	for _, file := range reader.File {
		// Sanitize file path to prevent directory traversal
		cleanPath := filepath.Clean(file.Name)
		if strings.HasPrefix(cleanPath, ".."+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", file.Name)
		}

		destPath := filepath.Join(destDir, cleanPath)

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Handle directories
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		// Extract file
		srcFile, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in archive: %w", err)
		}
		defer srcFile.Close()

		destFile, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer destFile.Close()

		if _, err := io.Copy(destFile, srcFile); err != nil {
			return fmt.Errorf("failed to extract file: %w", err)
		}

		// Preserve file permissions
		if err := os.Chmod(destPath, file.Mode()); err != nil {
			return fmt.Errorf("failed to set file permissions: %w", err)
		}
	}

	return nil
}

// ExtractTarGZ extracts a tar.gz archive
func (e *Extractor) ExtractTarGZ(archivePath, destDir string) error {
	// Open the gzipped file
	gzFile, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open tar.gz file: %w", err)
	}
	defer gzFile.Close()

	// Create gzip reader
	gzipReader, err := gzip.NewReader(gzFile)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzipReader)

	// Ensure destination directory exists
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Extract each file
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// Sanitize file path
		cleanPath := filepath.Clean(header.Name)
		if strings.HasPrefix(cleanPath, ".."+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", header.Name)
		}

		destPath := filepath.Join(destDir, cleanPath)

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(destPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			// Create parent directories
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}

			// Create file
			destFile, err := os.Create(destPath)
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}

			// Copy file contents
			if _, err := io.Copy(destFile, tarReader); err != nil {
				destFile.Close()
				return fmt.Errorf("failed to extract file: %w", err)
			}
			destFile.Close()

			// Preserve file permissions
			if err := os.Chmod(destPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to set file permissions: %w", err)
			}
		}
	}

	return nil
}

// GetFormat returns the archive format based on file extension
func (e *Extractor) GetFormat(path string) ArchiveFormat {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".zip":
		return FormatZIP
	case ".gz":
		// Check if it's .tar.gz
		base := strings.TrimSuffix(path, ext)
		if strings.HasSuffix(base, ".tar") {
			return FormatTarGZ
		}
		return FormatUnknown
	default:
		return FormatUnknown
	}
}

// ArchiveFormat represents the archive format
type ArchiveFormat int

const (
	FormatUnknown ArchiveFormat = iota
	FormatZIP
	FormatTarGZ
)

// String returns the string representation of the format
func (f ArchiveFormat) String() string {
	switch f {
	case FormatZIP:
		return "zip"
	case FormatTarGZ:
		return "tar.gz"
	default:
		return "unknown"
	}
}
