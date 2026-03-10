package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMigrateCurrentToDefaults_JDK verifies migration of jdk.current to defaults.jdk
func TestMigrateCurrentToDefaults_JDK(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		JDK: JDKConfig{
			Current: "temurin-17",
		},
		Gradle: GradleConfig{
			Current: "",
		},
		Defaults: DefaultsConfig{
			JDK:    "",
			Gradle: "",
		},
	}

	err := MigrateCurrentToDefaults(cfg, tmpDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify defaults.jdk was set
	if cfg.Defaults.JDK != "temurin-17" {
		t.Errorf("Expected defaults.jdk to be 'temurin-17', got: %s", cfg.Defaults.JDK)
	}

	// Verify jdk.current was preserved (backward compatibility)
	if cfg.JDK.Current != "temurin-17" {
		t.Errorf("Expected jdk.current to be preserved, got: %s", cfg.JDK.Current)
	}
}

// TestMigrateCurrentToDefaults_Gradle verifies migration of gradle.current to defaults.gradle
func TestMigrateCurrentToDefaults_Gradle(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		JDK: JDKConfig{
			Current: "",
		},
		Gradle: GradleConfig{
			Current: "8.5",
		},
		Defaults: DefaultsConfig{
			JDK:    "",
			Gradle: "",
		},
	}

	err := MigrateCurrentToDefaults(cfg, tmpDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify defaults.gradle was set
	if cfg.Defaults.Gradle != "8.5" {
		t.Errorf("Expected defaults.gradle to be '8.5', got: %s", cfg.Defaults.Gradle)
	}

	// Verify gradle.current was preserved (backward compatibility)
	if cfg.Gradle.Current != "8.5" {
		t.Errorf("Expected gradle.current to be preserved, got: %s", cfg.Gradle.Current)
	}
}

// TestMigrateCurrentToDefaults_Both verifies migration of both jdk and gradle
func TestMigrateCurrentToDefaults_Both(t *testing.T) {
	tmpDir := t.TempDir()

	// Create target directories for symlinks
	jdkDir := filepath.Join(tmpDir, "jdks", "temurin-17")
	gradleDir := filepath.Join(tmpDir, "gradles", "8.5")
	os.MkdirAll(jdkDir, 0755)
	os.MkdirAll(gradleDir, 0755)

	cfg := &Config{
		JDK: JDKConfig{
			Current: "temurin-17",
		},
		Gradle: GradleConfig{
			Current: "8.5",
		},
		Defaults: DefaultsConfig{
			JDK:    "",
			Gradle: "",
		},
	}

	err := MigrateCurrentToDefaults(cfg, tmpDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify both defaults were set
	if cfg.Defaults.JDK != "temurin-17" {
		t.Errorf("Expected defaults.jdk to be 'temurin-17', got: %s", cfg.Defaults.JDK)
	}
	if cfg.Defaults.Gradle != "8.5" {
		t.Errorf("Expected defaults.gradle to be '8.5', got: %s", cfg.Defaults.Gradle)
	}

	// Verify symlinks were created
	javaSymlink := filepath.Join(tmpDir, "current", "java")
	gradleSymlink := filepath.Join(tmpDir, "current", "gradle")

	if _, err := os.Lstat(javaSymlink); os.IsNotExist(err) {
		t.Error("Expected java symlink to be created")
	}
	if _, err := os.Lstat(gradleSymlink); os.IsNotExist(err) {
		t.Error("Expected gradle symlink to be created")
	}
}

// TestMigrateCurrentToDefaults_AlreadySet verifies no migration when defaults already set
func TestMigrateCurrentToDefaults_AlreadySet(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		JDK: JDKConfig{
			Current: "temurin-21",
		},
		Gradle: GradleConfig{
			Current: "8.6",
		},
		Defaults: DefaultsConfig{
			JDK:    "temurin-17", // Already set
			Gradle: "8.5",        // Already set
		},
	}

	err := MigrateCurrentToDefaults(cfg, tmpDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify defaults were NOT changed
	if cfg.Defaults.JDK != "temurin-17" {
		t.Errorf("Expected defaults.jdk to remain 'temurin-17', got: %s", cfg.Defaults.JDK)
	}
	if cfg.Defaults.Gradle != "8.5" {
		t.Errorf("Expected defaults.gradle to remain '8.5', got: %s", cfg.Defaults.Gradle)
	}
}

// TestMigrateCurrentToDefaults_NoCurrent verifies no migration when no current set
func TestMigrateCurrentToDefaults_NoCurrent(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		JDK: JDKConfig{
			Current: "",
		},
		Gradle: GradleConfig{
			Current: "",
		},
		Defaults: DefaultsConfig{
			JDK:    "",
			Gradle: "",
		},
	}

	err := MigrateCurrentToDefaults(cfg, tmpDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify defaults remain empty
	if cfg.Defaults.JDK != "" {
		t.Errorf("Expected defaults.jdk to remain empty, got: %s", cfg.Defaults.JDK)
	}
	if cfg.Defaults.Gradle != "" {
		t.Errorf("Expected defaults.gradle to remain empty, got: %s", cfg.Defaults.Gradle)
	}
}

// TestMigrateCurrentToDefaults_TargetNotExists verifies migration works when target doesn't exist
func TestMigrateCurrentToDefaults_TargetNotExists(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		JDK: JDKConfig{
			Current: "nonexistent-jdk",
		},
		Defaults: DefaultsConfig{
			JDK: "",
		},
	}

	err := MigrateCurrentToDefaults(cfg, tmpDir)
	if err != nil {
		t.Errorf("Expected no error when target doesn't exist, got: %v", err)
	}

	// Verify defaults was set
	if cfg.Defaults.JDK != "nonexistent-jdk" {
		t.Errorf("Expected defaults.jdk to be set, got: %s", cfg.Defaults.JDK)
	}

	// Verify no symlink was created (target doesn't exist)
	symlinkPath := filepath.Join(tmpDir, "current", "java")
	if _, err := os.Lstat(symlinkPath); !os.IsNotExist(err) {
		t.Error("Expected no symlink to be created when target doesn't exist")
	}
}

