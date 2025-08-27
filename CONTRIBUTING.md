# 🤝 Contributing to Architect

Thank you for your interest in contributing to Architect! This document provides guidelines and information for contributors.

## 🚀 Getting Started

### Prerequisites

- Go 1.25.0 or later
- Git
- Basic understanding of CLI development

### Development Setup

```bash
# 1. Fork and clone the repository
git clone https://github.com/faisalahmedsifat/architect.git
cd architect

# 2. Install dependencies
go mod tidy

# 3. Build the project
go build ./cmd/architect

# 4. Run tests
go test ./...

# 5. Install locally for testing
go install ./cmd/architect
```

## 🏗️ Project Structure

```
architect/
├── cmd/architect/           # Main application entry point
├── internal/
│   ├── commands/           # CLI command implementations
│   │   ├── init.go        # Initialize project
│   │   ├── import.go      # Import API specs
│   │   ├── export.go      # Export to different formats
│   │   └── ...
│   ├── importers/         # Format-specific importers
│   │   ├── interface.go   # Importer interface
│   │   ├── openapi.go     # OpenAPI 3.0 importer
│   │   ├── postman.go     # Postman collection importer
│   │   └── architect.go   # Native format importer
│   ├── models/           # Data structures
│   ├── generator/        # Cursor rules generation
│   ├── parser/          # YAML/JSON parsing utilities
│   └── utils/           # Common utilities
├── README.md            # Main documentation
├── go.mod              # Go module definition
└── go.sum              # Go module checksums
```

## 🎯 How to Contribute

### 1. 🐛 Bug Reports

Before submitting a bug report:

1. **Search existing issues** to avoid duplicates
2. **Use the latest version** of Architect
3. **Provide clear reproduction steps**

**Bug Report Template:**
```markdown
## Bug Description
Clear description of the bug

## Steps to Reproduce
1. Run `architect init -n "test" -d "test"`
2. Run `architect import invalid.json`
3. See error...

## Expected Behavior
What should happen

## Actual Behavior  
What actually happens

## Environment
- OS: [e.g., Ubuntu 22.04]
- Go version: [e.g., 1.25.0]
- Architect version: [e.g., 1.0.0]

## Additional Context
Any other relevant information
```

### 2. ✨ Feature Requests

For new features:

1. **Check if it aligns** with Architect's goals
2. **Describe the use case** clearly
3. **Consider implementation complexity**

**Feature Request Template:**
```markdown
## Feature Description
Clear description of the proposed feature

## Use Case
Why is this feature needed? Who would use it?

## Proposed Solution
How should this feature work?

## Alternatives Considered
What other solutions did you consider?

## Additional Context
Mockups, examples, or other relevant information
```

### 3. 🔧 Code Contributions

#### Development Workflow

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Make your changes** following our coding standards
4. **Write/update tests** for your changes
5. **Test thoroughly** including edge cases
6. **Update documentation** if needed
7. **Submit a pull request**

#### Pull Request Guidelines

- **Clear title**: Describe what the PR does in one line
- **Detailed description**: Explain the changes and why they're needed  
- **Link related issues**: Use "Fixes #123" or "Closes #123"
- **Include tests**: All new code should have tests
- **Update docs**: Update README or other docs if needed

**Pull Request Template:**
```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)  
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## How Has This Been Tested?
Describe the tests you ran and their results

## Checklist
- [ ] My code follows the project's style guidelines
- [ ] I have performed a self-review of my code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
```

## 📝 Coding Standards

### Go Style Guidelines

Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) and these additional guidelines:

#### 1. Naming Conventions
```go
// ✅ Good: Clear, descriptive names
func ParseOpenAPISpec(filename string) (*models.API, error)
func (i *OpenAPIImporter) Import(filename string) (*models.API, error)

// ❌ Bad: Unclear abbreviations  
func ParseOAS(f string) (*models.API, error)
func (o *OASImp) Imp(f string) (*models.API, error)
```

#### 2. Error Handling
```go
// ✅ Good: Descriptive error messages
if err != nil {
    return nil, fmt.Errorf("failed to parse OpenAPI specification: %w", err)
}

// ❌ Bad: Generic error messages
if err != nil {
    return nil, err
}
```

#### 3. Package Organization
```go
// ✅ Good: Clear package responsibilities
package importers  // Handles importing from different formats
package commands   // Handles CLI command implementations
package models     // Defines data structures

// ❌ Bad: Mixed responsibilities
package utils      // Contains importers, commands, and models
```

#### 4. Interface Design
```go
// ✅ Good: Small, focused interfaces
type Importer interface {
    Import(filename string) (*models.API, error)
    Validate(api *models.API) error
    GetSupportedExtensions() []string
}

// ❌ Bad: Large interfaces with many responsibilities
type Handler interface {
    Import(string) (*models.API, error)
    Export(string, string) error
    Validate(*models.API) error
    Sync() error
    Watch() error
}
```

### Documentation Standards

#### 1. Package Documentation
```go
// Package importers provides functionality for importing API specifications
// from various formats including OpenAPI 3.0, Postman Collections, and
// Architect's native YAML format.
package importers
```

#### 2. Function Documentation
```go
// Import parses the given file and converts it to our internal API model.
// It automatically detects the format based on file content and extension.
//
// Supported formats:
//   - OpenAPI 3.0 (JSON/YAML)
//   - Postman Collections (JSON)
//   - Architect native format (YAML)
//
// Returns an error if the file cannot be parsed or the format is unsupported.
func (i *OpenAPIImporter) Import(filename string) (*models.API, error) {
```

