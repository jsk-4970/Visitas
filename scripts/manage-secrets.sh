#!/bin/bash

# Visitas Secret Manager Management Script
# Usage: ./scripts/manage-secrets.sh [command] [options]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ID="stunning-grin-480914-n1"
REGION="asia-northeast1"

# Function to print colored messages
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

# Function to display usage
usage() {
    cat << EOF
Visitas Secret Manager Management Script

Usage: $0 [command] [options]

Commands:
  list [env]              List all secrets for environment (dev/staging/prod)
  view <secret-name>      View the latest version of a secret
  update <secret-name>    Update a secret value (interactive)
  create [env]            Create all required secrets for environment
  grant [env]             Grant Cloud Run service account access to secrets
  sync-local [env]        Sync secrets to local .env file (for development)
  delete <secret-name>    Delete a secret (use with caution!)

Environments:
  dev      - Development environment
  staging  - Staging environment
  prod     - Production environment

Examples:
  $0 list dev
  $0 view google-maps-api-key-dev
  $0 update google-maps-api-key-dev
  $0 create dev
  $0 grant dev
  $0 sync-local dev

EOF
    exit 1
}

# Function to list secrets
list_secrets() {
    local ENV=$1
    print_info "Listing secrets for environment: $ENV"
    echo ""

    if [ -z "$ENV" ]; then
        gcloud secrets list --project=$PROJECT_ID
    else
        gcloud secrets list --project=$PROJECT_ID --filter="name:*-${ENV}"
    fi
}

# Function to view secret value
view_secret() {
    local SECRET_NAME=$1

    if [ -z "$SECRET_NAME" ]; then
        print_error "Secret name is required"
        usage
    fi

    print_info "Viewing latest version of secret: $SECRET_NAME"
    echo ""

    gcloud secrets versions access latest \
        --secret="$SECRET_NAME" \
        --project=$PROJECT_ID
}

# Function to update secret
update_secret() {
    local SECRET_NAME=$1

    if [ -z "$SECRET_NAME" ]; then
        print_error "Secret name is required"
        usage
    fi

    print_info "Updating secret: $SECRET_NAME"
    echo ""
    print_warning "Current value:"
    gcloud secrets versions access latest \
        --secret="$SECRET_NAME" \
        --project=$PROJECT_ID 2>/dev/null || echo "(No current value)"
    echo ""

    read -p "Enter new value: " NEW_VALUE

    if [ -z "$NEW_VALUE" ]; then
        print_error "Value cannot be empty"
        exit 1
    fi

    echo "$NEW_VALUE" | gcloud secrets versions add "$SECRET_NAME" \
        --project=$PROJECT_ID \
        --data-file=-

    print_success "Secret updated successfully!"
}

# Function to create all secrets
create_secrets() {
    local ENV=$1

    if [ -z "$ENV" ]; then
        print_error "Environment is required (dev/staging/prod)"
        usage
    fi

    print_info "Creating secrets for environment: $ENV"
    echo ""

    # Firebase Service Account
    print_info "Creating firebase-service-account-${ENV}..."
    gcloud secrets create firebase-service-account-${ENV} \
        --project=$PROJECT_ID \
        --replication-policy="user-managed" \
        --locations="$REGION" \
        --labels="environment=$ENV,project=visitas" \
        2>/dev/null || print_warning "Secret already exists"

    # Google Maps API Key
    print_info "Creating google-maps-api-key-${ENV}..."
    echo "placeholder_google_maps_api_key" | gcloud secrets create google-maps-api-key-${ENV} \
        --project=$PROJECT_ID \
        --replication-policy="user-managed" \
        --locations="$REGION" \
        --labels="environment=$ENV,project=visitas" \
        --data-file=- \
        2>/dev/null || print_warning "Secret already exists"

    # Gemini API Key
    print_info "Creating gemini-api-key-${ENV}..."
    echo "placeholder_gemini_api_key" | gcloud secrets create gemini-api-key-${ENV} \
        --project=$PROJECT_ID \
        --replication-policy="user-managed" \
        --locations="$REGION" \
        --labels="environment=$ENV,project=visitas" \
        --data-file=- \
        2>/dev/null || print_warning "Secret already exists"

    # CORS Allowed Origins
    print_info "Creating cors-allowed-origins-${ENV}..."
    echo "http://localhost:3000,http://localhost:8080" | gcloud secrets create cors-allowed-origins-${ENV} \
        --project=$PROJECT_ID \
        --replication-policy="user-managed" \
        --locations="$REGION" \
        --labels="environment=$ENV,project=visitas" \
        --data-file=- \
        2>/dev/null || print_warning "Secret already exists"

    print_success "Secrets created successfully!"
    echo ""
    print_info "Remember to update placeholder values with actual keys"
}

