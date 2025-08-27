package importers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/faisalahmedsifat/architect/internal/models"
	"gopkg.in/yaml.v3"
)

// OpenAPIImporter handles importing OpenAPI 3.0 specifications
type OpenAPIImporter struct{}

// OpenAPI represents a simplified OpenAPI 3.0 specification structure
type OpenAPI struct {
	OpenAPI string                 `json:"openapi" yaml:"openapi"`
	Info    OpenAPIInfo            `json:"info" yaml:"info"`
	Servers []OpenAPIServer        `json:"servers,omitempty" yaml:"servers,omitempty"`
	Paths   map[string]OpenAPIPath `json:"paths" yaml:"paths"`
}

type OpenAPIInfo struct {
	Title       string `json:"title" yaml:"title"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Version     string `json:"version" yaml:"version"`
}

type OpenAPIServer struct {
	URL         string `json:"url" yaml:"url"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

type OpenAPIPath map[string]OpenAPIOperation

type OpenAPIOperation struct {
	Summary     string                     `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string                     `json:"description,omitempty" yaml:"description,omitempty"`
	Parameters  []OpenAPIParameter         `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody *OpenAPIRequestBody        `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses   map[string]OpenAPIResponse `json:"responses,omitempty" yaml:"responses,omitempty"`
	Security    []map[string][]string      `json:"security,omitempty" yaml:"security,omitempty"`
	Tags        []string                   `json:"tags,omitempty" yaml:"tags,omitempty"`
}

type OpenAPIParameter struct {
	Name        string      `json:"name" yaml:"name"`
	In          string      `json:"in" yaml:"in"`
	Description string      `json:"description,omitempty" yaml:"description,omitempty"`
	Required    bool        `json:"required,omitempty" yaml:"required,omitempty"`
	Schema      interface{} `json:"schema,omitempty" yaml:"schema,omitempty"`
}

type OpenAPIRequestBody struct {
	Description string                      `json:"description,omitempty" yaml:"description,omitempty"`
	Content     map[string]OpenAPIMediaType `json:"content,omitempty" yaml:"content,omitempty"`
	Required    bool                        `json:"required,omitempty" yaml:"required,omitempty"`
}

type OpenAPIResponse struct {
	Description string                      `json:"description" yaml:"description"`
	Content     map[string]OpenAPIMediaType `json:"content,omitempty" yaml:"content,omitempty"`
}

type OpenAPIMediaType struct {
	Schema interface{} `json:"schema,omitempty" yaml:"schema,omitempty"`
}

// Import parses an OpenAPI file and converts it to our internal API model
func (i *OpenAPIImporter) Import(filename string) (*models.API, error) {
	// Read file
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	// Parse based on file extension
	var openAPI OpenAPI
	ext := filepath.Ext(filename)

	switch ext {
	case ".json":
		if err := json.Unmarshal(content, &openAPI); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(content, &openAPI); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}

	// Convert to our internal format
	api := &models.API{
		BaseURL:   i.extractBaseURL(openAPI.Servers),
		AuthType:  i.determineAuthType(openAPI),
		Endpoints: []models.Endpoint{},
	}

	// Convert paths to endpoints
	for path, pathItem := range openAPI.Paths {
		for method, operation := range pathItem {
			endpoint := i.convertOperation(path, strings.ToUpper(method), operation)
			api.Endpoints = append(api.Endpoints, endpoint)
		}
	}

	return api, nil
}

// Validate checks if the imported API is valid
func (i *OpenAPIImporter) Validate(api *models.API) error {
	if api == nil {
		return fmt.Errorf("API cannot be nil")
	}

	if api.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}

	// Validate endpoints
	for idx, endpoint := range api.Endpoints {
		if endpoint.Path == "" {
			return fmt.Errorf("endpoint %d: path is required", idx)
		}
		if endpoint.Method == "" {
			return fmt.Errorf("endpoint %d: method is required", idx)
		}
	}

	return nil
}

// GetSupportedExtensions returns supported file extensions
func (i *OpenAPIImporter) GetSupportedExtensions() []string {
	return []string{".json", ".yaml", ".yml"}
}

// extractBaseURL extracts base URL from servers array
func (i *OpenAPIImporter) extractBaseURL(servers []OpenAPIServer) string {
	if len(servers) == 0 {
		return "/api/v1" // Default
	}

	// Use first server URL
	url := servers[0].URL

	// Clean up the URL to extract just the path
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		// Extract path from full URL
		parts := strings.SplitN(url, "/", 4)
		if len(parts) >= 4 {
			return "/" + parts[3]
		}
		return "/api/v1"
	}

	// Already a path
	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}

	return url
}

// determineAuthType analyzes the OpenAPI spec to determine auth type
func (i *OpenAPIImporter) determineAuthType(openAPI OpenAPI) string {
	// Simple heuristic: check if any endpoint has security requirements
	for _, pathItem := range openAPI.Paths {
		for _, operation := range pathItem {
			if len(operation.Security) > 0 {
				return "bearer" // Default to bearer if security is present
			}
		}
	}
	return "none"
}

// convertOperation converts an OpenAPI operation to our endpoint format
func (i *OpenAPIImporter) convertOperation(path, method string, operation OpenAPIOperation) models.Endpoint {
	endpoint := models.Endpoint{
		Path:        path,
		Method:      method,
		Description: operation.Summary,
		Auth:        len(operation.Security) > 0,
	}

	// If no summary, use description
	if endpoint.Description == "" {
		endpoint.Description = operation.Description
	}

	// Convert request parameters and body
	if operation.RequestBody != nil || len(operation.Parameters) > 0 {
		endpoint.Request = &models.EndpointRequest{
			Params: make(map[string]string),
			Query:  make(map[string]string),
			Body:   make(map[string]string),
		}

		// Handle parameters
		for _, param := range operation.Parameters {
			paramType := i.convertSchemaType(param.Schema)
			if param.Required {
				paramType += ", required"
			} else {
				paramType += ", optional"
			}

			switch param.In {
			case "path":
				endpoint.Request.Params[param.Name] = paramType
			case "query":
				endpoint.Request.Query[param.Name] = paramType
			}
		}

		// Handle request body
		if operation.RequestBody != nil {
			bodyFields := i.extractSchemaFields(operation.RequestBody.Content)
			for name, fieldType := range bodyFields {
				endpoint.Request.Body[name] = fieldType
			}
		}
	}

	// Convert responses
	if len(operation.Responses) > 0 {
		// Use first successful response (200, 201, etc.)
		for statusCode, response := range operation.Responses {
			if strings.HasPrefix(statusCode, "2") { // 2xx responses
				endpoint.Response = &models.EndpointResponse{
					Status: i.parseStatusCode(statusCode),
					Body:   i.extractSchemaFields(response.Content),
				}
				break
			}
		}
	}

	return endpoint
}

// convertSchemaType converts OpenAPI schema types to our format
func (i *OpenAPIImporter) convertSchemaType(schema interface{}) string {
	if schema == nil {
		return "string"
	}

	// Handle map[string]interface{} from JSON parsing
	if schemaMap, ok := schema.(map[string]interface{}); ok {
		if typeVal, exists := schemaMap["type"]; exists {
			if typeStr, ok := typeVal.(string); ok {
				switch typeStr {
				case "integer", "number":
					return "integer"
				case "boolean":
					return "boolean"
				case "array":
					return "array"
				case "object":
					return "object"
				default:
					return "string"
				}
			}
		}

		// Check for format field for more specific types
		if formatVal, exists := schemaMap["format"]; exists {
			if formatStr, ok := formatVal.(string); ok {
				switch formatStr {
				case "uuid":
					return "uuid"
				case "date-time":
					return "datetime"
				case "email":
					return "string" // We treat email as string with validation
				}
			}
		}
	}

	return "string"
}

// extractSchemaFields extracts field definitions from content schemas
func (i *OpenAPIImporter) extractSchemaFields(content map[string]OpenAPIMediaType) map[string]string {
	fields := make(map[string]string)

	// Look for application/json content first
	for contentType, mediaType := range content {
		if strings.Contains(contentType, "json") {
			fields = i.parseSchemaProperties(mediaType.Schema)
			break
		}
	}

	// If no JSON content, use first available
	if len(fields) == 0 && len(content) > 0 {
		for _, mediaType := range content {
			fields = i.parseSchemaProperties(mediaType.Schema)
			break
		}
	}

	return fields
}

// parseSchemaProperties recursively parses schema properties
func (i *OpenAPIImporter) parseSchemaProperties(schema interface{}) map[string]string {
	fields := make(map[string]string)

	if schemaMap, ok := schema.(map[string]interface{}); ok {
		if properties, exists := schemaMap["properties"]; exists {
			if propMap, ok := properties.(map[string]interface{}); ok {
				// Get required fields
				requiredFields := make(map[string]bool)
				if required, exists := schemaMap["required"]; exists {
					if reqArray, ok := required.([]interface{}); ok {
						for _, field := range reqArray {
							if fieldStr, ok := field.(string); ok {
								requiredFields[fieldStr] = true
							}
						}
					}
				}

				// Convert properties
				for propName, propSchema := range propMap {
					fieldType := i.convertSchemaType(propSchema)
					if requiredFields[propName] {
						fieldType += ", required"
					} else {
						fieldType += ", optional"
					}
					fields[propName] = fieldType
				}
			}
		}
	}

	return fields
}

// parseStatusCode converts string status code to integer
func (i *OpenAPIImporter) parseStatusCode(statusCode string) int {
	switch statusCode {
	case "200":
		return 200
	case "201":
		return 201
	case "204":
		return 204
	case "400":
		return 400
	case "401":
		return 401
	case "403":
		return 403
	case "404":
		return 404
	case "500":
		return 500
	default:
		return 200 // Default
	}
}
