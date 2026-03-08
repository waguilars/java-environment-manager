package config

// ConfigRepository interface defines the configuration persistence contract
type ConfigRepository interface {
	Load() (*Config, error)
	Save(config *Config) error
	Reload() error

	// General operations
	GetDefaultProvider() string
	SetDefaultProvider(provider string) error

	// JDK operations
	GetJDKCurrent() string
	SetJDKCurrent(name string) error
	ListInstalledJDKs() []JDKInfo
	ListDetectedJDKs() []JDKInfo
	AddInstalledJDK(info JDKInfo) error
	RemoveInstalledJDK(name string) error
	AddDetectedJDK(info JDKInfo) error
	ClearDetectedJDKs() error

	// Gradle operations
	GetGradleCurrent() string
	SetGradleCurrent(name string) error
	ListInstalledGradles() []GradleInfo
	ListDetectedGradles() []GradleInfo
	AddInstalledGradle(info GradleInfo) error
	RemoveInstalledGradle(name string) error
	AddDetectedGradle(info GradleInfo) error
	ClearDetectedGradles() error
}
