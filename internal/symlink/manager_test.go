package symlink

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/waguilars/java-environment-manager/internal/platform"
)

// MockPlatformForSymlink creates a mock platform for symlink tests
type MockPlatformForSymlink struct {
	platform.LinuxPlatform
	homeDir string
}

func (m *MockPlatformForSymlink) HomeDir() string {
	if m.homeDir != "" {
		return m.homeDir
	}
	return "/tmp"
}

func TestNewSymlinkManager(t *testing.T) {
	mockPlatform := &MockPlatformForSymlink{}
	manager := NewSymlinkManager(mockPlatform)

	if manager == nil {
		t.Error("Expected NewSymlinkManager to return a non-nil manager")
	}

	if manager.platform != mockPlatform {
		t.Error("Expected platform to be set correctly")
	}
}

func TestUpdateCurrentJava(t *testing.T) {
	tmpDir := t.TempDir()
	mockPlatform := &MockPlatformForSymlink{homeDir: tmpDir}
	manager := NewSymlinkManager(mockPlatform)

	// Create the jem structure
	jemDir := filepath.Join(tmpDir, ".jem")
	jdkDir := filepath.Join(jemDir, "jdks", "21")
	currentDir := filepath.Join(jemDir, "current")

	os.MkdirAll(jdkDir, 0755)

	// Update current Java
	err := manager.UpdateCurrentJava("21", jdkDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify symlink was created
	javaLink := filepath.Join(currentDir, "java")
	if _, err := os.Lstat(javaLink); err != nil {
		t.Errorf("Expected java symlink to exist, got error: %v", err)
	}

	// Verify symlink points to correct target
	target, err := os.Readlink(javaLink)
	if err != nil {
		t.Errorf("Expected to read symlink target, got error: %v", err)
	}
	if target != jdkDir {
		t.Errorf("Expected symlink to point to %s, got %s", jdkDir, target)
	}
}

func TestUpdateCurrentJava_VersionNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	mockPlatform := &MockPlatformForSymlink{homeDir: tmpDir}
	manager := NewSymlinkManager(mockPlatform)

	// Try to update to non-existent version
	err := manager.UpdateCurrentJava("99", filepath.Join(tmpDir, ".jem", "jdks", "99"))
	if err == nil {
		t.Error("Expected error for non-existent version")
	}

	if !contains(err.Error(), "not found") {
		t.Errorf("Expected error message to contain 'not found', got: %v", err)
	}
}

func TestUpdateCurrentGradle(t *testing.T) {
	tmpDir := t.TempDir()
	mockPlatform := &MockPlatformForSymlink{homeDir: tmpDir}
	manager := NewSymlinkManager(mockPlatform)

	// Create the jem structure
	jemDir := filepath.Join(tmpDir, ".jem")
	gradleDir := filepath.Join(jemDir, "gradles", "8.5")
	currentDir := filepath.Join(jemDir, "current")

	os.MkdirAll(gradleDir, 0755)

	// Update current Gradle
	err := manager.UpdateCurrentGradle("8.5", gradleDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify symlink was created
	gradleLink := filepath.Join(currentDir, "gradle")
	if _, err := os.Lstat(gradleLink); err != nil {
		t.Errorf("Expected gradle symlink to exist, got error: %v", err)
	}

	// Verify symlink points to correct target
	target, err := os.Readlink(gradleLink)
	if err != nil {
		t.Errorf("Expected to read symlink target, got error: %v", err)
	}
	if target != gradleDir {
		t.Errorf("Expected symlink to point to %s, got %s", gradleDir, target)
	}
}

func TestUpdateCurrentGradle_VersionNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	mockPlatform := &MockPlatformForSymlink{homeDir: tmpDir}
	manager := NewSymlinkManager(mockPlatform)

	// Try to update to non-existent version
	err := manager.UpdateCurrentGradle("99.9", filepath.Join(tmpDir, ".jem", "gradles", "99.9"))
	if err == nil {
		t.Error("Expected error for non-existent version")
	}

	if !contains(err.Error(), "not found") {
		t.Errorf("Expected error message to contain 'not found', got: %v", err)
	}
}

func TestGetCurrentJava(t *testing.T) {
	tmpDir := t.TempDir()
	mockPlatform := &MockPlatformForSymlink{homeDir: tmpDir}
	manager := NewSymlinkManager(mockPlatform)

	// Create the jem structure
	jemDir := filepath.Join(tmpDir, ".jem")
	jdkDir := filepath.Join(jemDir, "jdks", "21")
	currentDir := filepath.Join(jemDir, "current")
	javaLink := filepath.Join(currentDir, "java")

	os.MkdirAll(jdkDir, 0755)
	os.MkdirAll(currentDir, 0755)
	os.Symlink(jdkDir, javaLink)

	// Get current Java
	current, err := manager.GetCurrentJava()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if current != jdkDir {
		t.Errorf("Expected current Java to be %s, got %s", jdkDir, current)
	}
}

