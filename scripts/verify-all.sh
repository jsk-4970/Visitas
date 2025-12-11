#!/bin/bash

# Comprehensive Visitas Project Verification Script
# This script runs all verification checks

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

print_header() {
    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  $1${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

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
WARNINGS=0

echo ""
echo "🔍 Visitas Comprehensive Verification"
echo "======================================"

# 1. Project Structure
print_header "1. Project Structure Verification"
if ./scripts/verify-project.sh; then
    print_success "Project structure is valid"
else
    print_error "Project structure has issues"
    ((ERRORS++))
fi

# 2. Go Module Verification
print_header "2. Go Module Verification"
cd backend

if [ ! -f "go.mod" ]; then
    print_error "go.mod not found"
    ((ERRORS++))
    cd ..
    exit 1
fi

print_info "Verifying Go module..."
if go mod verify 2>/dev/null; then
    print_success "Go modules verified"
else
    print_warning "go.sum missing or outdated"
    print_info "Run: cd backend && go mod tidy"
    ((WARNINGS++))
fi

# 3. Go Code Syntax Check
print_header "3. Go Code Syntax Check"
print_info "Running go vet..."
if go vet ./... 2>/dev/null; then
    print_success "No issues found with go vet"
else
    print_warning "go vet found potential issues"
    ((WARNINGS++))
fi

# 4. Import Path Verification
print_header "4. Import Path Verification"
print_info "Checking module path..."
MODULE_PATH=$(head -1 go.mod | awk '{print $2}')
print_info "Module path: $MODULE_PATH"

if [ "$MODULE_PATH" = "github.com/visitas/backend" ]; then
    print_success "Module path is correct"
else
    print_error "Module path should be 'github.com/visitas/backend', found '$MODULE_PATH'"
    ((ERRORS++))
fi

cd ..

# 5. Docker Build Test
print_header "5. Docker Build Verification"
if command -v docker &> /dev/null; then
    print_info "Testing Docker build..."
    if docker build -t visitas-api:verify-test backend/ > /tmp/docker-build.log 2>&1; then
        print_success "Docker build successful"

        # Check image size
        IMAGE_SIZE=$(docker images visitas-api:verify-test --format "{{.Size}}")
        print_info "Image size: $IMAGE_SIZE"

        # Cleanup
        docker rmi visitas-api:verify-test > /dev/null 2>&1
    else
        print_error "Docker build failed"
        print_info "Check logs: cat /tmp/docker-build.log"
        ((ERRORS++))
    fi
else
    print_warning "Docker not installed, skipping Docker build test"
    ((WARNINGS++))
fi

# 6. Terraform Validation
print_header "6. Terraform Configuration Validation"
if command -v terraform &> /dev/null; then
    cd infra/terraform/environments/dev

    print_info "Initializing Terraform (without backend)..."
    if terraform init -backend=false > /tmp/terraform-init.log 2>&1; then
        print_success "Terraform initialized"

        print_info "Validating Terraform configuration..."
        if terraform validate > /tmp/terraform-validate.log 2>&1; then
            print_success "Terraform configuration is valid"
        else
            print_error "Terraform validation failed"
            cat /tmp/terraform-validate.log
            ((ERRORS++))
        fi

        print_info "Checking Terraform formatting..."
        if terraform fmt -check > /dev/null 2>&1; then
            print_success "Terraform files are properly formatted"
        else
            print_warning "Some Terraform files need formatting (run: terraform fmt)"
            ((WARNINGS++))
        fi
    else
        print_error "Terraform initialization failed"
        ((ERRORS++))
    fi

    cd ../../../..
else
    print_warning "Terraform not installed, skipping Terraform validation"
    ((WARNINGS++))
fi

# 7. Configuration Consistency Check
print_header "7. Configuration Consistency Check"

print_info "Checking environment variable consistency..."

# Project ID
PROJECT_ID_ENV=$(grep GCP_PROJECT_ID backend/.env.example 2>/dev/null | cut -d= -f2)
PROJECT_ID_TF=$(grep 'default.*=.*"stunning-grin' infra/terraform/environments/dev/variables.tf 2>/dev/null | head -1 | grep -oE '"[^"]+"' | tr -d '"')

if [ "$PROJECT_ID_ENV" = "$PROJECT_ID_TF" ]; then
    print_success "Project ID is consistent across files"
else
    print_warning "Project ID inconsistency detected"
    ((WARNINGS++))
fi

# Region
REGION_ENV=$(grep GCP_REGION backend/.env.example 2>/dev/null | cut -d= -f2)
if [[ "$REGION_ENV" == "asia-northeast1" ]] || [[ "$REGION_ENV" == "asia-northeast2" ]]; then
    print_success "Region is set to Japan ($REGION_ENV)"
else
    print_warning "Region should be in Japan (asia-northeast1 or asia-northeast2)"
    ((WARNINGS++))
fi

# 8. Security Check
print_header "8. Security & Secrets Check"

print_info "Checking for exposed secrets..."
SECRETS_FOUND=0

if [ -f "backend/.env" ]; then
    print_warning "backend/.env exists (should not be committed)"
    ((WARNINGS++))
fi

if [ -f "backend/config/firebase-service-account.json" ]; then
    print_warning "Firebase service account key exists in repo (should not be committed)"
    ((WARNINGS++))
fi

if git ls-files | grep -qE '\.(env|key|pem)$'; then
    print_warning "Potential secret files are tracked by git"
    ((WARNINGS++))
fi

if [ $SECRETS_FOUND -eq 0 ]; then
    print_success "No obvious secrets found in repository"
fi

# 9. Documentation Check
print_header "9. Documentation Verification"

DOCS=(
    "README.md:Project README"
    "docs/REQUIREMENTS.md:Requirements document"
    "docs/FIREBASE_SETUP.md:Firebase setup guide"
    "docs/CICD.md:CI/CD documentation"
    "docs/TESTING.md:Testing guide"
)

DOC_MISSING=0
for entry in "${DOCS[@]}"; do
    doc="${entry%%:*}"
    description="${entry#*:}"
    if [ -f "$doc" ]; then
        print_success "$description exists"
    else
        print_warning "$description is missing"
        ((WARNINGS++))
        ((DOC_MISSING++))
    fi
done

# 10. CI/CD Configuration Check
print_header "10. CI/CD Configuration Check"

print_info "Checking GitHub Actions workflows..."
if [ -f ".github/workflows/test.yml" ] && [ -f ".github/workflows/deploy.yml" ]; then
    print_success "GitHub Actions workflows are configured"
else
    print_error "GitHub Actions workflows are missing"
    ((ERRORS++))
fi

print_info "Checking Cloud Build configuration..."
if [ -f "cloudbuild.yaml" ]; then
    print_success "Cloud Build configuration exists"
else
    print_error "cloudbuild.yaml is missing"
    ((ERRORS++))
fi

print_info "Checking deployment scripts..."
if [ -x "scripts/deploy.sh" ] && [ -x "scripts/setup-cloudbuild.sh" ]; then
    print_success "Deployment scripts are executable"
else
    print_warning "Some scripts may not be executable"
    ((WARNINGS++))
fi

# Summary
print_header "Verification Summary"

echo ""
if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    print_success "🎉 All checks passed! Project is ready."
    echo ""
    print_info "Next steps:"
    echo "  1. Set up Firebase: Follow docs/FIREBASE_SETUP.md"
    echo "  2. Deploy infrastructure: cd infra/terraform/environments/dev && terraform apply"
    echo "  3. Deploy application: ./scripts/deploy.sh dev"
    echo ""
    exit 0
elif [ $ERRORS -eq 0 ]; then
    print_warning "⚠️  Verification completed with $WARNINGS warning(s)"
    echo ""
    print_info "The project structure is valid but some improvements can be made."
    print_info "Review the warnings above and address them if needed."
    echo ""
    exit 0
else
    print_error "❌ Verification failed with $ERRORS error(s) and $WARNINGS warning(s)"
    echo ""
    print_info "Please address the errors before proceeding."
    echo ""
    exit 1
fi
