package importers

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/faisalahmedsifat/architect/internal/models"
)

// PostmanImporter handles importing Postman collections
type PostmanImporter struct{}

// PostmanCollection represents a simplified Postman collection structure
type PostmanCollection struct {
	Info     PostmanInfo       `json:"info"`
	Item     []PostmanItem     `json:"item"`
	Auth     *PostmanAuth      `json:"auth,omitempty"`
	Variable []PostmanVariable `json:"variable,omitempty"`
}

type PostmanInfo struct {
	PostmanID   string              `json:"_postman_id"`
	Name        string              `json:"name"`
	Description *PostmanDescription `json:"description,omitempty"`
	Schema      string              `json:"schema"`
}

type PostmanItem struct {
	Name        string              `json:"name"`
	Description *PostmanDescription `json:"description,omitempty"`
	Request     *PostmanRequest     `json:"request,omitempty"`
	Item        []PostmanItem       `json:"item,omitempty"` // For folders
	Response    []interface{}       `json:"response,omitempty"`
}

type PostmanRequest struct {
	Method      string              `json:"method"`
	Header      []PostmanHeader     `json:"header,omitempty"`
	Body        *PostmanBody        `json:"body,omitempty"`
	URL         *PostmanURL         `json:"url,omitempty"`
	Auth        *PostmanAuth        `json:"auth,omitempty"`
	Description *PostmanDescription `json:"description,omitempty"`
}

type PostmanDescription struct {
	Content string `json:"content,omitempty"`
	Type    string `json:"type,omitempty"`
}

type PostmanHeader struct {
	Key         string              `json:"key"`
	Value       string              `json:"value"`
	Description *PostmanDescription `json:"description,omitempty"`
	Disabled    bool                `json:"disabled,omitempty"`
}

type PostmanBody struct {
	Mode       string                 `json:"mode"`
	Raw        string                 `json:"raw,omitempty"`
	URLEncoded []PostmanFormParameter `json:"urlencoded,omitempty"`
	FormData   []PostmanFormParameter `json:"formdata,omitempty"`
	Options    map[string]interface{} `json:"options,omitempty"`
}

type PostmanFormParameter struct {
	Key         string              `json:"key"`
	Value       string              `json:"value"`
	Description *PostmanDescription `json:"description,omitempty"`
	Type        string              `json:"type,omitempty"`
	Disabled    bool                `json:"disabled,omitempty"`
}

type PostmanURL struct {
	Raw      string            `json:"raw"`
	Protocol string            `json:"protocol,omitempty"`
	Host     []string          `json:"host,omitempty"`
	Port     string            `json:"port,omitempty"`
	Path     []string          `json:"path,omitempty"`
	Query    []PostmanQuery    `json:"query,omitempty"`
	Variable []PostmanVariable `json:"variable,omitempty"`
}

type PostmanQuery struct {
	Key         string              `json:"key"`
	Value       string              `json:"value"`
	Description *PostmanDescription `json:"description,omitempty"`
	Disabled    bool                `json:"disabled,omitempty"`
}

type PostmanVariable struct {
	Key         string              `json:"key"`
	Value       string              `json:"value"`
	Type        string              `json:"type,omitempty"`
	Description *PostmanDescription `json:"description,omitempty"`
}

type PostmanAuth struct {
	Type   string              `json:"type"`
	Bearer []PostmanAuthBearer `json:"bearer,omitempty"`
	Basic  []PostmanAuthBasic  `json:"basic,omitempty"`
	APIKey []PostmanAuthAPIKey `json:"apikey,omitempty"`
}

type PostmanAuthBearer struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

type PostmanAuthBasic struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

type PostmanAuthAPIKey struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	In    string `json:"in"`
}

// Import parses a Postman collection file and converts it to our internal API model
func (i *PostmanImporter) Import(filename string) (*models.API, error) {
	// Read file
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	// Parse JSON
	var collection PostmanCollection
	if err := json.Unmarshal(content, &collection); err != nil {
		return nil, fmt.Errorf("failed to parse Postman collection JSON: %w", err)
	}

	// Validate it's a Postman collection
	if collection.Info.Schema == "" {
		return nil, fmt.Errorf("invalid Postman collection: missing schema")
	}

	// Convert to our internal format
	api := &models.API{
		BaseURL:   i.extractBaseURL(&collection),
		AuthType:  i.determineAuthType(&collection),
		Endpoints: []models.Endpoint{},
	}

	// Process all items (including nested folders)
	endpoints := i.processItems(collection.Item, &collection)
	api.Endpoints = endpoints

	return api, nil
}