func TestGetCurrentJava_NoSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	mockPlatform := &MockPlatformForSymlink{homeDir: tmpDir}
	manager := NewSymlinkManager(mockPlatform)

	// Get current Java without symlink
	_, err := manager.GetCurrentJava()
	if err == nil {
		t.Error("Expected error when no symlink exists")
	}

	if !contains(err.Error(), "no current Java") {
		t.Errorf("Expected error message to contain 'no current Java', got: %v", err)
	}
}

func TestGetCurrentGradle(t *testing.T) {
	tmpDir := t.TempDir()
	mockPlatform := &MockPlatformForSymlink{homeDir: tmpDir}
	manager := NewSymlinkManager(mockPlatform)

	// Create the jem structure
	jemDir := filepath.Join(tmpDir, ".jem")
	gradleDir := filepath.Join(jemDir, "gradles", "8.5")
	currentDir := filepath.Join(jemDir, "current")
	gradleLink := filepath.Join(currentDir, "gradle")

	os.MkdirAll(gradleDir, 0755)
	os.MkdirAll(currentDir, 0755)
	os.Symlink(gradleDir, gradleLink)

	// Get current Gradle
	current, err := manager.GetCurrentGradle()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if current != gradleDir {
		t.Errorf("Expected current Gradle to be %s, got %s", gradleDir, current)
	}
}

func TestGetCurrentGradle_NoSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	mockPlatform := &MockPlatformForSymlink{homeDir: tmpDir}
	manager := NewSymlinkManager(mockPlatform)

	// Get current Gradle without symlink
	_, err := manager.GetCurrentGradle()
	if err == nil {
		t.Error("Expected error when no symlink exists")
	}

	if !contains(err.Error(), "no current Gradle") {
		t.Errorf("Expected error message to contain 'no current Gradle', got: %v", err)
	}
}

func TestValidateVersionExists_Java(t *testing.T) {
	tmpDir := t.TempDir()
	mockPlatform := &MockPlatformForSymlink{homeDir: tmpDir}
	manager := NewSymlinkManager(mockPlatform)

	// Create JDK directory
	jdkDir := filepath.Join(tmpDir, ".jem", "jdks", "21")
	os.MkdirAll(jdkDir, 0755)

	// Validate existing version
	err := manager.ValidateVersionExists("java", "21")
	if err != nil {
		t.Errorf("Expected no error for existing version, got: %v", err)
	}

	// Validate non-existing version
	err = manager.ValidateVersionExists("java", "99")
	if err == nil {
		t.Error("Expected error for non-existing version")
	}
}

func TestValidateVersionExists_Gradle(t *testing.T) {
	tmpDir := t.TempDir()
	mockPlatform := &MockPlatformForSymlink{homeDir: tmpDir}
	manager := NewSymlinkManager(mockPlatform)

	// Create Gradle directory
	gradleDir := filepath.Join(tmpDir, ".jem", "gradles", "8.5")
	os.MkdirAll(gradleDir, 0755)

	// Validate existing version
	err := manager.ValidateVersionExists("gradle", "8.5")
	if err != nil {
		t.Errorf("Expected no error for existing version, got: %v", err)
	}

	// Validate non-existing version
	err = manager.ValidateVersionExists("gradle", "99.9")
	if err == nil {
		t.Error("Expected error for non-existing version")
	}
}

func TestValidateVersionExists_UnknownTool(t *testing.T) {
	tmpDir := t.TempDir()
	mockPlatform := &MockPlatformForSymlink{homeDir: tmpDir}
	manager := NewSymlinkManager(mockPlatform)

	// Validate unknown tool
	err := manager.ValidateVersionExists("unknown", "1.0")
	if err == nil {
		t.Error("Expected error for unknown tool")
	}

	if !contains(err.Error(), "unknown tool") {
		t.Errorf("Expected error message to contain 'unknown tool', got: %v", err)
	}
}

func TestRemoveCurrentJava(t *testing.T) {
	tmpDir := t.TempDir()
	mockPlatform := &MockPlatformForSymlink{homeDir: tmpDir}
	manager := NewSymlinkManager(mockPlatform)

	// Create the jem structure with symlink
	jemDir := filepath.Join(tmpDir, ".jem")
	jdkDir := filepath.Join(jemDir, "jdks", "21")
	currentDir := filepath.Join(jemDir, "current")
	javaLink := filepath.Join(currentDir, "java")

	os.MkdirAll(jdkDir, 0755)
	os.MkdirAll(currentDir, 0755)
	os.Symlink(jdkDir, javaLink)

	// Remove current Java
	err := manager.RemoveCurrentJava()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify symlink was removed
	if _, err := os.Lstat(javaLink); err == nil {
		t.Error("Expected java symlink to be removed")
	}
}

