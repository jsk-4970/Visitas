# Visitas - 在宅医療特化型AIプラットフォーム

## プロジェクト概要

Visitasは、日本の在宅医療（訪問診療）の課題を解決するためのAI駆動型クラウドプラットフォームです。医師の認知負荷を低減し、移動効率を最適化し、カルテ記載業務と法的書類作成を自動化することで、医師が患者と向き合う時間を最大化します。

## コアバリュー

1. **Ambient Clinical Intelligence**: Gemini 1.5 Proによる診療会話の自動構造化（SOAP形式）
2. **AI-Powered Documentation**: 生成AIによる法的書類（訪問看護指示書等）の自動下書き作成
3. **Dynamic Logistics**: Google Maps Route Optimization APIによる訪問ルートの最適化
4. **Secure Mobility**: オフラインファースト設計と3省2ガイドライン準拠のセキュリティ

## 技術スタック

### バックエンド
- **言語**: Go 1.22+ (Goroutinesによる高並行処理)
- **Webフレームワーク**: Chi / Gin
- **API基盤**: Cloud Run (サーバーレスコンテナ)
- **メインDB**: Cloud Spanner (強整合性、99.99%可用性)
- **リアルタイムDB**: Firestore (チャット、位置情報更新)
- **認証**: Firebase Authentication / Identity Platform
- **AI/ML**: Vertex AI (Gemini 1.5 Pro/Flash)
- **ストレージ**: Cloud Storage (音声、画像、バックアップ)
- **依存管理**: Go Modules
- **文書生成**: gofpdf / excelize (PDF/Excel生成)

### フロントエンド
- **モバイル**: Flutter (iOS/Android)
- **Web管理画面**: Flutter Web / React
- **ローカルDB**: Isar / sqflite (オフライン対応)

### インフラ・セキュリティ
- **プラットフォーム**: Google Cloud Platform (GCP)
- **リージョン**: asia-northeast1 (東京) / asia-northeast2 (大阪)
- **セキュリティ**: Cloud Armor, Identity-Aware Proxy, CMEK暗号化
- **コンプライアンス**: 3省2ガイドライン準拠（医療情報システム安全管理）

### 外部API連携
- **地図**: Google Maps Platform (Maps API, Routes API, Route Optimization API)
- **医療標準規格**: HL7 FHIR (Cloud Healthcare API経由)

## プロジェクト構造

```
Visitas/
├── backend/                 # Goバックエンドサービス
│   ├── cmd/
│   │   └── api/            # アプリケーションエントリーポイント
│   │       └── main.go
│   ├── internal/           # プライベートアプリケーションコード
│   │   ├── handlers/      # HTTPハンドラー
│   │   ├── services/      # ビジネスロジック
│   │   ├── models/        # データモデル
│   │   ├── repository/    # データアクセス層
│   │   │   ├── spanner/  # Spanner実装
│   │   │   └── firestore/ # Firestore実装
│   │   ├── middleware/    # 認証、ログ、CORS等
│   │   └── ai/           # Vertex AI / Gemini統合
│   ├── pkg/              # 外部パッケージで使用可能なコード
│   │   ├── auth/        # 認証ユーティリティ
│   │   ├── logger/      # ログ
│   │   └── validator/   # バリデーション
│   ├── migrations/       # Spannerマイグレーション
│   ├── scripts/         # ビルド、デプロイスクリプト
│   ├── tests/          # テストコード
│   ├── go.mod
│   ├── go.sum
│   ├── Dockerfile
│   ├── document_templates/  # 法的書類テンプレート
│   │   ├── nursing_instruction.json
│   │   ├── home_care_instruction.json
│   │   ├── special_nursing_instruction.json
│   │   ├── medication_instruction.json
│   │   └── point_instruction.json
│   └── .air.toml       # ホットリロード設定
├── mobile/              # Flutterモバイルアプリ
│   ├── lib/
│   │   ├── main.dart
│   │   ├── screens/    # 画面UI
│   │   ├── models/     # データモデル
│   │   ├── services/   # API通信、オフライン同期
│   │   ├── providers/  # 状態管理 (Riverpod)
│   │   └── utils/
│   ├── pubspec.yaml
│   └── android/
│       └── ios/
├── web/                # Web管理画面
├── infra/             # IaC (Terraform)
│   ├── terraform/
│   │   ├── environments/
│   │   │   ├── dev/
│   │   │   ├── staging/
│   │   │   └── prod/
│   │   └── modules/
│   └── cloudbuild.yaml
├── docs/              # ドキュメント
│   ├── REQUIREMENTS.md      # 要件定義書
│   ├── ARCHITECTURE.md      # アーキテクチャ設計
│   ├── API_SPEC.md         # API仕様書
│   ├── SECURITY.md         # セキュリティガイドライン
│   └── DEPLOYMENT.md       # デプロイ手順
├── .github/
│   └── workflows/          # CI/CD
├── claude.md              # このファイル
└── README.md
```

