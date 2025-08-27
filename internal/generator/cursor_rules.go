package generator

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/faisalahmedsifat/architect/internal/models"
)

type Generator struct {
	Project        *models.Project
	ProjectContent string
	API            *models.API
}

func New(project *models.Project, api *models.API) *Generator {
	return &Generator{
		Project: project,
		API:     api,
	}
}

func NewFromContent(projectContent string, api *models.API) *Generator {
	return &Generator{
		ProjectContent: projectContent,
		API:            api,
	}
}

func (g *Generator) GenerateCursorRules() string {
	tmpl := `# {{ .ProjectName }} Implementation Guide

## 📁 Source Specifications
All project specifications are in the ` + "`" + `.architect/` + "`" + ` directory:
- **Project Details & Business Logic**: ` + "`" + `.architect/project.md` + "`" + `
- **API Specifications**: ` + "`" + `.architect/api.yaml` + "`" + `

## 🚨 CRITICAL: Always Read Specifications First
Before implementing ANY feature, check:
1. ` + "`" + `.architect/project.md` + "`" + ` for business logic and rules
2. ` + "`" + `.architect/api.yaml` + "`" + ` for exact API contracts

## Project Overview
{{ .ProjectDescription }}

## Authentication
{{ if .RequiresAuth }}All endpoints except ` + "`" + `/auth/*` + "`" + ` require {{ .AuthType }} authentication:
` + "```" + `
Authorization: Bearer <token>
` + "```" + `
{{ else }}No authentication required for this API.{{ end }}

## API Implementation Requirements

### Endpoint Structure
Base URL: ` + "`" + `{{ .BaseURL }}` + "`" + `

### Available Endpoints
Check ` + "`" + `.architect/api.yaml` + "`" + ` for complete specifications.

{{ .EndpointsList }}

## Request/Response Formats

### IMPORTANT: Follow exact schema from ` + "`" + `.architect/api.yaml` + "`" + `

{{ .EndpointExamples }}

## Business Logic Implementation

### CRITICAL: Read ` + "`" + `.architect/project.md` + "`" + ` for all business rules

{{ .BusinessLogicSummary }}

## Error Handling
All errors must follow this format:
` + "```json" + `
{
    "error": {
        "code": "ERROR_CODE",
        "message": "Human readable message",
        "details": {},
        "timestamp": "{{ .CurrentTime }}"
    }
}
` + "```" + `

## Implementation Pattern
` + "```python" + `
# Route Handler
@router.post("{{ .SampleEndpoint }}")
async def handler(
    request: RequestDTO,
    current_user: User = Depends(get_current_user)
):
    # Business logic validation
    # Service layer call
    # Return response matching schema
    pass
` + "```" + `

## Validation Requirements
- All UUIDs must be valid format
- Dates must be ISO 8601
- Email must be valid format
- String length limits as specified in ` + "`" + `.architect/api.yaml` + "`" + `

## Before Committing Code
Always verify:
- [ ] Endpoints match ` + "`" + `.architect/api.yaml` + "`" + ` exactly
- [ ] Business logic follows ` + "`" + `.architect/project.md` + "`" + `
- [ ] Request/response schemas match specifications
- [ ] Error responses use standard format
- [ ] Authentication required where specified
- [ ] All validations implemented

## Quick Reference Commands
` + "```bash" + `
# View project description
cat .architect/project.md

# View API specifications  
cat .architect/api.yaml

# Validate implementation
architect validate
` + "```" + `

---
*This file is auto-generated from ` + "`" + `.architect/` + "`" + ` specifications. Do not edit manually.*
*Last updated: {{ .UpdateTime }}*`

	data := g.prepareTemplateData()

	t := template.Must(template.New("rules").Parse(tmpl))
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return fmt.Sprintf("Error generating rules: %v", err)
	}

	return buf.String()
}