// Validate checks if the imported API is valid
func (i *PostmanImporter) Validate(api *models.API) error {
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
func (i *PostmanImporter) GetSupportedExtensions() []string {
	return []string{".json"}
}

// processItems recursively processes Postman items (handles folders)
func (i *PostmanImporter) processItems(items []PostmanItem, collection *PostmanCollection) []models.Endpoint {
	var endpoints []models.Endpoint

	for _, item := range items {
		if item.Request != nil {
			// It's a request
			endpoint := i.convertRequest(item, collection)
			endpoints = append(endpoints, endpoint)
		} else if len(item.Item) > 0 {
			// It's a folder, process recursively
			subEndpoints := i.processItems(item.Item, collection)
			endpoints = append(endpoints, subEndpoints...)
		}
	}

	return endpoints
}

// convertRequest converts a Postman request to our endpoint format
func (i *PostmanImporter) convertRequest(item PostmanItem, collection *PostmanCollection) models.Endpoint {
	request := item.Request
	endpoint := models.Endpoint{
		Method:      strings.ToUpper(request.Method),
		Description: item.Name,
		Auth:        i.requestRequiresAuth(request, collection),
	}

	// If item has description, prefer it over name
	if request.Description != nil && request.Description.Content != "" {
		endpoint.Description = request.Description.Content
	}

	// Parse URL and path
	if request.URL != nil {
		endpoint.Path = i.extractPath(request.URL, collection)

		// Handle path parameters
		endpoint.Request = &models.EndpointRequest{
			Params: make(map[string]interface{}),
			Query:  make(map[string]interface{}),
			Body:   make(map[string]interface{}),
		}

		// Extract path variables
		for _, variable := range request.URL.Variable {
			endpoint.Request.Params[variable.Key] = "string, required"
		}

		// Extract query parameters
		for _, query := range request.URL.Query {
			if !query.Disabled {
				fieldType := "string, optional"
				if query.Value != "" {
					fieldType = "string, required"
				}
				endpoint.Request.Query[query.Key] = fieldType
			}
		}
	}

	// Handle request body
	if request.Body != nil && endpoint.Method != "GET" && endpoint.Method != "DELETE" {
		if endpoint.Request == nil {
			endpoint.Request = &models.EndpointRequest{
				Params: make(map[string]interface{}),
				Query:  make(map[string]interface{}),
				Body:   make(map[string]interface{}),
			}
		}

		bodyFields := i.parseRequestBody(request.Body)
		for key, value := range bodyFields {
			endpoint.Request.Body[key] = value
		}
	}

	// Set default response for all endpoints
	endpoint.Response = &models.EndpointResponse{
		Status: 200,
		Body:   make(map[string]interface{}),
	}

	// Adjust status code for POST requests
	if endpoint.Method == "POST" {
		endpoint.Response.Status = 201
	} else if endpoint.Method == "DELETE" {
		endpoint.Response.Status = 204
	}

	return endpoint
}

// extractBaseURL extracts base URL from Postman collection
func (i *PostmanImporter) extractBaseURL(collection *PostmanCollection) string {
	// Look for common base URL in variables
	for _, variable := range collection.Variable {
		if strings.Contains(strings.ToLower(variable.Key), "url") ||
			strings.Contains(strings.ToLower(variable.Key), "host") ||
			strings.Contains(strings.ToLower(variable.Key), "base") {

			// Parse URL to extract path
			if parsedURL, err := url.Parse(variable.Value); err == nil {
				if parsedURL.Path != "" && parsedURL.Path != "/" {
					return parsedURL.Path
				}
			}
		}
	}

	// Analyze first few requests to find common base path
	var paths []string
	items := i.getAllRequests(collection.Item)

	for idx, item := range items {
		if idx >= 5 { // Only check first 5 requests
			break
		}
		if item.Request != nil && item.Request.URL != nil {
			path := i.extractPath(item.Request.URL, collection)
			if path != "" {
				paths = append(paths, path)
			}
		}
	}

	// Find common prefix
	if len(paths) > 0 {
		commonPath := i.findCommonPathPrefix(paths)
		if commonPath != "" {
			return commonPath
		}
	}

	return "/api/v1" // Default
}

// getAllRequests recursively collects all requests from items
func (i *PostmanImporter) getAllRequests(items []PostmanItem) []PostmanItem {
	var requests []PostmanItem

	for _, item := range items {
		if item.Request != nil {
			requests = append(requests, item)
		} else if len(item.Item) > 0 {
			subRequests := i.getAllRequests(item.Item)
			requests = append(requests, subRequests...)
		}
	}

	return requests
}

// findCommonPathPrefix finds the common prefix among API paths
func (i *PostmanImporter) findCommonPathPrefix(paths []string) string {
	if len(paths) == 0 {
		return ""
	}

	// Split each path into segments
	var pathSegments [][]string
	for _, path := range paths {
		segments := strings.Split(strings.Trim(path, "/"), "/")
		if len(segments) > 0 && segments[0] != "" {
			pathSegments = append(pathSegments, segments)
		}
	}

	if len(pathSegments) == 0 {
		return ""
	}

	// Find common prefix segments
	var commonSegments []string
	minLength := len(pathSegments[0])
	for _, segments := range pathSegments[1:] {
		if len(segments) < minLength {
			minLength = len(segments)
		}
	}

	for idx := 0; idx < minLength; idx++ {
		segment := pathSegments[0][idx]
		isCommon := true

		for _, segments := range pathSegments[1:] {
			if segments[idx] != segment {
				isCommon = false
				break
			}
		}

		if !isCommon {
			break
		}

		// Skip variable segments like {{variable}}
		if !strings.Contains(segment, "{{") {
			commonSegments = append(commonSegments, segment)
		}
	}

	if len(commonSegments) > 0 {
		return "/" + strings.Join(commonSegments, "/")
	}

	return ""
}

// extractPath extracts the API path from Postman URL
func (i *PostmanImporter) extractPath(postmanURL *PostmanURL, collection *PostmanCollection) string {
	if postmanURL.Raw != "" {
		// Parse raw URL
		if parsedURL, err := url.Parse(postmanURL.Raw); err == nil {
			path := parsedURL.Path

			// Replace Postman variables with path parameters
			path = i.replacePostmanVariables(path, collection)

			return path
		}
	}

	// Fallback: construct from path segments
	if len(postmanURL.Path) > 0 {
		path := "/" + strings.Join(postmanURL.Path, "/")
		path = i.replacePostmanVariables(path, collection)
		return path
	}

	return ""
}

// replacePostmanVariables converts Postman variables to OpenAPI-style path parameters
func (i *PostmanImporter) replacePostmanVariables(path string, collection *PostmanCollection) string {
	// Replace {{variable}} with {variable}
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	path = re.ReplaceAllString(path, "{$1}")

	// Replace collection variables with their default values if available
	for _, variable := range collection.Variable {
		placeholder := "{" + variable.Key + "}"
		if strings.Contains(path, placeholder) && variable.Value != "" {
			// If it's a URL component, keep as parameter
			// Otherwise, replace with actual value
			if i.isPathParameter(variable.Value) {
				continue
			} else {
				path = strings.ReplaceAll(path, placeholder, variable.Value)
			}
		}
	}

	return path
}

// isPathParameter determines if a value looks like a path parameter
func (i *PostmanImporter) isPathParameter(value string) bool {
	// Simple heuristic: if it looks like an ID or parameter
	patterns := []string{
		`^\d+$`,        // numeric ID
		`^[a-f0-9-]+$`, // UUID-like
		`^\{\w+\}$`,    // already a parameter
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, value); matched {
			return true
		}
	}

	return false
}

