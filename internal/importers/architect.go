package importers

import (
	"fmt"

	"github.com/faisalahmedsifat/architect/internal/models"
	"github.com/faisalahmedsifat/architect/internal/parser"
)

// ArchitectImporter handles importing existing Architect API specifications
type ArchitectImporter struct{}

// Import parses an Architect API file and returns it as our internal API model
func (i *ArchitectImporter) Import(filename string) (*models.API, error) {
	// Leverage existing parser
	return parser.ParseAPIYAML(filename)
}

// Validate checks if the imported API is valid
func (i *ArchitectImporter) Validate(api *models.API) error {
	if api == nil {
		return fmt.Errorf("API cannot be nil")
	}

	if api.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}

	return nil
}

// GetSupportedExtensions returns supported file extensions
func (i *ArchitectImporter) GetSupportedExtensions() []string {
	return []string{".yaml", ".yml"}
}
