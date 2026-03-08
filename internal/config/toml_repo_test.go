package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTOMLConfigRepository_Load_NewConfig(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := NewTOMLConfigRepository(configPath)
	config, err := repo.Load()

	if err != nil {
		t.Fatalf("Load() should not error on new config: %v", err)
	}

	if config == nil {
		t.Fatal("Load() should return a config")
	}

	// Verify default values
	if config.General.DefaultProvider != "temurin" {
		t.Errorf("Expected default_provider to be 'temurin', got '%s'", config.General.DefaultProvider)
	}

	if config.JDK.Current != "" {
		t.Errorf("Expected JDK.Current to be empty, got '%s'", config.JDK.Current)
	}

	if config.InstalledJDKs == nil {
		t.Error("Expected InstalledJDKs to be initialized")
	}

	if config.DetectedJDKs == nil {
		t.Error("Expected DetectedJDKs to be initialized")
	}

	if config.InstalledGradles == nil {
		t.Error("Expected InstalledGradles to be initialized")
	}

	if config.DetectedGradles == nil {
		t.Error("Expected DetectedGradles to be initialized")
	}
}

func TestTOMLConfigRepository_Load_ExistingConfig(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create a config file with specific values
	configContent := `[general]
default_provider = "temurin"

[jdk]
current = "temurin-21"

[gradle]
current = "gradle-8.5"

[jdks.installed]
[jdks.detected]
[gradles.installed]
[gradles.detected]
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	repo := NewTOMLConfigRepository(configPath)
	config, err := repo.Load()

	if err != nil {
		t.Fatalf("Load() should not error on existing config: %v", err)
	}

	if config == nil {
		t.Fatal("Load() should return a config")
	}

	if config.General.DefaultProvider != "temurin" {
		t.Errorf("Expected default_provider to be 'temurin', got '%s'", config.General.DefaultProvider)
	}

	if config.JDK.Current != "temurin-21" {
		t.Errorf("Expected JDK.Current to be 'temurin-21', got '%s'", config.JDK.Current)
	}

	if config.Gradle.Current != "gradle-8.5" {
		t.Errorf("Expected Gradle.Current to be 'gradle-8.5', got '%s'", config.Gradle.Current)
	}
}

func TestTOMLConfigRepository_Save(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := NewTOMLConfigRepository(configPath)

	// Create a config to save
	config := &Config{
		General: GeneralConfig{
			DefaultProvider: "temurin",
		},
		JDK: JDKConfig{
			Current: "temurin-21",
		},
		Gradle: GradleConfig{
			Current: "gradle-8.5",
		},
		InstalledJDKs: map[string]JDKInfo{
			"/path/to/temurin-21": {
				Path:     "/path/to/temurin-21",
				Version:  "21.0.2",
				Provider: "temurin",
				Managed:  true,
			},
		},
		DetectedJDKs: map[string]JDKInfo{
			"/usr/lib/jvm/java-17": {
				Path:     "/usr/lib/jvm/java-17",
				Version:  "17.0.10",
				Provider: "corretto",
				Managed:  false,
			},
		},
		InstalledGradles: make(map[string]GradleInfo),
		DetectedGradles:  make(map[string]GradleInfo),
	}

	err := repo.Save(config)
	if err != nil {
		t.Fatalf("Save() should not error: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file should exist after Save()")
	}

	// Verify we can reload the config
	repo2 := NewTOMLConfigRepository(configPath)
	loadedConfig, err := repo2.Load()
	if err != nil {
		t.Fatalf("Failed to reload config: %v", err)
	}

	if loadedConfig.JDK.Current != "temurin-21" {
		t.Errorf("Expected reloaded JDK.Current to be 'temurin-21', got '%s'", loadedConfig.JDK.Current)
	}

	if len(loadedConfig.InstalledJDKs) != 1 {
		t.Errorf("Expected 1 installed JDK, got %d", len(loadedConfig.InstalledJDKs))
	}
}

func TestTOMLConfigRepository_GetJDKCurrent(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := NewTOMLConfigRepository(configPath)

	// Test with empty config
	current := repo.GetJDKCurrent()
	if current != "" {
		t.Errorf("Expected empty current JDK, got '%s'", current)
	}

	// Test with config that has current JDK
	config := &Config{
		JDK: JDKConfig{
			Current: "temurin-21",
		},
	}
	repo.config = config

	current = repo.GetJDKCurrent()
	if current != "temurin-21" {
		t.Errorf("Expected current JDK to be 'temurin-21', got '%s'", current)
	}
}

func TestTOMLConfigRepository_SetJDKCurrent(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := NewTOMLConfigRepository(configPath)

	err := repo.SetJDKCurrent("temurin-21")
	if err != nil {
		t.Fatalf("SetJDKCurrent() should not error: %v", err)
	}

	// Verify it was saved
	repo2 := NewTOMLConfigRepository(configPath)
	loadedConfig, err := repo2.Load()
	if err != nil {
		t.Fatalf("Failed to reload config: %v", err)
	}

	if loadedConfig.JDK.Current != "temurin-21" {
		t.Errorf("Expected JDK.Current to be 'temurin-21', got '%s'", loadedConfig.JDK.Current)
	}
}

func TestTOMLConfigRepository_ListInstalledJDKs(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := NewTOMLConfigRepository(configPath)

	// Test with empty config
	jdkList := repo.ListInstalledJDKs()
	if len(jdkList) != 0 {
		t.Errorf("Expected 0 installed JDKs, got %d", len(jdkList))
	}

	// Test with config that has installed JDKs
	config := &Config{
		InstalledJDKs: map[string]JDKInfo{
			"/path/to/temurin-21": {
				Path:     "/path/to/temurin-21",
				Version:  "21.0.2",
				Provider: "temurin",
				Managed:  true,
			},
			"/path/to/temurin-17": {
				Path:     "/path/to/temurin-17",
				Version:  "17.0.10",
				Provider: "temurin",
				Managed:  true,
			},
		},
	}
	repo.config = config

	jdkList = repo.ListInstalledJDKs()
	if len(jdkList) != 2 {
		t.Errorf("Expected 2 installed JDKs, got %d", len(jdkList))
	}
}

func TestTOMLConfigRepository_AddInstalledJDK(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := NewTOMLConfigRepository(configPath)

	jdkInfo := JDKInfo{
		Path:     "/path/to/temurin-21",
		Version:  "21.0.2",
		Provider: "temurin",
		Managed:  true,
	}

	err := repo.AddInstalledJDK(jdkInfo)
	if err != nil {
		t.Fatalf("AddInstalledJDK() should not error: %v", err)
	}

	// Verify it was added
	jdkList := repo.ListInstalledJDKs()
	if len(jdkList) != 1 {
		t.Errorf("Expected 1 installed JDK, got %d", len(jdkList))
	}

	// Verify it was persisted
	repo2 := NewTOMLConfigRepository(configPath)
	loadedConfig, err := repo2.Load()
	if err != nil {
		t.Fatalf("Failed to reload config: %v", err)
	}

	if len(loadedConfig.InstalledJDKs) != 1 {
		t.Errorf("Expected 1 installed JDK in persisted config, got %d", len(loadedConfig.InstalledJDKs))
	}
}

func TestTOMLConfigRepository_ListDetectedJDKs(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := NewTOMLConfigRepository(configPath)

	// Test with empty config
	jdkList := repo.ListDetectedJDKs()
	if len(jdkList) != 0 {
		t.Errorf("Expected 0 detected JDKs, got %d", len(jdkList))
	}

	// Test with config that has detected JDKs
	config := &Config{
		DetectedJDKs: map[string]JDKInfo{
			"/usr/lib/jvm/java-17": {
				Path:     "/usr/lib/jvm/java-17",
				Version:  "17.0.10",
				Provider: "corretto",
				Managed:  false,
			},
		},
	}
	repo.config = config

	jdkList = repo.ListDetectedJDKs()
	if len(jdkList) != 1 {
		t.Errorf("Expected 1 detected JDK, got %d", len(jdkList))
	}
}

func TestTOMLConfigRepository_AddDetectedJDK(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := NewTOMLConfigRepository(configPath)

	jdkInfo := JDKInfo{
		Path:     "/usr/lib/jvm/java-17",
		Version:  "17.0.10",
		Provider: "corretto",
		Managed:  false,
	}

	err := repo.AddDetectedJDK(jdkInfo)
	if err != nil {
		t.Fatalf("AddDetectedJDK() should not error: %v", err)
	}

	// Verify it was added
	jdkList := repo.ListDetectedJDKs()
	if len(jdkList) != 1 {
		t.Errorf("Expected 1 detected JDK, got %d", len(jdkList))
	}

	// Verify it was persisted
	repo2 := NewTOMLConfigRepository(configPath)
	loadedConfig, err := repo2.Load()
	if err != nil {
		t.Fatalf("Failed to reload config: %v", err)
	}

	if len(loadedConfig.DetectedJDKs) != 1 {
		t.Errorf("Expected 1 detected JDK in persisted config, got %d", len(loadedConfig.DetectedJDKs))
	}
}

func TestTOMLConfigRepository_Load_CorruptedConfig(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create a corrupted config file
	corruptedContent := `this is not valid toml {{{
	[general]
	default_provider = "temurin"
`
	err := os.WriteFile(configPath, []byte(corruptedContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write corrupted config: %v", err)
	}

	repo := NewTOMLConfigRepository(configPath)
	_, err = repo.Load()

	if err != nil {
		t.Fatalf("Load() should not error on corrupted config: %v", err)
	}

	// Verify backup was created
	backupPath := configPath + ".backup"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Backup file should exist after loading corrupted config")
	}

	// Verify default values are used
	repo2 := NewTOMLConfigRepository(configPath)
	_, err = repo2.Load()
	if err != nil {
		t.Fatalf("Failed to reload after corruption: %v", err)
	}
}

func TestTOMLConfigRepository_GetGradleCurrent(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := NewTOMLConfigRepository(configPath)

	// Test with empty config
	current := repo.GetGradleCurrent()
	if current != "" {
		t.Errorf("Expected empty current Gradle, got '%s'", current)
	}

	// Test with config that has current Gradle
	config := &Config{
		Gradle: GradleConfig{
			Current: "gradle-8.5",
		},
	}
	repo.config = config

	current = repo.GetGradleCurrent()
	if current != "gradle-8.5" {
		t.Errorf("Expected current Gradle to be 'gradle-8.5', got '%s'", current)
	}
}

func TestTOMLConfigRepository_SetGradleCurrent(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := NewTOMLConfigRepository(configPath)

	err := repo.SetGradleCurrent("gradle-8.5")
	if err != nil {
		t.Fatalf("SetGradleCurrent() should not error: %v", err)
	}

	// Verify it was saved
	repo2 := NewTOMLConfigRepository(configPath)
	loadedConfig, err := repo2.Load()
	if err != nil {
		t.Fatalf("Failed to reload config: %v", err)
	}

	if loadedConfig.Gradle.Current != "gradle-8.5" {
		t.Errorf("Expected Gradle.Current to be 'gradle-8.5', got '%s'", loadedConfig.Gradle.Current)
	}
}

func TestTOMLConfigRepository_ListInstalledGradles(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := NewTOMLConfigRepository(configPath)

	// Test with empty config
	gradleList := repo.ListInstalledGradles()
	if len(gradleList) != 0 {
		t.Errorf("Expected 0 installed Gradles, got %d", len(gradleList))
	}

	// Test with config that has installed Gradles
	config := &Config{
		InstalledGradles: map[string]GradleInfo{
			"/path/to/gradle-8.5": {
				Path:    "/path/to/gradle-8.5",
				Version: "8.5.0",
				Managed: true,
			},
		},
	}
	repo.config = config

	gradleList = repo.ListInstalledGradles()
	if len(gradleList) != 1 {
		t.Errorf("Expected 1 installed Gradle, got %d", len(gradleList))
	}
}

func TestTOMLConfigRepository_AddInstalledGradle(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := NewTOMLConfigRepository(configPath)

	gradleInfo := GradleInfo{
		Path:    "/path/to/gradle-8.5",
		Version: "8.5.0",
		Managed: true,
	}

	err := repo.AddInstalledGradle(gradleInfo)
	if err != nil {
		t.Fatalf("AddInstalledGradle() should not error: %v", err)
	}

	// Verify it was added
	gradleList := repo.ListInstalledGradles()
	if len(gradleList) != 1 {
		t.Errorf("Expected 1 installed Gradle, got %d", len(gradleList))
	}
}

func TestTOMLConfigRepository_ListDetectedGradles(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := NewTOMLConfigRepository(configPath)

	// Test with empty config
	gradleList := repo.ListDetectedGradles()
	if len(gradleList) != 0 {
		t.Errorf("Expected 0 detected Gradles, got %d", len(gradleList))
	}

	// Test with config that has detected Gradles
	config := &Config{
		DetectedGradles: map[string]GradleInfo{
			"/opt/gradle-8.0": {
				Path:    "/opt/gradle-8.0",
				Version: "8.0.2",
				Managed: false,
			},
		},
	}
	repo.config = config

	gradleList = repo.ListDetectedGradles()
	if len(gradleList) != 1 {
		t.Errorf("Expected 1 detected Gradle, got %d", len(gradleList))
	}
}

func TestTOMLConfigRepository_AddDetectedGradle(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := NewTOMLConfigRepository(configPath)

	gradleInfo := GradleInfo{
		Path:    "/opt/gradle-8.0",
		Version: "8.0.2",
		Managed: false,
	}

	err := repo.AddDetectedGradle(gradleInfo)
	if err != nil {
		t.Fatalf("AddDetectedGradle() should not error: %v", err)
	}

	// Verify it was added
	gradleList := repo.ListDetectedGradles()
	if len(gradleList) != 1 {
		t.Errorf("Expected 1 detected Gradle, got %d", len(gradleList))
	}
}

func TestTOMLConfigRepository_RemoveInstalledJDK(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := NewTOMLConfigRepository(configPath)

	// Add some JDKs
	repo.AddInstalledJDK(JDKInfo{
		Path:     "/path/to/temurin-21",
		Version:  "21.0.2",
		Provider: "temurin",
		Managed:  true,
	})
	repo.AddInstalledJDK(JDKInfo{
		Path:     "/path/to/temurin-17",
		Version:  "17.0.10",
		Provider: "temurin",
		Managed:  true,
	})

	// Remove one
	err := repo.RemoveInstalledJDK("temurin-21")
	if err != nil {
		t.Fatalf("RemoveInstalledJDK() should not error: %v", err)
	}

	// Verify it was removed
	jdkList := repo.ListInstalledJDKs()
	if len(jdkList) != 1 {
		t.Errorf("Expected 1 installed JDK after removal, got %d", len(jdkList))
	}
}

func TestTOMLConfigRepository_RemoveInstalledGradle(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := NewTOMLConfigRepository(configPath)

	// Add some Gradles
	repo.AddInstalledGradle(GradleInfo{
		Path:    "/path/to/gradle-8.5",
		Version: "8.5.0",
		Managed: true,
	})
	repo.AddInstalledGradle(GradleInfo{
		Path:    "/path/to/gradle-8.0",
		Version: "8.0.2",
		Managed: true,
	})

	// Remove one
	err := repo.RemoveInstalledGradle("gradle-8.5")
	if err != nil {
		t.Fatalf("RemoveInstalledGradle() should not error: %v", err)
	}

	// Verify it was removed
	gradleList := repo.ListInstalledGradles()
	if len(gradleList) != 1 {
		t.Errorf("Expected 1 installed Gradle after removal, got %d", len(gradleList))
	}
}

func TestTOMLConfigRepository_ClearDetectedJDKs(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := NewTOMLConfigRepository(configPath)

	// Add some detected JDKs
	repo.AddDetectedJDK(JDKInfo{
		Path:     "/usr/lib/jvm/java-17",
		Version:  "17.0.10",
		Provider: "corretto",
		Managed:  false,
	})
	repo.AddDetectedJDK(JDKInfo{
		Path:     "/usr/lib/jvm/java-21",
		Version:  "21.0.2",
		Provider: "temurin",
		Managed:  false,
	})

	// Clear them
	err := repo.ClearDetectedJDKs()
	if err != nil {
		t.Fatalf("ClearDetectedJDKs() should not error: %v", err)
	}

	// Verify they were cleared
	jdkList := repo.ListDetectedJDKs()
	if len(jdkList) != 0 {
		t.Errorf("Expected 0 detected JDKs after clear, got %d", len(jdkList))
	}
}

func TestTOMLConfigRepository_ClearDetectedGradles(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := NewTOMLConfigRepository(configPath)

	// Add some detected Gradles
	repo.AddDetectedGradle(GradleInfo{
		Path:    "/opt/gradle-8.0",
		Version: "8.0.2",
		Managed: false,
	})
	repo.AddDetectedGradle(GradleInfo{
		Path:    "/opt/gradle-8.5",
		Version: "8.5.0",
		Managed: false,
	})

	// Clear them
	err := repo.ClearDetectedGradles()
	if err != nil {
		t.Fatalf("ClearDetectedGradles() should not error: %v", err)
	}

	// Verify they were cleared
	gradleList := repo.ListDetectedGradles()
	if len(gradleList) != 0 {
		t.Errorf("Expected 0 detected Gradles after clear, got %d", len(gradleList))
	}
}

func TestTOMLConfigRepository_Reload(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	repo := NewTOMLConfigRepository(configPath)

	// Initial load
	_, err := repo.Load()
	if err != nil {
		t.Fatalf("Initial Load() should not error: %v", err)
	}

	// Modify config through another repo
	repo2 := NewTOMLConfigRepository(configPath)
	repo2.config = &Config{
		JDK: JDKConfig{
			Current: "temurin-21",
		},
	}
	err = repo2.Save(repo2.config)
	if err != nil {
		t.Fatalf("Save() should not error: %v", err)
	}

	// Reload
	err = repo.Reload()
	if err != nil {
		t.Fatalf("Reload() should not error: %v", err)
	}

	// Verify reload worked
	if repo.config.JDK.Current != "temurin-21" {
		t.Errorf("Expected JDK.Current to be 'temurin-21' after reload, got '%s'", repo.config.JDK.Current)
	}
}
