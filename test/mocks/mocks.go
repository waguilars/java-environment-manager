package mocks

import (
	"context"
	"github.com/user/jem/internal/config"
	"github.com/user/jem/internal/provider"
)

// MockPlatform is a mock implementation of the Platform interface
type MockPlatform struct {
	NameFunc                 func() string
	HomeDirFunc              func() string
	DetectShellFunc          func() config.Shell
	CreateLinkFunc           func(target, link string) error
	RemoveLinkFunc           func(link string) error
	IsLinkFunc               func(path string) bool
	CanCreateSymlinksFunc    func() bool
	ShellConfigPathFunc      func(shell config.Shell) string
	JDKDetectionPathsFunc    func() []string
	GradleDetectionPathsFunc func() []string
}

func (m *MockPlatform) Name() string {
	if m.NameFunc != nil {
		return m.NameFunc()
	}
	return "mock"
}

func (m *MockPlatform) HomeDir() string {
	if m.HomeDirFunc != nil {
		return m.HomeDirFunc()
	}
	return "/tmp"
}

func (m *MockPlatform) DetectShell() config.Shell {
	if m.DetectShellFunc != nil {
		return m.DetectShellFunc()
	}
	return config.ShellBash
}

func (m *MockPlatform) CreateLink(target, link string) error {
	if m.CreateLinkFunc != nil {
		return m.CreateLinkFunc(target, link)
	}
	return nil
}

func (m *MockPlatform) RemoveLink(link string) error {
	if m.RemoveLinkFunc != nil {
		return m.RemoveLinkFunc(link)
	}
	return nil
}

func (m *MockPlatform) IsLink(path string) bool {
	if m.IsLinkFunc != nil {
		return m.IsLinkFunc(path)
	}
	return false
}

func (m *MockPlatform) CanCreateSymlinks() bool {
	if m.CanCreateSymlinksFunc != nil {
		return m.CanCreateSymlinksFunc()
	}
	return true
}

func (m *MockPlatform) ShellConfigPath(shell config.Shell) string {
	if m.ShellConfigPathFunc != nil {
		return m.ShellConfigPathFunc(shell)
	}
	return "/tmp/.bashrc"
}

func (m *MockPlatform) JDKDetectionPaths() []string {
	if m.JDKDetectionPathsFunc != nil {
		return m.JDKDetectionPathsFunc()
	}
	return []string{}
}

func (m *MockPlatform) GradleDetectionPaths() []string {
	if m.GradleDetectionPathsFunc != nil {
		return m.GradleDetectionPathsFunc()
	}
	return []string{}
}

// MockJDKProvider is a mock implementation of the JDKProvider interface
type MockJDKProvider struct {
	NameFunc          func() string
	DisplayNameFunc   func() string
	ListAvailableFunc func(ctx context.Context, opts provider.ListOptions) ([]provider.JDKRelease, error)
	GetLatestLTSFunc  func(ctx context.Context) (*provider.JDKRelease, error)
	GetLatestFunc     func(ctx context.Context, majorVersion int) (*provider.JDKRelease, error)
	DownloadFunc      func(ctx context.Context, release provider.JDKRelease, dest string, progress provider.ProgressFunc) error
	GetChecksumFunc   func(release provider.JDKRelease) string
}

func (m *MockJDKProvider) Name() string {
	if m.NameFunc != nil {
		return m.NameFunc()
	}
	return "mock-provider"
}

func (m *MockJDKProvider) DisplayName() string {
	if m.DisplayNameFunc != nil {
		return m.DisplayNameFunc()
	}
	return "Mock Provider"
}

func (m *MockJDKProvider) ListAvailable(ctx context.Context, opts provider.ListOptions) ([]provider.JDKRelease, error) {
	if m.ListAvailableFunc != nil {
		return m.ListAvailableFunc(ctx, opts)
	}
	return []provider.JDKRelease{}, nil
}

func (m *MockJDKProvider) GetLatestLTS(ctx context.Context) (*provider.JDKRelease, error) {
	if m.GetLatestLTSFunc != nil {
		return m.GetLatestLTSFunc(ctx)
	}
	return nil, nil
}

func (m *MockJDKProvider) GetLatest(ctx context.Context, majorVersion int) (*provider.JDKRelease, error) {
	if m.GetLatestFunc != nil {
		return m.GetLatestFunc(ctx, majorVersion)
	}
	return nil, nil
}

func (m *MockJDKProvider) Download(ctx context.Context, release provider.JDKRelease, dest string, progress provider.ProgressFunc) error {
	if m.DownloadFunc != nil {
		return m.DownloadFunc(ctx, release, dest, progress)
	}
	return nil
}

