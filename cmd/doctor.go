package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/waguilars/java-environment-manager/internal/config"
	"github.com/waguilars/java-environment-manager/internal/platform"
)

// DoctorCommand handles the 'jem doctor' command
type DoctorCommand struct {
	platform   platform.Platform
	configRepo config.ConfigRepository
}

// CheckResult represents the result of a single diagnostic check
type CheckResult struct {
	Name        string
	Status      CheckStatus
	Message     string
	Remediation string
}

// CheckStatus represents the status of a diagnostic check
type CheckStatus int

const (
	// StatusPass indicates the check passed
	StatusPass CheckStatus = iota
	// StatusWarn indicates a warning (non-critical issue)
	StatusWarn
	// StatusFail indicates a critical failure
	StatusFail
)

// Execute runs diagnostic checks and returns exit code
func (c *DoctorCommand) Execute() int {
	fmt.Println("Running jem diagnostics...")

	results := []CheckResult{
		c.checkCurrentSymlink(),
		c.checkBinDirectory(),
		c.checkPathConfiguration(),
		c.checkVersionConsistency(),
	}

	// Print results
	hasFailures := false
	for _, result := range results {
		c.printResult(result)
		if result.Status == StatusFail {
			hasFailures = true
		}
	}

	if hasFailures {
		fmt.Println("\nSome checks failed. Please review the issues above.")
		return 1
	}
	fmt.Println("\nAll checks passed!")
	return 0
}

// checkCurrentSymlink verifies ~/.jem/jdks/current exists and points to valid directory
func (c *DoctorCommand) checkCurrentSymlink() CheckResult {
	currentLink := filepath.Join(c.platform.HomeDir(), ".jem", "jdks", "current")

	// Check if symlink exists
	info, err := os.Lstat(currentLink)
	if os.IsNotExist(err) {
		return CheckResult{
			Name:        "Current JDK Symlink",
			Status:      StatusWarn,
			Message:     "~/.jem/jdks/current does not exist",
			Remediation: "Run 'jem use jdk <version>' to create it",
		}
	}

	// Check if it's a symlink
	if info.Mode()&os.ModeSymlink == 0 {
		return CheckResult{
			Name:        "Current JDK Symlink",
			Status:      StatusFail,
			Message:     "~/.jem/jdks/current exists but is not a symlink",
			Remediation: "Remove it manually and run 'jem use jdk <version>'",
		}
	}

	// Check if target exists
	target, err := os.Readlink(currentLink)
	if err != nil {
		return CheckResult{
			Name:        "Current JDK Symlink",
			Status:      StatusFail,
			Message:     fmt.Sprintf("Cannot read symlink: %v", err),
			Remediation: "Remove it manually and run 'jem use jdk <version>'",
		}
	}

	if _, err := os.Stat(target); os.IsNotExist(err) {
		return CheckResult{
			Name:        "Current JDK Symlink",
			Status:      StatusFail,
			Message:     fmt.Sprintf("Symlink target does not exist: %s", target),
			Remediation: "Run 'jem use jdk <version>' to fix",
		}
	}

	return CheckResult{
		Name:    "Current JDK Symlink",
		Status:  StatusPass,
		Message: fmt.Sprintf("Points to valid JDK: %s", target),
	}
}

// checkBinDirectory verifies ~/.jem/bin exists and contains java binary
func (c *DoctorCommand) checkBinDirectory() CheckResult {
	binDir := filepath.Join(c.platform.HomeDir(), ".jem", "bin")

	// Check if bin exists
	info, err := os.Lstat(binDir)
	if os.IsNotExist(err) {
		return CheckResult{
			Name:        "Bin Directory",
			Status:      StatusFail,
			Message:     "~/.jem/bin does not exist",
			Remediation: "Run 'jem use jdk <version>' to create it",
		}
	}

	// Check if it's a symlink (expected case)
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(binDir)
		if err != nil {
			return CheckResult{
				Name:        "Bin Directory",
				Status:      StatusFail,
				Message:     "Cannot read bin symlink",
				Remediation: "Remove it manually and run 'jem use jdk <version>'",
			}
		}

		if _, err := os.Stat(target); os.IsNotExist(err) {
			return CheckResult{
				Name:        "Bin Directory",
				Status:      StatusFail,
				Message:     fmt.Sprintf("Bin symlink target does not exist: %s", target),
				Remediation: "Run 'jem use jdk <version>' to fix",
			}
		}

		// Check for java binary
		javaPath := filepath.Join(binDir, "java")
		if _, err := os.Stat(javaPath); os.IsNotExist(err) {
			return CheckResult{
				Name:        "Bin Directory",
				Status:      StatusWarn,
				Message:     "Bin exists but java binary not found",
				Remediation: "Run 'jem use jdk <version>' to fix",
			}
		}

		return CheckResult{
			Name:    "Bin Directory",
			Status:  StatusPass,
			Message: fmt.Sprintf("Contains valid Java binaries (%s)", target),
		}
	}

	// It's a regular directory (also acceptable)
	javaPath := filepath.Join(binDir, "java")
	if _, err := os.Stat(javaPath); os.IsNotExist(err) {
		return CheckResult{
			Name:        "Bin Directory",
			Status:      StatusWarn,
			Message:     "Bin directory exists but java binary not found",
			Remediation: "Run 'jem use jdk <version>' to fix",
		}
	}

	return CheckResult{
		Name:    "Bin Directory",
		Status:  StatusPass,
		Message: "Contains valid Java binaries",
	}
}

