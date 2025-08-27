# Architect CLI - Documentation & Examples 📚

## What is Architect?

**Architect** is a CLI tool that creates a specification layer between your project planning and AI-assisted development. It ensures AI coding assistants (like Cursor) follow your exact API contracts and business logic, preventing architectural drift during implementation.

### The Problem It Solves

When using AI to build applications:
- ❌ AI forgets your API structure after a few files
- ❌ Business logic gets implemented inconsistently  
- ❌ You repeatedly explain the same requirements
- ❌ Request/response contracts drift from original design

### The Solution

Architect maintains your specifications in `.architect/` and auto-generates rules that AI assistants follow, ensuring consistent implementation across your entire codebase.

## Installation

```bash
npm install -g architect-cli
# or
pip install architect-cli
```

## Quick Start

```bash
# 1. Initialize your project specifications
architect init

# 2. AI assistant now sees your specs via .cursor/rules/architect.mdc
# Start coding with Cursor - it follows your specifications automatically!

# 3. Keep specs and rules in sync
architect sync
```

## Core Commands

### `architect init` - Initialize Project

Creates the `.architect/` directory with your project specifications.

```bash
$ architect init

📋 Architect - Project Specification Setup
─────────────────────────────────────────

? Project name: TaskFlow
? Brief description: Task management API for team collaboration
? Tech stack (backend): FastAPI
? Database: PostgreSQL
? Authentication type: JWT Bearer
? Would you like to add API endpoints now? Yes

? Endpoint path: /api/v1/auth/register
? Method: POST
? Requires auth? No
? Description: Register new user
? Define request body? Yes
  ? Field name: email
  ? Type: string
  ? Required? Yes
  ? Validation: email
  ? Add another field? Yes
  ? Field name: password
  ? Type: string  
  ? Required? Yes
  ? Validation: min:8
  ? Add another field? No
? Define response body? Yes
  [... similar prompts ...]
? Add another endpoint? No

✅ Created .architect/project.md
✅ Created .architect/api.yaml
✅ Created .cursor/rules/architect.mdc

🎉 Project specifications initialized!
Next step: Start coding with your AI assistant - it will follow your specs automatically.
```

### `architect sync` - Sync Specifications to Cursor

Regenerates `.cursor/rules/architect.mdc` from your specifications.

```bash
$ architect sync

🔄 Syncing specifications to Cursor rules...
📖 Reading .architect/project.md
📖 Reading .architect/api.yaml
✅ Updated .cursor/rules/architect.mdc

✨ Cursor rules synchronized with latest specifications!
```

**Use when:**
- You manually edit `.architect/` files
- You want to ensure rules are up-to-date
- After pulling changes from git

### `architect add-endpoint` - Add New API Endpoint

Interactively add a new endpoint to your specifications.

```bash
$ architect add-endpoint

? Endpoint path: /api/v1/tasks/{task_id}/comments
? Method: POST
? Requires authentication? Yes
? Description: Add comment to task
? Define request body? Yes
  ? Field name: content
  ? Type: string
  ? Required? Yes
  ? Validation: max:1000
  ? Add another field? No
? Define response body? Yes
  ? Field name: id
  ? Type: uuid
  [...]

✅ Added endpoint to .architect/api.yaml
✅ Updated .cursor/rules/architect.mdc

New endpoint available: POST /api/v1/tasks/{task_id}/comments
```

### `architect validate` - Validate Implementation

Checks if your code follows the specifications (requires code analysis).

```bash
$ architect validate

🔍 Validating implementation against specifications...

Checking endpoints...
✅ POST /api/v1/auth/register - Implemented correctly
✅ POST /api/v1/auth/login - Implemented correctly  
⚠️  GET /api/v1/projects - Missing pagination parameters
❌ POST /api/v1/projects - Response schema mismatch
   Expected: {id, name, slug, description, owner_id}
   Found: {id, name, description}
❌ GET /api/v1/tasks - Endpoint not implemented

Summary:
- ✅ 2 endpoints correct
- ⚠️  1 endpoint with warnings  
- ❌ 2 endpoints with errors

Run 'architect validate --fix' for suggestions on fixing these issues.
```

### `architect watch` - Watch Mode

Auto-syncs rules when specifications change.

```bash
$ architect watch

👀 Watching .architect/ for changes...
[10:34:22] Changed: .architect/api.yaml
[10:34:22] Syncing specifications...
[10:34:23] ✅ Updated .cursor/rules/architect.mdc
[10:45:11] Changed: .architect/project.md
[10:45:11] Syncing specifications...
[10:45:12] ✅ Updated .cursor/rules/architect.mdc

Press Ctrl+C to stop watching
```

### `architect edit` - Edit Specifications

Opens your specifications in the default editor.

```bash
$ architect edit

? What would you like to edit?
> Project description (.architect/project.md)
  API specifications (.architect/api.yaml)
  Both files

[Opens in your $EDITOR]
```

### `architect show` - Display Current Specifications

View your specifications in the terminal.

```bash
$ architect show

📁 .architect/project.md
────────────────────────
# TaskFlow - Task Management System

## Overview
TaskFlow is a task management API that allows teams...
[...]

📁 .architect/api.yaml
────────────────────────
base_url: "/api/v1"
auth_type: "bearer"
endpoints:
  - path: "/auth/register"
    method: POST
[...]

$ architect show --endpoints

API Endpoints:
├── 🔓 POST   /api/v1/auth/register       Register new user
├── 🔓 POST   /api/v1/auth/login          Login user  
├── 🔒 GET    /api/v1/projects            List user's projects
├── 🔒 POST   /api/v1/projects            Create new project
├── 🔒 GET    /api/v1/projects/{id}       Get project details
├── 🔒 GET    /api/v1/projects/{id}/tasks List project tasks
├── 🔒 POST   /api/v1/projects/{id}/tasks Create new task
├── 🔒 PUT    /api/v1/tasks/{id}          Update task
└── 🔒 DELETE /api/v1/tasks/{id}          Delete task

🔒 = Requires authentication
🔓 = Public endpoint
```

