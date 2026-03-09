package cmd

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/waguilars/java-environment-manager/internal/config"
	"github.com/waguilars/java-environment-manager/internal/jdk"
)

// ListCommand handles the 'jem list' command
type ListCommand struct {
	configRepo config.ConfigRepository
}

// Execute runs the list command (shows all by default)
func (c *ListCommand) Execute() error {
	c.listJDKs()
	c.listGradles()
	return nil
}

// ExecuteJDK lists only JDKs
func (c *ListCommand) ExecuteJDK() error {
	c.listJDKs()
	return nil
}

// ExecuteGradle lists only Gradles
func (c *ListCommand) ExecuteGradle() error {
	c.listGradles()
	return nil
}

// listJDKs prints all JDKs (installed and detected)
func (c *ListCommand) listJDKs() {
	installedJDKs := c.configRepo.ListInstalledJDKs()
	detectedJDKs := c.configRepo.ListDetectedJDKs()
	currentJDK := c.configRepo.GetJDKCurrent()

	// If no JDKs in config, detect from system
	if len(installedJDKs) == 0 && len(detectedJDKs) == 0 {
		systemJava := jdk.DetectSystemJava()
		if systemJava != nil {
			fmt.Println("\n=== System JDK ===")
			fmt.Printf("%s\t(%s)\t[system]\n", systemJava.Version, systemJava.Path)
			return
		}
	}

	// Create tab writer for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Print installed JDKs
	fmt.Println("\n=== Installed JDKs ===")
	if len(installedJDKs) == 0 {
		fmt.Println("No JDKs installed. Use 'jem install jdk <version>' to install.")
	} else {
		// Sort by version
		sort.Slice(installedJDKs, func(i, j int) bool {
			return installedJDKs[i].Version < installedJDKs[j].Version
		})
		for _, jdkInfo := range installedJDKs {
			currentMarker := ""
			if jdkInfo.Version == currentJDK {
				currentMarker = " *"
			}
			fmt.Fprintf(w, "%s\t(%s)%s\n", jdkInfo.Version, jdkInfo.Path, currentMarker)
		}
		w.Flush()
	}

	// Print detected JDKs
	fmt.Println("\n=== Detected JDKs ===")
	if len(detectedJDKs) == 0 {
		fmt.Println("No JDKs detected. Use 'jem scan' to detect JDKs on your system.")
	} else {
		// Sort by version
		sort.Slice(detectedJDKs, func(i, j int) bool {
			return detectedJDKs[i].Version < detectedJDKs[j].Version
		})
		for _, jdkInfo := range detectedJDKs {
			currentMarker := ""
			if jdkInfo.Version == currentJDK {
				currentMarker = " *"
			}
			fmt.Fprintf(w, "%s\t(%s)%s\n", jdkInfo.Version, jdkInfo.Path, currentMarker)
		}
		w.Flush()
	}
}

// listGradles prints all Gradles (installed and detected)
func (c *ListCommand) listGradles() {
	installedGradles := c.configRepo.ListInstalledGradles()
	detectedGradles := c.configRepo.ListDetectedGradles()
	currentGradle := c.configRepo.GetGradleCurrent()

	// If no Gradles in config, detect from system
	if len(installedGradles) == 0 && len(detectedGradles) == 0 {
		systemGradle := detectSystemGradle()
		if systemGradle != nil {
			fmt.Println("\n=== System Gradle ===")
			fmt.Printf("%s\t(%s)\t[system]\n", systemGradle.Version, systemGradle.Path)
			return
		}
		fmt.Println("\n=== Gradles ===")
		fmt.Println("No Gradles found. Use 'jem scan' to detect Gradles on your system.")
		return
	}

	// Create tab writer for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Print installed Gradles
	fmt.Println("\n=== Installed Gradles ===")
	if len(installedGradles) == 0 {
		fmt.Println("No Gradles installed.")
	} else {
		// Sort by version
		sort.Slice(installedGradles, func(i, j int) bool {
			return installedGradles[i].Version < installedGradles[j].Version
		})
		for _, gradleInfo := range installedGradles {
			currentMarker := ""
			if gradleInfo.Version == currentGradle {
				currentMarker = " *"
			}
			fmt.Fprintf(w, "%s\t(%s)%s\n", gradleInfo.Version, gradleInfo.Path, currentMarker)
		}
		w.Flush()
	}

	// Print detected Gradles
	fmt.Println("\n=== Detected Gradles ===")
	if len(detectedGradles) == 0 {
		fmt.Println("No Gradles detected. Use 'jem scan' to detect Gradles on your system.")
	} else {
		// Sort by version
		sort.Slice(detectedGradles, func(i, j int) bool {
			return detectedGradles[i].Version < detectedGradles[j].Version
		})
		for _, gradleInfo := range detectedGradles {
			currentMarker := ""
			if gradleInfo.Version == currentGradle {
				currentMarker = " *"
			}
			fmt.Fprintf(w, "%s\t(%s)%s\n", gradleInfo.Version, gradleInfo.Path, currentMarker)
		}
		w.Flush()
	}
}

// listCommand creates the list subcommand with jdk/gradle subcommands
func listCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed and detected tools",
		Long: `List installed and detected JDKs and Gradles.

Without a subcommand, shows both JDKs and Gradles.
Use 'list jdk' or 'list gradle' to show only one tool.`,
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

	// Add subcommands
	cmd.AddCommand(listJDKCommand())
	cmd.AddCommand(listGradleCommand())

	return cmd
}

// listJDKCommand creates the 'list jdk' subcommand
func listJDKCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "jdk",
		Short: "List installed and detected JDKs",
		Long:  `List all installed and detected JDKs on your system.`,
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
			return factory.CreateListCommand().ExecuteJDK()
		},
	}
}

// listGradleCommand creates the 'list gradle' subcommand
func listGradleCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "gradle",
		Short: "List installed and detected Gradles",
		Long:  `List all installed and detected Gradles on your system.`,
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
			return factory.CreateListCommand().ExecuteGradle()
		},
	}
}
