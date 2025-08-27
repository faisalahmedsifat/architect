# 📚 Architect Examples

This directory contains examples and sample files to help you get started with Architect.

## 🏃‍♂️ Quick Start Examples

### 1. Lightning-Fast Setup
```bash
# Create a new API project in 1 second
architect init -n "MyAPI" -d "My awesome API" --quiet
```

### 2. Import Real-World APIs
```bash
# Import the sample OpenAPI specification
architect import examples/sample-openapi.yaml

# Import the sample Postman collection
architect import examples/sample-postman.json
```

### 3. Export to Different Formats
```bash
# Export to OpenAPI for Swagger UI
architect export --format openapi --output docs/swagger.json

# Export to Markdown for documentation
architect export --format markdown --output docs/API.md
```

## 📁 Sample Files

### OpenAPI Examples
- `sample-openapi.yaml` - Basic OpenAPI 3.0 specification
- `petstore-openapi.json` - Classic Petstore API example
- `ecommerce-openapi.yaml` - E-commerce API with authentication

### Postman Examples  
- `sample-postman.json` - Basic Postman collection
- `testing-collection.json` - Collection focused on API testing
- `auth-workflow.json` - Authentication workflow example

### Architect Native
- `sample-architect.yaml` - Native Architect format example
- `microservice.yaml` - Microservice API specification
- `complete-api.yaml` - Full-featured API example

## 🎯 Use Case Examples

### Microservice Setup
```bash
# Quick microservice initialization
architect init \
  -n "UserService" \
  -d "User management microservice" \
  --backend "FastAPI" \
  --database "PostgreSQL" \
  --auth "JWT Bearer" \
  --quiet

# Import additional endpoints
architect import examples/auth-endpoints.yaml --merge
```

### Team Collaboration
```bash
# Developer A: Create initial spec
architect init -n "TeamAPI" -d "Shared team API"
git add .architect/ .cursor/
git commit -m "📋 Initial API specification"

# Developer B: Import and extend  
git pull && architect sync
architect import examples/payment-endpoints.yaml --merge
```

### Documentation Generation
```bash
# Generate complete documentation suite
architect export --format openapi --output docs/openapi.json
architect export --format markdown --output docs/README.md
architect export --format postman --output testing/collection.json
```

## 🔧 Advanced Examples

### CI/CD Integration
```bash
# .github/workflows/api-spec.yml example
architect init -n "CI-API" -d "CI/CD API" --force --quiet
architect import specs/*.json --merge --quiet
architect export --format openapi --output dist/api.json --quiet
```

### Batch Operations
```bash
# Import multiple microservices
for service in auth users payments; do
  architect import specs/$service.json --merge
done

# Export to all formats
formats=(openapi postman markdown)
for format in "${formats[@]}"; do
  architect export --format $format --output dist/api.$format
done
```

## 🧪 Testing Examples

### Validation Workflow
```bash
# Initialize with specifications
architect init -n "TestAPI" -d "API for testing"

# Import comprehensive test spec
architect import examples/complete-api.yaml

# Validate implementation (when available)
architect validate
```

### Round-Trip Testing
```bash
# Test format conversion integrity
architect import examples/sample-openapi.yaml
architect export --format postman --output temp.json
architect import temp.json --format postman --overwrite
# Should maintain all data integrity! ✅
```

## 🎨 Template Examples

Browse the template files to understand:
- **Project structure**: How to organize your API specifications
- **Endpoint definitions**: Best practices for defining API endpoints  
- **Authentication patterns**: Different auth schemes and implementations
- **Documentation standards**: How to write clear, maintainable specs

## 🚀 Getting Help

- **Main documentation**: See [../README.md](../README.md)
- **Contributing**: See [../CONTRIBUTING.md](../CONTRIBUTING.md)
- **Issues**: [GitHub Issues](https://github.com/faisalahmedsifat/architect/issues)

---

Start with the examples that match your use case, then customize them for your specific needs!
