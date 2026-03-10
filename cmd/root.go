package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/waguilars/java-environment-manager/internal/menu"
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
		Version: "", // Version is injected at build time via main.Version
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
			// Default behavior: show help
			cmd.Help()
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
	rootCmd.AddCommand(importCommand())
	rootCmd.AddCommand(doctorCommand())
	rootCmd.AddCommand(tuiCommand())
	rootCmd.AddCommand(initCommand())

	return rootCmd
}

// executeMenuAction executes a command based on the menu selection
func executeMenuAction(action string) {
	factory, err := NewCommandFactory()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	ctx := context.Background()

	fmt.Println() // Add newline for better output

	switch action {
	case "setup":
		if err := factory.CreateSetupCommand().Execute(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	case "scan":
		if err := factory.CreateScanCommand().Execute(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	case "list":
		if err := factory.CreateListCommand().Execute(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	case "current":
		if err := factory.CreateCurrentCommand().Execute(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	case "use":
		fmt.Println("Use: jem use <version>")
		fmt.Println("Please specify a JDK version to use.")
	case "install":
		fmt.Println("Use: jem install jdk <version>")
		fmt.Println("Please specify a JDK version to install.")
	}
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

// useCommand creates the use subcommand with jdk/gradle subcommands
func useCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use",
		Short: "Switch to a different version",
		Long: `Switch to a different JDK or Gradle version.

You can use either an installed version or a detected one (it will be automatically imported).

Examples:
  # Session mode - output environment variables
  jem use jdk 21 --output-env
  eval "$(jem use jdk 21 --output-env)"

  # Default mode - update config and symlinks
  jem use jdk 21 --default
  jem use default jdk 21`,
	}

	// Add subcommands
	cmd.AddCommand(useJDKCommand())
	cmd.AddCommand(useGradleCommand())
	cmd.AddCommand(useDefaultCommand())

	return cmd
}

// useJDKCommand creates the 'use jdk' subcommand
func useJDKCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jdk <version>",
		Short: "Switch to a different JDK version",
		Long: `Switch to a different JDK version for the current session.

By default, this updates symlinks to switch the JDK immediately without changing
the default configuration. The change only affects the current session.

Use --default to persist the change as the new default JDK.
Use --output-env to output environment variables for shell eval instead.`,
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
			useCmd := factory.CreateUseCommand()
			useCmd.SetForce(getFlagBool(cmd, "force"))
			useCmd.SetOutputEnv(getFlagBool(cmd, "output-env"))

			// Determine mode based on flags
			// Default: update symlinks for current session only
			// --default: update symlinks and config (persist)
			// --output-env: output exports for eval
			mode := UseModeSessionSymlink
			if getFlagBool(cmd, "output-env") {
				mode = UseModeSession
			} else if getFlagBool(cmd, "default") {
				mode = UseModeDefault
			}

			return useCmd.ExecuteJDK(ctx, args[0], mode)
		},
	}

	// Add flags
	cmd.Flags().Bool("output-env", false, "Output environment variables for shell eval instead of updating symlinks")
	cmd.Flags().Bool("default", false, "Set as default JDK (updates config and symlinks)")

	return cmd
}

// useGradleCommand creates the 'use gradle' subcommand
func useGradleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gradle <version>",
		Short: "Switch to a different Gradle version",
		Long: `Switch to a different Gradle version for the current session.

By default, this updates symlinks to switch the Gradle version immediately without
changing the default configuration. The change only affects the current session.

Use --default to persist the change as the new default Gradle.
Use --output-env to output environment variables for shell eval instead.`,
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
			useCmd := factory.CreateUseCommand()
			useCmd.SetForce(getFlagBool(cmd, "force"))
			useCmd.SetOutputEnv(getFlagBool(cmd, "output-env"))

			// Determine mode based on flags
			// Default: update symlinks for current session only
			// --default: update symlinks and config (persist)
			// --output-env: output exports for eval
			mode := UseModeSessionSymlink
			if getFlagBool(cmd, "output-env") {
				mode = UseModeSession
			} else if getFlagBool(cmd, "default") {
				mode = UseModeDefault
			}

			return useCmd.ExecuteGradle(ctx, args[0], mode)
		},
	}

	// Add flags
	cmd.Flags().Bool("output-env", false, "Output environment variables for shell eval instead of updating symlinks")
	cmd.Flags().Bool("default", false, "Set as default Gradle (updates config and symlinks)")

	return cmd
}

// useDefaultCommand creates the 'use default' subcommand
func useDefaultCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "default",
		Short: "Set default JDK or Gradle version",
		Long: `Set the default JDK or Gradle version.

This updates the configuration and symlinks to make the specified version the default.`,
	}

	// Add subcommands
	cmd.AddCommand(useDefaultJDKCommand())
	cmd.AddCommand(useDefaultGradleCommand())

	return cmd
}

// useDefaultJDKCommand creates the 'use default jdk' subcommand
func useDefaultJDKCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "jdk <version>",
		Short: "Set the default JDK version",
		Long: `Set the default JDK version.

This updates the configuration and symlinks to make the specified JDK the default.`,
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
			useCmd := factory.CreateUseCommand()
			useCmd.SetForce(getFlagBool(cmd, "force"))
			return useCmd.ExecuteJDK(ctx, args[0], UseModeDefault)
		},
	}
}