## 開発フェーズ

### Phase 1: MVP (1-3ヶ月) - **現在のフォーカス**

#### Sprint 1: 基本インフラとDB設計 (Week 1-2)
- GCPプロジェクトセットアップ
- Cloud Spannerインスタンス作成・スキーマ設計
- Go APIサーバー雛形（Chi/Gin + Cloud Run）
- Firebase Authentication統合
- CI/CDパイプライン基礎

#### Sprint 2: 患者管理機能 (Week 3-4)
- 患者CRUD API（Go）
- Spannerへの患者データ保存
- 基本的なバリデーション
- ユニットテスト

#### Sprint 3: スケジュール管理機能 (Week 5-6)
- 訪問スケジュールCRUD API
- 医師・患者の割り当て
- 日付・時間管理

#### Sprint 4: モバイルアプリ基本機能 (Week 7-9)
- Flutter雛形プロジェクト
- ログイン・認証画面
- 患者一覧表示
- スケジュール表示（タイムライン）
- オフラインDB（Isar）統合

#### Sprint 5: 地図連携とルート表示 (Week 10-12)
- Google Maps API統合
- 患者住所の地図表示
- Googleマップアプリへのディープリンク
- 基本的な移動距離計算

### Phase 2: AI Integration (4-6ヶ月)

#### Sprint 6: 法的書類AI生成（高優先度） (Week 13-16)
- **対象書類**: 訪問看護指示書、居宅療養指導書、特別訪問看護指示書、訪問薬剤管理指導書、点的指示書
- **書類テンプレート設計**: JSON形式で各書類の構造定義
- **データ統合API**:
  - Spannerから患者基本情報（氏名、生年月日、要介護度）を取得
  - 過去1ヶ月分の訪問ログ、バイタルデータ、構造化された雑談データを取得
- **Gemini統合**:
  - プロンプトエンジニアリング（「訪問看護指示書の『病状・治療状態』欄の下書きを作成」等）
  - グラウンディング機能（生成文の根拠となったログへのリンク表示）
  - 医療専門職向けの適切な文体制御
- **出力機能**:
  - PDF生成（gofpdf使用）
  - Excel生成（excelize使用）
  - テンプレートへの動的データ埋め込み
- **編集・承認UI**（Web管理画面）:
  - AI生成下書きの表示
  - インライン編集機能
  - 参照元データの監査ビュー
  - 最終承認ワークフロー

#### Sprint 7: SOAP自動生成 (Week 17-20)
- 音声録音・アップロード機能
- Gemini 1.5 Proによる診療会話の自動SOAP生成
- AIサマリー確認・修正UI
- プロンプトエンジニアリング最適化

### Phase 3: Optimization (7-9ヶ月)
- Route Optimization API実装
- FHIR API連携
- IoT/ウェアラブル連携
- データ分析基盤（BigQuery）

## 開発ガイドライン

### コーディング規約

#### Go
- **スタイル**: Effective Go準拠
- **フォーマット**: `gofmt`または`goimports`で自動整形
- **Lint**: `golangci-lint`使用
- **ネーミング**:
  - パッケージ名: 小文字、単数形（例: `patient`, `schedule`）
  - インターフェース: `～er`形式（例: `Repository`, `Handler`）
  - エクスポート: 大文字開始（例: `GetPatient()`）
