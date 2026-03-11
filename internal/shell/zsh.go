package shell

import (
	"fmt"
	"strings"
)

// ZshGenerator generates Zsh shell scripts
type ZshGenerator struct{}

// NewZshGenerator creates a new Zsh generator
func NewZshGenerator() *ZshGenerator {
	return &ZshGenerator{}
}

// Name returns the shell name
func (g *ZshGenerator) Name() string {
	return "zsh"
}

// GenerateWrapperFunction generates a Zsh function wrapper for jem use auto-execution
func (g *ZshGenerator) GenerateWrapperFunction() string {
	return `jem() {
    case "$1" in
        use)
            shift
            eval "$(command jem use "$@" --output-env)"
            ;;
        *)
            command jem "$@"
            ;;
    esac
}`
}
func (g *ZshGenerator) GenerateInitScript(envVars map[string]string) string {
	var lines []string

	// Add header comment
	lines = append(lines, "# jem environment initialization")

	// Generate export statements for each environment variable
	for key, value := range envVars {
		lines = append(lines, fmt.Sprintf(`export %s="%s"`, key, value))
	}

	// Update PATH to include jem bin directory if JAVA_HOME is set
	if _, hasJavaHome := envVars["JAVA_HOME"]; hasJavaHome {
		lines = append(lines, `export PATH="$JAVA_HOME/bin:$PATH"`)
	}

	return strings.Join(lines, "\n")
}
