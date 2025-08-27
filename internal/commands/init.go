package commands

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/faisalahmedsifat/architect/internal/generator"
	"github.com/faisalahmedsifat/architect/internal/models"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func InitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize project specifications",
		Long:  "Creates .architect/ directory with project specifications",
		RunE:  runInit,
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	color.Cyan("📋 Architect - Project Specification Setup")
	fmt.Println("─────────────────────────────────────────\n")

	// Check if .architect already exists
	if _, err := os.Stat(".architect"); err == nil {
		color.Yellow("⚠️  .architect/ directory already exists")
		overwrite := false
		prompt := &survey.Confirm{
			Message: "Do you want to overwrite existing specifications?",
		}
		survey.AskOne(prompt, &overwrite)
		if !overwrite {
			return nil
		}
	}

	// Collect project information
	project := &models.Project{}
	api := &models.API{
		BaseURL: "/api/v1",
	}

	// Project questions
	var qs = []*survey.Question{
		{
			Name:     "name",
			Prompt:   &survey.Input{Message: "Project name:"},
			Validate: survey.Required,
		},
		{
			Name:     "description",
			Prompt:   &survey.Input{Message: "Brief description:"},
			Validate: survey.Required,
		},
		{
			Name: "backend",
			Prompt: &survey.Select{
				Message: "Tech stack (backend):",
				Options: []string{"FastAPI", "Express", "Django", "Spring Boot", "Rails", "Other"},
				Default: "FastAPI",
			},
		},
		{
			Name: "database",
			Prompt: &survey.Select{
				Message: "Database:",
				Options: []string{"PostgreSQL", "MySQL", "MongoDB", "SQLite", "Other"},
				Default: "PostgreSQL",
			},
		},
		{
			Name: "auth",
			Prompt: &survey.Select{
				Message: "Authentication type:",
				Options: []string{"JWT Bearer", "API Key", "OAuth2", "Basic Auth", "None"},
				Default: "JWT Bearer",
			},
		},
	}

	answers := struct {
		Name        string
		Description string
		Backend     string
		Database    string
		Auth        string
	}{}

	if err := survey.Ask(qs, &answers); err != nil {
		return err
	}

	project.Name = answers.Name
	project.Description = answers.Description
	project.TechStack.Backend = answers.Backend
	project.TechStack.Database = answers.Database
	project.TechStack.Auth = answers.Auth
	api.AuthType = convertAuthType(answers.Auth)

	// Ask about business logic
	addBusinessLogic := false
	prompt := &survey.Confirm{
		Message: "Would you like to add business logic descriptions?",
		Default: true,
	}
	survey.AskOne(prompt, &addBusinessLogic)

	if addBusinessLogic {
		project.BusinessLogic = collectBusinessLogic()
	}

	// Ask about API endpoints
	addEndpoints := false
	endpointPrompt := &survey.Confirm{
		Message: "Would you like to add API endpoints now?",
		Default: true,
	}
	survey.AskOne(endpointPrompt, &addEndpoints)

	if addEndpoints {
		api.Endpoints = collectEndpoints()
	}

	// Create directories
	if err := os.MkdirAll(".architect", 0755); err != nil {
		return fmt.Errorf("failed to create .architect directory: %w", err)
	}

	if err := os.MkdirAll(".cursor/rules", 0755); err != nil {
		return fmt.Errorf("failed to create .cursor/rules directory: %w", err)
	}

	// Save project.md
	projectMD := project.ToMarkdown()
	if err := os.WriteFile(".architect/project.md", []byte(projectMD), 0644); err != nil {
		return fmt.Errorf("failed to write project.md: %w", err)
	}
	color.Green("✅ Created .architect/project.md")

	// Save api.yaml
	apiData, err := yaml.Marshal(api)
	if err != nil {
		return fmt.Errorf("failed to marshal API: %w", err)
	}
	if err := os.WriteFile(".architect/api.yaml", apiData, 0644); err != nil {
		return fmt.Errorf("failed to write api.yaml: %w", err)
	}
	color.Green("✅ Created .architect/api.yaml")

	// Generate cursor rules
	gen := generator.New(project, api)
	rules := gen.GenerateCursorRules()
	if err := os.WriteFile(".cursor/rules/architect.mdc", []byte(rules), 0644); err != nil {
		return fmt.Errorf("failed to write cursor rules: %w", err)
	}
	color.Green("✅ Created .cursor/rules/architect.mdc")

	color.Green("\n🎉 Project specifications initialized!")
	fmt.Println("Next step: Start coding with your AI assistant - it will follow your specs automatically.")

	return nil
}