// determineAuthType analyzes the collection to determine auth type
func (i *PostmanImporter) determineAuthType(collection *PostmanCollection) string {
	// Check collection-level auth
	if collection.Auth != nil {
		return i.mapPostmanAuthType(collection.Auth.Type)
	}

	// Check individual requests for auth
	requests := i.getAllRequests(collection.Item)
	for _, item := range requests {
		if item.Request != nil && item.Request.Auth != nil {
			return i.mapPostmanAuthType(item.Request.Auth.Type)
		}

		// Check headers for auth patterns
		if item.Request != nil {
			for _, header := range item.Request.Header {
				if strings.ToLower(header.Key) == "authorization" {
					if strings.Contains(strings.ToLower(header.Value), "bearer") {
						return "bearer"
					}
					if strings.Contains(strings.ToLower(header.Value), "basic") {
						return "basic"
					}
				}
				if strings.ToLower(header.Key) == "x-api-key" {
					return "apikey"
				}
			}
		}
	}

	return "none"
}

// mapPostmanAuthType maps Postman auth types to our format
func (i *PostmanImporter) mapPostmanAuthType(postmanType string) string {
	switch strings.ToLower(postmanType) {
	case "bearer":
		return "bearer"
	case "basic":
		return "basic"
	case "apikey":
		return "apikey"
	case "oauth1", "oauth2":
		return "bearer" // Map OAuth to bearer
	default:
		return "none"
	}
}

