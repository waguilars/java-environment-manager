package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// TOMLConfigRepository implements ConfigRepository using TOML files
type TOMLConfigRepository struct {
	configPath string
	config     *Config
}

// NewTOMLConfigRepository creates a new TOML config repository
func NewTOMLConfigRepository(configPath string) *TOMLConfigRepository {
	return &TOMLConfigRepository{
		configPath: configPath,
		config:     &Config{},
	}
}

// Load reads the configuration from the TOML file
func (r *TOMLConfigRepository) Load() (*Config, error) {
	// Create default config if file doesn't exist
	if _, err := os.Stat(r.configPath); os.IsNotExist(err) {
		r.config = r.getDefaultConfig()
		return r.config, nil
	}

	// Try to read and parse the config file
	_, err := os.Stat(r.configPath)
	if err != nil {
		// File doesn't exist, create default
		r.config = r.getDefaultConfig()
		return r.config, nil
	}

	// Parse the TOML file
	_, err = toml.DecodeFile(r.configPath, r.config)
	if err != nil {
		// Config is corrupted, create backup and default
		r.handleCorruptedConfig(err)
		r.config = r.getDefaultConfig()
		return r.config, nil
	}

	// Ensure maps are initialized
	if r.config.InstalledJDKs == nil {
		r.config.InstalledJDKs = make(map[string]JDKInfo)
	}
	if r.config.DetectedJDKs == nil {
		r.config.DetectedJDKs = make(map[string]JDKInfo)
	}
	if r.config.InstalledGradles == nil {
		r.config.InstalledGradles = make(map[string]GradleInfo)
	}
	if r.config.DetectedGradles == nil {
		r.config.DetectedGradles = make(map[string]GradleInfo)
	}

	// Run migration for old config format (jdk.current -> defaults.jdk, etc.)
	jemDir := filepath.Dir(r.configPath)
	if err := MigrateCurrentToDefaults(r.config, jemDir); err != nil {
		// Log error but don't fail - migration is best-effort
		fmt.Fprintf(os.Stderr, "Warning: config migration failed: %v\n", err)
	}

	return r.config, nil
}