// TestCreateCurrentSymlink_Java verifies symlink creation for java
func TestCreateCurrentSymlink_Java(t *testing.T) {
	tmpDir := t.TempDir()

	// Create target directory
	targetDir := filepath.Join(tmpDir, "jdks", "temurin-17")
	os.MkdirAll(targetDir, 0755)

	err := createCurrentSymlink(tmpDir, "java", "temurin-17")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify symlink exists
	symlinkPath := filepath.Join(tmpDir, "current", "java")
	info, err := os.Lstat(symlinkPath)
	if err != nil {
		t.Errorf("Expected symlink to exist: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("Expected path to be a symlink")
	}

	// Verify symlink points to correct target
	target, err := os.Readlink(symlinkPath)
	if err != nil {
		t.Errorf("Failed to read symlink: %v", err)
	}
	expectedTarget := filepath.Join(tmpDir, "jdks", "temurin-17")
	if target != expectedTarget {
		t.Errorf("Expected symlink to point to '%s', got: '%s'", expectedTarget, target)
	}
}

// TestCreateCurrentSymlink_Gradle verifies symlink creation for gradle
func TestCreateCurrentSymlink_Gradle(t *testing.T) {
	tmpDir := t.TempDir()

	// Create target directory
	targetDir := filepath.Join(tmpDir, "gradles", "8.5")
	os.MkdirAll(targetDir, 0755)

	err := createCurrentSymlink(tmpDir, "gradle", "8.5")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify symlink exists
	symlinkPath := filepath.Join(tmpDir, "current", "gradle")
	info, err := os.Lstat(symlinkPath)
	if err != nil {
		t.Errorf("Expected symlink to exist: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("Expected path to be a symlink")
	}

	// Verify symlink points to correct target
	target, err := os.Readlink(symlinkPath)
	if err != nil {
		t.Errorf("Failed to read symlink: %v", err)
	}
	expectedTarget := filepath.Join(tmpDir, "gradles", "8.5")
	if target != expectedTarget {
		t.Errorf("Expected symlink to point to '%s', got: '%s'", expectedTarget, target)
	}
}

// TestCreateCurrentSymlink_UpdateExisting verifies updating an existing symlink
func TestCreateCurrentSymlink_UpdateExisting(t *testing.T) {
	tmpDir := t.TempDir()

	// Create old target
	oldTarget := filepath.Join(tmpDir, "jdks", "temurin-11")
	newTarget := filepath.Join(tmpDir, "jdks", "temurin-17")
	os.MkdirAll(oldTarget, 0755)
	os.MkdirAll(newTarget, 0755)

	// Create initial symlink
	currentDir := filepath.Join(tmpDir, "current")
	os.MkdirAll(currentDir, 0755)
	symlinkPath := filepath.Join(currentDir, "java")
	os.Symlink(oldTarget, symlinkPath)

	// Update symlink
	err := createCurrentSymlink(tmpDir, "java", "temurin-17")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify symlink was updated
	target, err := os.Readlink(symlinkPath)
	if err != nil {
		t.Errorf("Failed to read symlink: %v", err)
	}
	if target != newTarget {
		t.Errorf("Expected symlink to point to '%s', got: '%s'", newTarget, target)
	}
}

// TestUpdateCurrentSymlinks_Both verifies updating both symlinks
func TestUpdateCurrentSymlinks_Both(t *testing.T) {
	tmpDir := t.TempDir()

	// Create targets
	jdkTarget := filepath.Join(tmpDir, "jdks", "temurin-21")
	gradleTarget := filepath.Join(tmpDir, "gradles", "8.6")
	os.MkdirAll(jdkTarget, 0755)
	os.MkdirAll(gradleTarget, 0755)

	cfg := &Config{
		Defaults: DefaultsConfig{
			JDK:    "temurin-21",
			Gradle: "8.6",
		},
	}

	err := UpdateCurrentSymlinks(cfg, tmpDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify both symlinks exist
	javaSymlink := filepath.Join(tmpDir, "current", "java")
	gradleSymlink := filepath.Join(tmpDir, "current", "gradle")

	if _, err := os.Lstat(javaSymlink); os.IsNotExist(err) {
		t.Error("Expected java symlink to exist")
	}
	if _, err := os.Lstat(gradleSymlink); os.IsNotExist(err) {
		t.Error("Expected gradle symlink to exist")
	}
}

// TestUpdateCurrentSymlinks_OnlyJDK verifies updating only JDK symlink
func TestUpdateCurrentSymlinks_OnlyJDK(t *testing.T) {
	tmpDir := t.TempDir()

	// Create target
	jdkTarget := filepath.Join(tmpDir, "jdks", "temurin-17")
	os.MkdirAll(jdkTarget, 0755)

	cfg := &Config{
		Defaults: DefaultsConfig{
			JDK:    "temurin-17",
			Gradle: "", // Not set
		},
	}

	err := UpdateCurrentSymlinks(cfg, tmpDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify only java symlink exists
	javaSymlink := filepath.Join(tmpDir, "current", "java")
	gradleSymlink := filepath.Join(tmpDir, "current", "gradle")

	if _, err := os.Lstat(javaSymlink); os.IsNotExist(err) {
		t.Error("Expected java symlink to exist")
	}
	if _, err := os.Lstat(gradleSymlink); !os.IsNotExist(err) {
		t.Error("Expected gradle symlink to NOT exist")
	}
}

// TestUpdateCurrentSymlinks_OnlyGradle verifies updating only Gradle symlink
func TestUpdateCurrentSymlinks_OnlyGradle(t *testing.T) {
	tmpDir := t.TempDir()

	// Create target
	gradleTarget := filepath.Join(tmpDir, "gradles", "8.5")
	os.MkdirAll(gradleTarget, 0755)

	cfg := &Config{
		Defaults: DefaultsConfig{
			JDK:    "", // Not set
			Gradle: "8.5",
		},
	}

	err := UpdateCurrentSymlinks(cfg, tmpDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify only gradle symlink exists
	javaSymlink := filepath.Join(tmpDir, "current", "java")
	gradleSymlink := filepath.Join(tmpDir, "current", "gradle")

	if _, err := os.Lstat(javaSymlink); !os.IsNotExist(err) {
		t.Error("Expected java symlink to NOT exist")
	}
	if _, err := os.Lstat(gradleSymlink); os.IsNotExist(err) {
		t.Error("Expected gradle symlink to exist")
	}
}

// TestUpdateCurrentSymlinks_None verifies no symlinks created when defaults are empty
func TestUpdateCurrentSymlinks_None(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		Defaults: DefaultsConfig{
			JDK:    "",
			Gradle: "",
		},
	}

	err := UpdateCurrentSymlinks(cfg, tmpDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify no current directory was created
	currentDir := filepath.Join(tmpDir, "current")
	if _, err := os.Stat(currentDir); !os.IsNotExist(err) {
		t.Error("Expected current directory to NOT exist")
	}
}