// checkPathConfiguration verifies ~/.jem/bin is in PATH with correct priority
func (c *DoctorCommand) checkPathConfiguration() CheckResult {
	path := os.Getenv("PATH")
	homeDir := c.platform.HomeDir()
	jemBin := filepath.Join(homeDir, ".jem", "bin")

	// Check if jem/bin is in PATH
	if !strings.Contains(path, jemBin) && !strings.Contains(path, ".jem/bin") {
		return CheckResult{
			Name:        "PATH Configuration",
			Status:      StatusFail,
			Message:     "~/.jem/bin is not in PATH",
			Remediation: "Run 'jem setup' to configure PATH",
		}
	}

	// Check priority (jem bin should be first)
	// This is a heuristic check
	paths := strings.Split(path, string(os.PathListSeparator))
	if len(paths) > 0 {
		firstPath := paths[0]
		if strings.Contains(firstPath, ".jem/bin") || strings.Contains(firstPath, jemBin) {
			return CheckResult{
				Name:    "PATH Configuration",
				Status:  StatusPass,
				Message: "~/.jem/bin is first in PATH (correct priority)",
			}
		}
	}

	return CheckResult{
		Name:        "PATH Configuration",
		Status:      StatusWarn,
		Message:     "~/.jem/bin is in PATH but not first",
		Remediation: "Re-run 'jem setup' or manually move ~/.jem/bin to PATH start",
	}
}

// checkVersionConsistency compares configured version with actual java -version output
func (c *DoctorCommand) checkVersionConsistency() CheckResult {
	// Get configured version
	configuredVersion := c.configRepo.GetJDKCurrent()
	if configuredVersion == "" {
		return CheckResult{
			Name:    "Version Consistency",
			Status:  StatusWarn,
			Message: "No JDK configured in jem",
		}
	}

	// Get actual java version
	cmd := exec.Command("java", "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return CheckResult{
			Name:        "Version Consistency",
			Status:      StatusWarn,
			Message:     "java command not found or failed to execute",
			Remediation: "Run 'jem use jdk <version>' or check PATH",
		}
	}

	// Parse version from output (handles both OpenJDK and Oracle formats)
	actualVersion := parseJavaVersionFromDoctor(string(output))
	if actualVersion == "" {
		return CheckResult{
			Name:    "Version Consistency",
			Status:  StatusPass,
			Message: fmt.Sprintf("Configured: %s (could not detect actual version)", configuredVersion),
		}
	}

	// Compare versions
	if strings.Contains(configuredVersion, actualVersion) || strings.Contains(actualVersion, configuredVersion) {
		return CheckResult{
			Name:    "Version Consistency",
			Status:  StatusPass,
			Message: fmt.Sprintf("Configured: %s, Active: %s (match)", configuredVersion, actualVersion),
		}
	}

	return CheckResult{
		Name:        "Version Consistency",
		Status:      StatusWarn,
		Message:     fmt.Sprintf("Configured: %s, Active: %s (mismatch)", configuredVersion, actualVersion),
		Remediation: "Check PATH priority or run 'jem doctor'",
	}
}

// parseJavaVersionFromDoctor parses Java version from java -version output
func parseJavaVersionFromDoctor(output string) string {
	// Parse OpenJDK format: openjdk version "21.0.2" 2024-01-16 LTS
	// Parse Oracle format: java version "21.0.2" 2024-01-16 LTS
	lines := strings.Split(output, "\n")
	if len(lines) == 0 {
		return ""
	}

	firstLine := lines[0]

	// Extract version between quotes
	start := strings.Index(firstLine, "\"")
	if start == -1 {
		return ""
	}
	end := strings.Index(firstLine[start+1:], "\"")
	if end == -1 {
		return ""
	}

	return firstLine[start+1 : start+1+end]
}

// extractMajorVersionFromDoctor extracts the major version number from a version string
func extractMajorVersionFromDoctor(version string) string {
	re := regexp.MustCompile(`(\d+)`)
	match := re.FindStringSubmatch(version)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

// printResult formats and prints a check result
func (c *DoctorCommand) printResult(result CheckResult) {
	var statusIcon string
	switch result.Status {
	case StatusPass:
		statusIcon = "✓ [PASS]"
	case StatusWarn:
		statusIcon = "⚠ [WARN]"
	case StatusFail:
		statusIcon = "✗ [FAIL]"
	}

	fmt.Printf("%s %s\n", statusIcon, result.Name)
	fmt.Printf("  %s\n", result.Message)
	if result.Remediation != "" {
		fmt.Printf("  → %s\n", result.Remediation)
	}
	fmt.Println()
}
