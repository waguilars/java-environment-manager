package jdk

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJDKSymlinker_UpdateCurrentLink(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a fake JDK directory
	jdkPath := filepath.Join(tmpDir, "temurin-21")
	if err := os.MkdirAll(jdkPath, 0755); err != nil {
		t.Fatalf("Failed to create JDK directory: %v", err)
	}

	// Create the .jem/jdks directory
	jdksDir := filepath.Join(tmpDir, ".jem", "jdks")
	if err := os.MkdirAll(jdksDir, 0755); err != nil {
		t.Fatalf("Failed to create jdks directory: %v", err)
	}

	// Create mock platform
	platform := &MockPlatform{
		HomeDirFunc: func() string {
			return tmpDir
		},
		CreateLinkFunc: func(target, link string) error {
			// Simulate creating a symlink
			return os.Symlink(target, link)
		},
		IsLinkFunc: func(path string) bool {
			info, err := os.Lstat(path)
			if err != nil {
				return false
			}
			return info.Mode()&os.ModeSymlink != 0
		},
		RemoveLinkFunc: func(link string) error {
			return os.Remove(link)
		},
	}

	symlinker := NewJDKSymlinker(platform)

	err := symlinker.UpdateCurrentLink("temurin-21", jdkPath)

	if err != nil {
		t.Fatalf("UpdateCurrentLink() should not error: %v", err)
	}

	// Verify the symlink was created
	currentLink := filepath.Join(tmpDir, ".jem", "jdks", "current")
	if _, err := os.Stat(currentLink); os.IsNotExist(err) {
		t.Error("Symlink was not created")
	}
}

