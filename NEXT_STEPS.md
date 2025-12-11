# Visitas プロジェクト - 次のステップ

## 完了したタスク ✅

### 1. データベースマイグレーションファイル作成
- ✅ `backend/migrations/` ディレクトリに15個のマイグレーションファイルを作成
- ✅ 患者マスター、識別子、社会的背景、保険情報、病名、アレルギー、バイタル、ケア計画、ACP、処方、訪問スケジュール、ロジスティクス、ルート最適化、監査ログ、スタッフ管理テーブル
- ✅ PostgreSQL構文でCloud Spanner PostgreSQLインターフェース対応
- ✅ マイグレーションREADME作成

### 2. Go環境のセットアップ手順ドキュメント化
- ✅ [`docs/SETUP.md`](docs/SETUP.md) 作成
- ✅ Homebrew、Go、GCP CLI、開発ツールのインストール手順
- ✅ データベースセットアップ（PostgreSQL/Spanner）
- ✅ トラブルシューティングガイド

### 3. Firebase設定ファイルの配置手順ドキュメント化
- ✅ `backend/config/README.md` 作成
- ✅ Firebase Console から秘密鍵を取得する手順
- ✅ `firebase-config.example.json` テンプレート作成
- ✅ セキュリティガイドライン (.gitignore、ファイル権限)

### 4. OpenAPI仕様書の作成
- ✅ [`docs/openapi.yaml`](docs/openapi.yaml) 作成
- ✅ Patients、Social Profiles、Coverages、Observations APIの定義
- ✅ セキュリティ設定 (Firebase Authentication Bearer Token)
- ✅ スキーマ定義（Patient、SocialProfile、Coverage等）

### 5. ユニットテストの実装とドキュメント化
- ✅ `backend/internal/models/patient_test.go` (既存ファイル確認)
- ✅ `backend/tests/README.md` 作成
- ✅ テスト戦略、実行方法、モック生成、統合テスト、パフォーマンステストのガイド

## 次のステップ (優先度順)

### 優先度: CRITICAL (デプロイ前に必須)

#### 1. Go環境のインストール
現在、Goがインストールされていないため、まずこれを完了する必要があります。

```bash
# Homebrewのインストール
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Goのインストール
brew install go@1.22

# 確認
go version
```

詳細: [docs/SETUP.md](docs/SETUP.md) の「1. Homebrewのインストール」セクション参照

#### 2. Go依存関係の整理
```bash
cd backend
go mod download
go mod tidy
```

