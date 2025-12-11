# CI/CD パイプライン ドキュメント

このドキュメントでは、Visitas プロジェクトの継続的インテグレーション（CI）と継続的デリバリー（CD）パイプラインについて説明します。

## 目次

1. [概要](#概要)
2. [GitHub Actions](#github-actions)
3. [Cloud Build](#cloud-build)
4. [手動デプロイ](#手動デプロイ)
5. [環境設定](#環境設定)
6. [デプロイフロー](#デプロイフロー)
7. [トラブルシューティング](#トラブルシューティング)

## 概要

Visitas プロジェクトでは、以下の2つのCI/CDシステムを使用します：

1. **GitHub Actions** - プルリクエストのテスト、Linting、自動デプロイ
2. **Cloud Build** - GCP ネイティブなビルド＆デプロイ（オプション）

### パイプラインの流れ

```
コード変更
  ↓
Git Push / PR作成
  ↓
自動テスト & Lint (GitHub Actions)
  ↓
テスト成功
  ↓
Docker イメージビルド
  ↓
Artifact Registry にプッシュ
  ↓
Cloud Run にデプロイ
  ↓
ヘルスチェック
  ↓
デプロイ完了 🎉
```

## GitHub Actions

GitHub Actions は `.github/workflows/` ディレクトリに定義されています。

### ワークフロー一覧

#### 1. Test ワークフロー ([.github/workflows/test.yml](../.github/workflows/test.yml))

**トリガー:**
- Pull Request が `main` または `develop` ブランチに作成された時
- `backend/**` または `.github/workflows/test.yml` に変更があった時

**ジョブ:**

1. **Lint** - コード品質チェック
   - `go vet`: Go の静的解析
   - `golangci-lint`: 包括的な Lint ツール

2. **Test** - ユニットテスト実行
   - `go test -race`: データ競合検出付きテスト
   - カバレッジレポート生成
   - Codecov へのアップロード

3. **Build** - Docker イメージビルド検証
   - `docker build`: イメージが正しくビルドできるか確認
   - キャッシュ活用による高速化

4. **Security** - セキュリティスキャン
   - `gosec`: Go セキュリティスキャナー
   - `trivy`: 脆弱性スキャナー

**使用例:**

```bash
# PR を作成すると自動実行
git checkout -b feature/new-feature
git add .
git commit -m "Add new feature"
git push origin feature/new-feature
# → GitHub で PR 作成 → 自動的にテストが実行される
```

#### 2. Deploy ワークフロー ([.github/workflows/deploy.yml](../.github/workflows/deploy.yml))

**トリガー:**
- `main` ブランチへの Push → Production 環境へデプロイ
- `develop` ブランチへの Push → Development 環境へデプロイ
- 手動トリガー（workflow_dispatch）

**ジョブ:**

1. **Setup** - 環境変数の設定
   - ブランチに応じて環境（dev/staging/prod）を決定

2. **Build and Push** - Docker イメージのビルドとプッシュ
   - マルチタグ付与（commit SHA、latest、環境別 latest）
   - Artifact Registry へプッシュ

3. **Deploy** - Cloud Run へのデプロイ
   - サービスアカウント設定
   - 環境変数とシークレット設定
   - リソース制限設定

4. **Notify** - デプロイ結果の通知

**環境別設定:**

| 環境 | ブランチ | Min Instances | Max Instances | CPU | Memory |
|------|----------|---------------|---------------|-----|--------|
| dev | develop | 0 | 3 | 1 | 512Mi |
| staging | staging | 0 | 5 | 1 | 512Mi |
| prod | main | 1 | 20 | 2 | 1Gi |

### GitHub Actions のセットアップ

#### 1. Workload Identity Federation の設定

GitHub Actions から GCP にアクセスするために、Workload Identity Federation を使用します（サービスアカウントキーよりセキュア）。

```bash
# プロジェクト変数
PROJECT_ID="stunning-grin-480914-n1"
PROJECT_NUMBER=$(gcloud projects describe $PROJECT_ID --format="value(projectNumber)")
REPO="your-github-username/visitas"

# Workload Identity Pool の作成
gcloud iam workload-identity-pools create "github-actions" \
  --project="$PROJECT_ID" \
  --location="global" \
  --display-name="GitHub Actions Pool"

# Workload Identity Provider の作成
gcloud iam workload-identity-pools providers create-oidc "github-actions-provider" \
  --project="$PROJECT_ID" \
  --location="global" \
  --workload-identity-pool="github-actions" \
  --display-name="GitHub Actions Provider" \
  --attribute-mapping="google.subject=assertion.sub,attribute.actor=assertion.actor,attribute.repository=assertion.repository" \
  --issuer-uri="https://token.actions.githubusercontent.com"

# サービスアカウントの作成
gcloud iam service-accounts create github-actions \
  --project="$PROJECT_ID" \
  --display-name="GitHub Actions Service Account"

# 必要な権限を付与
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:github-actions@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/run.admin"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:github-actions@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/artifactregistry.writer"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:github-actions@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/iam.serviceAccountUser"

# Workload Identity Binding の設定
gcloud iam service-accounts add-iam-policy-binding \
  "github-actions@$PROJECT_ID.iam.gserviceaccount.com" \
  --project="$PROJECT_ID" \
  --role="roles/iam.workloadIdentityUser" \
  --member="principalSet://iam.googleapis.com/projects/$PROJECT_NUMBER/locations/global/workloadIdentityPools/github-actions/attribute.repository/$REPO"
```

#### 2. GitHub Secrets の設定

GitHub リポジトリに以下のシークレットを設定：

1. **Settings** → **Secrets and variables** → **Actions** → **New repository secret**

| Secret Name | Value | 説明 |
|-------------|-------|------|
| `GCP_WORKLOAD_IDENTITY_PROVIDER` | `projects/PROJECT_NUMBER/locations/global/workloadIdentityPools/github-actions/providers/github-actions-provider` | Workload Identity Provider のリソース名 |
| `GCP_SERVICE_ACCOUNT` | `github-actions@stunning-grin-480914-n1.iam.gserviceaccount.com` | サービスアカウントのメールアドレス |

#### 3. 環境の設定（オプション）

GitHub リポジトリに環境を設定して、承認フローを追加できます。

1. **Settings** → **Environments** → **New environment**
2. 環境名を入力（`dev`, `staging`, `prod`）
3. **Protection rules** を設定:
   - `prod` 環境: Required reviewers（承認者を指定）
   - デプロイ前に手動承認が必要

## Cloud Build

Cloud Build は GCP ネイティブな CI/CD サービスです。GitHub Actions の代替として使用できます。

### Cloud Build 設定ファイル

[cloudbuild.yaml](../cloudbuild.yaml) にビルド設定が定義されています。

**ステップ:**

1. **Test** - Go テストの実行
2. **Lint** - golangci-lint の実行
3. **Build** - Docker イメージのビルド
4. **Push** - Artifact Registry へのプッシュ
5. **Deploy** - Cloud Run へのデプロイ
6. **Test Deployment** - ヘルスチェック

### Cloud Build のセットアップ

#### 1. セットアップスクリプトの実行

```bash
# スクリプトを編集して REPO_OWNER を設定
vim scripts/setup-cloudbuild.sh

# スクリプトを実行
./scripts/setup-cloudbuild.sh
```

#### 2. GitHub リポジトリの接続

1. [Cloud Build Console](https://console.cloud.google.com/cloud-build/triggers/connect?project=stunning-grin-480914-n1) を開く
2. 「リポジトリを接続」をクリック
3. GitHub を選択
4. リポジトリを認証・選択
5. 接続を確認

#### 3. トリガーの確認

```bash
gcloud builds triggers list --project=stunning-grin-480914-n1
```

作成されたトリガー:
- `visitas-backend-dev`: `develop` ブランチへの Push
- `visitas-backend-staging`: `staging` ブランチへの Push
- `visitas-backend-prod`: `main` ブランチへの Push

### 手動でビルドを実行

```bash
# 開発環境へデプロイ
gcloud builds submit \
  --config=cloudbuild.yaml \
  --substitutions=_ENVIRONMENT=dev \
  --project=stunning-grin-480914-n1

# 本番環境へデプロイ
gcloud builds submit \
  --config=cloudbuild.yaml \
  --substitutions=_ENVIRONMENT=prod \
  --project=stunning-grin-480914-n1
```

## 手動デプロイ

緊急時やローカルからのデプロイには、デプロイスクリプトを使用します。

### デプロイスクリプトの使用

```bash
# 開発環境へデプロイ
./scripts/deploy.sh dev

# ステージング環境へデプロイ
./scripts/deploy.sh staging

# 本番環境へデプロイ（確認プロンプトあり）
./scripts/deploy.sh prod
```

**スクリプトの処理内容:**

1. テストの実行
2. Docker イメージのビルド
3. Artifact Registry へのプッシュ
4. Cloud Run へのデプロイ
5. ヘルスチェック
6. デプロイ結果の表示

### Docker を使用した手動デプロイ

```bash
cd backend

# イメージをビルド
docker build -t visitas-api:local .

# ローカルでテスト
docker run -p 8080:8080 \
  -e GCP_PROJECT_ID=stunning-grin-480914-n1 \
  -e SPANNER_EMULATOR_HOST=host.docker.internal:9010 \
  visitas-api:local

# Artifact Registry へプッシュ
docker tag visitas-api:local \
  asia-northeast1-docker.pkg.dev/stunning-grin-480914-n1/visitas-dev/api:manual

docker push asia-northeast1-docker.pkg.dev/stunning-grin-480914-n1/visitas-dev/api:manual

# Cloud Run へデプロイ
gcloud run deploy visitas-api-dev \
  --image asia-northeast1-docker.pkg.dev/stunning-grin-480914-n1/visitas-dev/api:manual \
  --platform managed \
  --region asia-northeast1
```

## 環境設定

### 環境変数

各環境で設定される環境変数：

| 変数名 | 説明 | 例 |
|--------|------|-----|
| `GCP_PROJECT_ID` | GCP プロジェクト ID | `stunning-grin-480914-n1` |
| `GCP_REGION` | GCP リージョン | `asia-northeast1` |
| `SPANNER_INSTANCE` | Spanner インスタンス名 | `stunning-grin-480914-n1-instance` |
| `SPANNER_DATABASE` | Spanner データベース名 | `stunning-grin-480914-n1-db` |
| `ENV` | 環境名 | `dev` / `staging` / `prod` |
| `LOG_LEVEL` | ログレベル | `debug` / `info` / `error` |

### シークレット

Secret Manager に保存されるシークレット：

| シークレット名 | 説明 | マウント先 |
|---------------|------|-----------|
| `firebase-service-account-dev` | Firebase サービスアカウントキー（開発） | `FIREBASE_CONFIG_PATH` |
| `firebase-service-account-staging` | Firebase サービスアカウントキー（ステージング） | `FIREBASE_CONFIG_PATH` |
| `firebase-service-account-prod` | Firebase サービスアカウントキー（本番） | `FIREBASE_CONFIG_PATH` |

## デプロイフロー

### 開発フロー（Feature → Develop）

```bash
# 1. Feature ブランチを作成
git checkout -b feature/new-api-endpoint

# 2. コードを変更
# ... コーディング ...

# 3. コミット & プッシュ
git add .
git commit -m "feat(api): add new endpoint"
git push origin feature/new-api-endpoint

# 4. Pull Request を作成
# → GitHub で PR を作成
# → 自動的にテスト & Lint が実行される

# 5. レビュー & マージ
# → PR が承認されたら develop にマージ

# 6. 自動デプロイ
# → develop ブランチへのマージで自動的に dev 環境へデプロイ
```

### リリースフロー（Develop → Main）

```bash
# 1. Develop から Main へ PR を作成
git checkout main
git pull origin main
git checkout -b release/v1.0.0

# 2. バージョン番号を更新（必要に応じて）
# ... version ファイルを更新 ...

git add .
git commit -m "chore: bump version to v1.0.0"
git push origin release/v1.0.0

# 3. PR を作成（develop → main）
# → 承認フローが必要（prod 環境の設定による）

# 4. マージ
# → main ブランチへマージ
# → 自動的に prod 環境へデプロイ

# 5. タグを作成
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### ホットフィックスフロー（Main → Hotfix → Main）

```bash
# 1. Hotfix ブランチを作成
git checkout main
git pull origin main
git checkout -b hotfix/critical-bug

# 2. バグ修正
# ... コーディング ...

git add .
git commit -m "fix: resolve critical security issue"
git push origin hotfix/critical-bug

# 3. PR を作成 & マージ（main へ）
# → 承認後にマージ
# → 自動的に prod 環境へデプロイ

# 4. Develop へもバックポート
git checkout develop
git merge hotfix/critical-bug
git push origin develop
```

## トラブルシューティング

### GitHub Actions のエラー

#### 問題: Workload Identity 認証エラー

```
Error: google-github-actions/auth failed with: retry function failed after 1 attempt(s)
```

**解決策:**

1. Workload Identity Provider が正しく設定されているか確認

```bash
gcloud iam workload-identity-pools describe github-actions \
  --location=global \
  --project=stunning-grin-480914-n1
```

2. GitHub Secrets が正しく設定されているか確認
   - `GCP_WORKLOAD_IDENTITY_PROVIDER`
   - `GCP_SERVICE_ACCOUNT`

3. サービスアカウントに必要な権限があるか確認

```bash
gcloud projects get-iam-policy stunning-grin-480914-n1 \
  --flatten="bindings[].members" \
  --filter="bindings.members:serviceAccount:github-actions@*"
```

#### 問題: テスト失敗

```
Error: Process completed with exit code 1.
```

**解決策:**

1. ローカルでテストを実行して原因を特定

```bash
cd backend
go test -v ./...
```

2. 依存関係の問題を確認

```bash
go mod tidy
go mod verify
```

3. キャッシュをクリア（GitHub Actions）
   - Settings → Actions → Caches → Delete cache

### Cloud Build のエラー

#### 問題: 権限エラー

```
ERROR: (gcloud.run.deploy) PERMISSION_DENIED: Permission denied on resource project
```

**解決策:**

1. Cloud Build サービスアカウントに権限を付与

```bash
PROJECT_NUMBER=$(gcloud projects describe stunning-grin-480914-n1 --format="value(projectNumber)")
CLOUDBUILD_SA="${PROJECT_NUMBER}@cloudbuild.gserviceaccount.com"

gcloud projects add-iam-policy-binding stunning-grin-480914-n1 \
  --member="serviceAccount:${CLOUDBUILD_SA}" \
  --role="roles/run.admin"
```

#### 問題: Docker ビルドエラー

```
ERROR: failed to solve: failed to compute cache key
```

**解決策:**

1. Dockerfile の構文を確認
2. `.dockerignore` ファイルを確認
3. キャッシュをクリアして再ビルド

```bash
gcloud builds submit --no-cache
```

### Cloud Run デプロイエラー

#### 問題: サービス起動失敗

```
ERROR: (gcloud.run.deploy) Cloud Run error: Container failed to start.
```

**解決策:**

1. ログを確認

```bash
gcloud run logs read visitas-api-dev --region asia-northeast1 --limit 50
```

2. ローカルで Docker イメージをテスト

```bash
docker run -p 8080:8080 visitas-api:local
curl http://localhost:8080/health
```

3. 環境変数とシークレットが正しく設定されているか確認

```bash
gcloud run services describe visitas-api-dev \
  --region asia-northeast1 \
  --format yaml
```

#### 問題: ヘルスチェック失敗

```
ERROR: Health check failed
```

**解決策:**

1. `/health` エンドポイントが実装されているか確認
2. ポート設定が正しいか確認（8080）
3. タイムアウト設定を確認

## ベストプラクティス

### 1. コミットメッセージ規約

Conventional Commits を使用：

```
feat(api): add patient search endpoint
fix(auth): resolve token expiration issue
docs: update deployment guide
chore: bump dependencies
```

### 2. ブランチ戦略

- `main`: 本番環境（prod）
- `develop`: 開発環境（dev）
- `staging`: ステージング環境（オプション）
- `feature/*`: 機能開発
- `hotfix/*`: 緊急修正
- `release/*`: リリース準備

### 3. タグ付け

セマンティックバージョニングを使用：

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### 4. ロールバック

問題が発生した場合のロールバック手順：

```bash
# 前のバージョンのイメージを取得
IMAGE_TAG="previous-commit-sha"
IMAGE="asia-northeast1-docker.pkg.dev/stunning-grin-480914-n1/visitas-prod/api:$IMAGE_TAG"

# ロールバック
gcloud run deploy visitas-api-prod \
  --image $IMAGE \
  --region asia-northeast1
```

### 5. モニタリング

デプロイ後のモニタリング：

```bash
# ログの確認
gcloud run logs read visitas-api-prod --region asia-northeast1 --limit 100

# メトリクスの確認
gcloud run services describe visitas-api-prod \
  --region asia-northeast1 \
  --format="value(status.url)"

# Cloud Console でメトリクスを表示
# https://console.cloud.google.com/run/detail/asia-northeast1/visitas-api-prod/metrics
```

## 参考リソース

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Cloud Build Documentation](https://cloud.google.com/build/docs)
- [Cloud Run Documentation](https://cloud.google.com/run/docs)
- [Workload Identity Federation](https://cloud.google.com/iam/docs/workload-identity-federation)
- [Artifact Registry Documentation](https://cloud.google.com/artifact-registry/docs)
