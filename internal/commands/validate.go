package commands

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/faisalahmedsifat/architect/internal/parser"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func ValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate implementation against specifications",
		Long:  "Checks if your code follows the specifications",
		RunE:  runValidate,
	}

	cmd.Flags().Bool("fix", false, "Show fix suggestions")

	return cmd
}

func runValidate(cmd *cobra.Command, args []string) error {
	color.Cyan("🔍 Validating implementation against specifications...\n")

	api, err := parser.ParseAPIYAML(".architect/api.yaml")
	if err != nil {
		return fmt.Errorf("failed to parse api.yaml: %w", err)
	}

	// Basic validation - check if endpoint files exist
	// This is a simplified version - real implementation would parse actual code

	fmt.Println("Checking endpoints...")

	valid := 0
	warnings := 0
	errors := 0

	for _, endpoint := range api.Endpoints {
		// Simplified check - look for route definition in common locations
		found := checkEndpointImplemented(endpoint.Method, endpoint.Path)

		if found {
			color.Green("✅ %s %s - Implemented correctly", endpoint.Method, endpoint.Path)
			valid++
		} else {
			color.Red("❌ %s %s - Endpoint not implemented", endpoint.Method, endpoint.Path)
			errors++
		}
	}

	fmt.Printf("\nSummary:\n")
	fmt.Printf("- ✅ %d endpoints correct\n", valid)
	if warnings > 0 {
		fmt.Printf("- ⚠️  %d endpoints with warnings\n", warnings)
	}
	if errors > 0 {
		fmt.Printf("- ❌ %d endpoints with errors\n", errors)
	}

	showFix, _ := cmd.Flags().GetBool("fix")
	if showFix && errors > 0 {
		fmt.Println("\nRun 'architect validate --fix' for suggestions on fixing these issues.")
	}

	if errors > 0 {
		return fmt.Errorf("validation failed with %d errors", errors)
	}

	return nil
}

func checkEndpointImplemented(method, path string) bool {
	// This is a very basic check - real implementation would use AST parsing
	// Check common directories for route definitions

	searchDirs := []string{
		"app", "src", "api", "routes", "routers", "handlers", "controllers",
	}

	for _, dir := range searchDirs {
		if files, err := ioutil.ReadDir(dir); err == nil {
			for _, file := range files {
				if strings.HasSuffix(file.Name(), ".py") ||
					strings.HasSuffix(file.Name(), ".js") ||
					strings.HasSuffix(file.Name(), ".ts") {

					content, _ := ioutil.ReadFile(filepath.Join(dir, file.Name()))
					contentStr := string(content)

					// Very basic check - look for method and path
					if strings.Contains(contentStr, method) &&
						strings.Contains(contentStr, cleanPath(path)) {
						return true
					}
				}
			}
		}
	}

	return false
}

func cleanPath(path string) string {
	// Remove parameter placeholders for basic matching
	path = strings.ReplaceAll(path, "{", "")
	path = strings.ReplaceAll(path, "}", "")
	return path
}
