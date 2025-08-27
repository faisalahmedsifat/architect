package importers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/faisalahmedsifat/architect/internal/models"
)

// Importer defines the interface for importing API specifications from different formats
type Importer interface {
	// Import parses the given file and converts it to our internal API model
	Import(filename string) (*models.API, error)

	// Validate checks if the imported API is valid
	Validate(api *models.API) error

	// GetSupportedExtensions returns the file extensions this importer supports
	GetSupportedExtensions() []string
}

// ImporterFactory creates the appropriate importer based on file format
type ImporterFactory struct{}

// CreateImporter returns an importer instance based on the format
func (f *ImporterFactory) CreateImporter(format string) (Importer, error) {
	switch format {
	case "openapi", "swagger", "json", "yaml", "yml":
		return &OpenAPIImporter{}, nil
	case "postman":
		return &PostmanImporter{}, nil
	case "architect":
		return &ArchitectImporter{}, nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// DetectFormat attempts to detect the format from file extension and content
func (f *ImporterFactory) DetectFormat(filename string) (string, error) {
	ext := filepath.Ext(filename)

	switch ext {
	case ".json":
		// Need to check content to distinguish between OpenAPI and Postman
		content, err := os.ReadFile(filename)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}

		if strings.Contains(string(content), "openapi") || strings.Contains(string(content), "swagger") {
			return "openapi", nil
		}
		if strings.Contains(string(content), "postman") || strings.Contains(string(content), "collection") {
			return "postman", nil
		}
		return "openapi", nil // Default to OpenAPI for JSON

	case ".yaml", ".yml":
		return "openapi", nil

	default:
		return "", fmt.Errorf("unable to detect format from extension: %s", ext)
	}
}