# Function to grant access
grant_access() {
    local ENV=$1

    if [ -z "$ENV" ]; then
        print_error "Environment is required (dev/staging/prod)"
        usage
    fi

    local SERVICE_ACCOUNT="visitas-${ENV}-run@${PROJECT_ID}.iam.gserviceaccount.com"

    print_info "Granting access to Cloud Run service account: $SERVICE_ACCOUNT"
    echo ""

    for SECRET in firebase-service-account google-maps-api-key gemini-api-key cors-allowed-origins; do
        print_info "Granting access to ${SECRET}-${ENV}..."
        gcloud secrets add-iam-policy-binding ${SECRET}-${ENV} \
            --project=$PROJECT_ID \
            --member="serviceAccount:${SERVICE_ACCOUNT}" \
            --role="roles/secretmanager.secretAccessor" \
            --quiet
    done

    print_success "Access granted successfully!"
}

# Function to sync secrets to local .env
sync_local() {
    local ENV=$1

    if [ -z "$ENV" ]; then
        print_error "Environment is required (dev/staging/prod)"
        usage
    fi

    local ENV_FILE="backend/.env"

    print_info "Syncing secrets to local .env file..."
    print_warning "This will overwrite your current .env file!"
    read -p "Continue? (yes/no): " CONFIRM

    if [ "$CONFIRM" != "yes" ]; then
        print_info "Cancelled"
        exit 0
    fi

    # Copy .env.example
    cp backend/.env.example "$ENV_FILE"

    # Fetch and append secrets
    {
        echo ""
        echo "# Secrets from Secret Manager (synced on $(date))"
        echo "GOOGLE_MAPS_API_KEY=$(gcloud secrets versions access latest --secret=google-maps-api-key-${ENV} --project=$PROJECT_ID)"
        echo "GEMINI_API_KEY=$(gcloud secrets versions access latest --secret=gemini-api-key-${ENV} --project=$PROJECT_ID)"
        echo "ALLOWED_ORIGINS=$(gcloud secrets versions access latest --secret=cors-allowed-origins-${ENV} --project=$PROJECT_ID)"
    } >> "$ENV_FILE"

    # Fetch Firebase service account
    gcloud secrets versions access latest \
        --secret="firebase-service-account-${ENV}" \
        --project=$PROJECT_ID > backend/config/firebase-service-account.json

    print_success "Secrets synced to $ENV_FILE"
}

# Function to delete secret
delete_secret() {
    local SECRET_NAME=$1

    if [ -z "$SECRET_NAME" ]; then
        print_error "Secret name is required"
        usage
    fi

    print_warning "This will PERMANENTLY delete the secret: $SECRET_NAME"
    read -p "Are you sure? (type 'DELETE' to confirm): " CONFIRM

    if [ "$CONFIRM" != "DELETE" ]; then
        print_info "Cancelled"
        exit 0
    fi

    gcloud secrets delete "$SECRET_NAME" \
        --project=$PROJECT_ID \
        --quiet

    print_success "Secret deleted"
}

# Main script logic
if [ -z "$1" ]; then
    usage
fi

COMMAND=$1
shift

case $COMMAND in
    list)
        list_secrets "$1"
        ;;
    view)
        view_secret "$1"
        ;;
    update)
        update_secret "$1"
        ;;
    create)
        create_secrets "$1"
        ;;
    grant)
        grant_access "$1"
        ;;
    sync-local)
        sync_local "$1"
        ;;
    delete)
        delete_secret "$1"
        ;;
    *)
        print_error "Unknown command: $COMMAND"
        usage
        ;;
esac
