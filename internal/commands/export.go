package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/faisalahmedsifat/architect/internal/models"
	"github.com/faisalahmedsifat/architect/internal/parser"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func ExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export specifications",
		Long:  "Export specifications in different formats",
		RunE:  runExport,
	}

	cmd.Flags().String("format", "openapi", "Export format (openapi, markdown, postman)")
	cmd.Flags().String("output", "", "Output file (default: stdout)")

	return cmd
}

func runExport(cmd *cobra.Command, args []string) error {
	format, _ := cmd.Flags().GetString("format")
	output, _ := cmd.Flags().GetString("output")

	api, err := parser.ParseAPIYAML(".architect/api.yaml")
	if err != nil {
		return fmt.Errorf("failed to parse api.yaml: %w", err)
	}

	var content string
	var filename string

	switch format {
	case "openapi":
		content = exportOpenAPI(api)
		if output == "" {
			filename = "openapi.json"
		}
	case "markdown":
		content = exportMarkdown(api)
		if output == "" {
			filename = "API_DOCUMENTATION.md"
		}
	case "postman":
		content = exportPostman(api)
		if output == "" {
			filename = "postman_collection.json"
		}
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if output != "" {
		filename = output
	}

	if filename != "" {
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		color.Green("✅ Exported to %s", filename)
	} else {
		fmt.Print(content)
	}

	return nil
}

func exportOpenAPI(api *models.API) string {
	// Simplified OpenAPI 3.0 export
	openapi := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]string{
			"title":   "API Documentation",
			"version": "1.0.0",
		},
		"servers": []map[string]string{
			{"url": api.BaseURL},
		},
		"paths": make(map[string]interface{}),
	}

	paths := openapi["paths"].(map[string]interface{})

	for _, endpoint := range api.Endpoints {
		path := endpoint.Path
		if _, exists := paths[path]; !exists {
			paths[path] = make(map[string]interface{})
		}

		method := strings.ToLower(endpoint.Method)
		paths[path].(map[string]interface{})[method] = map[string]interface{}{
			"summary":   endpoint.Description,
			"security":  []map[string][]string{},
			"responses": buildResponses(endpoint),
		}

		if endpoint.Auth {
			paths[path].(map[string]interface{})[method].(map[string]interface{})["security"] = []map[string][]string{
				{"bearerAuth": []string{}},
			}
		}

		if endpoint.Request != nil && endpoint.Request.Body != nil {
			paths[path].(map[string]interface{})[method].(map[string]interface{})["requestBody"] = buildRequestBody(endpoint.Request)
		}
	}

	if api.AuthType == "bearer" {
		openapi["components"] = map[string]interface{}{
			"securitySchemes": map[string]interface{}{
				"bearerAuth": map[string]string{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "JWT",
				},
			},
		}
	}

	data, _ := json.MarshalIndent(openapi, "", "  ")
	return string(data)
}

func buildResponses(endpoint models.Endpoint) map[string]interface{} {
	responses := make(map[string]interface{})

	if endpoint.Response != nil {
		status := fmt.Sprintf("%d", endpoint.Response.Status)
		responses[status] = map[string]interface{}{
			"description": "Success",
		}

		if endpoint.Response.Body != nil {
			responses[status].(map[string]interface{})["content"] = map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": buildSchema(endpoint.Response.Body),
				},
			}
		}
	}

	for _, err := range endpoint.Errors {
		status := fmt.Sprintf("%d", err.Status)
		responses[status] = map[string]interface{}{
			"description": err.Message,
		}
	}

	return responses
}

func buildRequestBody(request *models.EndpointRequest) map[string]interface{} {
	return map[string]interface{}{
		"required": true,
		"content": map[string]interface{}{
			"application/json": map[string]interface{}{
				"schema": buildSchema(request.Body),
			},
		},
	}
}

