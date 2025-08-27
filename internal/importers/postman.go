package importers

import (
	"fmt"

	"github.com/faisalahmedsifat/architect/internal/models"
)

// PostmanImporter handles importing Postman collections
type PostmanImporter struct{}

// Import parses a Postman collection file and converts it to our internal API model
func (i *PostmanImporter) Import(filename string) (*models.API, error) {
	// TODO: Implement Postman collection parsing
	return nil, fmt.Errorf("Postman import not yet implemented")
}

// Validate checks if the imported API is valid
func (i *PostmanImporter) Validate(api *models.API) error {
	if api == nil {
		return fmt.Errorf("API cannot be nil")
	}
	return nil
}

// GetSupportedExtensions returns supported file extensions
func (i *PostmanImporter) GetSupportedExtensions() []string {
	return []string{".json"}
}
