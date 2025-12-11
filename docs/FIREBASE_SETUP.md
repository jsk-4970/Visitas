# Firebase Authentication セットアップガイド

このドキュメントでは、Visitas プロジェクトに Firebase Authentication を統合する手順を説明します。

## 目次

1. [前提条件](#前提条件)
2. [Firebase プロジェクトのセットアップ](#firebase-プロジェクトのセットアップ)
3. [Identity Platform の有効化](#identity-platform-の有効化)
4. [Terraform による自動構築](#terraform-による自動構築)
5. [サービスアカウントキーの取得](#サービスアカウントキーの取得)
6. [ローカル開発環境の設定](#ローカル開発環境の設定)
7. [認証のテスト](#認証のテスト)
8. [トラブルシューティング](#トラブルシューティング)

## 前提条件

- GCP プロジェクト: `stunning-grin-480914-n1`
- GCP CLI (`gcloud`) がインストールされていること
- Terraform がインストールされていること
- Firebase Console へのアクセス権限

## Firebase プロジェクトのセットアップ

### 1. Firebase Console でプロジェクトを追加

1. [Firebase Console](https://console.firebase.google.com/) にアクセス
2. 「プロジェクトを追加」をクリック
3. **既存の GCP プロジェクトを選択**: `stunning-grin-480914-n1`
4. Firebase の利用規約に同意
5. Google Analytics は任意（開発環境では不要）

### 2. Authentication の有効化

1. Firebase Console で「Authentication」を選択
2. 「始める」をクリック
3. 「Sign-in method」タブを開く
4. 以下のプロバイダーを有効化：
   - **メール/パスワード**: 有効化（医師・スタッフのログイン用）
   - **Google**: 有効化（オプション：簡易ログイン用）

## Identity Platform の有効化

### GCP Console での設定

1. [GCP Console の Identity Platform](https://console.cloud.google.com/customer-identity) にアクセス
2. プロジェクト `stunning-grin-480914-n1` を選択
3. 「Identity Platform を有効にする」をクリック
4. プロバイダーの確認（Firebase で設定したものが表示される）

### API の有効化（CLI）

```bash
gcloud services enable identitytoolkit.googleapis.com \
  --project=stunning-grin-480914-n1
```

## Terraform による自動構築

### 1. Terraform 初期化

```bash
cd infra/terraform/environments/dev
terraform init
```

### 2. Terraform Plan の確認

```bash
terraform plan
```

以下のリソースが作成されることを確認：
- Identity Platform API の有効化
- Firebase Admin サービスアカウント
- サービスアカウントキーの作成
- Secret Manager への保存
- Cloud Run サービスアカウントへの権限付与

### 3. Terraform Apply

```bash
terraform apply
```

> **注意**: サービスアカウントキーが自動的に作成され、Secret Manager に保存されます。

### 4. 出力の確認

```bash
terraform output firebase_service_account_email
terraform output firebase_secret_name
```

## サービスアカウントキーの取得

### 方法1: Secret Manager から取得（推奨）

```bash
# Secret Manager からキーを取得
gcloud secrets versions access latest \
  --secret="firebase-service-account-dev" \
  --project="stunning-grin-480914-n1" \
  > backend/config/firebase-service-account.json

# パーミッション設定
chmod 600 backend/config/firebase-service-account.json
```

### 方法2: Firebase Console から手動ダウンロード

1. [Firebase Console](https://console.firebase.google.com/) を開く
2. プロジェクト設定 → サービスアカウント
3. 「新しい秘密鍵の生成」をクリック
4. ダウンロードした JSON ファイルを `backend/config/firebase-service-account.json` に配置

```bash
# ディレクトリ作成
mkdir -p backend/config

# ファイルを配置
mv ~/Downloads/stunning-grin-480914-n1-firebase-adminsdk-*.json \
   backend/config/firebase-service-account.json

# パーミッション設定
chmod 600 backend/config/firebase-service-account.json

# .gitignore に追加（既に追加済み）
echo "backend/config/firebase-service-account.json" >> .gitignore
```

## ローカル開発環境の設定

### 1. 環境変数の設定

`.env` ファイルを作成（`.env.example` をコピー）：

```bash
cd backend
cp .env.example .env
```

`.env` ファイルを編集：

```bash
# GCP Settings
GCP_PROJECT_ID=stunning-grin-480914-n1
GCP_REGION=asia-northeast1
SPANNER_INSTANCE=stunning-grin-480914-n1-instance
SPANNER_DATABASE=stunning-grin-480914-n1-db

# Firebase (Admin SDK)
FIREBASE_CONFIG_PATH=./config/firebase-service-account.json

# Server Settings
PORT=8080
ENV=development
LOG_LEVEL=debug

# CORS Settings
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
```

### 2. Go モジュールの依存関係をインストール

```bash
cd backend
go mod tidy
```

Firebase Admin SDK が追加されます：
- `firebase.google.com/go/v4`
- `firebase.google.com/go/v4/auth`

### 3. サーバーの起動

```bash
# 通常起動
go run cmd/api/main.go

# ホットリロード（開発時）
air
```

期待される起動ログ：

```
2024/12/12 12:00:00 No .env file found, using environment variables
2024/12/12 12:00:01 Firebase Authentication initialized successfully
2024/12/12 12:00:01 Starting server on :8080 (env: development)
```

Firebase が正しく初期化されていることを確認してください。

## 認証のテスト

### 1. テストユーザーの作成

Firebase Console または Admin SDK を使用してテストユーザーを作成：

#### Firebase Console で作成

1. Firebase Console → Authentication → Users
2. 「ユーザーを追加」をクリック
3. メールアドレスとパスワードを入力
4. 「ユーザーを追加」をクリック

#### curl でテストユーザーを作成（Admin SDK 経由）

```bash
# TODO: ユーザー作成APIを実装後に更新
```

### 2. ID トークンの取得

Firebase Authentication REST API を使用：

```bash
# ログイン
curl -X POST \
  'https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=YOUR_WEB_API_KEY' \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "test@example.com",
    "password": "testpassword123",
    "returnSecureToken": true
  }'
```

レスポンスから `idToken` を取得します。

> **Web API Key の取得**: Firebase Console → プロジェクト設定 → 全般 → ウェブ API キー

### 3. 認証付き API リクエストのテスト

```bash
# ID トークンを環境変数に設定
export ID_TOKEN="eyJhbGciOiJSUzI1NiIsImtpZCI6..."

# ヘルスチェック（認証不要）
curl http://localhost:8080/health

# 患者一覧取得（認証必要）
curl -X GET \
  http://localhost:8080/api/v1/patients \
  -H "Authorization: Bearer $ID_TOKEN"
```

期待されるレスポンス：
- **認証成功**: HTTP 200 + データ
- **トークンなし**: HTTP 401 "Missing authorization header"
- **無効なトークン**: HTTP 401 "Invalid or expired token"

### 4. ミドルウェアの動作確認

```bash
# 無効なトークンでテスト
curl -X GET \
  http://localhost:8080/api/v1/patients \
  -H "Authorization: Bearer invalid_token"

# Expected: HTTP 401 Unauthorized
```

## 実装されている機能

### 1. Firebase Client (`pkg/auth/firebase.go`)

- `NewFirebaseClient`: Firebase Admin SDK の初期化
- `VerifyIDToken`: ID トークンの検証
- `GetUser`: UID によるユーザー情報取得
- `CreateUser`: 新規ユーザー作成
- `UpdateUser`: ユーザー情報更新
- `DeleteUser`: ユーザー削除
- `SetCustomUserClaims`: カスタムクレーム設定（ロール・権限）

### 2. 認証ミドルウェア (`internal/middleware/auth.go`)

- `RequireAuth`: 認証必須ミドルウェア（全エンドポイント保護）
- `OptionalAuth`: 認証オプショナル（トークンがあれば検証）
- `RequireRole`: ロールベースアクセス制御（RBAC）

### 3. Context ヘルパー関数

- `GetUserIDFromContext`: リクエストコンテキストから UID を取得
- `GetUserEmailFromContext`: リクエストコンテキストからメールを取得
- `GetUserClaimsFromContext`: カスタムクレームを取得

### 使用例（ハンドラー内）

```go
func (h *PatientHandler) Get(w http.ResponseWriter, r *http.Request) {
    // 認証済みユーザーの UID を取得
    userID, ok := middleware.GetUserIDFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    log.Printf("User %s is accessing patient data", userID)
    // ... 処理続行
}
```

## Cloud Run へのデプロイ

### 1. Secret Manager の設定

Secret Manager に Firebase サービスアカウントキーが保存されていることを確認：

```bash
gcloud secrets list --project=stunning-grin-480914-n1
```

### 2. Cloud Run サービスのデプロイ

```bash
gcloud run deploy visitas-api \
  --image gcr.io/stunning-grin-480914-n1/visitas-api:latest \
  --platform managed \
  --region asia-northeast1 \
  --service-account visitas-dev-run@stunning-grin-480914-n1.iam.gserviceaccount.com \
  --set-secrets FIREBASE_CONFIG_PATH=firebase-service-account-dev:latest \
  --allow-unauthenticated=false
```

### 3. 環境変数の設定

Cloud Run の環境変数として設定：

```bash
gcloud run services update visitas-api \
  --region asia-northeast1 \
  --set-env-vars GCP_PROJECT_ID=stunning-grin-480914-n1,\
GCP_REGION=asia-northeast1,\
SPANNER_INSTANCE=stunning-grin-480914-n1-instance,\
SPANNER_DATABASE=stunning-grin-480914-n1-db,\
ENV=production
```

## セキュリティベストプラクティス

### 1. サービスアカウントキーの管理

- ✅ Secret Manager に保存（Terraform で自動化済み）
- ✅ `.gitignore` に追加済み
- ✅ ファイルパーミッション 600
- ❌ コードにハードコードしない
- ❌ 公開リポジトリにコミットしない

### 2. トークンの有効期限

Firebase ID トークンは **1時間** で自動的に期限切れになります。クライアント側でリフレッシュトークンを使用して更新してください。

### 3. カスタムクレーム（ロール・権限）

医師、看護師、管理者など、ロールベースのアクセス制御を実装：

```go
// 例: 医師のみアクセス可能なエンドポイント
r.Route("/admin", func(r chi.Router) {
    r.Use(authMiddleware.RequireRole("doctor"))
    r.Get("/reports", adminHandler.GetReports)
})
```

カスタムクレームの設定：

```go
claims := map[string]interface{}{
    "role": "doctor",
    "clinic_id": "clinic_123",
}
firebaseClient.SetCustomUserClaims(ctx, userUID, claims)
```

### 4. CORS 設定

本番環境では、信頼できるオリジンのみ許可：

```bash
ALLOWED_ORIGINS=https://visitas.app,https://admin.visitas.app
```

## トラブルシューティング

### 問題: Firebase client の初期化失敗

**エラーメッセージ**:
```
Failed to initialize Firebase client: failed to initialize Firebase app: ...
```

**解決策**:
1. `FIREBASE_CONFIG_PATH` が正しく設定されているか確認
2. サービスアカウントキーファイルが存在するか確認
3. JSON ファイルの形式が正しいか確認

```bash
cat backend/config/firebase-service-account.json | jq
```

### 問題: トークン検証失敗

**エラーメッセージ**:
```
Invalid or expired token
```

**解決策**:
1. トークンの有効期限を確認（1時間）
2. Web API Key が正しいプロジェクトのものか確認
3. Firebase Console で該当ユーザーが存在するか確認

### 問題: 権限エラー

**エラーメッセージ**:
```
Permission denied
```

**解決策**:
1. サービスアカウントに `roles/firebase.admin` が付与されているか確認

```bash
gcloud projects get-iam-policy stunning-grin-480914-n1 \
  --flatten="bindings[].members" \
  --filter="bindings.members:serviceAccount:firebase-admin-dev@*"
```

2. Cloud Run サービスアカウントに Secret Manager へのアクセス権限があるか確認

```bash
gcloud secrets get-iam-policy firebase-service-account-dev \
  --project=stunning-grin-480914-n1
```

### 問題: ローカルで動くが Cloud Run で動かない

**解決策**:
1. Secret Manager から正しくシークレットがマウントされているか確認
2. Cloud Run のログを確認

```bash
gcloud run logs read visitas-api --region asia-northeast1 --limit 50
```

3. サービスアカウントの権限を確認

## 参考リソース

- [Firebase Authentication ドキュメント](https://firebase.google.com/docs/auth)
- [Firebase Admin SDK (Go)](https://firebase.google.com/docs/admin/setup?hl=ja#go)
- [Identity Platform ドキュメント](https://cloud.google.com/identity-platform/docs)
- [GCP Secret Manager](https://cloud.google.com/secret-manager/docs)

## 次のステップ

1. **ユーザー管理 API の実装**: ユーザー作成・更新・削除エンドポイント
2. **ロールベースアクセス制御**: 医師・看護師・管理者の権限管理
3. **モバイルアプリ統合**: Flutter での Firebase Authentication 統合
4. **パスワードリセット**: メール経由のパスワードリセット機能
5. **多要素認証（MFA）**: Identity Platform の MFA 機能を有効化