- **エラーハンドリング**: エラーは常に返却し、呼び出し元で処理
- **コンテキスト**: すべてのI/O操作に`context.Context`を渡す

#### Dart/Flutter
- **スタイル**: Effective Dart準拠
- **Lint**: `flutter_lints`使用
- **状態管理**: Riverpod推奨
- **ネーミング**: キャメルケース

#### コミットメッセージ
- **形式**: Conventional Commits
- **例**:
  - `feat(api): add patient CRUD endpoints`
  - `fix(mobile): resolve offline sync conflict`
  - `docs: update API specification`

### セキュリティ原則

1. **データレジデンシー**: 全データは日本国内リージョン（asia-northeast1/2）に保存
2. **ゼロトラスト**: 全アクセスを認証・認可（Firebase Auth + IAP）
3. **暗号化**:
   - 転送時: TLS 1.3必須
   - 保存時: CMEK（Customer-Managed Encryption Keys）
4. **最小権限**: IAMロールは最小権限の原則で設定
5. **監査ログ**: 全アクセスをCloud Audit Logsで記録（5年保存）
6. **個人情報保護**: PHI（Protected Health Information）は全て暗号化、マスキング

### オフライン対応戦略

- **楽観的UI**: ユーザーアクションは即座にローカルに反映
- **バックグラウンド同期**: 接続回復時に自動同期（exponential backoff）
- **コンフリクト解決**:
  - 医療記録: LWW（Last Write Wins）または手動マージ
  - スケジュール: サーバー優先
- **データ圧縮**: 同期データはgzip圧縮

### テスト戦略

#### バックエンド（Go）
- **ユニットテスト**: `testing`パッケージ、カバレッジ80%以上
- **モック**: `gomock`または`testify/mock`
- **統合テスト**: Testcontainersでローカルエミュレータ使用
- **実行**: `go test ./... -v -cover`

#### モバイル（Flutter）
- **ユニットテスト**: `flutter_test`
- **ウィジェットテスト**: 主要画面をテスト
- **統合テスト**: `integration_test`でE2Eテスト
- **実行**: `flutter test`

#### 負荷テスト
- **ツール**: Locust / k6
- **目標**: 1000同時接続、レスポンスタイム<200ms

## 環境設定

### 開発環境セットアップ

```bash
# 1. GCP認証
gcloud auth login
gcloud config set project visitas-dev

# 2. Goバックエンド
cd backend
go mod download
go run cmd/api/main.go

# ホットリロード（開発時）
go install github.com/cosmtrek/air@latest
air

# 3. Flutterモバイル
cd mobile
flutter pub get
flutter run

# 4. Terraform
cd infra/terraform/environments/dev
terraform init
terraform plan
```

### 環境変数

```bash
# .env.local (ローカル開発用)
export GCP_PROJECT_ID=visitas-dev
export GCP_REGION=asia-northeast1
export SPANNER_INSTANCE=visitas-dev-instance
export SPANNER_DATABASE=visitas-dev-db
export FIREBASE_CONFIG_PATH=./firebase-config.json
export GEMINI_API_KEY=your_api_key_here
export GOOGLE_MAPS_API_KEY=your_maps_key_here
export PORT=8080
export LOG_LEVEL=debug
```

## よくある開発タスク

### 新しいAPIエンドポイントを追加

1. `internal/models/`にGoの構造体を定義
2. `internal/repository/`にデータアクセス層を実装
3. `internal/services/`にビジネスロジックを実装
4. `internal/handlers/`にHTTPハンドラーを実装
5. `cmd/api/main.go`でルーティングを追加
6. `tests/`にテストを追加
7. `docs/API_SPEC.md`を更新

### Spannerスキーマ変更

1. `migrations/`に`.sql`ファイルを作成
   ```sql
   -- 001_create_patients.sql
   CREATE TABLE Patients (
     patient_id STRING(36) NOT NULL,
     name STRING(100) NOT NULL,
     created_at TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
   ) PRIMARY KEY (patient_id);
   ```
