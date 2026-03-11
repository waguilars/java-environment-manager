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

// GenerateWrapperFunction generates a PowerShell function wrapper for jem use auto-execution
func (g *PowerShellGenerator) GenerateWrapperFunction() string {
	return `function jem {
    param([Parameter(ValueFromRemainingArguments)]$Args)
    if ($Args[0] -eq 'use') {
        $output = & jem $Args --output-env 2>&1
        if ($LASTEXITCODE -eq 0) {
            Invoke-Expression $output
        } else {
            Write-Host $output
        }
    } else {
        & jem @Args
    }
}`
}
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