#### 3. Complex Logic Comments
```go
// Parse OpenAPI paths and convert to our endpoint format.
// OpenAPI uses a nested structure: paths -> path -> method -> operation
// We flatten this to a simple array of endpoints for easier processing.
for path, pathItem := range spec.Paths {
    for method, operation := range pathItem.Operations() {
        endpoint := convertOperation(path, method, operation)
        api.Endpoints = append(api.Endpoints, endpoint)
    }
}
```

## 🧪 Testing Guidelines

### Test Structure

```go
func TestOpenAPIImporter_Import(t *testing.T) {
    tests := []struct {
        name     string
        filename string
        want     *models.API
        wantErr  bool
    }{
        {
            name:     "valid OpenAPI spec",
            filename: "testdata/valid-openapi.json",
            want:     expectedAPI,
            wantErr:  false,
        },
        {
            name:     "invalid JSON",
            filename: "testdata/invalid.json", 
            want:     nil,
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            i := &OpenAPIImporter{}
            got, err := i.Import(tt.filename)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("Import() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("Import() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Test Data

- Store test data in `testdata/` directories
- Use realistic but minimal examples
- Include both positive and negative test cases
- Test edge cases and error conditions

### Integration Tests

```go
func TestRoundTripImportExport(t *testing.T) {
    // Test that import -> export -> import preserves data
    originalAPI := loadTestAPI(t)
    
    // Export to OpenAPI
    exported := exportToOpenAPI(t, originalAPI)
    
    // Import back
    imported := importFromOpenAPI(t, exported)
    
    // Verify data integrity
    assert.Equal(t, originalAPI.BaseURL, imported.BaseURL)
    assert.Equal(t, len(originalAPI.Endpoints), len(imported.Endpoints))
}
```

## 🎯 Specific Contribution Areas

### 1. Adding New Import Formats

To add support for a new API specification format:

1. **Create a new importer** in `internal/importers/`
2. **Implement the Importer interface**
3. **Add format detection** to `DetectFormat()` in `interface.go`
4. **Update the factory** in `CreateImporter()` 
5. **Add comprehensive tests**
6. **Update documentation**

Example:
```go
// internal/importers/swagger2.go
type Swagger2Importer struct{}

func (i *Swagger2Importer) Import(filename string) (*models.API, error) {
    // Implementation here
}

func (i *Swagger2Importer) Validate(api *models.API) error {
    // Validation logic
}

func (i *Swagger2Importer) GetSupportedExtensions() []string {
    return []string{".json", ".yaml", ".yml"}
}
```

### 2. Adding New Export Formats

To add a new export format:

1. **Add the format** to `exporters` package (if needed)
2. **Update export command** in `internal/commands/export.go`
3. **Add format-specific export function**
4. **Add comprehensive tests**
5. **Update documentation**

### 3. Improving CLI Experience

Areas for improvement:
- Better error messages
- Progress indicators for large operations
- Improved command help text
- Shell completions
- Interactive prompts enhancements

### 4. Performance Optimizations

Performance improvement areas:
- Faster parsing for large specifications
- Memory usage optimization
- Concurrent processing
- Caching mechanisms

## 🔍 Code Review Process

### Review Checklist

**Functionality:**
- [ ] Does the code do what it's supposed to do?
- [ ] Are edge cases handled properly?
- [ ] Is error handling comprehensive?

**Code Quality:**
- [ ] Is the code readable and well-documented?
- [ ] Are variable and function names descriptive?
- [ ] Is the code properly structured?

**Testing:**
- [ ] Are there sufficient tests?
- [ ] Do tests cover edge cases?
- [ ] Are integration tests included where appropriate?

**Performance:**
- [ ] Are there any obvious performance issues?
- [ ] Is memory usage reasonable?
- [ ] Are large files handled efficiently?

**Documentation:**
- [ ] Is the README updated if needed?
- [ ] Are code comments adequate?
- [ ] Are examples provided for new features?

### Review Timeline

- **Initial review**: Within 2-3 days
- **Follow-up reviews**: Within 1-2 days
- **Final approval**: When all requirements are met

## 🐛 Debugging Tips

### Common Issues

1. **Import failures**: Check format detection logic
2. **Export errors**: Verify data structure mapping
3. **Test failures**: Ensure test data is up to date
4. **Build issues**: Check Go version and dependencies

### Debugging Tools

```bash
# Enable verbose logging
export ARCHITECT_DEBUG=true

# Run with debugging
go run ./cmd/architect import test.json

# Run specific tests
go test -v ./internal/importers -run TestOpenAPI

# Run with race detection
go test -race ./...

# Profile memory usage
go test -memprofile mem.prof
go tool pprof mem.prof
```

## 📞 Getting Help

- **Documentation**: Check the README and this guide first
- **Issues**: Search existing issues on GitHub
- **Discussions**: Use GitHub Discussions for questions
- **Discord**: Join our community Discord (link in README)

## 🎉 Recognition

Contributors are recognized in:
- **CONTRIBUTORS.md**: All contributors listed
- **Release notes**: Major contributions highlighted  
- **README**: Top contributors featured

Thank you for contributing to Architect! Your efforts help make API development better for everyone. 🚀
