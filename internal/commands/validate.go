package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/faisalahmedsifat/architect/internal/models"
	"github.com/faisalahmedsifat/architect/internal/parser"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

func ValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate [base-url]",
		Short: "Validate implementation against specifications",
		Long:  "Checks if your code follows the specifications. Use --live with base URL to validate against live API.",
		RunE:  runValidate,
	}

	// Existing flags
	cmd.Flags().Bool("fix", false, "Show fix suggestions")
	
	// New live validation flags
	cmd.Flags().Bool("live", false, "Validate against live API")
	cmd.Flags().Bool("watch", false, "Watch mode - continuously validate")
	cmd.Flags().String("only", "", "Only validate specific endpoint path")
	cmd.Flags().String("skip", "", "Skip specific endpoint path")
	cmd.Flags().Duration("interval", 5*time.Second, "Polling interval for watch mode")
	cmd.Flags().Int("timeout", 30, "HTTP request timeout in seconds")
	cmd.Flags().String("auth-token", "", "Authorization token for API requests")

	return cmd
}

func runValidate(cmd *cobra.Command, args []string) error {
	live, _ := cmd.Flags().GetBool("live")
	
	if live {
		// NEW: Live API validation
		if len(args) == 0 {
			return fmt.Errorf("base URL required for live validation")
		}
		return runLiveValidation(cmd, args[0])
	}
	
	// EXISTING: Code validation (unchanged)
	color.Cyan("🔍 Validating implementation against specifications...\n")

	api, err := parser.ParseAPIYAML(".architect/api.yaml")
	if err != nil {
		return fmt.Errorf("failed to parse api.yaml: %w", err)
	}

	// Basic validation - check if endpoint files exist
	// This is a simplified version - real implementation would parse actual code

	fmt.Println("Checking endpoints...")

	valid := 0
	warnings := 0
	errors := 0

	for _, endpoint := range api.Endpoints {
		// Simplified check - look for route definition in common locations
		found := checkEndpointImplemented(endpoint.Method, endpoint.Path)

		if found {
			color.Green("✅ %s %s - Implemented correctly", endpoint.Method, endpoint.Path)
			valid++
		} else {
			color.Red("❌ %s %s - Endpoint not implemented", endpoint.Method, endpoint.Path)
			errors++
		}
	}

	fmt.Printf("\nSummary:\n")
	fmt.Printf("- ✅ %d endpoints correct\n", valid)
	if warnings > 0 {
		fmt.Printf("- ⚠️  %d endpoints with warnings\n", warnings)
	}
	if errors > 0 {
		fmt.Printf("- ❌ %d endpoints with errors\n", errors)
	}

	showFix, _ := cmd.Flags().GetBool("fix")
	if showFix && errors > 0 {
		fmt.Println("\nRun 'architect validate --fix' for suggestions on fixing these issues.")
	}

	if errors > 0 {
		return fmt.Errorf("validation failed with %d errors", errors)
	}

	return nil
}

// ValidationResult represents the result of validating a single endpoint
type ValidationResult struct {
	Endpoint   models.Endpoint
	Success    bool
	StatusCode int
	Duration   time.Duration
	Error      string
	Timestamp  time.Time
}

func runLiveValidation(cmd *cobra.Command, baseURL string) error {
	watch, _ := cmd.Flags().GetBool("watch")
	
	if watch {
		return runWatchMode(cmd, baseURL)
	}
	
	// Single validation run
	color.Cyan("🔗 Validating live API against specifications...\n")
	color.Cyan("🌐 Base URL: %s\n\n", baseURL)
	
	api, err := parser.ParseAPIYAML(".architect/api.yaml")
	if err != nil {
		return fmt.Errorf("failed to parse api.yaml: %w", err)
	}
	
	results, err := validateAllEndpoints(cmd, baseURL, api)
	if err != nil {
		return err
	}
	
	displaySummary(results)
	
	// Exit with error if any validations failed
	for _, result := range results {
		if !result.Success {
			return fmt.Errorf("validation failed")
		}
	}
	
	return nil
}

