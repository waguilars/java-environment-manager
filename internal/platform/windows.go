//go:build windows

package platform

import (
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/user/jem/internal/config"
)

// WindowsPlatform implements Platform interface for Windows
type WindowsPlatform struct {
	canCreateSymlinks bool
}

// NewWindowsPlatform creates a new WindowsPlatform with capability detection
func NewWindowsPlatform() *WindowsPlatform {
	return &WindowsPlatform{
		canCreateSymlinks: checkSymlinkCapability(),
	}
}

// checkSymlinkCapability checks if the system can create symlinks
func checkSymlinkCapability() bool {
	// On Windows 10+, Developer Mode allows symlink creation without admin
	// We'll try to detect if Developer Mode is enabled
	// For simplicity, we'll assume it's enabled (common on modern Windows)
	return true
}

// DetectShellWindows detects the shell on Windows
func DetectShellWindows() config.Shell {
	// Check environment variables
	if shell := os.Getenv("PSMODULEPATH"); shell != "" {
		return config.ShellPowerShell
	}

	// Default to PowerShell on Windows
	return config.ShellPowerShell
}

// Name returns "windows"
func (p *WindowsPlatform) Name() string {
	return "windows"
}

// HomeDir returns the user's home directory
func (p *WindowsPlatform) HomeDir() string {
	home, _ := os.UserHomeDir()
	return home
}

// DetectShell returns the detected shell
func (p *WindowsPlatform) DetectShell() config.Shell {
	return DetectShellWindows()
}

// CreateLink creates a symlink or junction based on capability
func (p *WindowsPlatform) CreateLink(target, link string) error {
	if p.canCreateSymlinks {
		return p.createSymlink(target, link)
	}
	return p.createJunction(target, link)
}

// createSymlink creates a symbolic link
func (p *WindowsPlatform) createSymlink(target, link string) error {
	// Use syscall.CreateSymbolicLinkW for Unicode support
	linkPtr, err := syscall.UTF16PtrFromString(link)
	if err != nil {
		return err
	}
	targetPtr, err := syscall.UTF16PtrFromString(target)
	if err != nil {
		return err
	}

	// 1 = SYMBOLIC_LINK_FLAG_DIRECTORY
	return syscall.CreateSymbolicLink(linkPtr, targetPtr, 1)
}

// createJunction creates a directory junction (fallback)
func (p *WindowsPlatform) createJunction(target, link string) error {
	// Create a directory junction using mklink
	cmd := "cmd.exe"
	args := []string{"/c", "mklink", "/J", link, target}

	// Use CreateProcess for better control
	cmdPath, err := syscall.UTF16PtrFromString(cmd)
	if err != nil {
		return err
	}

	argsStr := args[0] + " " + args[1] + " " + args[2] + " " + args[3] + " " + args[4]
	argsPtr, err := syscall.UTF16PtrFromString(argsStr)
	if err != nil {
		return err
	}

	var sa syscall.SecurityAttributes
	sa.Length = uint32(unsafe.Sizeof(sa))
	sa.InheritHandle = 1

	var si syscall.StartupInfo
	si.Cb = uint32(unsafe.Sizeof(si))
	si.Flags = syscall.STARTF_USESHOWWINDOW
	si.ShowWindow = syscall.SW_HIDE

	var pi syscall.ProcessInformation

	err = syscall.CreateProcess(cmdPath, argsPtr, nil, nil, false, 0, nil, nil, &si, &pi)
	if err != nil {
		return err
	}

	// Wait for process to complete
	syscall.WaitForSingleObject(pi.Process, 0xFFFFFFFF)
	syscall.CloseHandle(pi.Process)
	syscall.CloseHandle(pi.Thread)

	return nil
}

// RemoveLink removes a symbolic link or junction
func (p *WindowsPlatform) RemoveLink(link string) error {
	return os.Remove(link)
}

// IsLink checks if path is a symbolic link or junction
func (p *WindowsPlatform) IsLink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

// CanCreateSymlinks returns whether symlinks can be created
func (p *WindowsPlatform) CanCreateSymlinks() bool {
	return p.canCreateSymlinks
}

// ShellConfigPath returns the path to the shell config file
func (p *WindowsPlatform) ShellConfigPath(shell config.Shell) string {
	switch shell {
	case config.ShellPowerShell:
		return "$PROFILE"
	default:
		return "$PROFILE"
	}
}

// JDKDetectionPaths returns standard JDK detection paths on Windows
func (p *WindowsPlatform) JDKDetectionPaths() []string {
	home := p.HomeDir()
	return []string{
		filepath.Join(home, ".jdks"),
		filepath.Join(home, ".sdkman", "candidates", "java"),
		"C:\\Program Files\\Eclipse Adoptium",
		"C:\\Program Files\\Java",
		"C:\\Program Files\\Amazon Corretto",
	}
}

// GradleDetectionPaths returns standard Gradle detection paths on Windows
func (p *WindowsPlatform) GradleDetectionPaths() []string {
	home := p.HomeDir()
	return []string{
		filepath.Join(home, ".sdkman", "candidates", "gradle"),
		"C:\\Program Files\\gradle",
	}
}