func TestRemoveCurrentGradle(t *testing.T) {
	tmpDir := t.TempDir()
	mockPlatform := &MockPlatformForSymlink{homeDir: tmpDir}
	manager := NewSymlinkManager(mockPlatform)

	// Create the jem structure with symlink
	jemDir := filepath.Join(tmpDir, ".jem")
	gradleDir := filepath.Join(jemDir, "gradles", "8.5")
	currentDir := filepath.Join(jemDir, "current")
	gradleLink := filepath.Join(currentDir, "gradle")

	os.MkdirAll(gradleDir, 0755)
	os.MkdirAll(currentDir, 0755)
	os.Symlink(gradleDir, gradleLink)

	// Remove current Gradle
	err := manager.RemoveCurrentGradle()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify symlink was removed
	if _, err := os.Lstat(gradleLink); err == nil {
		t.Error("Expected gradle symlink to be removed")
	}
}

func TestHasCurrentJava(t *testing.T) {
	tmpDir := t.TempDir()
	mockPlatform := &MockPlatformForSymlink{homeDir: tmpDir}
	manager := NewSymlinkManager(mockPlatform)

	// Initially should not have current Java
	if manager.HasCurrentJava() {
		t.Error("Expected HasCurrentJava to return false initially")
	}

	// Create symlink
	jemDir := filepath.Join(tmpDir, ".jem")
	jdkDir := filepath.Join(jemDir, "jdks", "21")
	currentDir := filepath.Join(jemDir, "current")
	javaLink := filepath.Join(currentDir, "java")

	os.MkdirAll(jdkDir, 0755)
	os.MkdirAll(currentDir, 0755)
	os.Symlink(jdkDir, javaLink)

	// Should have current Java now
	if !manager.HasCurrentJava() {
		t.Error("Expected HasCurrentJava to return true after creating symlink")
	}
}

func TestHasCurrentGradle(t *testing.T) {
	tmpDir := t.TempDir()
	mockPlatform := &MockPlatformForSymlink{homeDir: tmpDir}
	manager := NewSymlinkManager(mockPlatform)

	// Initially should not have current Gradle
	if manager.HasCurrentGradle() {
		t.Error("Expected HasCurrentGradle to return false initially")
	}

	// Create symlink
	jemDir := filepath.Join(tmpDir, ".jem")
	gradleDir := filepath.Join(jemDir, "gradles", "8.5")
	currentDir := filepath.Join(jemDir, "current")
	gradleLink := filepath.Join(currentDir, "gradle")

	os.MkdirAll(gradleDir, 0755)
	os.MkdirAll(currentDir, 0755)
	os.Symlink(gradleDir, gradleLink)

	// Should have current Gradle now
	if !manager.HasCurrentGradle() {
		t.Error("Expected HasCurrentGradle to return true after creating symlink")
	}
}

func TestUpdateCurrentJava_ReplaceExisting(t *testing.T) {
	tmpDir := t.TempDir()
	mockPlatform := &MockPlatformForSymlink{homeDir: tmpDir}
	manager := NewSymlinkManager(mockPlatform)

	// Create the jem structure with multiple JDKs
	jemDir := filepath.Join(tmpDir, ".jem")
	jdk21Dir := filepath.Join(jemDir, "jdks", "21")
	jdk17Dir := filepath.Join(jemDir, "jdks", "17")
	currentDir := filepath.Join(jemDir, "current")

	os.MkdirAll(jdk21Dir, 0755)
	os.MkdirAll(jdk17Dir, 0755)

	// Set initial version
	err := manager.UpdateCurrentJava("21", jdk21Dir)
	if err != nil {
		t.Fatalf("Failed to set initial Java version: %v", err)
	}

	// Verify initial symlink
	javaLink := filepath.Join(currentDir, "java")
	target, _ := os.Readlink(javaLink)
	if target != jdk21Dir {
		t.Errorf("Expected initial symlink to point to %s, got %s", jdk21Dir, target)
	}

	// Update to different version
	err = manager.UpdateCurrentJava("17", jdk17Dir)
	if err != nil {
		t.Errorf("Expected no error when replacing symlink, got: %v", err)
	}

	// Verify symlink was updated
	target, _ = os.Readlink(javaLink)
	if target != jdk17Dir {
		t.Errorf("Expected updated symlink to point to %s, got %s", jdk17Dir, target)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsInternal(s, substr))
}

func containsInternal(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