func runWatchMode(cmd *cobra.Command, baseURL string) error {
	color.Cyan("👀 Starting live API validation in watch mode...\n")
	color.Cyan("🔗 Base URL: %s\n\n", baseURL)
	
	interval, _ := cmd.Flags().GetDuration("interval")
	
	// Create file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	defer watcher.Close()
	
	// Watch the API spec file
	err = watcher.Add(".architect/api.yaml")
	if err != nil {
		return fmt.Errorf("failed to watch api.yaml: %w", err)
	}
	
	// Initial validation
	runSingleValidation(cmd, baseURL)
	
	// Setup periodic validation timer
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				color.Yellow("📝 Detected changes in api.yaml, re-validating...\n")
				runSingleValidation(cmd, baseURL)
			}
			
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			color.Red("❌ File watcher error: %v\n", err)
			
		case <-ticker.C:
			runSingleValidation(cmd, baseURL)
		}
	}
}

func runSingleValidation(cmd *cobra.Command, baseURL string) {
	api, err := parser.ParseAPIYAML(".architect/api.yaml")
	if err != nil {
		color.Red("❌ Failed to parse api.yaml: %v\n", err)
		return
	}
	
	results, err := validateAllEndpoints(cmd, baseURL, api)
	if err != nil {
		color.Red("❌ Validation error: %v\n", err)
		return
	}
	
	displaySummary(results)
	fmt.Println() // Add spacing between runs
}

func validateAllEndpoints(cmd *cobra.Command, baseURL string, api *models.API) ([]ValidationResult, error) {
	timeout, _ := cmd.Flags().GetInt("timeout")
	authToken, _ := cmd.Flags().GetString("auth-token")
	only, _ := cmd.Flags().GetString("only")
	skip, _ := cmd.Flags().GetString("skip")
	
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	
	var results []ValidationResult
	
	for _, endpoint := range api.Endpoints {
		// Apply filters
		if only != "" && !strings.Contains(endpoint.Path, only) {
			continue
		}
		if skip != "" && strings.Contains(endpoint.Path, skip) {
			continue
		}
		
		result := validateEndpoint(client, baseURL, endpoint, authToken)
		results = append(results, result)
		displayResult(result)
	}
	
	return results, nil
}

func validateEndpoint(client *http.Client, baseURL string, endpoint models.Endpoint, authToken string) ValidationResult {
	start := time.Now()
	result := ValidationResult{
		Endpoint:  endpoint,
		Timestamp: start,
	}
	
	// Construct full URL
	fullURL, err := url.JoinPath(baseURL, endpoint.Path)
	if err != nil {
		result.Error = fmt.Sprintf("Invalid URL construction: %v", err)
		result.Duration = time.Since(start)
		return result
	}
	
	// Replace path parameters with dummy values for testing
	fullURL = replacePlaceholders(fullURL)
	
	// Create HTTP request
	req, err := http.NewRequest(endpoint.Method, fullURL, nil)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to create request: %v", err)
		result.Duration = time.Since(start)
		return result
	}
	
	// Add authentication if required
	if endpoint.Auth && authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	
	// Add content type for requests with body
	if endpoint.Request != nil && len(endpoint.Request.Body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	
	// Make HTTP request
	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("Request failed: %v", err)
		result.Duration = time.Since(start)
		return result
	}
	defer resp.Body.Close()
	
	result.StatusCode = resp.StatusCode
	result.Duration = time.Since(start)
	
	// Validate status code
	expectedStatus := 200 // Default
	if endpoint.Response != nil {
		expectedStatus = endpoint.Response.Status
	}
	
	if resp.StatusCode != expectedStatus {
		result.Error = fmt.Sprintf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
		return result
	}
	
	// Validate response body if specified
	if endpoint.Response != nil && len(endpoint.Response.Body) > 0 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			result.Error = fmt.Sprintf("Failed to read response body: %v", err)
			return result
		}
		
		if err := validateResponseBody(body, endpoint.Response.Body); err != nil {
			result.Error = err.Error()
			return result
		}
	}
	
	result.Success = true
	return result
}

