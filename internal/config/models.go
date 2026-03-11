package config

// Config represents the root configuration structure
type Config struct {
	General          GeneralConfig         `toml:"general"`
	JDK              JDKConfig             `toml:"jdk"`
	Gradle           GradleConfig          `toml:"gradle"`
	Defaults         DefaultsConfig        `toml:"defaults"`
	InstalledJDKs    map[string]JDKInfo    `toml:"jdks.installed"`
	DetectedJDKs     map[string]JDKInfo    `toml:"jdks.detected"`
	InstalledGradles map[string]GradleInfo `toml:"gradles.installed"`
	DetectedGradles  map[string]GradleInfo `toml:"gradles.detected"`
}

// GeneralConfig contains general settings
type GeneralConfig struct {
	DefaultProvider string `toml:"default_provider"`
}

// JDKConfig contains JDK-specific settings
type JDKConfig struct {
	// Current is the current session JDK version (may differ from default)
	// Set when running 'jem use jdk <version>' without --default flag
	Current string `toml:"current"`
}

// GradleConfig contains Gradle-specific settings
type GradleConfig struct {
	// Current is the current session Gradle version (may differ from default)
	// Set when running 'jem use gradle <version>' without --default flag
	Current string `toml:"current"`
}

// JDKInfo represents information about a JDK installation
type JDKInfo struct {
	Path     string `toml:"path"`
	Version  string `toml:"version"`
	Provider string `toml:"provider,omitempty"`
	Managed  bool   `toml:"managed"` // true if installed by jem
}

// GradleInfo represents information about a Gradle installation
type GradleInfo struct {
	Path    string `toml:"path"`
	Version string `toml:"version"`
	Managed bool   `toml:"managed"` // true if installed by jem
}

// DefaultsConfig contains default version settings
// These values persist across sessions and are set with 'jem use <tool> <version> --default'
type DefaultsConfig struct {
	// JDK is the persistent default JDK version (survives shell restarts)
	// Set when running 'jem use jdk <version> --default' or 'jem use default jdk <version>'
	JDK string `toml:"jdk"`
	// Gradle is the persistent default Gradle version (survives shell restarts)
	// Set when running 'jem use gradle <version> --default' or 'jem use default gradle <version>'
	Gradle string `toml:"gradle"`
}

// UseMode represents the mode for version selection
type UseMode string

const (
	UseModeSession UseMode = "session"
	UseModeDefault UseMode = "default"
)

// Shell represents the shell type
type Shell string

const (
	ShellBash       Shell = "bash"
	ShellZsh        Shell = "zsh"
	ShellPowerShell Shell = "powershell"
	ShellFish       Shell = "fish"
)