// Save writes the configuration to the TOML file
func (r *TOMLConfigRepository) Save(config *Config) error {
	// Ensure maps are not nil
	if config.InstalledJDKs == nil {
		config.InstalledJDKs = make(map[string]JDKInfo)
	}
	if config.DetectedJDKs == nil {
		config.DetectedJDKs = make(map[string]JDKInfo)
	}
	if config.InstalledGradles == nil {
		config.InstalledGradles = make(map[string]GradleInfo)
	}
	if config.DetectedGradles == nil {
		config.DetectedGradles = make(map[string]GradleInfo)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(r.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Encode to TOML and write
	file, err := os.Create(r.configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	return encoder.Encode(config)
}

// Reload reloads the configuration from the file
func (r *TOMLConfigRepository) Reload() error {
	config, err := r.Load()
	if err != nil {
		return err
	}
	r.config = config
	return nil
}

// GetDefaultProvider returns the default JDK provider
func (r *TOMLConfigRepository) GetDefaultProvider() string {
	return r.config.General.DefaultProvider
}

// SetDefaultProvider sets the default JDK provider
func (r *TOMLConfigRepository) SetDefaultProvider(provider string) error {
	r.config.General.DefaultProvider = provider
	return r.Save(r.config)
}

// GetJDKCurrent returns the currently active JDK name
func (r *TOMLConfigRepository) GetJDKCurrent() string {
	return r.config.JDK.Current
}

// SetJDKCurrent sets the currently active JDK name
func (r *TOMLConfigRepository) SetJDKCurrent(name string) error {
	r.config.JDK.Current = name
	return r.Save(r.config)
}

// ListInstalledJDKs returns all installed JDKs
func (r *TOMLConfigRepository) ListInstalledJDKs() []JDKInfo {
	jdkList := make([]JDKInfo, 0, len(r.config.InstalledJDKs))
	for _, info := range r.config.InstalledJDKs {
		jdkList = append(jdkList, info)
	}
	return jdkList
}

// ListDetectedJDKs returns all detected JDKs
func (r *TOMLConfigRepository) ListDetectedJDKs() []JDKInfo {
	jdkList := make([]JDKInfo, 0, len(r.config.DetectedJDKs))
	for _, info := range r.config.DetectedJDKs {
		jdkList = append(jdkList, info)
	}
	return jdkList
}

// AddInstalledJDK adds an installed JDK
func (r *TOMLConfigRepository) AddInstalledJDK(info JDKInfo) error {
	if r.config.InstalledJDKs == nil {
		r.config.InstalledJDKs = make(map[string]JDKInfo)
	}
	r.config.InstalledJDKs[info.Path] = info
	return r.Save(r.config)
}

// RemoveInstalledJDK removes an installed JDK
func (r *TOMLConfigRepository) RemoveInstalledJDK(name string) error {
	for path, info := range r.config.InstalledJDKs {
		if info.Version == name || contains(path, name) {
			delete(r.config.InstalledJDKs, path)
			return r.Save(r.config)
		}
	}
	return nil
}

// AddDetectedJDK adds a detected JDK
func (r *TOMLConfigRepository) AddDetectedJDK(info JDKInfo) error {
	if r.config.DetectedJDKs == nil {
		r.config.DetectedJDKs = make(map[string]JDKInfo)
	}
	r.config.DetectedJDKs[info.Path] = info
	return r.Save(r.config)
}

// ClearDetectedJDKs removes all detected JDKs
func (r *TOMLConfigRepository) ClearDetectedJDKs() error {
	r.config.DetectedJDKs = make(map[string]JDKInfo)
	return r.Save(r.config)
}

// GetGradleCurrent returns the currently active Gradle name
func (r *TOMLConfigRepository) GetGradleCurrent() string {
	return r.config.Gradle.Current
}

// SetGradleCurrent sets the currently active Gradle name
func (r *TOMLConfigRepository) SetGradleCurrent(name string) error {
	r.config.Gradle.Current = name
	return r.Save(r.config)
}

// ListInstalledGradles returns all installed Gradles
func (r *TOMLConfigRepository) ListInstalledGradles() []GradleInfo {
	gradleList := make([]GradleInfo, 0, len(r.config.InstalledGradles))
	for _, info := range r.config.InstalledGradles {
		gradleList = append(gradleList, info)
	}
	return gradleList
}

// ListDetectedGradles returns all detected Gradles
func (r *TOMLConfigRepository) ListDetectedGradles() []GradleInfo {
	gradleList := make([]GradleInfo, 0, len(r.config.DetectedGradles))
	for _, info := range r.config.DetectedGradles {
		gradleList = append(gradleList, info)
	}
	return gradleList
}

// AddInstalledGradle adds an installed Gradle
func (r *TOMLConfigRepository) AddInstalledGradle(info GradleInfo) error {
	if r.config.InstalledGradles == nil {
		r.config.InstalledGradles = make(map[string]GradleInfo)
	}
	r.config.InstalledGradles[info.Path] = info
	return r.Save(r.config)
}

// RemoveInstalledGradle removes an installed Gradle
func (r *TOMLConfigRepository) RemoveInstalledGradle(name string) error {
	for path, info := range r.config.InstalledGradles {
		if info.Version == name || contains(path, name) {
			delete(r.config.InstalledGradles, path)
			return r.Save(r.config)
		}
	}
	return nil
}

// AddDetectedGradle adds a detected Gradle
func (r *TOMLConfigRepository) AddDetectedGradle(info GradleInfo) error {
	if r.config.DetectedGradles == nil {
		r.config.DetectedGradles = make(map[string]GradleInfo)
	}
	r.config.DetectedGradles[info.Path] = info
	return r.Save(r.config)
}

// ClearDetectedGradles removes all detected Gradles
func (r *TOMLConfigRepository) ClearDetectedGradles() error {
	r.config.DetectedGradles = make(map[string]GradleInfo)
	return r.Save(r.config)
}

// GetDefaultJDK returns the default JDK version
func (r *TOMLConfigRepository) GetDefaultJDK() string {
	return r.config.Defaults.JDK
}

// SetDefaultJDK sets the default JDK version
func (r *TOMLConfigRepository) SetDefaultJDK(version string) error {
	r.config.Defaults.JDK = version
	return r.Save(r.config)
}

// GetDefaultGradle returns the default Gradle version
func (r *TOMLConfigRepository) GetDefaultGradle() string {
	return r.config.Defaults.Gradle
}

// SetDefaultGradle sets the default Gradle version
func (r *TOMLConfigRepository) SetDefaultGradle(version string) error {
	r.config.Defaults.Gradle = version
	return r.Save(r.config)
}

// getDefaultConfig returns a default configuration
func (r *TOMLConfigRepository) getDefaultConfig() *Config {
	return &Config{
		General: GeneralConfig{
			DefaultProvider: "temurin",
		},
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
		InstalledJDKs:    make(map[string]JDKInfo),
		DetectedJDKs:     make(map[string]JDKInfo),
		InstalledGradles: make(map[string]GradleInfo),
		DetectedGradles:  make(map[string]GradleInfo),
	}
}

// handleCorruptedConfig handles corrupted config files
func (r *TOMLConfigRepository) handleCorruptedConfig(err error) {
	// Create backup
	backupPath := r.configPath + ".backup"
	os.Rename(r.configPath, backupPath)
}

// contains is a helper to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
