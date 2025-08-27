package commands

import (
	"fmt"
	"os"

	"github.com/faisalahmedsifat/architect/internal/parser"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func ShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Display current specifications",
		Long:  "View your specifications in the terminal",
		RunE:  runShow,
	}

	cmd.Flags().Bool("endpoints", false, "Show only endpoints")
	cmd.Flags().Bool("project", false, "Show only project description")

	return cmd
}

func runShow(cmd *cobra.Command, args []string) error {
	showEndpoints, _ := cmd.Flags().GetBool("endpoints")
	showProject, _ := cmd.Flags().GetBool("project")

	if !showEndpoints && !showProject {
		// Show both
		showProject = true
		showEndpoints = true
	}

	if showProject {
		color.Cyan("📁 .architect/project.md")
		fmt.Println("────────────────────────")
		content, err := os.ReadFile(".architect/project.md")
		if err != nil {
			color.Yellow("No project.md found")
		} else {
			fmt.Println(string(content))
		}
		fmt.Println()
	}

	if showEndpoints {
		api, err := parser.ParseAPIYAML(".architect/api.yaml")
		if err != nil {
			return fmt.Errorf("failed to parse api.yaml: %w", err)
		}

		color.Cyan("API Endpoints:")
		fmt.Println()

		// Simple table display without tablewriter for now
		fmt.Printf("%-4s %-8s %-30s %s\n", "Auth", "Method", "Path", "Description")
		fmt.Println("────────────────────────────────────────────────────────────────")

		for _, endpoint := range api.Endpoints {
			auth := "🔓"
			if endpoint.Auth {
				auth = "🔒"
			}

			method := colorMethod(endpoint.Method)
			fmt.Printf("%-4s %-8s %-30s %s\n", auth, method, endpoint.Path, endpoint.Description)
		}

		fmt.Println("\n🔒 = Requires authentication")
		fmt.Println("🔓 = Public endpoint")
	}

	return nil
}

func colorMethod(method string) string {
	switch method {
	case "GET":
		return color.GreenString(method)
	case "POST":
		return color.YellowString(method)
	case "PUT":
		return color.BlueString(method)
	case "DELETE":
		return color.RedString(method)
	case "PATCH":
		return color.MagentaString(method)
	default:
		return method
	}
}
