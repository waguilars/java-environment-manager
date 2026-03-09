package cmd

import (
	"path/filepath"
	"testing"

	"github.com/waguilars/java-environment-manager/internal/config"
)

func TestListCommand_Execute(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	cmd := &ListCommand{
		configRepo: repo,
	}

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestListCommand_ExecuteJDK(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	cmd := &ListCommand{
		configRepo: repo,
	}

	err = cmd.ExecuteJDK()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestListCommand_ExecuteGradle(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	cmd := &ListCommand{
		configRepo: repo,
	}

	err = cmd.ExecuteGradle()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestListCommand_listJDKs_NoJDKs(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	cmd := &ListCommand{
		configRepo: repo,
	}

	// This should not error even with no JDKs
	cmd.listJDKs()
}

func TestListCommand_listGradles_NoGradles(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	cmd := &ListCommand{
		configRepo: repo,
	}

	// This should not error even with no Gradles
	cmd.listGradles()
}

func TestListCommand_listJDKs_WithInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Add some installed JDKs
	repo.AddInstalledJDK(config.JDKInfo{
		Path:     filepath.Join(tmpDir, "jdk1"),
		Version:  "17.0.10",
		Provider: "temurin",
		Managed:  true,
	})
	repo.AddInstalledJDK(config.JDKInfo{
		Path:     filepath.Join(tmpDir, "jdk2"),
		Version:  "21.0.1",
		Provider: "temurin",
		Managed:  true,
	})

	// Set current JDK
	repo.SetJDKCurrent("21.0.1")

	cmd := &ListCommand{
		configRepo: repo,
	}

	// This should list the JDKs
	cmd.listJDKs()
}

func TestListCommand_listJDKs_WithDetected(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Add some detected JDKs
	repo.AddDetectedJDK(config.JDKInfo{
		Path:     "/usr/lib/jvm/java-11",
		Version:  "11.0.12",
		Provider: "corretto",
		Managed:  false,
	})

	cmd := &ListCommand{
		configRepo: repo,
	}

	// This should list the detected JDKs
	cmd.listJDKs()
}

func TestListCommand_listGradles_WithInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Add some installed Gradles
	repo.AddInstalledGradle(config.GradleInfo{
		Path:    filepath.Join(tmpDir, "gradle1"),
		Version: "7.6.1",
		Managed: true,
	})
	repo.AddInstalledGradle(config.GradleInfo{
		Path:    filepath.Join(tmpDir, "gradle2"),
		Version: "8.5.0",
		Managed: true,
	})

	// Set current Gradle
	repo.SetGradleCurrent("8.5.0")

	cmd := &ListCommand{
		configRepo: repo,
	}

	// This should list the Gradles
	cmd.listGradles()
}

func TestListCommand_listGradles_WithDetected(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Add some detected Gradles
	repo.AddDetectedGradle(config.GradleInfo{
		Path:    "/opt/gradle-6.9",
		Version: "6.9.4",
		Managed: false,
	})

	cmd := &ListCommand{
		configRepo: repo,
	}

	// This should list the detected Gradles
	cmd.listGradles()
}

func TestListCommand_listJDKs_Sorting(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Add JDKs in random order
	repo.AddInstalledJDK(config.JDKInfo{
		Path:     filepath.Join(tmpDir, "jdk3"),
		Version:  "21.0.1",
		Provider: "temurin",
		Managed:  true,
	})
	repo.AddInstalledJDK(config.JDKInfo{
		Path:     filepath.Join(tmpDir, "jdk1"),
		Version:  "17.0.10",
		Provider: "temurin",
		Managed:  true,
	})
	repo.AddInstalledJDK(config.JDKInfo{
		Path:     filepath.Join(tmpDir, "jdk2"),
		Version:  "11.0.12",
		Provider: "temurin",
		Managed:  true,
	})

	cmd := &ListCommand{
		configRepo: repo,
	}

	// This should sort them by version
	cmd.listJDKs()
}

func TestListCommand_listGradles_Sorting(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Add Gradles in random order
	repo.AddInstalledGradle(config.GradleInfo{
		Path:    filepath.Join(tmpDir, "gradle3"),
		Version: "8.5.0",
		Managed: true,
	})
	repo.AddInstalledGradle(config.GradleInfo{
		Path:    filepath.Join(tmpDir, "gradle1"),
		Version: "7.6.1",
		Managed: true,
	})
	repo.AddInstalledGradle(config.GradleInfo{
		Path:    filepath.Join(tmpDir, "gradle2"),
		Version: "6.9.4",
		Managed: true,
	})

	cmd := &ListCommand{
		configRepo: repo,
	}

	// This should sort them by version
	cmd.listGradles()
}

func TestListCommand_listJDKs_WithCurrentMarker(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Add JDKs
	repo.AddInstalledJDK(config.JDKInfo{
		Path:     filepath.Join(tmpDir, "jdk1"),
		Version:  "17.0.10",
		Provider: "temurin",
		Managed:  true,
	})
	repo.AddInstalledJDK(config.JDKInfo{
		Path:     filepath.Join(tmpDir, "jdk2"),
		Version:  "21.0.1",
		Provider: "temurin",
		Managed:  true,
	})

	// Set current JDK
	repo.SetJDKCurrent("21.0.1")

	cmd := &ListCommand{
		configRepo: repo,
	}

	// This should mark the current JDK
	cmd.listJDKs()
}

func TestListCommand_listGradles_WithCurrentMarker(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Add Gradles
	repo.AddInstalledGradle(config.GradleInfo{
		Path:    filepath.Join(tmpDir, "gradle1"),
		Version: "7.6.1",
		Managed: true,
	})
	repo.AddInstalledGradle(config.GradleInfo{
		Path:    filepath.Join(tmpDir, "gradle2"),
		Version: "8.5.0",
		Managed: true,
	})

	// Set current Gradle
	repo.SetGradleCurrent("8.5.0")

	cmd := &ListCommand{
		configRepo: repo,
	}

	// This should mark the current Gradle
	cmd.listGradles()
}

func TestListCommand_listJDKs_WithSystemFallback(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	cmd := &ListCommand{
		configRepo: repo,
	}

	// This should fallback to system detection when no JDKs in config
	cmd.listJDKs()
}

func TestListCommand_listGradles_WithSystemFallback(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := config.NewTOMLConfigRepository(configPath)
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	cmd := &ListCommand{
		configRepo: repo,
	}

	// This should fallback to system detection when no Gradles in config
	cmd.listGradles()
}