// useDefaultGradleCommand creates the 'use default gradle' subcommand
func useDefaultGradleCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "gradle <version>",
		Short: "Set the default Gradle version",
		Long: `Set the default Gradle version.

This updates the configuration and symlinks to make the specified Gradle the default.`,
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
			useCmd := factory.CreateUseCommand()
			useCmd.SetForce(getFlagBool(cmd, "force"))
			return useCmd.ExecuteGradle(ctx, args[0], UseModeDefault)
		},
	}
}

// installCommand creates the install subcommand
func installCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install [jdk|gradle] <version>",
		Short: "Install a JDK or Gradle version",
		Long: `Install a JDK or Gradle version from a provider (default: Temurin for JDK, Gradle for Gradle).
		
Examples:
  jem install jdk 21
  jem install jdk --lts
  jem install gradle 8.5
  jem install gradle latest`,
	}

	// Add subcommands
	cmd.AddCommand(installJDKCommand())
	cmd.AddCommand(installGradleCommand())

	return cmd
}

// importCommand creates the import subcommand
func importCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import [jdk|gradle] <path>",
		Short: "Import an external JDK or Gradle installation",
		Long: `Import an external JDK or Gradle installation into jem management.
		
Examples:
  jem import jdk /opt/jdk-21
  jem import gradle /opt/gradle-8.5`,
	}

	// Add subcommands
	cmd.AddCommand(importJDKCommand())
	cmd.AddCommand(importGradleCommand())

	return cmd
}

// importJDKCommand creates the 'import jdk' subcommand
func importJDKCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "jdk <path>",
		Short: "Import an external JDK installation",
		Long: `Import an external JDK installation into jem management.
		
Examples:
  jem import jdk /opt/jdk-21`,
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

			path := args[0]
			importCmd := factory.CreateImportCommand()

			// Use the directory name as the name if not specified
			name := filepath.Base(path)

			return importCmd.ExecuteJDK(ctx, path, name)
		},
	}
}

// importGradleCommand creates the 'import gradle' subcommand
func importGradleCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "gradle <path>",
		Short: "Import an external Gradle installation",
		Long: `Import an external Gradle installation into jem management.
		
Examples:
  jem import gradle /opt/gradle-8.5`,
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

			path := args[0]
			importCmd := factory.CreateImportGradleCommand()

			// Use the directory name as the name if not specified
			name := filepath.Base(path)

			return importCmd.ExecuteGradle(ctx, path, name)
		},
	}
}

// installJDKCommand creates the 'install jdk' subcommand
func installJDKCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "jdk <version>",
		Short: "Install a JDK version",
		Long: `Install a JDK version from a provider (default: Temurin).
		
Examples:
  jem install jdk 21
  jem install jdk --lts`,
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

			version := args[0]
			installCmd := factory.CreateInstallCommand()
			installCmd.SetForce(getFlagBool(cmd, "force"))
			installCmd.SetOnlyLTS(getFlagBool(cmd, "lts"))

			return installCmd.ExecuteJDK(ctx, version)
		},
	}
}

// installGradleCommand creates the 'install gradle' subcommand
func installGradleCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "gradle <version>",
		Short: "Install a Gradle version",
		Long: `Install a Gradle version from the official Gradle provider.
		
Examples:
  jem install gradle 8.5
  jem install gradle latest`,
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

			version := args[0]
			installCmd := factory.CreateInstallCommand()
			installCmd.SetForce(getFlagBool(cmd, "force"))

			return installCmd.ExecuteGradle(ctx, version)
		},
	}
}

// doctorCommand creates the doctor subcommand
func doctorCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose jem environment issues",
		Long: `Run diagnostics to check the jem environment for common issues.

Checks include:
  - Current JDK symlink validity
  - Bin directory existence and contents
  - PATH configuration
  - Version consistency between config and actual Java`,
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
			exitCode := factory.CreateDoctorCommand().Execute()
			if exitCode != 0 {
				os.Exit(exitCode)
			}
			return nil
		},
	}
}

// tuiCommand creates the tui subcommand
func tuiCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Launch interactive menu",
		Long:  `Launch the interactive terminal user interface (TUI) for managing JDKs.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Show interactive menu and get selected action
			action, err := menu.Run()
			if err != nil {
				return fmt.Errorf("error running menu: %w", err)
			}

			// Execute the selected action
			if action != "" {
				executeMenuAction(action)
			}
			return nil
		},
	}
}

// initCommand creates the init subcommand
func initCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [shell]",
		Short: "Initialize shell environment for jem",
		Long: `Initialize the shell environment for jem.

Outputs shell commands to set up JAVA_HOME, GRADLE_HOME, and PATH.
Run this in your shell with: eval "$(jem init)"

Supported shells: bash, zsh, powershell, fish`,
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

			shell, _ := cmd.Flags().GetString("shell")
			return factory.CreateInitCommand().Execute(ctx, shell)
		},
	}

	// Add flags
	cmd.Flags().String("shell", "", "Shell type (bash|zsh|powershell|fish)")

	return cmd
}
