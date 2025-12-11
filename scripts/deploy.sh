#!/bin/bash

# Visitas Deployment Script
# Usage: ./scripts/deploy.sh [environment]
# Environments: dev, staging, prod

set -e

# Add Go to PATH
export PATH=$PATH:/usr/local/go/bin

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
    echo "Usage: $0 [environment]"
    echo ""
    echo "Environments:"
    echo "  dev      - Development environment"
    echo "  staging  - Staging environment"
    echo "  prod     - Production environment"
    echo ""
    echo "Example:"
    echo "  $0 dev"
    exit 1
}

# Check if environment is provided
if [ -z "$1" ]; then
    print_error "Environment not specified"
    usage
fi

ENV=$1

# Validate environment
case $ENV in
    dev|staging|prod)
        print_info "Deploying to $ENV environment"
        ;;
    *)
        print_error "Invalid environment: $ENV"
        usage
        ;;
esac

# Set environment-specific variables
SERVICE_NAME="visitas-api-$ENV"
ARTIFACT_REGISTRY="$REGION-docker.pkg.dev/$PROJECT_ID/visitas-$ENV"
SERVICE_ACCOUNT="visitas-$ENV-run@$PROJECT_ID.iam.gserviceaccount.com"
SPANNER_INSTANCE="stunning-grin-480914-n1-instance"
SPANNER_DATABASE="stunning-grin-480914-n1-db"

# Set resource limits based on environment
case $ENV in
    prod)
        MIN_INSTANCES=1
        MAX_INSTANCES=20
        CPU=2
        MEMORY="1Gi"
        LOG_LEVEL="info"
        ;;
    staging)
        MIN_INSTANCES=0
        MAX_INSTANCES=5
        CPU=1
        MEMORY="512Mi"
        LOG_LEVEL="debug"
        ;;
    dev)
        MIN_INSTANCES=0
        MAX_INSTANCES=3
        CPU=1
        MEMORY="512Mi"
        LOG_LEVEL="debug"
        ;;
esac

# Check if gcloud is installed
if ! command -v gcloud &> /dev/null; then
    print_error "gcloud CLI is not installed"
    exit 1
fi

# Check if user is authenticated
if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" &> /dev/null; then
    print_error "Not authenticated with gcloud. Run: gcloud auth login"
    exit 1
fi

# Set project
print_info "Setting GCP project to $PROJECT_ID"
gcloud config set project $PROJECT_ID

# Confirm deployment for production
if [ "$ENV" == "prod" ]; then
    print_warning "You are about to deploy to PRODUCTION!"
    read -p "Are you sure? (yes/no): " CONFIRM
    if [ "$CONFIRM" != "yes" ]; then
        print_info "Deployment cancelled"
        exit 0
    fi
fi

# Change to backend directory
cd "$(dirname "$0")/../backend" || exit 1

# Run tests
print_info "Running tests..."
if go test -v ./...; then
    print_success "Tests passed"
else
    print_error "Tests failed"
    exit 1
fi

# Build Docker image
IMAGE_TAG=$(git rev-parse --short HEAD)
IMAGE_NAME="$ARTIFACT_REGISTRY/api:$IMAGE_TAG"
IMAGE_LATEST="$ARTIFACT_REGISTRY/api:latest"

print_info "Building Docker image: $IMAGE_NAME"
docker build --platform linux/amd64 -t $IMAGE_NAME -t $IMAGE_LATEST -f Dockerfile .

if [ $? -eq 0 ]; then
    print_success "Docker image built successfully"
else
    print_error "Docker build failed"
    exit 1
fi

# Configure Docker to use Artifact Registry
print_info "Configuring Docker authentication..."
gcloud auth configure-docker $REGION-docker.pkg.dev --quiet

# Push Docker image
print_info "Pushing Docker image to Artifact Registry..."
docker push $IMAGE_NAME
docker push $IMAGE_LATEST

if [ $? -eq 0 ]; then
    print_success "Docker image pushed successfully"
else
    print_error "Docker push failed"
    exit 1
fi

# Deploy to Cloud Run
print_info "Deploying to Cloud Run..."
gcloud run deploy $SERVICE_NAME \
    --image $IMAGE_NAME \
    --platform managed \
    --region $REGION \
    --service-account $SERVICE_ACCOUNT \
    --set-env-vars "GCP_PROJECT_ID=$PROJECT_ID" \
    --set-env-vars "GCP_REGION=$REGION" \
    --set-env-vars "SPANNER_INSTANCE=$SPANNER_INSTANCE" \
    --set-env-vars "SPANNER_DATABASE=$SPANNER_DATABASE" \
    --set-env-vars "ENV=$ENV" \
    --set-env-vars "LOG_LEVEL=$LOG_LEVEL" \
    --set-env-vars "FIREBASE_CONFIG_PATH=/secrets/firebase.json" \
    --set-secrets "/secrets/firebase.json=firebase-service-account-$ENV:latest" \
    --allow-unauthenticated \
    --min-instances $MIN_INSTANCES \
    --max-instances $MAX_INSTANCES \
    --cpu $CPU \
    --memory $MEMORY \
    --timeout 60s \
    --concurrency 80 \
    --port 8080

if [ $? -eq 0 ]; then
    print_success "Deployment successful"
else
    print_error "Deployment failed"
    exit 1
fi

# Get service URL
SERVICE_URL=$(gcloud run services describe $SERVICE_NAME \
    --platform managed \
    --region $REGION \
    --format 'value(status.url)')

print_success "Service deployed at: $SERVICE_URL"

# Test deployment
print_info "Testing deployment..."
sleep 5

for i in {1..10}; do
    HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" $SERVICE_URL/health)
    if [ "$HTTP_STATUS" -eq 200 ]; then
        print_success "Health check passed (HTTP $HTTP_STATUS)"
        break
    fi
    if [ $i -eq 10 ]; then
        print_error "Health check failed after 10 attempts"
        exit 1
    fi
    print_info "Waiting for service to be ready... (attempt $i/10)"
    sleep 3
done

# Display deployment summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
print_success "Deployment Complete!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Environment:    $ENV"
echo "Service:        $SERVICE_NAME"
echo "URL:            $SERVICE_URL"
echo "Image:          $IMAGE_NAME"
echo "Region:         $REGION"
echo ""
print_info "View logs: gcloud run logs read $SERVICE_NAME --region $REGION"
print_info "View service: https://console.cloud.google.com/run/detail/$REGION/$SERVICE_NAME/metrics?project=$PROJECT_ID"
echo ""
