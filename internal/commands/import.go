package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/faisalahmedsifat/architect/internal/importers"
	"github.com/faisalahmedsifat/architect/internal/models"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func ImportCmd() *cobra.Command {
	var (
		format    string
		merge     bool
		overwrite bool
	)

	cmd := &cobra.Command{
		Use:   "import [file]",
		Short: "Import API specification from external formats",
		Long: `Import API specifications from various formats including:
- OpenAPI 3.0 (JSON/YAML)
- Postman Collections (JSON) [Coming Soon]
- Existing Architect specifications (YAML)

The import will convert the external format to Architect's specification format.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImport(args[0], format, merge, overwrite)
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "", "Force specific format (openapi, postman, architect)")
	cmd.Flags().BoolVarP(&merge, "merge", "m", false, "Merge with existing specification instead of replacing")
	cmd.Flags().BoolVarP(&overwrite, "overwrite", "o", false, "Overwrite existing files without confirmation")

	return cmd
}

func runImport(filename, format string, merge, overwrite bool) error {
	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filename)
	}

	// Check if .architect directory exists
	if _, err := os.Stat(".architect"); os.IsNotExist(err) {
		return fmt.Errorf(".architect/ directory not found. Run 'architect init' first")
	}

	// Create importer factory
	factory := &importers.ImporterFactory{}

	// Detect format if not specified
	if format == "" {
		var err error
		format, err = factory.DetectFormat(filename)
		if err != nil {
			return fmt.Errorf("failed to detect format: %w", err)
		}
		color.Blue("🔍 Detected format: %s", format)
	}

	// Create appropriate importer
	importer, err := factory.CreateImporter(format)
	if err != nil {
		return fmt.Errorf("failed to create importer: %w", err)
	}

	// Import the API specification
	color.Blue("📥 Importing from %s...", filename)
	importedAPI, err := importer.Import(filename)
	if err != nil {
		return fmt.Errorf("failed to import: %w", err)
	}

	// Validate imported API
	if err := importer.Validate(importedAPI); err != nil {
		return fmt.Errorf("imported API is invalid: %w", err)
	}

	// Handle merge vs replace
	var finalAPI *models.API
	if merge {
		color.Blue("🔄 Merging with existing specification...")
		finalAPI, err = mergeWithExisting(importedAPI)
		if err != nil {
			return fmt.Errorf("failed to merge: %w", err)
		}
	} else {
		finalAPI = importedAPI
	}

	// Check if files exist and need overwrite confirmation
	if !overwrite {
		if _, err := os.Stat(".architect/api.yaml"); err == nil {
			if !merge {
				color.Yellow("⚠️  .architect/api.yaml already exists. Use --overwrite to replace or --merge to combine")
				return fmt.Errorf("file exists, operation cancelled")
			}
		}
	}

	// Write the API specification
	if err := writeAPISpec(finalAPI); err != nil {
		return fmt.Errorf("failed to write API specification: %w", err)
	}

	// Generate basic project.md if it doesn't exist
	if _, err := os.Stat(".architect/project.md"); os.IsNotExist(err) {
		if err := writeBasicProjectMd(importedAPI); err != nil {
			color.Yellow("⚠️  Failed to create project.md: %v", err)
		} else {
			color.Green("📝 Created basic project.md")
		}
	}

	// Show import summary
	color.Green("✅ Successfully imported %d endpoints from %s", len(finalAPI.Endpoints), filename)

	// List imported endpoints
	if len(finalAPI.Endpoints) > 0 {
		color.Blue("\n📋 Imported endpoints:")
		for _, endpoint := range finalAPI.Endpoints {
			authIndicator := ""
			if endpoint.Auth {
				authIndicator = " 🔒"
			}
			fmt.Printf("  %s %s%s\n",
				color.CyanString(endpoint.Method),
				endpoint.Path,
				authIndicator)
		}
	}

	// Sync cursor rules
	color.Blue("\n🔄 Syncing Cursor rules...")
	return runSync(nil, []string{})
}

func mergeWithExisting(importedAPI *models.API) (*models.API, error) {
	// Try to load existing API
	existingAPI := &models.API{
		BaseURL:   "/api/v1",
		AuthType:  "none",
		Endpoints: []models.Endpoint{},
	}

	if _, err := os.Stat(".architect/api.yaml"); err == nil {
		// File exists, try to parse it
		content, err := os.ReadFile(".architect/api.yaml")
		if err != nil {
			return nil, fmt.Errorf("failed to read existing api.yaml: %w", err)
		}

		if err := yaml.Unmarshal(content, existingAPI); err != nil {
			return nil, fmt.Errorf("failed to parse existing api.yaml: %w", err)
		}
	}

	// Merge strategy:
	// 1. Keep existing base URL if imported one is default
	// 2. Upgrade auth type if imported is more secure
	// 3. Merge endpoints (avoid duplicates by path+method)

	mergedAPI := &models.API{
		BaseURL:   existingAPI.BaseURL,
		AuthType:  existingAPI.AuthType,
		Endpoints: []models.Endpoint{},
	}

	// Update base URL if imported has non-default value
	if importedAPI.BaseURL != "" && importedAPI.BaseURL != "/api/v1" {
		mergedAPI.BaseURL = importedAPI.BaseURL
	}

	// Update auth type if imported is not "none"
	if importedAPI.AuthType != "" && importedAPI.AuthType != "none" {
		mergedAPI.AuthType = importedAPI.AuthType
	}

	// Create endpoint map for deduplication
	endpointMap := make(map[string]models.Endpoint)

	// Add existing endpoints
	for _, endpoint := range existingAPI.Endpoints {
		key := fmt.Sprintf("%s:%s", endpoint.Method, endpoint.Path)
		endpointMap[key] = endpoint
	}

	// Add imported endpoints (will overwrite duplicates)
	for _, endpoint := range importedAPI.Endpoints {
		key := fmt.Sprintf("%s:%s", endpoint.Method, endpoint.Path)
		endpointMap[key] = endpoint
	}

	// Convert map back to slice
	for _, endpoint := range endpointMap {
		mergedAPI.Endpoints = append(mergedAPI.Endpoints, endpoint)
	}

	return mergedAPI, nil
}

func writeAPISpec(api *models.API) error {
	// Marshal to YAML
	apiData, err := yaml.Marshal(api)
	if err != nil {
		return fmt.Errorf("failed to marshal API: %w", err)
	}

	// Write to file
	if err := os.WriteFile(".architect/api.yaml", apiData, 0644); err != nil {
		return fmt.Errorf("failed to write api.yaml: %w", err)
	}

	color.Green("💾 Updated .architect/api.yaml")
	return nil
}

func writeBasicProjectMd(api *models.API) error {
	// Determine tech stack based on endpoints
	techStack := "- Backend: Other\n- Database: Other\n- Auth: "
	if api.AuthType == "none" {
		techStack += "None"
	} else {
		techStack += strings.Title(api.AuthType)
	}

	// Generate basic content
	content := fmt.Sprintf(`# Imported API Project

## Overview
This project was imported from an external API specification.

## Tech Stack
%s

## Business Logic

### Endpoints Overview
This API contains %d endpoints with base URL: %s

%s

## Authentication
%s
`,
		techStack,
		len(api.Endpoints),
		api.BaseURL,
		generateEndpointSummary(api.Endpoints),
		generateAuthDescription(api.AuthType))

	// Write to file
	return os.WriteFile(".architect/project.md", []byte(content), 0644)
}

func generateEndpointSummary(endpoints []models.Endpoint) string {
	if len(endpoints) == 0 {
		return "No endpoints defined yet."
	}

	var summary strings.Builder
	summary.WriteString("### Endpoint Summary\n")

	// Group by method
	methods := make(map[string][]string)
	for _, endpoint := range endpoints {
		methods[endpoint.Method] = append(methods[endpoint.Method], endpoint.Path)
	}

	for method, paths := range methods {
		summary.WriteString(fmt.Sprintf("- **%s**: %d endpoints\n", method, len(paths)))
	}

	return summary.String()
}

func generateAuthDescription(authType string) string {
	switch authType {
	case "none":
		return "No authentication required for this API."
	case "bearer":
		return "Bearer token authentication required."
	case "basic":
		return "Basic authentication required."
	case "apikey":
		return "API key authentication required."
	default:
		return fmt.Sprintf("%s authentication required.", strings.Title(authType))
	}
}
