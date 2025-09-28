# 📁 Repository Restructure Plan

## Current Issues
- ❌ `main.go` in root (should be in `/cmd/`)
- ❌ Multiple loose documentation files in root
- ❌ Sensitive data in `.env` (needs `.env.example`)
- ❌ No proper `/tests/` directory
- ❌ IDE-specific folders (`.idea/`, `.vscode/`) should be gitignored
- ❌ Development artifacts (`.kiro/`, debug files) in repo

## Proposed Structure

```
gocraft/
├── cmd/
│   └── gocraft/
│       └── main.go              # Main application entrypoint
├── internal/                    # ✅ Already good structure
│   ├── api/
│   ├── auth/
│   ├── builder/                 # Core project generation logic
│   ├── config/
│   ├── database/
│   ├── handlers/
│   ├── middleware/
│   ├── models/
│   ├── services/
│   ├── templates/               # Project templates
│   ├── utils/
│   └── validation/
├── pkg/                         # Reusable packages (if any)
├── docs/
│   ├── api/                     # Swagger docs (moved from /docs/)
│   │   ├── docs.go
│   │   ├── swagger.json
│   │   └── swagger.yaml
│   ├── ARCHITECTURE.md          # System architecture
│   ├── FEATURES.md              # Feature documentation
│   └── ROADMAP.md               # Future plans
├── tests/
│   ├── integration/             # Integration tests
│   ├── unit/                    # Unit tests
│   └── fixtures/                # Test data
├── scripts/
│   ├── migrate.go               # Database migrations
│   ├── optimize_database.go     # DB optimization
│   ├── build.sh                 # Build script
│   └── setup.sh                 # Development setup
├── examples/                    # ✅ Keep example projects
├── output/                      # ✅ Generated projects (gitignored)
├── .env.example                 # Environment template
├── .gitignore                   # Updated gitignore
├── README.md                    # Polished README
├── CONTRIBUTING.md              # Contribution guidelines
├── LICENSE                      # Open source license
├── Makefile                     # Build automation
├── docker-compose.yml           # Development environment
├── go.mod
└── go.sum
```

## Files to Move/Create/Delete

### 🔄 Move Files
```bash
# Move main.go to cmd structure
mkdir -p cmd/gocraft
mv main.go cmd/gocraft/

# Move docs to proper structure
mkdir -p docs/api
mv docs/* docs/api/

# Consolidate scripts
# (scripts already in good location)
```

### 🗑️ Delete Files
```bash
# Remove development artifacts
rm -rf .kiro/
rm -rf .idea/
rm -rf .vscode/
rm .DS_Store
rm debug_download.html
rm qodana.yaml

# Remove loose documentation (consolidate into /docs/)
rm ADMIN_DASHBOARD_API_GUIDE.md
rm BUG_FIX_FIBER_FRAMEWORK.md
rm BUG_FIXES_SUMMARY.md
rm DEPLOYMENT_FIXES.md
rm FRAMEWORK_FIELD_FIX.md
rm FRONTEND_ADMIN_INTEGRATION_PROMPT.md
rm FRONTEND_INTEGRATION_GUIDE.md
rm FRONTEND_SUMMARY.md
rm PROJECT_HISTORY_FRONTEND_GUIDE.md
rm TROUBLESHOOTING.md
rm VALIDATION_SYSTEM.md
```

### 📝 Create Files
```bash
# Core documentation
touch CONTRIBUTING.md
touch LICENSE
touch Makefile
touch .env.example

# Documentation structure
mkdir -p docs
touch docs/ARCHITECTURE.md
touch docs/FEATURES.md
touch docs/ROADMAP.md

# Test structure
mkdir -p tests/{unit,integration,fixtures}
touch tests/README.md
```

### 🔒 Security Fixes
```bash
# Move sensitive data to example
mv .env .env.local
# Create .env.example (see below)
```