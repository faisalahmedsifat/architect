package commands

import (
	"fmt"
	"os"

	"github.com/faisalahmedsifat/architect/internal/generator"
	"github.com/faisalahmedsifat/architect/internal/parser"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func SyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Sync specifications to Cursor rules",
		Long:  "Regenerates .cursor/rules/architect.mdc from your specifications",
		RunE:  runSync,
	}
}

func runSync(cmd *cobra.Command, args []string) error {
	color.Cyan("🔄 Syncing specifications to Cursor rules...\n")

	// Check if .architect exists
	if _, err := os.Stat(".architect"); os.IsNotExist(err) {
		return fmt.Errorf(".architect/ directory not found. Run 'architect init' first")
	}

	// Parse project.md
	color.Blue("📖 Reading .architect/project.md")
	projectContent, err := os.ReadFile(".architect/project.md")
	if err != nil {
		return fmt.Errorf("failed to read project.md: %w", err)
	}

	// Parse api.yaml
	color.Blue("📖 Reading .architect/api.yaml")
	api, err := parser.ParseAPIYAML(".architect/api.yaml")
	if err != nil {
		return fmt.Errorf("failed to parse api.yaml: %w", err)
	}

	// Create .cursor/rules directory if it doesn't exist
	if err := os.MkdirAll(".cursor/rules", 0755); err != nil {
		return fmt.Errorf("failed to create .cursor/rules directory: %w", err)
	}

	// Generate cursor rules
	gen := generator.NewFromContent(string(projectContent), api)
	rules := gen.GenerateCursorRules()

	// Write rules
	if err := os.WriteFile(".cursor/rules/architect.mdc", []byte(rules), 0644); err != nil {
		return fmt.Errorf("failed to write cursor rules: %w", err)
	}

	color.Green("✅ Updated .cursor/rules/architect.mdc")
	color.Green("\n✨ Cursor rules synchronized with latest specifications!")

	return nil
}
