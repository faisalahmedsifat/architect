package commands

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/faisalahmedsifat/architect/internal/models"
	"github.com/faisalahmedsifat/architect/internal/parser"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func AddEndpointCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add-endpoint",
		Short: "Add new API endpoint",
		Long:  "Interactively add a new endpoint to your specifications",
		RunE:  runAddEndpoint,
	}
}

func runAddEndpoint(cmd *cobra.Command, args []string) error {
	// Check if .architect exists
	if _, err := os.Stat(".architect"); os.IsNotExist(err) {
		return fmt.Errorf(".architect/ directory not found. Run 'architect init' first")
	}

	// Load existing API
	api, err := parser.ParseAPIYAML(".architect/api.yaml")
	if err != nil {
		return fmt.Errorf("failed to parse api.yaml: %w", err)
	}

	// Collect endpoint information
	endpoint := models.Endpoint{}

	pathPrompt := &survey.Input{
		Message: "Endpoint path:",
		Default: api.BaseURL + "/",
	}
	survey.AskOne(pathPrompt, &endpoint.Path)

	methodPrompt := &survey.Select{
		Message: "Method:",
		Options: []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
	}
	survey.AskOne(methodPrompt, &endpoint.Method)

	authPrompt := &survey.Confirm{
		Message: "Requires authentication?",
		Default: true,
	}
	survey.AskOne(authPrompt, &endpoint.Auth)

	descPrompt := &survey.Input{
		Message: "Description:",
	}
	survey.AskOne(descPrompt, &endpoint.Description)

	// Request body for non-GET methods
	if endpoint.Method != "GET" && endpoint.Method != "DELETE" {
		hasBody := false
		bodyPrompt := &survey.Confirm{
			Message: "Define request body?",
			Default: true,
		}
		survey.AskOne(bodyPrompt, &hasBody)

		if hasBody {
			endpoint.Request = &models.EndpointRequest{
				Body: collectEndpointFields("request body"),
			}
		}
	}

	// Response body
	hasResponse := false
	responsePrompt := &survey.Confirm{
		Message: "Define response body?",
		Default: true,
	}
	survey.AskOne(responsePrompt, &hasResponse)

	if hasResponse {
		endpoint.Response = &models.EndpointResponse{
			Status: 200,
			Body:   collectEndpointFields("response"),
		}
		if endpoint.Method == "POST" {
			endpoint.Response.Status = 201
		}
	}

	// Add to API
	api.Endpoints = append(api.Endpoints, endpoint)

	// Save updated API
	apiData, err := yaml.Marshal(api)
	if err != nil {
		return fmt.Errorf("failed to marshal API: %w", err)
	}

	if err := os.WriteFile(".architect/api.yaml", apiData, 0644); err != nil {
		return fmt.Errorf("failed to write api.yaml: %w", err)
	}

	color.Green("✅ Added endpoint to .architect/api.yaml")

	// Sync cursor rules
	fmt.Println()
	return runSync(cmd, args)
}

func collectEndpointFields(context string) map[string]interface{} {
	fields := make(map[string]interface{})

	for {
		var fieldName string
		namePrompt := &survey.Input{
			Message: fmt.Sprintf("Field name for %s (empty to finish):", context),
		}
		survey.AskOne(namePrompt, &fieldName)

		if fieldName == "" {
			break
		}

		var fieldType string
		typePrompt := &survey.Select{
			Message: "Type:",
			Options: []string{"string", "integer", "boolean", "uuid", "datetime", "number", "object", "array"},
			Default: "string",
		}
		survey.AskOne(typePrompt, &fieldType)

		var required bool
		reqPrompt := &survey.Confirm{
			Message: "Required?",
			Default: true,
		}
		survey.AskOne(reqPrompt, &required)

		fieldDef := fieldType
		if required {
			fieldDef += ", required"
		} else {
			fieldDef += ", optional"
		}

		fields[fieldName] = fieldDef

		var addMore bool
		morePrompt := &survey.Confirm{
			Message: "Add another field?",
			Default: true,
		}
		survey.AskOne(morePrompt, &addMore)
		if !addMore {
			break
		}
	}

	return fields
}