#### 3. Firebase設定ファイルの配置
1. [Firebase Console](https://console.firebase.google.com/) にアクセス
2. プロジェクト「stunning-grin-480914-n1」の秘密鍵をダウンロード
3. `backend/config/firebase-config.json` に配置

詳細: [backend/config/README.md](backend/config/README.md)

#### 4. 環境変数の設定
```bash
cd backend
cp .env.example .env
# .envファイルを編集して、API keysを設定
```

### 優先度: HIGH (MVP実装に必要)

#### 5. データベースのセットアップ

**ローカル開発環境 (PostgreSQL)**
```bash
# PostgreSQLのインストール
brew install postgresql@15
brew services start postgresql@15

# データベース作成
createdb visitas_dev

# マイグレーション適用
cd backend/migrations
for file in *.sql; do
  psql -U $(whoami) -d visitas_dev -f "$file"
done
```

**本番環境 (Cloud Spanner)**
```bash
# Spannerインスタンス作成
gcloud spanner instances create visitas-instance \
  --config=regional-asia-northeast1 \
  --description="Visitas Production Instance" \
  --nodes=1

# データベース作成
gcloud spanner databases create visitas-db \
  --instance=visitas-instance \
  --database-dialect=POSTGRESQL

# マイグレーション適用
cd backend/migrations
gcloud spanner databases ddl update visitas-db \
  --instance=visitas-instance \
  --ddl="$(cat 001_create_patients.sql)"
```

#### 6. Social Profiles & Coverages APIの実装

**実装する機能:**
- `POST /patients/{patientId}/social-profiles` - 社会的背景登録
- `GET /patients/{patientId}/social-profiles` - 社会的背景取得
- `POST /patients/{patientId}/coverages` - 保険情報登録
- `GET /patients/{patientId}/coverages` - 保険情報取得

**実装ファイル:**
```
backend/internal/
├── handlers/
│   ├── social_profile_handler.go
│   └── coverage_handler.go
├── services/
│   ├── social_profile_service.go
│   └── coverage_service.go
├── repository/
│   ├── social_profile_repository.go
│   └── coverage_repository.go
└── models/
    ├── social_profile.go
    └── coverage.go
```

#### 7. 基本的なCRUD APIの実装

**Patients API (Sprint 2の一部)**
- ✅ モデル定義済み (`backend/internal/models/patient.go`)
- ⏳ Handler実装
- ⏳ Service実装
- ⏳ Repository実装 (Spanner接続)

#### 8. 認証・認可ミドルウェアの実装

```go
// backend/internal/middleware/auth.go
func FirebaseAuthMiddleware() func(http.Handler) http.Handler {
    // Firebase ID Tokenの検証
    // Contextにユーザー情報を格納
}
```

### 優先度: MEDIUM (品質向上)

#### 9. ユニットテストの拡充
- ハンドラーのテスト
- サービスのテスト
- リポジトリのテスト (モック使用)
- カバレッジ80%以上を目指す

#### 10. 統合テストの実装
```bash
# Testcontainersでローカルテスト環境構築
cd backend/tests/integration
go test -v
```

#### 11. API仕様書の拡充
現在の `openapi.yaml` に以下を追加:
- Medical Conditions API
- Allergies API
- Care Plans API
- ACP API
- Medications API
- Visit Schedules API

#### 12. ロギング・エラーハンドリングの標準化
```go
// backend/pkg/logger/logger.go
// 構造化ログの実装 (Zerolog / Zap)

// backend/pkg/errors/errors.go
// カスタムエラー型の定義
```

### 優先度: LOW (最適化・拡張)

#### 13. Dockerコンテナ化
```bash
# Dockerfileのビルド確認
cd backend
docker build -t visitas-api .

# docker-compose.ymlの更新
docker-compose up -d
```

#### 14. CI/CDパイプラインの整備
`.github/workflows/` の実装:
- `test.yml` - テスト自動実行
- `lint.yml` - コード品質チェック
- `deploy-staging.yml` - Staging環境へのデプロイ
- `deploy-production.yml` - 本番環境へのデプロイ

#### 15. パフォーマンス最適化
- Spannerクエリの最適化
- Generated Columnsの活用
- キャッシング戦略 (Redis検討)

#### 16. セキュリティ監査
- [ ] CMEK暗号化の有効化確認
- [ ] Cloud Armorの設定
- [ ] Identity-Aware Proxy (IAP)の設定
- [ ] 監査ログの動作確認
- [ ] 脆弱性スキャン (Snyk / Trivy)

## 現在のプロジェクト状態

### ✅ 完了
- データベーススキーマ設計
- マイグレーションファイル作成
- 開発環境セットアップドキュメント
- OpenAPI仕様書 (基本版)
- テストフレームワーク準備

### 🚧 進行中
- Go環境のセットアップ (ユーザー側で実施必要)
- Firebase設定 (ユーザー側で実施必要)

### ⏳ 未着手
- API実装 (Handlers, Services, Repositories)
- 認証・認可ミドルウェア
- テストコードの拡充
- デプロイ準備

## 推奨される実装順序

### Week 1: 環境構築
1. ✅ Go環境インストール
2. ✅ Firebase設定
3. ✅ ローカルDB (PostgreSQL) セットアップ
4. ✅ 依存関係インストール (`go mod tidy`)
5. ✅ サーバー起動確認 (`go run cmd/api/main.go`)

### Week 2-3: MVP実装 (Sprint 2相当)
1. Patients CRUD API実装
2. Social Profiles API実装
3. Coverages API実装
4. 認証ミドルウェア実装
5. ユニットテスト実装 (カバレッジ60%以上)

### Week 4: テスト・品質向上
1. 統合テスト実装
2. OpenAPI仕様書完成
3. ロギング・エラーハンドリング標準化
4. コードレビュー・リファクタリング

### Week 5-6: デプロイ準備
1. Cloud Spannerへの移行
2. Firebase Authenticationの本格統合
3. CI/CDパイプライン構築
4. Staging環境デプロイ
5. セキュリティ監査

## よくある質問

### Q1: Goがインストールされていない場合は?
**A:** [docs/SETUP.md](docs/SETUP.md) の手順に従って、Homebrew経由でインストールしてください。

### Q2: Cloud Spannerの代わりにローカルで開発したい
**A:** PostgreSQL 15をインストールし、マイグレーションファイルを適用してください。Spanner PostgreSQLインターフェースと互換性があります。

### Q3: Firebase設定ファイルが手に入らない
**A:** プロジェクトオーナーまたはFirebaseコンソールへのアクセス権を持つ管理者に依頼してください。

### Q4: マイグレーションファイルの適用方法は?
**A:**
- PostgreSQL: `psql -U user -d db -f 001_create_patients.sql`
- Spanner: `gcloud spanner databases ddl update ...`

詳細は [backend/migrations/README.md](backend/migrations/README.md) 参照。

### Q5: テストが実行できない
**A:**
```bash
# 依存関係を確認
go mod download

# キャッシュクリア
go clean -testcache

# 再実行
go test ./... -v
```

## 連絡先・リソース

- **ドキュメント**: [`docs/`](docs/) ディレクトリ
- **API仕様**: [docs/openapi.yaml](docs/openapi.yaml)
- **データベース要件**: [docs/DATABASE_REQUIREMENTS.md](docs/DATABASE_REQUIREMENTS.md)
- **セットアップガイド**: [docs/SETUP.md](docs/SETUP.md)

## 進捗管理

タスクの進捗は、GitHubのIssuesまたはプロジェクトボードで管理することを推奨します。

```bash
# 例: GitHub IssueをCLIから作成
gh issue create --title "Patients APIの実装" \
  --body "Handlers, Services, Repositoriesの実装" \
  --label "enhancement" \
  --milestone "Sprint 2"
```

---

**最終更新**: 2025-12-12
**作成者**: Claude (Anthropic)