// requestRequiresAuth determines if a specific request requires authentication
func (i *PostmanImporter) requestRequiresAuth(request *PostmanRequest, collection *PostmanCollection) bool {
	// Check request-level auth
	if request.Auth != nil {
		return request.Auth.Type != "noauth" && request.Auth.Type != ""
	}

	// Check collection-level auth
	if collection.Auth != nil {
		return collection.Auth.Type != "noauth" && collection.Auth.Type != ""
	}

	// Check headers for auth
	for _, header := range request.Header {
		if strings.ToLower(header.Key) == "authorization" ||
			strings.ToLower(header.Key) == "x-api-key" {
			return true
		}
	}

	return false
}

// parseRequestBody parses Postman request body into field definitions
func (i *PostmanImporter) parseRequestBody(body *PostmanBody) map[string]interface{} {
	fields := make(map[string]interface{})

	switch body.Mode {
	case "raw":
		// Try to parse JSON
		if body.Raw != "" {
			fields = i.parseJSONBody(body.Raw)
		}

	case "urlencoded":
		for _, param := range body.URLEncoded {
			if !param.Disabled {
				fieldType := "string, optional"
				if param.Value != "" {
					fieldType = "string, required"
				}
				fields[param.Key] = fieldType
			}
		}

	case "formdata":
		for _, param := range body.FormData {
			if !param.Disabled {
				fieldType := "string, optional"
				if param.Type == "file" {
					fieldType = "file, optional"
				}
				if param.Value != "" {
					fieldType = strings.Replace(fieldType, "optional", "required", 1)
				}
				fields[param.Key] = fieldType
			}
		}
	}

	return fields
}

// parseJSONBody attempts to parse JSON body and extract field types
func (i *PostmanImporter) parseJSONBody(rawBody string) map[string]interface{} {
	fields := make(map[string]interface{})

	// Try to parse as JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(rawBody), &jsonData); err != nil {
		// If parsing fails, create a generic body field
		fields["body"] = "object, required"
		return fields
	}

	// Extract field types from JSON
	for key, value := range jsonData {
		fieldType := i.inferJSONFieldType(value)
		fields[key] = fieldType + ", required"
	}

	return fields
}

// inferJSONFieldType infers the field type from JSON value
func (i *PostmanImporter) inferJSONFieldType(value interface{}) string {
	switch v := value.(type) {
	case string:
		// Check for special string formats
		if i.looksLikeUUID(v) {
			return "uuid"
		}
		if i.looksLikeDateTime(v) {
			return "datetime"
		}
		if i.looksLikeEmail(v) {
			return "string" // We treat email as string with validation
		}
		return "string"
	case float64:
		return "number"
	case int, int64:
		return "integer"
	case bool:
		return "boolean"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return "string"
	}
}

// looksLikeUUID checks if string looks like a UUID
func (i *PostmanImporter) looksLikeUUID(s string) bool {
	uuidPattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
	matched, _ := regexp.MatchString(uuidPattern, strings.ToLower(s))
	return matched
}

// looksLikeDateTime checks if string looks like a datetime
func (i *PostmanImporter) looksLikeDateTime(s string) bool {
	patterns := []string{
		`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`, // ISO 8601
		`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`, // SQL datetime
		`\d{4}/\d{2}/\d{2}`,                   // Date
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, s); matched {
			return true
		}
	}

	return false
}

// looksLikeEmail checks if string looks like an email
func (i *PostmanImporter) looksLikeEmail(s string) bool {
	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(emailPattern, s)
	return matched
}