func (m *MockJDKProvider) GetChecksum(release provider.JDKRelease) string {
	if m.GetChecksumFunc != nil {
		return m.GetChecksumFunc(release)
	}
	return "mock-checksum"
}

// MockGradleProvider is a mock implementation of the GradleProvider interface
type MockGradleProvider struct {
	NameFunc          func() string
	DisplayNameFunc   func() string
	ListAvailableFunc func(ctx context.Context) ([]provider.GradleRelease, error)
	GetLatestFunc     func(ctx context.Context) (*provider.GradleRelease, error)
	GetVersionFunc    func(ctx context.Context, version string) (*provider.GradleRelease, error)
	DownloadFunc      func(ctx context.Context, release provider.GradleRelease, dest string, progress provider.ProgressFunc) error
	GetChecksumFunc   func(ctx context.Context, release provider.GradleRelease) (string, error)
}

func (m *MockGradleProvider) Name() string {
	if m.NameFunc != nil {
		return m.NameFunc()
	}
	return "mock-gradle-provider"
}

func (m *MockGradleProvider) DisplayName() string {
	if m.DisplayNameFunc != nil {
		return m.DisplayNameFunc()
	}
	return "Mock Gradle Provider"
}

func (m *MockGradleProvider) ListAvailable(ctx context.Context) ([]provider.GradleRelease, error) {
	if m.ListAvailableFunc != nil {
		return m.ListAvailableFunc(ctx)
	}
	return []provider.GradleRelease{}, nil
}

func (m *MockGradleProvider) GetLatest(ctx context.Context) (*provider.GradleRelease, error) {
	if m.GetLatestFunc != nil {
		return m.GetLatestFunc(ctx)
	}
	return nil, nil
}

func (m *MockGradleProvider) GetVersion(ctx context.Context, version string) (*provider.GradleRelease, error) {
	if m.GetVersionFunc != nil {
		return m.GetVersionFunc(ctx, version)
	}
	return nil, nil
}

func (m *MockGradleProvider) Download(ctx context.Context, release provider.GradleRelease, dest string, progress provider.ProgressFunc) error {
	if m.DownloadFunc != nil {
		return m.DownloadFunc(ctx, release, dest, progress)
	}
	return nil
}

func (m *MockGradleProvider) GetChecksum(ctx context.Context, release provider.GradleRelease) (string, error) {
	if m.GetChecksumFunc != nil {
		return m.GetChecksumFunc(ctx, release)
	}
	return "mock-gradle-checksum", nil
}

// MockConfigRepository is a mock implementation of the ConfigRepository interface
type MockConfigRepository struct {
	LoadFunc                  func() (*config.Config, error)
	SaveFunc                  func(config *config.Config) error
	ReloadFunc                func() error
	GetDefaultProviderFunc    func() string
	SetDefaultProviderFunc    func(provider string) error
	GetJDKCurrentFunc         func() string
	SetJDKCurrentFunc         func(name string) error
	ListInstalledJDKsFunc     func() []config.JDKInfo
	ListDetectedJDKsFunc      func() []config.JDKInfo
	AddInstalledJDKFunc       func(info config.JDKInfo) error
	RemoveInstalledJDKFunc    func(name string) error
	AddDetectedJDKFunc        func(info config.JDKInfo) error
	ClearDetectedJDKsFunc     func() error
	GetGradleCurrentFunc      func() string
	SetGradleCurrentFunc      func(name string) error
	ListInstalledGradlesFunc  func() []config.GradleInfo
	ListDetectedGradlesFunc   func() []config.GradleInfo
	AddInstalledGradleFunc    func(info config.GradleInfo) error
	RemoveInstalledGradleFunc func(name string) error
	AddDetectedGradleFunc     func(info config.GradleInfo) error
	ClearDetectedGradlesFunc  func() error
}

func (m *MockConfigRepository) Load() (*config.Config, error) {
	if m.LoadFunc != nil {
		return m.LoadFunc()
	}
	return &config.Config{}, nil
}

func (m *MockConfigRepository) Save(config *config.Config) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(config)
	}
	return nil
}

func (m *MockConfigRepository) Reload() error {
	if m.ReloadFunc != nil {
		return m.ReloadFunc()
	}
	return nil
}

func (m *MockConfigRepository) GetDefaultProvider() string {
	if m.GetDefaultProviderFunc != nil {
		return m.GetDefaultProviderFunc()
	}
	return "temurin"
}

