package main

import (
	"fmt"
	"os"

	"github.com/user/jem/cmd"
)

var Version = "0.2.0-beta"

func main() {
	rootCmd := cmd.RootCommand()
	rootCmd.Version = Version

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