func buildSchema(fields map[string]interface{}) map[string]interface{} {
	schema := map[string]interface{}{
		"type":       "object",
		"properties": make(map[string]interface{}),
		"required":   []string{},
	}

	props := schema["properties"].(map[string]interface{})
	required := []string{}

	for field, def := range fields {
		// Simple field type parsing - convert interface{} to string
		defStr, ok := def.(string)
		if !ok {
			continue // Skip non-string field definitions
		}
		parts := strings.Split(defStr, ",")
		fieldType := strings.TrimSpace(parts[0])

		props[field] = map[string]string{
			"type": mapType(fieldType),
		}

		for _, part := range parts[1:] {
			part = strings.TrimSpace(part)
			if part == "required" {
				required = append(required, field)
			}
		}
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

func mapType(t string) string {
	switch t {
	case "uuid", "datetime":
		return "string"
	case "integer", "number":
		return "number"
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

func exportMarkdown(api *models.API) string {
	var sb strings.Builder

	sb.WriteString("# API Documentation\n\n")
	sb.WriteString("Base URL: `" + api.BaseURL + "`\n\n")

	if api.AuthType != "none" {
		sb.WriteString("## Authentication\n")
		sb.WriteString("This API uses " + api.AuthType + " authentication.\n\n")
	}

	sb.WriteString("## Endpoints\n\n")

	for _, endpoint := range api.Endpoints {
		sb.WriteString("### " + endpoint.Method + " " + endpoint.Path + "\n")
		sb.WriteString(endpoint.Description + "\n\n")

		if endpoint.Auth {
			sb.WriteString("**Authentication Required**\n\n")
		}

		if endpoint.Request != nil && endpoint.Request.Body != nil {
			sb.WriteString("**Request Body:**\n```json\n{\n")
			for field, def := range endpoint.Request.Body {
				sb.WriteString(fmt.Sprintf("  \"%s\": \"%s\",\n", field, def))
			}
			sb.WriteString("}\n```\n\n")
		}

		if endpoint.Response != nil && endpoint.Response.Body != nil {
			sb.WriteString(fmt.Sprintf("**Response (%d):**\n```json\n{\n", endpoint.Response.Status))
			for field, def := range endpoint.Response.Body {
				sb.WriteString(fmt.Sprintf("  \"%s\": \"%s\",\n", field, def))
			}
			sb.WriteString("}\n```\n\n")
		}

		if len(endpoint.Errors) > 0 {
			sb.WriteString("**Errors:**\n")
			for _, err := range endpoint.Errors {
				sb.WriteString(fmt.Sprintf("- %d %s: %s\n", err.Status, err.Code, err.Message))
			}
			sb.WriteString("\n")
		}

		sb.WriteString("---\n\n")
	}

	return sb.String()
}

func exportPostman(api *models.API) string {
	// Simplified Postman collection export
	collection := map[string]interface{}{
		"info": map[string]interface{}{
			"name":   "API Collection",
			"schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
		},
		"item": []interface{}{},
	}

	items := []interface{}{}

	for _, endpoint := range api.Endpoints {
		// Construct proper Postman URL object
		fullURL := api.BaseURL + endpoint.Path
		urlObject := map[string]interface{}{
			"raw": fullURL,
		}

		item := map[string]interface{}{
			"name": endpoint.Description,
			"request": map[string]interface{}{
				"method": endpoint.Method,
				"url":    urlObject,
				"header": []map[string]string{},
			},
		}

		if endpoint.Auth {
			item["request"].(map[string]interface{})["header"] = append(
				item["request"].(map[string]interface{})["header"].([]map[string]string),
				map[string]string{
					"key":   "Authorization",
					"value": "Bearer {{token}}",
				},
			)
		}

		if endpoint.Request != nil && endpoint.Request.Body != nil {
			body := make(map[string]interface{})
			for field := range endpoint.Request.Body {
				body[field] = ""
			}

			bodyJSON, _ := json.Marshal(body)
			item["request"].(map[string]interface{})["body"] = map[string]interface{}{
				"mode": "raw",
				"raw":  string(bodyJSON),
				"options": map[string]interface{}{
					"raw": map[string]string{
						"language": "json",
					},
				},
			}
		}

		items = append(items, item)
	}

	collection["item"] = items

	data, _ := json.MarshalIndent(collection, "", "  ")
	return string(data)
}
