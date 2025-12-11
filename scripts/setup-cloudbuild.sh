#!/bin/bash

# Cloud Build Trigger Setup Script
# This script sets up Cloud Build triggers for automated deployments

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

# Configuration
PROJECT_ID="stunning-grin-480914-n1"
REGION="asia-northeast1"
REPO_OWNER="your-github-username"  # TODO: Update this
REPO_NAME="visitas"

print_info "Setting up Cloud Build triggers for project: $PROJECT_ID"

# Check if gcloud is installed
if ! command -v gcloud &> /dev/null; then
    print_error "gcloud CLI is not installed"
    exit 1
fi

# Set project
gcloud config set project $PROJECT_ID

# Enable required APIs
print_info "Enabling Cloud Build API..."
gcloud services enable cloudbuild.googleapis.com --project=$PROJECT_ID

print_info "Enabling Secret Manager API..."
gcloud services enable secretmanager.googleapis.com --project=$PROJECT_ID

# Grant Cloud Build service account necessary permissions
CLOUDBUILD_SA="${PROJECT_NUMBER}@cloudbuild.gserviceaccount.com"
PROJECT_NUMBER=$(gcloud projects describe $PROJECT_ID --format="value(projectNumber)")
CLOUDBUILD_SA="${PROJECT_NUMBER}@cloudbuild.gserviceaccount.com"

print_info "Granting permissions to Cloud Build service account..."

# Grant Cloud Run Admin role
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:${CLOUDBUILD_SA}" \
    --role="roles/run.admin" \
    --quiet

# Grant Service Account User role
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:${CLOUDBUILD_SA}" \
    --role="roles/iam.serviceAccountUser" \
    --quiet

# Grant Secret Manager Secret Accessor role
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:${CLOUDBUILD_SA}" \
    --role="roles/secretmanager.secretAccessor" \
    --quiet

# Grant Artifact Registry Writer role
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:${CLOUDBUILD_SA}" \
    --role="roles/artifactregistry.writer" \
    --quiet

print_success "Permissions granted"

# Create Cloud Build triggers
print_info "Creating Cloud Build triggers..."

# Development trigger (on push to develop branch)
print_info "Creating development trigger..."
gcloud builds triggers create github \
    --name="visitas-backend-dev" \
    --repo-owner="$REPO_OWNER" \
    --repo-name="$REPO_NAME" \
    --branch-pattern="^develop$" \
    --build-config="cloudbuild.yaml" \
    --substitutions="_ENVIRONMENT=dev" \
    --description="Deploy backend to dev environment on push to develop branch" \
    --region="$REGION" \
    --project="$PROJECT_ID" \
    2>/dev/null || print_warning "Development trigger may already exist"

# Production trigger (on push to main branch)
print_info "Creating production trigger..."
gcloud builds triggers create github \
    --name="visitas-backend-prod" \
    --repo-owner="$REPO_OWNER" \
    --repo-name="$REPO_NAME" \
    --branch-pattern="^main$" \
    --build-config="cloudbuild.yaml" \
    --substitutions="_ENVIRONMENT=prod" \
    --description="Deploy backend to prod environment on push to main branch" \
    --region="$REGION" \
    --project="$PROJECT_ID" \
    2>/dev/null || print_warning "Production trigger may already exist"

# Staging trigger (on push to staging branch)
print_info "Creating staging trigger..."
gcloud builds triggers create github \
    --name="visitas-backend-staging" \
    --repo-owner="$REPO_OWNER" \
    --repo-name="$REPO_NAME" \
    --branch-pattern="^staging$" \
    --build-config="cloudbuild.yaml" \
    --substitutions="_ENVIRONMENT=staging" \
    --description="Deploy backend to staging environment on push to staging branch" \
    --region="$REGION" \
    --project="$PROJECT_ID" \
    2>/dev/null || print_warning "Staging trigger may already exist"

print_success "Cloud Build triggers created"

# Display created triggers
print_info "Listing Cloud Build triggers..."
gcloud builds triggers list --project=$PROJECT_ID --region=$REGION

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
print_success "Cloud Build Setup Complete!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
print_info "Triggers created:"
echo "  - visitas-backend-dev (develop branch)"
echo "  - visitas-backend-staging (staging branch)"
echo "  - visitas-backend-prod (main branch)"
echo ""
print_info "View triggers: https://console.cloud.google.com/cloud-build/triggers?project=$PROJECT_ID"
echo ""
print_warning "Note: Make sure to connect your GitHub repository in Cloud Build settings:"
print_warning "https://console.cloud.google.com/cloud-build/triggers/connect?project=$PROJECT_ID"
echo ""
