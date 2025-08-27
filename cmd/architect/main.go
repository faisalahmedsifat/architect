package main

import (
	"fmt"
	"os"

	"github.com/faisalahmedsifat/architect/internal/commands"
	"github.com/spf13/cobra"
)

var version = "1.0.0"

func main() {
	var rootCmd = &cobra.Command{
		Use:   "architect",
		Short: "Architect - AI Development Specification Manager",
		Long: `Architect creates a specification layer between your project planning 
and AI-assisted development. It ensures AI coding assistants follow your 
exact API contracts and business logic.`,
		Version: version,
	}

	// Add commands
	rootCmd.AddCommand(commands.InitCmd())
	rootCmd.AddCommand(commands.SyncCmd())
	rootCmd.AddCommand(commands.AddEndpointCmd())
	rootCmd.AddCommand(commands.ValidateCmd())
	rootCmd.AddCommand(commands.WatchCmd())
	rootCmd.AddCommand(commands.ShowCmd())
	rootCmd.AddCommand(commands.EditCmd())
	rootCmd.AddCommand(commands.ExportCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