2. 適用:
   ```bash
   gcloud spanner databases ddl update visitas-dev-db \
     --instance=visitas-dev-instance \
     --ddl="$(cat migrations/001_create_patients.sql)"
   ```
3. `internal/models/`のGoモデルを更新

### Flutterの新しい画面を追加

1. `lib/screens/`に画面ウィジェットを作成
2. `lib/models/`にデータモデルを定義
3. `lib/services/`にAPI通信ロジックを実装
4. `lib/providers/`に状態管理（Riverpod）を実装
5. ルーティング（GoRouter）を設定

### 法的書類テンプレートの追加

1. `backend/document_templates/`にJSON形式でテンプレート定義を作成
   ```json
   {
     "document_type": "nursing_instruction",
     "display_name": "訪問看護指示書",
     "sections": [
       {
         "field": "patient_info",
         "label": "患者情報",
         "source": "spanner.patients",
         "fields": ["name", "birth_date", "care_level"]
       },
       {
         "field": "medical_condition",
         "label": "病状・治療状態",
         "source": "gemini.summary",
         "prompt": "過去1ヶ月の訪問記録とバイタルデータに基づき、患者の病状と治療状態を200文字以内で要約してください。医療専門職向けの文体で、断定的表現は避けること。"
       },
       {
         "field": "nursing_plan",
         "label": "看護計画",
         "source": "gemini.summary",
         "prompt": "訪問看護で実施すべき内容を具体的に記載してください。"
       }
     ],
     "output_formats": ["pdf", "excel"]
   }
   ```
2. `internal/ai/prompts/`にプロンプトテンプレートを追加
3. `internal/services/document_generator.go`にドキュメント生成ロジックを実装
4. `pkg/pdfgen/`または`pkg/excelgen/`に出力ロジックを実装
5. `internal/handlers/documents.go`にHTTPハンドラーを実装
6. API仕様を`docs/API_SPEC.md`に追加

## デプロイ

### 本番環境へのデプロイ

```bash
# Cloud Buildトリガー経由（推奨）
git tag v1.0.0
git push origin v1.0.0

# 手動デプロイ（緊急時のみ）
cd backend
gcloud builds submit --tag gcr.io/visitas-prod/api
gcloud run deploy visitas-api \
  --image gcr.io/visitas-prod/api \
  --platform managed \
  --region asia-northeast1 \
  --allow-unauthenticated=false
```

## トラブルシューティング

### Spannerへの接続エラー
- IAMロールを確認: `roles/spanner.databaseUser`
- リージョン設定を確認: `asia-northeast1`
- Organization Policyでリージョン制限を確認
- 接続文字列の形式: `projects/PROJECT/instances/INSTANCE/databases/DB`

### Go実行時のエラー
```bash
# 依存関係の整理
go mod tidy

# キャッシュクリア
go clean -modcache

# ベンダーディレクトリ使用
go mod vendor
go build -mod=vendor
```

### オフライン同期の競合
- ローカルDBのバージョン番号を確認
- Firestoreの同期ステータスをログで確認
- 手動マージUIを使用

### Gemini APIのレート制限
- Vertex AIのQuotaを確認
- リトライロジック（Exponential Backoff）を確認
- Gemini 1.5 Flashへのダウングレードを検討

## 参考リソース

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Cloud Spanner Client](https://pkg.go.dev/cloud.google.com/go/spanner)
- [3省2ガイドライン](https://www.mhlw.go.jp/)
- [GCP Healthcare Solutions](https://cloud.google.com/solutions/healthcare-life-sciences)
- [Gemini API Documentation](https://cloud.google.com/vertex-ai/docs/generative-ai/model-reference/gemini)
- [Flutter Medical App Best Practices](https://flutter.dev)
- [HL7 FHIR Specification](https://www.hl7.org/fhir/)

## チーム連絡先

- **プロダクトオーナー**: [連絡先]
- **テックリード**: [連絡先]
- **医療監修**: [連絡先]

## ライセンス

[ライセンス情報]