func (g *Generator) prepareTemplateData() map[string]interface{} {
	data := make(map[string]interface{})

	// Basic info
	if g.Project != nil {
		data["ProjectName"] = g.Project.Name
		data["ProjectDescription"] = g.Project.Description
	} else {
		data["ProjectName"] = "Project"
		data["ProjectDescription"] = "See .architect/project.md for details"
	}

	data["BaseURL"] = g.API.BaseURL
	data["AuthType"] = g.API.AuthType
	data["RequiresAuth"] = g.API.AuthType != "none"
	data["CurrentTime"] = time.Now().Format(time.RFC3339)
	data["UpdateTime"] = time.Now().Format("2006-01-02 15:04:05")

	// Endpoints list
	data["EndpointsList"] = g.generateEndpointsList()
	data["EndpointExamples"] = g.generateEndpointExamples()
	data["BusinessLogicSummary"] = g.generateBusinessLogicSummary()

	// Sample endpoint
	if len(g.API.Endpoints) > 0 {
		data["SampleEndpoint"] = g.API.Endpoints[0].Path
	} else {
		data["SampleEndpoint"] = "/api/v1/example"
	}

	return data
}

func (g *Generator) generateEndpointsList() string {
	if len(g.API.Endpoints) == 0 {
		return "No endpoints defined yet."
	}

	var authEndpoints []string
	var publicEndpoints []string

	for _, ep := range g.API.Endpoints {
		line := fmt.Sprintf("- `%s %s` - %s", ep.Method, ep.Path, ep.Description)
		if ep.Auth {
			authEndpoints = append(authEndpoints, line)
		} else {
			publicEndpoints = append(publicEndpoints, line)
		}
	}

	var result strings.Builder
	if len(publicEndpoints) > 0 {
		result.WriteString("#### Public Endpoints (No auth required):\n")
		result.WriteString(strings.Join(publicEndpoints, "\n"))
		result.WriteString("\n\n")
	}

	if len(authEndpoints) > 0 {
		result.WriteString("#### Protected Endpoints (Auth required):\n")
		result.WriteString(strings.Join(authEndpoints, "\n"))
	}

	return result.String()
}

func (g *Generator) generateEndpointExamples() string {
	if len(g.API.Endpoints) == 0 {
		return "No endpoint examples available."
	}

	// Show first POST endpoint as example
	for _, ep := range g.API.Endpoints {
		if ep.Method == "POST" && ep.Request != nil && ep.Request.Body != nil {
			return g.formatEndpointExample(ep)
		}
	}

	// If no POST, show first endpoint
	return g.formatEndpointExample(g.API.Endpoints[0])
}

func (g *Generator) formatEndpointExample(ep models.Endpoint) string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("Example - %s:\n", ep.Description))
	result.WriteString("```python\n")
	result.WriteString(fmt.Sprintf("# Request\n%s %s\n", ep.Method, ep.Path))

	if ep.Request != nil && ep.Request.Body != nil {
		result.WriteString("Body: {\n")
		for field, ftype := range ep.Request.Body {
			result.WriteString(fmt.Sprintf("    \"%s\": \"%s\",\n", field, ftype))
		}
		result.WriteString("}\n")
	}

	result.WriteString("\n# Response")
	if ep.Response != nil {
		result.WriteString(fmt.Sprintf(" (%d)\n", ep.Response.Status))
		if ep.Response.Body != nil {
			result.WriteString("{\n")
			for field, ftype := range ep.Response.Body {
				result.WriteString(fmt.Sprintf("    \"%s\": \"%s\",\n", field, ftype))
			}
			result.WriteString("}\n")
		}
	}
	result.WriteString("```")

	return result.String()
}

func (g *Generator) generateBusinessLogicSummary() string {
	if g.Project != nil && len(g.Project.BusinessLogic) > 0 {
		var result strings.Builder
		result.WriteString("Key Rules:\n")
		i := 1
		for title, content := range g.Project.BusinessLogic {
			result.WriteString(fmt.Sprintf("%d. **%s**: %s\n", i, title,
				strings.Split(content, "\n")[0])) // First line only
			i++
		}
		return result.String()
	}
	return "See .architect/project.md for detailed business logic."
}
