package shell

import (
	"fmt"
	"strings"
)

// BashGenerator generates Bash shell scripts
type BashGenerator struct{}

// NewBashGenerator creates a new Bash generator
func NewBashGenerator() *BashGenerator {
	return &BashGenerator{}
}

// Name returns the shell name
func (g *BashGenerator) Name() string {
	return "bash"
}

// GenerateWrapperFunction generates a Bash function wrapper for jem use auto-execution
func (g *BashGenerator) GenerateWrapperFunction() string {
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
func (g *BashGenerator) GenerateInitScript(envVars map[string]string) string {
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
