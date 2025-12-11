# テスト・検証ガイド

このドキュメントでは、Visitas プロジェクトの整合性確認と動作検証の手順を説明します。

## 目次

1. [事前準備](#事前準備)
2. [ローカル環境での検証](#ローカル環境での検証)
3. [統合性チェック](#統合性チェック)
4. [ビルド検証](#ビルド検証)
5. [デプロイ前チェックリスト](#デプロイ前チェックリスト)
6. [本番デプロイ後の検証](#本番デプロイ後の検証)

## 事前準備

### 必要なツール

以下のツールがインストールされていることを確認してください：

```bash
# Go (1.22以上)
go version

# Docker
docker --version

# gcloud CLI
gcloud --version

# Terraform (オプション)
terraform --version
```

### 依存関係のインストール

```bash
# Go モジュールの依存関係をダウンロード
cd backend
go mod download
go mod tidy
go mod verify
```

## ローカル環境での検証

### 1. プロジェクト構造の確認

```bash
# プロジェクトルートで実行
./scripts/verify-project.sh
```

期待される出力：
- ✅ すべての必須ファイルが存在
- ✅ すべての必須ディレクトリが存在
- ✅ Go ファイル数が正常

### 2. 設定ファイルの整合性確認

```bash
# 環境変数の確認
cd backend
cat .env.example

# 実際の .env ファイルを作成（まだの場合）
cp .env.example .env

# 必要に応じて編集
vim .env
```

**確認項目：**
- [ ] `GCP_PROJECT_ID` が正しい
- [ ] `GCP_REGION` が `asia-northeast1` または `asia-northeast2`
- [ ] `SPANNER_INSTANCE` が正しい
- [ ] `SPANNER_DATABASE` が正しい
- [ ] `FIREBASE_CONFIG_PATH` が正しい（後で設定可能）

### 3. Go コードの構文チェック

```bash
cd backend

# Go の静的解析
go vet ./...

# 型チェック
go build -o /dev/null ./...

# インポートパスの確認
go list -m all
```

期待される結果：
```
✅ エラーなし
✅ すべてのパッケージがコンパイル可能
```

### 4. Lint の実行

```bash
# golangci-lint のインストール（まだの場合）
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Lint 実行
cd backend
golangci-lint run --timeout=5m
```

### 5. ユニットテストの実行

```bash
cd backend

# すべてのテストを実行
go test -v ./...

# カバレッジ付き
go test -v -race -coverprofile=coverage.out ./...

# カバレッジレポートを表示
go tool cover -html=coverage.out
```

**注意：** 現在は実際のテストコードがないため、テストはスキップされます。Sprint 2以降で実装予定。

### 6. Spanner Emulator でのテスト（オプション）

```bash
# Spanner Emulator を起動
docker run -d -p 9010:9010 gcr.io/cloud-spanner-emulator/emulator

# 環境変数を設定
export SPANNER_EMULATOR_HOST=localhost:9010

# アプリケーションを起動
cd backend
go run cmd/api/main.go
```

別のターミナルでテスト：

```bash
# ヘルスチェック
curl http://localhost:8080/health

# 患者 API（認証なしの場合はエラーになる）
curl http://localhost:8080/api/v1/patients
```

## 統合性チェック

### 自動チェックスクリプト

プロジェクトルートで統合チェックスクリプトを実行：

```bash
./scripts/verify-all.sh
```

このスクリプトは以下をチェックします：

1. **ファイル構造**
   - 必須ファイルの存在確認
   - ディレクトリ構造の確認

2. **設定の整合性**
   - 環境変数の一貫性
   - プロジェクト ID の一致
   - リージョン設定の一致

3. **Go コード**
   - インポートパスの正確性
   - 構文エラーのチェック
   - モジュール依存関係の検証

4. **Docker 設定**
   - Dockerfile の構文チェック
   - ポート設定の確認

5. **CI/CD 設定**
   - GitHub Actions ワークフローの検証
   - Cloud Build 設定の検証

## ビルド検証

### Docker イメージのビルド

```bash
cd backend

# イメージをビルド
docker build -t visitas-api:test .

# ビルドが成功したことを確認
docker images | grep visitas-api
```

期待される出力：
```
visitas-api   test   <IMAGE_ID>   X seconds ago   XX MB
```

### Docker コンテナの実行テスト

```bash
# コンテナを起動（Spanner Emulator 使用）
docker run -d \
  --name visitas-test \
  -p 8080:8080 \
  -e GCP_PROJECT_ID=test-project \
  -e SPANNER_EMULATOR_HOST=host.docker.internal:9010 \
  -e SPANNER_INSTANCE=test-instance \
  -e SPANNER_DATABASE=test-db \
  visitas-api:test

# ログを確認
docker logs visitas-test

# ヘルスチェック
curl http://localhost:8080/health

# コンテナを停止・削除
docker stop visitas-test
docker rm visitas-test
```

### マルチステージビルドの検証

```bash
# ビルドステージのサイズ確認
docker build --target builder -t visitas-builder:test backend/

# 最終イメージのサイズ確認（小さいはず）
docker images | grep visitas-api
```

期待される結果：
- Builder イメージ: ~500-800 MB
- 最終イメージ: ~20-50 MB

## デプロイ前チェックリスト

### Sprint 1 完了チェック

- [ ] **GCP プロジェクトセットアップ**
  - [ ] プロジェクト ID: `stunning-grin-480914-n1`
  - [ ] 必要な API が有効化されている
  - [ ] IAM 権限が正しく設定されている

- [ ] **Cloud Spanner**
  - [ ] インスタンスが作成されている
  - [ ] データベースが作成されている
  - [ ] マイグレーションが適用されている

- [ ] **Firebase Authentication**
  - [ ] Firebase プロジェクトが設定されている
  - [ ] Identity Platform が有効化されている
  - [ ] サービスアカウントキーが Secret Manager に保存されている

- [ ] **Go API サーバー**
  - [ ] コードがコンパイル可能
  - [ ] すべての依存関係が解決されている
  - [ ] 環境変数が正しく設定されている

- [ ] **Docker**
  - [ ] Dockerfile が正しくビルドできる
  - [ ] イメージサイズが適切（<100MB）
  - [ ] ヘルスチェックが実装されている

- [ ] **CI/CD**
  - [ ] GitHub Actions ワークフローが設定されている
  - [ ] Cloud Build 設定が完了している
  - [ ] デプロイスクリプトが動作する

### Terraform チェック

```bash
cd infra/terraform/environments/dev

# 初期化
terraform init

# プラン確認
terraform plan

# 構文チェック
terraform validate

# フォーマット確認
terraform fmt -check
```

期待される結果：
```
✅ Terraform initialized successfully
✅ Plan shows expected resources
✅ Configuration is valid
✅ All files are properly formatted
```

### Secret Manager チェック

```bash
# Firebase サービスアカウントキーの存在確認
gcloud secrets list --project=stunning-grin-480914-n1 | grep firebase

# シークレットのバージョン確認
gcloud secrets versions list firebase-service-account-dev \
  --project=stunning-grin-480914-n1
```

### Artifact Registry チェック

```bash
# リポジトリの存在確認
gcloud artifacts repositories list \
  --location=asia-northeast1 \
  --project=stunning-grin-480914-n1

# 認証設定
gcloud auth configure-docker asia-northeast1-docker.pkg.dev
```

## 本番デプロイ後の検証

### 1. Cloud Run サービスの確認

```bash
# サービスの存在確認
gcloud run services list \
  --region=asia-northeast1 \
  --project=stunning-grin-480914-n1

# サービスの詳細確認
gcloud run services describe visitas-api-dev \
  --region=asia-northeast1 \
  --project=stunning-grin-480914-n1

# URL の取得
SERVICE_URL=$(gcloud run services describe visitas-api-dev \
  --region=asia-northeast1 \
  --project=stunning-grin-480914-n1 \
  --format='value(status.url)')

echo "Service URL: $SERVICE_URL"
```

### 2. ヘルスチェック

```bash
# ヘルスエンドポイントの確認
curl -X GET $SERVICE_URL/health

# 期待されるレスポンス:
# {"status":"ok","timestamp":"..."}
```

### 3. 認証テスト

```bash
# 認証なしでアクセス（401 が返るべき）
curl -X GET $SERVICE_URL/api/v1/patients

# 期待されるレスポンス:
# Missing authorization header
```

### 4. Firebase 認証フローのテスト

```bash
# テストユーザーでログイン（Firebase REST API）
# Web API Key を取得: Firebase Console → プロジェクト設定 → Web API Key

WEB_API_KEY="your-web-api-key"

# ログイン
TOKEN_RESPONSE=$(curl -X POST \
  "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=$WEB_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "test@example.com",
    "password": "testpassword123",
    "returnSecureToken": true
  }')

# ID トークンを取得
ID_TOKEN=$(echo $TOKEN_RESPONSE | jq -r '.idToken')

# 認証付きリクエスト
curl -X GET $SERVICE_URL/api/v1/patients \
  -H "Authorization: Bearer $ID_TOKEN"

# 期待されるレスポンス:
# {"message":"List patients - TODO","data":[]}
```

### 5. ログの確認

```bash
# 最新のログを表示
gcloud run logs read visitas-api-dev \
  --region=asia-northeast1 \
  --project=stunning-grin-480914-n1 \
  --limit=50

# リアルタイムログ
gcloud run logs tail visitas-api-dev \
  --region=asia-northeast1 \
  --project=stunning-grin-480914-n1
```

### 6. メトリクスの確認

```bash
# Cloud Console でメトリクスを確認
echo "https://console.cloud.google.com/run/detail/asia-northeast1/visitas-api-dev/metrics?project=stunning-grin-480914-n1"

# または CLI で
gcloud run services describe visitas-api-dev \
  --region=asia-northeast1 \
  --project=stunning-grin-480914-n1 \
  --format=yaml | grep -A10 conditions
```

確認項目：
- [ ] リクエスト数
- [ ] レスポンスタイム
- [ ] エラー率
- [ ] インスタンス数
- [ ] CPU 使用率
- [ ] メモリ使用率

### 7. パフォーマンステスト

```bash
# Apache Bench を使用したシンプルな負荷テスト
ab -n 100 -c 10 $SERVICE_URL/health

# 期待される結果:
# - 成功率 100%
# - 平均レスポンスタイム < 200ms
# - エラー 0%
```

### 8. セキュリティチェック

```bash
# HTTPS 接続の確認
curl -I $SERVICE_URL/health | grep -i strict-transport

# CORS ヘッダーの確認
curl -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: GET" \
  -X OPTIONS \
  $SERVICE_URL/api/v1/patients \
  -v

# 期待されるヘッダー:
# Access-Control-Allow-Origin: http://localhost:3000
# Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
```

## トラブルシューティング

### ビルドエラー

#### 問題: `go.sum` が存在しない

```bash
cd backend
go mod tidy
```

#### 問題: パッケージが見つからない

```bash
cd backend
go mod download
go clean -modcache
go mod download
```

### Docker ビルドエラー

#### 問題: コンテキストが大きすぎる

`.dockerignore` を確認：

```bash
cat backend/.dockerignore
```

以下を追加：
```
.git
.env
.env.local
tmp/
vendor/
coverage.out
```

#### 問題: マルチステージビルドの失敗

```bash
# 各ステージを個別にビルド
docker build --target builder -t test-builder backend/
```

### デプロイエラー

#### 問題: 権限エラー

```bash
# サービスアカウントの権限確認
gcloud projects get-iam-policy stunning-grin-480914-n1 \
  --flatten="bindings[].members" \
  --filter="bindings.members:serviceAccount:*visitas*"
```

#### 問題: シークレットにアクセスできない

```bash
# シークレットの権限確認
gcloud secrets get-iam-policy firebase-service-account-dev \
  --project=stunning-grin-480914-n1
```

### ランタイムエラー

#### 問題: Firebase 初期化エラー

```bash
# ログを確認
gcloud run logs read visitas-api-dev \
  --region=asia-northeast1 \
  --limit=20 | grep -i firebase

# シークレットがマウントされているか確認
gcloud run services describe visitas-api-dev \
  --region=asia-northeast1 \
  --format=yaml | grep -A5 secrets
```

#### 問題: Spanner 接続エラー

```bash
# Spanner インスタンスの確認
gcloud spanner instances describe stunning-grin-480914-n1-instance \
  --project=stunning-grin-480914-n1

# データベースの確認
gcloud spanner databases list \
  --instance=stunning-grin-480914-n1-instance \
  --project=stunning-grin-480914-n1
```

## 自動テストスクリプト

### 完全検証スクリプト

```bash
#!/bin/bash
# scripts/verify-all.sh

set -e

echo "🔍 Visitas Project Verification"
echo "================================"
echo ""

# 1. Project structure
echo "1️⃣  Checking project structure..."
./scripts/verify-project.sh

# 2. Go code
echo ""
echo "2️⃣  Verifying Go code..."
cd backend
go mod verify
go vet ./...
cd ..

# 3. Docker build
echo ""
echo "3️⃣  Testing Docker build..."
docker build -t visitas-api:test backend/ > /dev/null 2>&1
echo "✅ Docker build successful"

# 4. Terraform validation
echo ""
echo "4️⃣  Validating Terraform..."
cd infra/terraform/environments/dev
terraform init -backend=false > /dev/null 2>&1
terraform validate
cd ../../../..

echo ""
echo "✅ All verifications passed!"
```

## 参考資料

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Docker Build Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Cloud Run Testing](https://cloud.google.com/run/docs/testing)
- [Terraform Testing](https://www.terraform.io/docs/language/modules/testing.html)