func collectBusinessLogic() map[string]string {
	logic := make(map[string]string)

	for {
		var title, content string
		titlePrompt := &survey.Input{
			Message: "Business logic title (e.g., 'User Registration'):",
		}
		survey.AskOne(titlePrompt, &title)

		if title == "" {
			break
		}

		contentPrompt := &survey.Multiline{
			Message: "Describe the business logic:",
		}
		survey.AskOne(contentPrompt, &content)

		logic[title] = content

		continuePrompt := &survey.Confirm{
			Message: "Add another business logic section?",
			Default: false,
		}
		var addMore bool
		survey.AskOne(continuePrompt, &addMore)
		if !addMore {
			break
		}
	}

	return logic
}

func collectEndpoints() []models.Endpoint {
	var endpoints []models.Endpoint

	for {
		endpoint := models.Endpoint{}

		// Basic endpoint info
		pathPrompt := &survey.Input{
			Message: "Endpoint path:",
			Default: "/api/v1/",
		}
		survey.AskOne(pathPrompt, &endpoint.Path)

		methodPrompt := &survey.Select{
			Message: "Method:",
			Options: []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		}
		survey.AskOne(methodPrompt, &endpoint.Method)

		descPrompt := &survey.Input{
			Message: "Description:",
		}
		survey.AskOne(descPrompt, &endpoint.Description)

		authPrompt := &survey.Confirm{
			Message: "Requires authentication?",
			Default: true,
		}
		survey.AskOne(authPrompt, &endpoint.Auth)

		// Request body
		if endpoint.Method != "GET" && endpoint.Method != "DELETE" {
			hasBody := false
			bodyPrompt := &survey.Confirm{
				Message: "Define request body?",
				Default: true,
			}
			survey.AskOne(bodyPrompt, &hasBody)

			if hasBody {
				endpoint.Request = &models.EndpointRequest{
					Body: collectFields("request body"),
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
				Body:   collectFields("response"),
			}
			if endpoint.Method == "POST" {
				endpoint.Response.Status = 201
			}
		}

		endpoints = append(endpoints, endpoint)

		// Continue?
		continuePrompt := &survey.Confirm{
			Message: "Add another endpoint?",
			Default: false,
		}
		var addMore bool
		survey.AskOne(continuePrompt, &addMore)
		if !addMore {
			break
		}
	}

	return endpoints
}

func collectFields(context string) map[string]string {
	fields := make(map[string]string)

	for {
		var fieldName, fieldType string

		namePrompt := &survey.Input{
			Message: fmt.Sprintf("Field name for %s:", context),
		}
		survey.AskOne(namePrompt, &fieldName)

		if fieldName == "" {
			break
		}

		typePrompt := &survey.Select{
			Message: "Type:",
			Options: []string{"string", "integer", "boolean", "uuid", "datetime", "object", "array"},
			Default: "string",
		}
		survey.AskOne(typePrompt, &fieldType)

		// Check if required
		required := false
		reqPrompt := &survey.Confirm{
			Message: "Required?",
			Default: true,
		}
		survey.AskOne(reqPrompt, &required)

		fieldDef := fieldType
		if required {
			fieldDef += ", required"
		}

		// Add validation for strings
		if fieldType == "string" {
			var validation string
			valPrompt := &survey.Input{
				Message: "Validation (e.g., 'email', 'min:8', 'max:100'):",
			}
			survey.AskOne(valPrompt, &validation)
			if validation != "" {
				fieldDef += ", " + validation
			}
		}

		fields[fieldName] = fieldDef

		// Continue?
		morePrompt := &survey.Confirm{
			Message: "Add another field?",
			Default: false,
		}
		var addMore bool
		survey.AskOne(morePrompt, &addMore)
		if !addMore {
			break
		}
	}

	return fields
}

func convertAuthType(auth string) string {
	switch auth {
	case "JWT Bearer":
		return "bearer"
	case "API Key":
		return "api_key"
	case "OAuth2":
		return "oauth2"
	case "Basic Auth":
		return "basic"
	default:
		return "none"
	}
}
