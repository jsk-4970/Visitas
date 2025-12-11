#!/bin/bash

# Visitas Project Structure Verification Script
# This script checks if all required files and directories exist

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_info() {
    echo -e "${BLUE}ℹ ${1}${NC}"
}

print_success() {
    echo -e "${GREEN}✅ ${1}${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  ${1}${NC}"
}

print_error() {
    echo -e "${RED}❌ ${1}${NC}"
}

ERRORS=0

echo "🔍 Visitas Project Structure Verification"
echo "=========================================="
echo ""

# Check essential files
print_info "Checking essential files..."
echo ""

files=(
    "backend/go.mod:Go module definition"
    "backend/Dockerfile:Docker configuration"
    "backend/.env.example:Environment variables template"
    "backend/cmd/api/main.go:Main application entry point"
    "backend/internal/config/config.go:Configuration loader"
    "backend/internal/handlers/health.go:Health check handler"
    "backend/internal/handlers/patient.go:Patient handler"
    "backend/internal/models/patient.go:Patient model"
    "backend/internal/repository/spanner.go:Spanner repository"
    "backend/internal/middleware/auth.go:Authentication middleware"
    "backend/pkg/auth/firebase.go:Firebase client"
    "cloudbuild.yaml:Cloud Build configuration"
    ".github/workflows/test.yml:Test workflow"
    ".github/workflows/deploy.yml:Deploy workflow"
    ".gitignore:Git ignore rules"
    "README.md:Project documentation"
    "docs/CICD.md:CI/CD documentation"
    "docs/FIREBASE_SETUP.md:Firebase setup guide"
    "docs/TESTING.md:Testing guide"
    "scripts/deploy.sh:Deployment script"
    "scripts/setup-cloudbuild.sh:Cloud Build setup script"
)

for entry in "${files[@]}"; do
    file="${entry%%:*}"
    description="${entry#*:}"
    if [ -f "$file" ]; then
        print_success "$description ($file)"
    else
        print_error "$description ($file) - NOT FOUND"
        ((ERRORS++))
    fi
done

echo ""
print_info "Checking directory structure..."
echo ""

dirs=(
    "backend/cmd/api:API entry point"
    "backend/internal/handlers:HTTP handlers"
    "backend/internal/models:Data models"
    "backend/internal/repository:Data access layer"
    "backend/internal/config:Configuration"
    "backend/internal/middleware:Middleware"
    "backend/pkg/auth:Authentication utilities"
    "backend/migrations:Database migrations"
    "backend/scripts:Backend scripts"
    "docs:Documentation"
    "scripts:Project scripts"
    "infra/terraform/environments/dev:Terraform dev environment"
    ".github/workflows:GitHub Actions workflows"
)

for entry in "${dirs[@]}"; do
    dir="${entry%%:*}"
    description="${entry#*:}"
    if [ -d "$dir" ]; then
        print_success "$description ($dir)"
    else
        print_error "$description ($dir) - NOT FOUND"
        ((ERRORS++))
    fi
done

echo ""
print_info "Checking Go files..."
echo ""

GO_FILES=$(find backend -name '*.go' 2>/dev/null | wc -l | tr -d ' ')
print_info "Total Go files: $GO_FILES"

if [ "$GO_FILES" -lt 8 ]; then
    print_warning "Expected at least 8 Go files, found $GO_FILES"
else
    print_success "Go files count is adequate"
fi

echo ""
print_info "Checking migrations..."
echo ""

SQL_FILES=$(find backend/migrations -name '*.sql' 2>/dev/null | wc -l | tr -d ' ')
print_info "SQL migration files: $SQL_FILES"

if [ "$SQL_FILES" -lt 4 ]; then
    print_warning "Expected at least 4 migration files, found $SQL_FILES"
else
    print_success "Migration files present"
    ls -1 backend/migrations/*.sql 2>/dev/null | while read file; do
        echo "  - $(basename $file)"
    done
fi

echo ""
print_info "Checking configuration consistency..."
echo ""

# Check if PROJECT_ID is consistent
PROJECT_ID_ENV=$(grep GCP_PROJECT_ID backend/.env.example 2>/dev/null | cut -d= -f2)
PROJECT_ID_TF=$(grep 'default.*=.*"stunning-grin' infra/terraform/environments/dev/variables.tf 2>/dev/null | head -1 | grep -oE '"[^"]+"' | tr -d '"')

if [ "$PROJECT_ID_ENV" = "$PROJECT_ID_TF" ]; then
    print_success "Project ID is consistent: $PROJECT_ID_ENV"
else
    print_warning "Project ID mismatch:"
    echo "  .env.example: $PROJECT_ID_ENV"
    echo "  Terraform: $PROJECT_ID_TF"
fi

# Check if REGION is consistent
REGION_ENV=$(grep GCP_REGION backend/.env.example 2>/dev/null | cut -d= -f2)
REGION_TF=$(grep 'default' infra/terraform/environments/dev/variables.tf 2>/dev/null | grep asia | head -1 | grep -oE 'asia-[a-z0-9-]+' | head -1)

if [ "$REGION_ENV" = "$REGION_TF" ]; then
    print_success "Region is consistent: $REGION_ENV"
else
    print_warning "Region mismatch:"
    echo "  .env.example: $REGION_ENV"
    echo "  Terraform: $REGION_TF"
fi

echo ""
print_info "Checking Docker configuration..."
echo ""

EXPOSE_PORT=$(grep "^EXPOSE" backend/Dockerfile 2>/dev/null | awk '{print $2}')
DEFAULT_PORT=$(grep 'getEnv.*PORT' backend/internal/config/config.go 2>/dev/null | grep -oE '[0-9]+' | head -1)

if [ "$EXPOSE_PORT" = "$DEFAULT_PORT" ]; then
    print_success "Port configuration is consistent: $EXPOSE_PORT"
else
    print_warning "Port mismatch:"
    echo "  Dockerfile EXPOSE: $EXPOSE_PORT"
    echo "  Config default: $DEFAULT_PORT"
fi

echo ""
print_info "Checking scripts permissions..."
echo ""

if [ -x "scripts/deploy.sh" ]; then
    print_success "deploy.sh is executable"
else
    print_warning "deploy.sh is not executable (run: chmod +x scripts/deploy.sh)"
fi

if [ -x "scripts/setup-cloudbuild.sh" ]; then
    print_success "setup-cloudbuild.sh is executable"
else
    print_warning "setup-cloudbuild.sh is not executable (run: chmod +x scripts/setup-cloudbuild.sh)"
fi

echo ""
echo "=========================================="

if [ $ERRORS -eq 0 ]; then
    print_success "All essential files and directories are present!"
    echo ""
    print_info "Next steps:"
    echo "  1. Run: cd backend && go mod tidy"
    echo "  2. Run: ./scripts/verify-all.sh (full verification)"
    echo "  3. See: docs/TESTING.md for testing instructions"
    exit 0
else
    print_error "Found $ERRORS missing files/directories"
    echo ""
    print_info "Please create missing files before proceeding"
    exit 1
fi