### `architect export` - Export Specifications

Export specifications in different formats.

```bash
# Export as OpenAPI/Swagger
$ architect export --format openapi
✅ Exported to openapi.json

# Export as Markdown documentation
$ architect export --format markdown
✅ Exported to API_DOCUMENTATION.md

# Export as Postman collection
$ architect export --format postman
✅ Exported to postman_collection.json
```

## Real-World Use Cases

### Use Case 1: Starting a New Project

```bash
# 1. Initialize with your API design
$ architect init
[Answer prompts about your project]

# 2. Start Cursor/VS Code
$ cursor .

# 3. Ask AI to implement
"Implement the authentication endpoints as specified in the architect rules"
# AI reads .cursor/rules/architect.mdc and implements exactly as specified

# 4. Add more endpoints as needed
$ architect add-endpoint
[Define new endpoint]
# AI immediately knows about the new endpoint
```

### Use Case 2: Team Collaboration

```bash
# Developer A creates specifications
$ architect init
$ git add .architect/
$ git commit -m "Add project specifications"
$ git push

# Developer B pulls and syncs
$ git pull
$ architect sync
# Now Developer B's AI assistant follows the same specs

# Developer B adds new endpoint
$ architect add-endpoint
$ git add .architect/
$ git commit -m "Add comments endpoint"
$ git push

# Developer A pulls and syncs
$ git pull  
$ architect sync
# Both developers' AI assistants now have the same specifications
```

### Use Case 3: Maintaining Existing Project

```bash
# You have an existing project that's getting messy
# Document what it SHOULD be
$ architect init
[Document your intended API structure]

# Validate current implementation
$ architect validate
❌ 5 endpoints don't match specifications

# Fix with AI assistance
"Update all endpoints to match the architect specifications"
# AI reads specs and fixes inconsistencies

# Keep watching for drift
$ architect watch
```

### Use Case 4: API Documentation

```bash
# Generate documentation from specs
$ architect export --format markdown

# Auto-generate OpenAPI for Swagger UI  
$ architect export --format openapi

# Create Postman collection for testing
$ architect export --format postman

# Your specs become your documentation!
```

## Configuration File (Optional)

`.architect/config.yaml`
```yaml
# Optional configuration
project:
  name: "TaskFlow"
  version: "1.0.0"

sync:
  auto_sync: true           # Auto-sync on spec changes
  watch_on_start: false     # Start watch mode automatically

validation:
  strict_mode: true         # Fail on any deviation
  ignore_paths:
    - "tests/*"
    - "migrations/*"

export:
  default_format: "openapi"
  output_dir: "./docs"
```

## Integration with Git Hooks

```bash
# .git/hooks/pre-commit
#!/bin/sh
architect validate --quiet || {
    echo "❌ Implementation doesn't match specifications"
    echo "Run 'architect validate' to see issues"
    exit 1
}
```

## Tips & Best Practices

### 1. Start with Specifications
```bash
# GOOD: Design first, implement second
$ architect init          # Define your API
$ cursor .                # Then implement

# BAD: Code first, document later
$ code app.py            # Write code
$ architect init         # Try to document after
```

### 2. Keep Specs Updated
```bash
# When requirements change, update specs first
$ architect edit
[Update specifications]
$ architect sync
# Now AI knows about the changes
```

### 3. Use Watch Mode During Development
```bash
# Terminal 1
$ architect watch

# Terminal 2  
$ cursor .
# Edit specs, rules auto-update
```

### 4. Commit Specifications to Git
```bash
# Always version control your specs
$ git add .architect/
$ git commit -m "Update API specifications"

# Team members can sync
$ git pull && architect sync
```

### 5. Let AI Reference Specifications
```prompt
"Implement the user registration endpoint as defined in the architect specifications"
"Create tests for all endpoints in the architect specs"
"Update the task endpoint to match the architect API specification"
```

## Common Issues & Solutions

### Issue: "Cursor isn't following my specifications"
```bash
# Solution: Ensure rules are synced
$ architect sync

# Verify rules file exists
$ ls .cursor/rules/architect.mdc
```

### Issue: "I changed my API design"
```bash
# Solution: Update specs and sync
$ architect edit
[Make changes]
$ architect sync
```

### Issue: "Multiple developers, inconsistent APIs"
```bash
# Solution: Share specs via git
$ git add .architect/
$ git commit -m "API specifications"
$ git push

# Other devs:
$ git pull && architect sync
```

## Environment Variables

```bash
# Optional environment variables
ARCHITECT_EDITOR=code        # Editor for 'architect edit'
ARCHITECT_AUTO_SYNC=true     # Auto-sync on spec changes
ARCHITECT_QUIET=true         # Suppress non-error output
```

## Summary

**Architect** bridges the gap between your API design and AI-assisted implementation:

1. **Define once**: Specify your API structure and business logic
2. **Implement consistently**: AI follows your specifications exactly  
3. **Stay in sync**: Changes to specs automatically update AI rules
4. **Validate compliance**: Ensure code matches specifications

No more explaining the same requirements repeatedly. No more architectural drift. Just consistent, specification-driven development with AI assistance.

---

**Quick Reference:**
```bash
architect init          # Initialize specifications
architect sync          # Sync specs to cursor rules
architect add-endpoint  # Add new API endpoint
architect validate      # Check implementation compliance  
architect watch        # Auto-sync on changes
architect show         # Display current specs
architect export       # Export documentation
architect edit         # Edit specifications
```