func TestJDKSymlinker_RemoveCurrentLink(t *testing.T) {
	tmpDir := t.TempDir()

	// Create the current symlink first
	currentLink := filepath.Join(tmpDir, ".jem", "jdks", "current")
	jdkPath := filepath.Join(tmpDir, "temurin-21")
	if err := os.MkdirAll(filepath.Dir(currentLink), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.Symlink(jdkPath, currentLink); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// Create mock platform
	platform := &MockPlatform{
		HomeDirFunc: func() string {
			return tmpDir
		},
		IsLinkFunc: func(path string) bool {
			info, err := os.Lstat(path)
			if err != nil {
				return false
			}
			return info.Mode()&os.ModeSymlink != 0
		},
		RemoveLinkFunc: func(link string) error {
			return os.Remove(link)
		},
	}

	symlinker := NewJDKSymlinker(platform)

	err := symlinker.RemoveCurrentLink()

	if err != nil {
		t.Fatalf("RemoveCurrentLink() should not error: %v", err)
	}

	// Verify the symlink was removed
	if _, err := os.Stat(currentLink); err == nil {
		t.Error("Symlink still exists after removal")
	}
}

func TestJDKSymlinker_UpdateBinLinks(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a fake JDK with bin directory
	jdkPath := filepath.Join(tmpDir, "temurin-21")
	binPath := filepath.Join(jdkPath, "bin")
	if err := os.MkdirAll(binPath, 0755); err != nil {
		t.Fatalf("Failed to create bin directory: %v", err)
	}

	// Create fake binary files
	binaries := []string{"java", "javac", "jar"}
	for _, binary := range binaries {
		binaryPath := filepath.Join(binPath, binary)
		if err := os.WriteFile(binaryPath, []byte("#!/bin/sh"), 0755); err != nil {
			t.Fatalf("Failed to create binary: %v", err)
		}
	}

	// Create mock platform
	platform := &MockPlatform{
		HomeDirFunc: func() string {
			return tmpDir
		},
		CreateLinkFunc: func(target, link string) error {
			return os.Symlink(target, link)
		},
		IsLinkFunc: func(path string) bool {
			info, err := os.Lstat(path)
			if err != nil {
				return false
			}
			return info.Mode()&os.ModeSymlink != 0
		},
		RemoveLinkFunc: func(link string) error {
			return os.Remove(link)
		},
	}

	symlinker := NewJDKSymlinker(platform)

	err := symlinker.UpdateBinLinks(jdkPath)

	if err != nil {
		t.Fatalf("UpdateBinLinks() should not error: %v", err)
	}

	// Verify symlinks were created
	binDir := filepath.Join(tmpDir, ".jem", "bin")
	for _, binary := range binaries {
		link := filepath.Join(binDir, binary)
		if _, err := os.Stat(link); os.IsNotExist(err) {
			t.Errorf("Binary symlink %s was not created", binary)
		}
	}
}

func TestJDKSymlinker_RemoveBinLinks(t *testing.T) {
	tmpDir := t.TempDir()

	// Create bin directory with symlinks
	binDir := filepath.Join(tmpDir, ".jem", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("Failed to create bin directory: %v", err)
	}

	binaries := []string{"java", "javac"}
	jdkPath := filepath.Join(tmpDir, "temurin-21", "bin")
	for _, binary := range binaries {
		link := filepath.Join(binDir, binary)
		source := filepath.Join(jdkPath, binary)
		if err := os.MkdirAll(filepath.Dir(source), 0755); err != nil {
			t.Fatalf("Failed to create jdk bin: %v", err)
		}
		if err := os.WriteFile(source, []byte("#!/bin/sh"), 0755); err != nil {
			t.Fatalf("Failed to create source: %v", err)
		}
		if err := os.Symlink(source, link); err != nil {
			t.Fatalf("Failed to create symlink: %v", err)
		}
	}

	// Create mock platform
	platform := &MockPlatform{
		HomeDirFunc: func() string {
			return tmpDir
		},
		IsLinkFunc: func(path string) bool {
			info, err := os.Lstat(path)
			if err != nil {
				return false
			}
			return info.Mode()&os.ModeSymlink != 0
		},
		RemoveLinkFunc: func(link string) error {
			return os.Remove(link)
		},
	}

	symlinker := NewJDKSymlinker(platform)

	err := symlinker.RemoveBinLinks()

	if err != nil {
		t.Fatalf("RemoveBinLinks() should not error: %v", err)
	}

	// Verify symlinks were removed
	for _, binary := range binaries {
		link := filepath.Join(binDir, binary)
		if _, err := os.Stat(link); err == nil {
			t.Errorf("Binary symlink %s still exists", binary)
		}
	}
}

func TestJDKSymlinker_GetCurrentJDK(t *testing.T) {
	tmpDir := t.TempDir()

	// Create the current symlink
	currentLink := filepath.Join(tmpDir, ".jem", "jdks", "current")
	jdkPath := filepath.Join(tmpDir, "temurin-21")
	if err := os.MkdirAll(filepath.Dir(currentLink), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.Symlink(jdkPath, currentLink); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// Create mock platform
	platform := &MockPlatform{
		HomeDirFunc: func() string {
			return tmpDir
		},
		IsLinkFunc: func(path string) bool {
			info, err := os.Lstat(path)
			if err != nil {
				return false
			}
			return info.Mode()&os.ModeSymlink != 0
		},
	}

	symlinker := NewJDKSymlinker(platform)

	target, err := symlinker.GetCurrentJDK()

	if err != nil {
		t.Fatalf("GetCurrentJDK() should not error: %v", err)
	}

	if target != jdkPath {
		t.Errorf("Expected '%s', got '%s'", jdkPath, target)
	}
}

func TestJDKSymlinker_GetCurrentJDK_NoLink(t *testing.T) {
	tmpDir := t.TempDir()

	// Create mock platform
	platform := &MockPlatform{
		HomeDirFunc: func() string {
			return tmpDir
		},
		IsLinkFunc: func(path string) bool {
			return false
		},
	}

	symlinker := NewJDKSymlinker(platform)

	_, err := symlinker.GetCurrentJDK()

	if err == nil {
		t.Error("Expected error when no symlink exists")
	}
}
