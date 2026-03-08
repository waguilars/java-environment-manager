package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/jem/internal/menu"
)

// CLIContext holds context information for CLI operations
type CLIContext struct {
	Verbose        bool
	Force          bool
	NonInteractive bool
	Platform       string
}

// NewCLIContext creates a new CLI context from command flags
func NewCLIContext(cmd *cobra.Command) *CLIContext {
	return &CLIContext{
		Verbose:        getFlagBool(cmd, "verbose"),
		Force:          getFlagBool(cmd, "force"),
		NonInteractive: getFlagBool(cmd, "non-interactive"),
		Platform:       getFlagString(cmd, "platform"),
	}
}

func getFlagBool(cmd *cobra.Command, name string) bool {
	val, _ := cmd.Flags().GetBool(name)
	return val
}

func getFlagString(cmd *cobra.Command, name string) string {
	val, _ := cmd.Flags().GetString(name)
	return val
}

// RootCommand creates the root Cobra command
func RootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "jem",
		Short: "Java Environment Manager - Manage your JDK installations",
		Long: `jem (Java Environment Manager) is a CLI tool for managing multiple
JDK versions on your local development machine.

Supports Windows and Linux with automatic platform detection.`,
		Version: "0.0.1-beta",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Load config and validate environment
			ctx := context.Background()
			factory, err := NewCommandFactory()
			if err != nil {
				return
			}

			// Store factory in context for subcommands
			cmd.SetContext(context.WithValue(ctx, "factory", factory))
		},
		Run: func(cmd *cobra.Command, args []string) {
			// If no arguments provided, show interactive menu
			if len(args) == 0 {
				// Check if non-interactive flag is set
				nonInteractive := getFlagBool(cmd, "non-interactive")
				if nonInteractive {
					cmd.Help()
					return
				}

				// Show interactive menu
				if err := menu.Run(); err != nil {
					fmt.Fprintf(os.Stderr, "Error running menu: %v\n", err)
					os.Exit(1)
				}
			}
		},
	}

	// Add persistent flags
	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose output")
	rootCmd.PersistentFlags().Bool("force", false, "Force operation without prompts")
	rootCmd.PersistentFlags().Bool("non-interactive", false, "Disable interactive prompts")
	rootCmd.PersistentFlags().String("platform", "", "Override platform detection (linux|windows)")

	// Add completion command
	rootCmd.AddCommand(completionCommand())

	// Add subcommands
	rootCmd.AddCommand(setupCommand())
	rootCmd.AddCommand(scanCommand())
	rootCmd.AddCommand(listCommand())
	rootCmd.AddCommand(currentCommand())
	rootCmd.AddCommand(useCommand())
	rootCmd.AddCommand(installCommand())

	return rootCmd
}

// completionCommand creates the completion subcommand
func completionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for jem.

To enable completions for bash:
  source <(jem completion bash)

To enable completions for zsh:
  jem completion zsh > "${fpath[1]}/_jem"

To enable completions for fish:
  jem completion fish > ~/.config/fish/completions/jem.fish

To enable completions for PowerShell:
  jem completion powershell > jem.ps1
  Import-Module ./jem.ps1`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			shell := args[0]
			switch shell {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			default:
				FatalErrorWithHint(
					fmt.Sprintf("Unsupported shell: %s", shell),
					"Supported shells: bash, zsh, fish, powershell",
				)
			}
		},
	}
}

// setupCommand creates the setup subcommand
func setupCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Initialize jem configuration",
		Long: `Initialize jem configuration for your system.
		
Creates the ~/.jem directory structure and configures your shell.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			factory, ok := ctx.Value("factory").(*CommandFactory)
			if !ok || factory == nil {
				var err error
				factory, err = NewCommandFactory()
				if err != nil {
					return err
				}
			}
			return factory.CreateSetupCommand().Execute(ctx)
		},
	}
}

// scanCommand creates the scan subcommand
func scanCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "scan",
		Short: "Scan for JDKs on your system",
		Long: `Scan for JDKs on your system and register them.
		
Detects JDKs in standard locations and adds them to your configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			factory, ok := ctx.Value("factory").(*CommandFactory)
			if !ok || factory == nil {
				var err error
				factory, err = NewCommandFactory()
				if err != nil {
					return err
				}
			}
			return factory.CreateScanCommand().Execute(ctx)
		},
	}
}

// listCommand creates the list subcommand
func listCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed and detected JDKs",
		Long: `List all installed and detected JDKs.
		
Shows both managed (installed by jem) and detected (found on system) JDKs.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			factory, ok := ctx.Value("factory").(*CommandFactory)
			if !ok || factory == nil {
				var err error
				factory, err = NewCommandFactory()
				if err != nil {
					return err
				}
			}
			return factory.CreateListCommand().Execute()
		},
	}
}

// currentCommand creates the current subcommand
func currentCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Show the currently active JDK",
		Long: `Show the currently active JDK and Gradle versions.
		
Displays which JDK is currently being used by your system.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			factory, ok := ctx.Value("factory").(*CommandFactory)
			if !ok || factory == nil {
				var err error
				factory, err = NewCommandFactory()
				if err != nil {
					return err
				}
			}
			return factory.CreateCurrentCommand().Execute()
		},
	}
}

// useCommand creates the use subcommand
func useCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "use <version>",
		Short: "Switch to a different JDK version",
		Long: `Switch to a different JDK version.
		
You can use either an installed JDK or a detected one (it will be automatically imported).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			factory, ok := ctx.Value("factory").(*CommandFactory)
			if !ok || factory == nil {
				var err error
				factory, err = NewCommandFactory()
				if err != nil {
					return err
				}
			}
			return factory.CreateUseCommand().Execute(ctx, args[0])
		},
	}
}

// installCommand creates the install subcommand
func installCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install jdk <version>",
		Short: "Install a JDK version",
		Long: `Install a JDK version from a provider (default: Temurin).
		
Examples:
  jem install jdk 21
  jem install jdk --lts
  jem install jdk temurin-17`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			factory, ok := ctx.Value("factory").(*CommandFactory)
			if !ok || factory == nil {
				var err error
				factory, err = NewCommandFactory()
				if err != nil {
					return err
				}
			}

			version := args[1]
			installCmd := factory.CreateInstallCommand()
			installCmd.SetForce(getFlagBool(cmd, "force"))
			installCmd.SetOnlyLTS(getFlagBool(cmd, "lts"))

			// Parse major version if not using --lts
			if !getFlagBool(cmd, "lts") {
				// Major version is already parsed in findRelease
			}

			return installCmd.Execute(ctx, version)
		},
	}

	cmd.Flags().Bool("lts", false, "Install the latest LTS version")
	cmd.Flags().Bool("force", false, "Force operation without prompts")

	return cmd
}
