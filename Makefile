.PHONY: help setup dev build test clean docker-up docker-down migrate-up migrate-down terraform-init terraform-plan terraform-apply

# デフォルトターゲット
.DEFAULT_GOAL := help

# 変数
PROJECT_ID := visitas-dev
REGION := asia-northeast1
SPANNER_INSTANCE := visitas-dev-instance
SPANNER_DATABASE := visitas-dev-db

help: ## ヘルプを表示
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

##@ 環境構築

setup: ## 開発環境のセットアップ（初回のみ）
	@echo "==> Setting up development environment..."
	@chmod +x scripts/setup-dev.sh
	@./scripts/setup-dev.sh

setup-check: ## 必要なツールがインストールされているか確認
	@echo "==> Checking required tools..."
	@command -v go >/dev/null 2>&1 || { echo "❌ Go is not installed"; exit 1; }
	@command -v docker >/dev/null 2>&1 || { echo "❌ Docker is not installed"; exit 1; }
	@command -v terraform >/dev/null 2>&1 || { echo "❌ Terraform is not installed"; exit 1; }
	@command -v gcloud >/dev/null 2>&1 || { echo "❌ gcloud is not installed"; exit 1; }
	@echo "✅ All required tools are installed"

##@ 開発

dev: ## ローカル開発環境を起動
	@echo "==> Starting local development environment..."
	docker-compose up -d spanner-emulator
	@sleep 3
	@echo "==> Setting up Spanner Emulator..."
	cd backend && bash scripts/create-spanner-emulator.sh
	@echo "==> Starting API server..."
	cd backend && cp .env.example .env || true
	cd backend && go run cmd/api/main.go

dev-docker: docker-up migrate-up ## Docker Composeで開発環境を起動
	@echo "✅ Development environment is ready at http://localhost:8080"
	@echo "📊 Spanner Emulator: localhost:9010"

##@ Docker

docker-up: ## Docker Composeを起動
	docker-compose up -d

docker-down: ## Docker Composeを停止
	docker-compose down

docker-logs: ## Docker Composeのログを表示
	docker-compose logs -f

docker-build: ## Dockerイメージをビルド
	docker-compose build

docker-rebuild: ## Dockerイメージを再ビルド（キャッシュなし）
	docker-compose build --no-cache

##@ データベース

migrate-up: ## Spannerマイグレーションを適用
	@echo "==> Applying Spanner migrations..."
	@export SPANNER_EMULATOR_HOST=localhost:9010 && \
	cd backend && bash scripts/create-spanner-emulator.sh

migrate-check: ## Spannerスキーマを確認
	@export SPANNER_EMULATOR_HOST=localhost:9010 && \
	gcloud spanner databases ddl describe $(SPANNER_DATABASE) \
		--instance=$(SPANNER_INSTANCE) \
		--project=$(PROJECT_ID)

##@ ビルド・テスト

build: ## APIサーバーをビルド
	@echo "==> Building API server..."
	cd backend && go build -o bin/api ./cmd/api

test: ## テストを実行
	@echo "==> Running tests..."
	cd backend && go test -v -cover ./...

test-coverage: ## テストカバレッジを生成
	@echo "==> Generating test coverage..."
	cd backend && go test -coverprofile=coverage.out ./...
	cd backend && go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: backend/coverage.html"

lint: ## コードをリント
	@echo "==> Running linter..."
	cd backend && golangci-lint run ./...

fmt: ## コードをフォーマット
	@echo "==> Formatting code..."
	cd backend && go fmt ./...
	cd backend && goimports -w .

##@ Terraform

terraform-init: ## Terraformを初期化
	@echo "==> Initializing Terraform..."
	cd infra/terraform/environments/dev && terraform init

terraform-plan: ## Terraformプランを確認
	@echo "==> Running Terraform plan..."
	cd infra/terraform/environments/dev && terraform plan

terraform-apply: ## Terraformを適用
	@echo "==> Applying Terraform configuration..."
	cd infra/terraform/environments/dev && terraform apply

terraform-destroy: ## Terraformリソースを削除（注意！）
	@echo "==> Destroying Terraform resources..."
	cd infra/terraform/environments/dev && terraform destroy

##@ クリーンアップ

clean: ## ビルド成果物を削除
	@echo "==> Cleaning up..."
	cd backend && rm -rf bin/ coverage.out coverage.html
	docker-compose down -v
	@echo "✅ Cleanup complete"

clean-all: clean ## すべてのキャッシュを削除
	cd backend && go clean -cache -testcache -modcache
	rm -rf backend/.env
	@echo "✅ All caches cleared"

##@ その他

logs-api: ## APIサーバーのログを表示
	docker-compose logs -f api

logs-spanner: ## Spanner Emulatorのログを表示
	docker-compose logs -f spanner-emulator

health: ## ヘルスチェック
	@curl -f http://localhost:8080/health || echo "❌ API server is not running"

version: ## バージョン情報を表示
	@echo "Go version:"
	@go version
	@echo ""
	@echo "Docker version:"
	@docker --version
	@echo ""
	@echo "Terraform version:"
	@terraform --version
	@echo ""
	@echo "gcloud version:"
	@gcloud --version
