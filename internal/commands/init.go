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
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize project specifications",
		Long:  "Creates .architect/ directory with project specifications",
		RunE:  runInit,
	}

	// Add flags for non-interactive mode
	cmd.Flags().StringP("name", "n", "", "Project name")
	cmd.Flags().StringP("description", "d", "", "Brief description")
	cmd.Flags().String("backend", "FastAPI", "Tech stack backend (FastAPI, Express, Django, Spring Boot, Rails, Other)")
	cmd.Flags().String("database", "PostgreSQL", "Database (PostgreSQL, MySQL, MongoDB, SQLite, Other)")
	cmd.Flags().String("auth", "JWT Bearer", "Authentication type (JWT Bearer, API Key, OAuth2, Basic Auth, None)")
	cmd.Flags().Bool("no-business-logic", false, "Skip adding business logic descriptions")
	cmd.Flags().Bool("no-endpoints", false, "Skip adding API endpoints")
	cmd.Flags().BoolP("force", "f", false, "Overwrite existing specifications without confirmation")
	cmd.Flags().Bool("quiet", false, "Suppress output and use all defaults for missing flags")

	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	// Get flag values
	flagName, _ := cmd.Flags().GetString("name")
	flagDescription, _ := cmd.Flags().GetString("description")
	flagBackend, _ := cmd.Flags().GetString("backend")
	flagDatabase, _ := cmd.Flags().GetString("database")
	flagAuth, _ := cmd.Flags().GetString("auth")
	flagNoBusinessLogic, _ := cmd.Flags().GetBool("no-business-logic")
	flagNoEndpoints, _ := cmd.Flags().GetBool("no-endpoints")
	flagForce, _ := cmd.Flags().GetBool("force")
	flagQuiet, _ := cmd.Flags().GetBool("quiet")

	// Determine if we're running in interactive mode
	isInteractive := flagName == "" || flagDescription == ""

	if !flagQuiet {
		color.Cyan("📋 Architect - Project Specification Setup")
		fmt.Println("─────────────────────────────────────────\n")
	}

	// Check if .architect already exists
	if _, err := os.Stat(".architect"); err == nil {
		if flagForce {
			// Force overwrite, no prompt needed
		} else if isInteractive {
			color.Yellow("⚠️  .architect/ directory already exists")
			overwrite := false
			prompt := &survey.Confirm{
				Message: "Do you want to overwrite existing specifications?",
			}
			survey.AskOne(prompt, &overwrite)
			if !overwrite {
				return nil
			}
		} else {
			return fmt.Errorf(".architect/ directory already exists. Use --force to overwrite")
		}
	}

	// Collect project information
	project := &models.Project{}
	api := &models.API{
		BaseURL: "/api/v1",
	}

	// Get project information from flags or interactive prompts
	var projectName, projectDescription, backend, database, auth string

	if isInteractive {
		// Interactive mode - use surveys for missing information
		var qs []*survey.Question

		if flagName == "" {
			qs = append(qs, &survey.Question{
				Name:     "name",
				Prompt:   &survey.Input{Message: "Project name:"},
				Validate: survey.Required,
			})
		}

		if flagDescription == "" {
			qs = append(qs, &survey.Question{
				Name:     "description",
				Prompt:   &survey.Input{Message: "Brief description:"},
				Validate: survey.Required,
			})
		}

		// Always ask for optional fields in interactive mode if not provided
		qs = append(qs, &survey.Question{
			Name: "backend",
			Prompt: &survey.Select{
				Message: "Tech stack (backend):",
				Options: []string{"FastAPI", "Express", "Django", "Spring Boot", "Rails", "Other"},
				Default: flagBackend,
			},
		})

		qs = append(qs, &survey.Question{
			Name: "database",
			Prompt: &survey.Select{
				Message: "Database:",
				Options: []string{"PostgreSQL", "MySQL", "MongoDB", "SQLite", "Other"},
				Default: flagDatabase,
			},
		})

		qs = append(qs, &survey.Question{
			Name: "auth",
			Prompt: &survey.Select{
				Message: "Authentication type:",
				Options: []string{"JWT Bearer", "API Key", "OAuth2", "Basic Auth", "None"},
				Default: flagAuth,
			},
		})

		answers := make(map[string]interface{})
		if err := survey.Ask(qs, &answers); err != nil {
			return err
		}

		// Use answers or flags
		if flagName != "" {
			projectName = flagName
		} else {
			projectName = answers["name"].(string)
		}

		if flagDescription != "" {
			projectDescription = flagDescription
		} else {
			projectDescription = answers["description"].(string)
		}

		backend = answers["backend"].(string)
		database = answers["database"].(string)
		auth = answers["auth"].(string)
	} else {
		// Non-interactive mode - use flags only
		projectName = flagName
		projectDescription = flagDescription
		backend = flagBackend
		database = flagDatabase
		auth = flagAuth
	}

	project.Name = projectName
	project.Description = projectDescription
	project.TechStack.Backend = backend
	project.TechStack.Database = database
	project.TechStack.Auth = auth
	api.AuthType = convertAuthType(auth)

	// Handle business logic
	addBusinessLogic := false
	if !flagNoBusinessLogic {
		if isInteractive && !flagQuiet {
			prompt := &survey.Confirm{
				Message: "Would you like to add business logic descriptions?",
				Default: true,
			}
			survey.AskOne(prompt, &addBusinessLogic)
		}
	}

	if addBusinessLogic && !flagQuiet {
		project.BusinessLogic = collectBusinessLogic()
	}

	// Handle API endpoints
	addEndpoints := false
	if !flagNoEndpoints {
		if isInteractive && !flagQuiet {
			endpointPrompt := &survey.Confirm{
				Message: "Would you like to add API endpoints now?",
				Default: true,
			}
			survey.AskOne(endpointPrompt, &addEndpoints)
		}
	}

	if addEndpoints && !flagQuiet {
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
	if !flagQuiet {
		color.Green("✅ Created .architect/project.md")
	}

	// Save api.yaml
	apiData, err := yaml.Marshal(api)
	if err != nil {
		return fmt.Errorf("failed to marshal API: %w", err)
	}
	if err := os.WriteFile(".architect/api.yaml", apiData, 0644); err != nil {
		return fmt.Errorf("failed to write api.yaml: %w", err)
	}
	if !flagQuiet {
		color.Green("✅ Created .architect/api.yaml")
	}

	// Generate cursor rules
	gen := generator.New(project, api)
	rules := gen.GenerateCursorRules()
	if err := os.WriteFile(".cursor/rules/architect.mdc", []byte(rules), 0644); err != nil {
		return fmt.Errorf("failed to write cursor rules: %w", err)
	}
	if !flagQuiet {
		color.Green("✅ Created .cursor/rules/architect.mdc")

		color.Green("\n🎉 Project specifications initialized!")
		fmt.Println("Next step: Start coding with your AI assistant - it will follow your specs automatically.")
	}

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

func collectFields(context string) map[string]interface{} {
	fields := make(map[string]interface{})

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
