#!/bin/bash

# 開発環境セットアップスクリプト
# このスクリプトは初回のみ実行してください

set -e

echo "=================================="
echo "  Visitas Development Setup"
echo "=================================="
echo ""

# カラー出力
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 必要なツールのチェック
echo "Step 1: Checking required tools..."
echo ""

check_command() {
  if command -v $1 >/dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} $1 is installed"
    return 0
  else
    echo -e "${RED}✗${NC} $1 is not installed"
    return 1
  fi
}

MISSING_TOOLS=false

check_command "go" || MISSING_TOOLS=true
check_command "docker" || MISSING_TOOLS=true
check_command "docker-compose" || MISSING_TOOLS=true
check_command "gcloud" || MISSING_TOOLS=true
check_command "terraform" || MISSING_TOOLS=true

echo ""

if [ "$MISSING_TOOLS" = true ]; then
  echo -e "${RED}Some required tools are missing. Please install them first.${NC}"
  echo ""
  echo "Installation guides:"
  echo "  - Go:             https://go.dev/doc/install"
  echo "  - Docker:         https://docs.docker.com/get-docker/"
  echo "  - gcloud CLI:     https://cloud.google.com/sdk/docs/install"
  echo "  - Terraform:      https://developer.hashicorp.com/terraform/downloads"
  echo ""
  exit 1
fi

# Go dependencies
echo "Step 2: Installing Go dependencies..."
cd backend
go mod download
cd ..
echo -e "${GREEN}✓${NC} Go dependencies installed"
echo ""

# .env ファイルの作成
echo "Step 3: Creating .env file..."
if [ ! -f backend/.env ]; then
  cp backend/.env.example backend/.env
  echo -e "${GREEN}✓${NC} .env file created (backend/.env)"
  echo -e "${YELLOW}⚠${NC}  Please edit backend/.env and set your API keys"
else
  echo -e "${YELLOW}⚠${NC}  .env file already exists"
fi
echo ""

# Terraform tfvars の作成
echo "Step 4: Creating Terraform variables..."
if [ ! -f infra/terraform/environments/dev/terraform.tfvars ]; then
  cp infra/terraform/environments/dev/terraform.tfvars.example infra/terraform/environments/dev/terraform.tfvars
  echo -e "${GREEN}✓${NC} terraform.tfvars created"
  echo -e "${YELLOW}⚠${NC}  Please edit infra/terraform/environments/dev/terraform.tfvars if needed"
else
  echo -e "${YELLOW}⚠${NC}  terraform.tfvars already exists"
fi
echo ""

# Docker ネットワークの作成
echo "Step 5: Creating Docker network..."
docker network create visitas-network 2>/dev/null || echo -e "${YELLOW}⚠${NC}  Network already exists"
echo ""

# gcloud 設定の確認
echo "Step 6: Checking gcloud configuration..."
CURRENT_PROJECT=$(gcloud config get-value project 2>/dev/null)
if [ -z "$CURRENT_PROJECT" ]; then
  echo -e "${YELLOW}⚠${NC}  No GCP project is set. Run: gcloud config set project YOUR_PROJECT_ID"
else
  echo -e "${GREEN}✓${NC} Current GCP project: $CURRENT_PROJECT"
fi
echo ""

# 完了メッセージ
echo "=================================="
echo -e "${GREEN}✓ Setup Complete!${NC}"
echo "=================================="
echo ""
echo "Next steps:"
echo ""
echo "  1. Edit backend/.env and add your API keys"
echo ""
echo "  2. Start local development environment:"
echo "     $ make dev-docker"
echo ""
echo "  3. Or start API server only (without Docker):"
echo "     $ make dev"
echo ""
echo "  4. Deploy to GCP (production):"
echo "     $ cd infra/terraform/environments/dev"
echo "     $ terraform init"
echo "     $ terraform apply"
echo ""
echo "For more commands, run: make help"
echo ""
