# Visitas - 在宅医療特化型AIプラットフォーム

Visitasは、日本の在宅医療（訪問診療）の課題を解決するためのAI駆動型クラウドプラットフォームです。

## 🎯 プロジェクト概要

### コアバリュー

1. **Ambient Clinical Intelligence**: Gemini 1.5 Proによる診療会話の自動構造化（SOAP形式）
2. **Dynamic Logistics**: Google Maps Route Optimization APIによる訪問ルートの最適化
3. **Secure Mobility**: オフラインファースト設計と3省2ガイドライン準拠のセキュリティ

### 技術スタック

- **バックエンド**: Go 1.22+, Chi Router, Cloud Run
- **データベース**: Cloud Spanner (メイン), Firestore (リアルタイム)
- **AI/ML**: Vertex AI (Gemini 1.5 Pro/Flash)
- **インフラ**: Google Cloud Platform (GCP)
- **モバイル**: Flutter (iOS/Android)
- **IaC**: Terraform

## 🚀 クイックスタート

### 必要な環境

- Go 1.22+
- Docker & Docker Compose
- gcloud CLI
- Terraform 1.5+

### 初回セットアップ

```bash
# リポジトリをクローン
git clone https://github.com/your-org/visitas.git
cd visitas

# 開発環境をセットアップ
make setup

# .envファイルを編集（APIキーなどを設定）
vim backend/.env
```

### ローカル開発環境の起動

#### オプション1: Docker Composeで起動（推奨）

```bash
# Spanner Emulator + APIサーバーを起動
make dev-docker

# APIサーバーのログを確認
make logs-api

# ヘルスチェック
make health
# または
curl http://localhost:8080/health
```

#### オプション2: ローカルで直接起動

```bash
# Spanner Emulatorのみ起動
docker-compose up -d spanner-emulator

# マイグレーション適用
make migrate-up

# APIサーバーを起動（Go）
make dev
```

### 停止

```bash
make docker-down
```

## 📁 プロジェクト構造

```
Visitas/
├── backend/                    # Goバックエンドサービス
│   ├── cmd/api/               # エントリーポイント
│   ├── internal/              # プライベートコード
│   │   ├── config/           # 設定管理
│   │   ├── handlers/         # HTTPハンドラー
│   │   ├── models/           # データモデル
│   │   ├── repository/       # データアクセス層
│   │   └── services/         # ビジネスロジック
│   ├── migrations/           # Spannerマイグレーション
│   ├── scripts/              # ビルド・デプロイスクリプト
│   ├── Dockerfile            # コンテナイメージ
│   └── go.mod                # Go dependencies
├── mobile/                    # Flutterモバイルアプリ（未実装）
├── web/                       # Web管理画面（未実装）
├── infra/                    # Infrastructure as Code
│   └── terraform/
│       └── environments/
│           ├── dev/
│           ├── staging/
│           └── prod/
├── docs/                     # ドキュメント
│   ├── REQUIREMENTS.md       # 要件定義書
│   └── ...
├── scripts/                  # 開発支援スクリプト
├── docker-compose.yml        # ローカル開発環境
├── Makefile                  # タスクランナー
├── claude.md                 # プロジェクトコンテキスト
└── README.md                 # このファイル
```

## 🛠️ 開発ワークフロー

### よく使うコマンド

```bash
# ヘルプを表示
make help

# テストを実行
make test

# コードをフォーマット
make fmt

# コードをリント
make lint

# Dockerイメージをビルド
make docker-build

# Spannerマイグレーションを適用
make migrate-up

# Terraformを初期化・適用
make terraform-init
make terraform-plan
make terraform-apply
```

### API エンドポイント

#### ヘルスチェック
```bash
GET /health
```

#### 患者管理
```bash
GET    /api/v1/patients       # 患者一覧
POST   /api/v1/patients       # 患者登録
GET    /api/v1/patients/:id   # 患者詳細
PUT    /api/v1/patients/:id   # 患者更新
DELETE /api/v1/patients/:id   # 患者削除（論理削除）
```

## 🗃️ データベース

### Spanner スキーマ

- **Patients**: 患者情報
- **Doctors**: 医師情報
- **Visits**: 訪問スケジュール
- **VisitRecords**: 訪問記録（診療内容）

### マイグレーション

```bash
# Spanner Emulatorでマイグレーション実行
export SPANNER_EMULATOR_HOST=localhost:9010
cd backend
bash scripts/create-spanner-emulator.sh
```

## 🌐 デプロイ

### GCPへのデプロイ

#### 1. Terraformでインフラ構築

```bash
cd infra/terraform/environments/dev

# 初回のみ
terraform init

# プランを確認
terraform plan

# 適用
terraform apply
```

#### 2. Cloud Runへデプロイ

```bash
# Dockerイメージをビルド & プッシュ
cd backend
gcloud builds submit --tag asia-northeast1-docker.pkg.dev/stunning-grin-480914-n1/visitas-dev/api:latest

# Cloud Runにデプロイ
gcloud run deploy visitas-api \
  --image asia-northeast1-docker.pkg.dev/stunning-grin-480914-n1/visitas-dev/api:latest \
  --platform managed \
  --region asia-northeast1 \
  --service-account visitas-dev-run@stunning-grin-480914-n1.iam.gserviceaccount.com
```

## 🔒 セキュリティ

### 3省2ガイドライン準拠

- **データレジデンシー**: 全データを日本国内リージョン（asia-northeast1/2）に保存
- **暗号化**: TLS 1.3（転送時）、CMEK（保存時）
- **アクセス制御**: IAP、Firebase Auth、RBAC
- **監査ログ**: Cloud Audit Logsで5年間保存

### 環境変数

機密情報は`.env`ファイルで管理（Gitには含めない）

```bash
# backend/.env
GCP_PROJECT_ID=stunning-grin-480914-n1
SPANNER_EMULATOR_HOST=localhost:9010
GOOGLE_MAPS_API_KEY=your_key_here
```

## 📊 モニタリング

- **ヘルスチェック**: `GET /health`
- **Cloud Monitoring**: GCPコンソールで確認
- **ログ**: Cloud Logging

## 🧪 テスト

```bash
# ユニットテスト
make test

# カバレッジレポート生成
make test-coverage
# => backend/coverage.html
```

## 📖 ドキュメント

- [claude.md](claude.md) - プロジェクト全体像とコンテキスト
- [docs/REQUIREMENTS.md](docs/REQUIREMENTS.md) - 要件定義書
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) - アーキテクチャ設計（TODO）
- [docs/API_SPEC.md](docs/API_SPEC.md) - API仕様書（TODO）

## 🤝 コントリビューション

### コーディング規約

- **Go**: Effective Go準拠、`gofmt`/`golangci-lint`
- **コミット**: Conventional Commits形式

### ブランチ戦略

- `main`: 本番環境
- `develop`: 開発環境
- `feature/*`: 機能開発
- `fix/*`: バグ修正

## 📝 ライセンス

[ライセンス情報]

## 👥 チーム

- **プロダクトオーナー**: [連絡先]
- **テックリード**: [連絡先]
- **医療監修**: [連絡先]

## 🔗 参考リソース

- [3省2ガイドライン](https://www.mhlw.go.jp/)
- [Google Cloud Healthcare Solutions](https://cloud.google.com/solutions/healthcare-life-sciences)
- [Cloud Spanner Documentation](https://cloud.google.com/spanner/docs)
- [Flutter Documentation](https://flutter.dev)

---

**Status**: 🚧 Phase 1 MVP開発中

**Last Updated**: 2025-12-11