func (m *MockConfigRepository) SetDefaultProvider(provider string) error {
	if m.SetDefaultProviderFunc != nil {
		return m.SetDefaultProviderFunc(provider)
	}
	return nil
}

func (m *MockConfigRepository) GetJDKCurrent() string {
	if m.GetJDKCurrentFunc != nil {
		return m.GetJDKCurrentFunc()
	}
	return ""
}

func (m *MockConfigRepository) SetJDKCurrent(name string) error {
	if m.SetJDKCurrentFunc != nil {
		return m.SetJDKCurrentFunc(name)
	}
	return nil
}

func (m *MockConfigRepository) ListInstalledJDKs() []config.JDKInfo {
	if m.ListInstalledJDKsFunc != nil {
		return m.ListInstalledJDKsFunc()
	}
	return []config.JDKInfo{}
}

func (m *MockConfigRepository) ListDetectedJDKs() []config.JDKInfo {
	if m.ListDetectedJDKsFunc != nil {
		return m.ListDetectedJDKsFunc()
	}
	return []config.JDKInfo{}
}

func (m *MockConfigRepository) AddInstalledJDK(info config.JDKInfo) error {
	if m.AddInstalledJDKFunc != nil {
		return m.AddInstalledJDKFunc(info)
	}
	return nil
}

func (m *MockConfigRepository) RemoveInstalledJDK(name string) error {
	if m.RemoveInstalledJDKFunc != nil {
		return m.RemoveInstalledJDKFunc(name)
	}
	return nil
}

func (m *MockConfigRepository) AddDetectedJDK(info config.JDKInfo) error {
	if m.AddDetectedJDKFunc != nil {
		return m.AddDetectedJDKFunc(info)
	}
	return nil
}

func (m *MockConfigRepository) ClearDetectedJDKs() error {
	if m.ClearDetectedJDKsFunc != nil {
		return m.ClearDetectedJDKsFunc()
	}
	return nil
}

func (m *MockConfigRepository) GetGradleCurrent() string {
	if m.GetGradleCurrentFunc != nil {
		return m.GetGradleCurrentFunc()
	}
	return ""
}

func (m *MockConfigRepository) SetGradleCurrent(name string) error {
	if m.SetGradleCurrentFunc != nil {
		return m.SetGradleCurrentFunc(name)
	}
	return nil
}

func (m *MockConfigRepository) ListInstalledGradles() []config.GradleInfo {
	if m.ListInstalledGradlesFunc != nil {
		return m.ListInstalledGradlesFunc()
	}
	return []config.GradleInfo{}
}

func (m *MockConfigRepository) ListDetectedGradles() []config.GradleInfo {
	if m.ListDetectedGradlesFunc != nil {
		return m.ListDetectedGradlesFunc()
	}
	return []config.GradleInfo{}
}

func (m *MockConfigRepository) AddInstalledGradle(info config.GradleInfo) error {
	if m.AddInstalledGradleFunc != nil {
		return m.AddInstalledGradleFunc(info)
	}
	return nil
}

func (m *MockConfigRepository) RemoveInstalledGradle(name string) error {
	if m.RemoveInstalledGradleFunc != nil {
		return m.RemoveInstalledGradleFunc(name)
	}
	return nil
}

func (m *MockConfigRepository) AddDetectedGradle(info config.GradleInfo) error {
	if m.AddDetectedGradleFunc != nil {
		return m.AddDetectedGradleFunc(info)
	}
	return nil
}

func (m *MockConfigRepository) ClearDetectedGradles() error {
	if m.ClearDetectedGradlesFunc != nil {
		return m.ClearDetectedGradlesFunc()
	}
	return nil
}

// MockPrompter is a mock implementation of the Prompter interface
type MockPrompter struct {
	ConfirmFunc func(message string, defaultValue bool) bool
	SelectFunc  func(message string, options []string, defaultValue string) string
	InputFunc   func(message string, defaultValue string) string
}

func (m *MockPrompter) Confirm(message string, defaultValue bool) bool {
	if m.ConfirmFunc != nil {
		return m.ConfirmFunc(message, defaultValue)
	}
	return defaultValue
}

func (m *MockPrompter) Select(message string, options []string, defaultValue string) string {
	if m.SelectFunc != nil {
		return m.SelectFunc(message, options, defaultValue)
	}
	return defaultValue
}

func (m *MockPrompter) Input(message string, defaultValue string) string {
	if m.InputFunc != nil {
		return m.InputFunc(message, defaultValue)
	}
	return defaultValue
}
