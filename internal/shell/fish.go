package shell

import (
	"fmt"
	"strings"
)

// FishGenerator generates Fish shell scripts
type FishGenerator struct{}

// NewFishGenerator creates a new Fish generator
func NewFishGenerator() *FishGenerator {
	return &FishGenerator{}
}

// Name returns the shell name
func (g *FishGenerator) Name() string {
	return "fish"
}

// GenerateWrapperFunction generates a Fish function wrapper (not supported - stub)
func (g *FishGenerator) GenerateWrapperFunction() string {
	return "# Fish shell wrapper not supported - use 'jem use default jdk <version>' for persistent changes"
}
func (g *FishGenerator) GenerateInitScript(envVars map[string]string) string {
	var lines []string

	// Add header comment
	lines = append(lines, "# jem environment initialization")

	// Generate set -x statements for each environment variable
	for key, value := range envVars {
		lines = append(lines, fmt.Sprintf(`set -x %s "%s"`, key, value))
	}

	// Update PATH if JAVA_HOME is set
	if _, hasJavaHome := envVars["JAVA_HOME"]; hasJavaHome {
		lines = append(lines, `set -x PATH "$JAVA_HOME/bin" $PATH`)
	}

	return strings.Join(lines, "\n")
}