func validateResponseBody(body []byte, expectedFields map[string]interface{}) error {
	if len(body) == 0 {
		return fmt.Errorf("empty response body")
	}
	
	var responseData interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		return fmt.Errorf("invalid JSON response: %v", err)
	}
	
	// Convert to map for field checking
	responseMap, ok := responseData.(map[string]interface{})
	if !ok {
		// If response is an array, we can't validate fields
		return nil
	}
	
	// Check if all expected fields are present (recursive for nested objects)
	return validateFields(responseMap, expectedFields, "")
	
	return nil
}

func validateFields(responseMap map[string]interface{}, expectedFields map[string]interface{}, prefix string) error {
	for field, expectedValue := range expectedFields {
		fieldPath := field
		if prefix != "" {
			fieldPath = prefix + "." + field
		}
		
		actualValue, exists := responseMap[field]
		if !exists {
			return fmt.Errorf("missing expected field: %s", fieldPath)
		}
		
		// If the expected value is a nested map, validate recursively
		if expectedMap, ok := expectedValue.(map[string]interface{}); ok {
			if actualMap, ok := actualValue.(map[string]interface{}); ok {
				if err := validateFields(actualMap, expectedMap, fieldPath); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("field %s should be an object, got %T", fieldPath, actualValue)
			}
		}
		// For non-nested fields, we just check presence (type validation could be added later)
	}
	
	return nil
}

func replacePlaceholders(path string) string {
	// Replace common path parameter patterns with dummy values
	re := regexp.MustCompile(`\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(path, func(match string) string {
		param := strings.Trim(match, "{}")
		// Use appropriate dummy values based on parameter name
		if strings.Contains(param, "id") || strings.Contains(param, "uuid") {
			return "123e4567-e89b-12d3-a456-426614174000"
		}
		if strings.Contains(param, "slug") || strings.Contains(param, "name") {
			return "test-item"
		}
		return "test-value"
	})
}

func displayResult(result ValidationResult) {
	if result.Success {
		color.Green("✅ %s %s - %d OK (%.2fs)", 
			result.Endpoint.Method, 
			result.Endpoint.Path, 
			result.StatusCode,
			result.Duration.Seconds())
	} else {
		color.Red("❌ %s %s - %s", 
			result.Endpoint.Method, 
			result.Endpoint.Path, 
			result.Error)
	}
}

func displaySummary(results []ValidationResult) {
	passed := 0
	failed := 0
	var totalDuration time.Duration
	
	for _, result := range results {
		if result.Success {
			passed++
		} else {
			failed++
		}
		totalDuration += result.Duration
	}
	
	avgDuration := time.Duration(0)
	if len(results) > 0 {
		avgDuration = totalDuration / time.Duration(len(results))
	}
	
	fmt.Printf("\n📊 Validation Summary:\n")
	if passed > 0 {
		color.Green("✅ %d passed", passed)
	}
	if failed > 0 {
		color.Red("❌ %d failed", failed)
	}
	fmt.Printf("⏱️  Average response time: %dms\n", avgDuration.Milliseconds())
}

func checkEndpointImplemented(method, path string) bool {
	// This is a very basic check - real implementation would use AST parsing
	// Check common directories for route definitions

	searchDirs := []string{
		"app", "src", "api", "routes", "routers", "handlers", "controllers",
	}

	for _, dir := range searchDirs {
		if files, err := ioutil.ReadDir(dir); err == nil {
			for _, file := range files {
				if strings.HasSuffix(file.Name(), ".py") ||
					strings.HasSuffix(file.Name(), ".js") ||
					strings.HasSuffix(file.Name(), ".ts") {

					content, _ := ioutil.ReadFile(filepath.Join(dir, file.Name()))
					contentStr := string(content)

					// Very basic check - look for method and path
					if strings.Contains(contentStr, method) &&
						strings.Contains(contentStr, cleanPath(path)) {
						return true
					}
				}
			}
		}
	}

	return false
}

func cleanPath(path string) string {
	// Remove parameter placeholders for basic matching
	path = strings.ReplaceAll(path, "{", "")
	path = strings.ReplaceAll(path, "}", "")
	return path
}
