package shell

import (
	"fmt"
	"strings"
)

// PowerShellGenerator generates PowerShell scripts
type PowerShellGenerator struct{}

// NewPowerShellGenerator creates a new PowerShell generator
func NewPowerShellGenerator() *PowerShellGenerator {
	return &PowerShellGenerator{}
}

// Name returns the shell name
func (g *PowerShellGenerator) Name() string {
	return "powershell"
}

// GenerateInitScript generates a PowerShell initialization script
func (g *PowerShellGenerator) GenerateInitScript(envVars map[string]string) string {
	var lines []string

	// Add header comment
	lines = append(lines, "# jem environment initialization")

	// Generate $env: statements for each environment variable
	for key, value := range envVars {
		lines = append(lines, fmt.Sprintf(`$env:%s = "%s"`, key, value))
	}

	// Update PATH if JAVA_HOME is set
	if _, hasJavaHome := envVars["JAVA_HOME"]; hasJavaHome {
		lines = append(lines, `$env:PATH = "$env:JAVA_HOME\bin;$env:PATH"`)
	}

	return strings.Join(lines, "\n")
}
