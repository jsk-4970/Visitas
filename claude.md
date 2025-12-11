# Visitas - 在宅医療特化型AIプラットフォーム

## プロジェクト概要

Visitasは、日本の在宅医療（訪問診療）の課題を解決するためのAI駆動型クラウドプラットフォームです。医師の認知負荷を低減し、移動効率を最適化し、カルテ記載業務と法的書類作成を自動化することで、医師が患者と向き合う時間を最大化します。

## コアバリュー

1. **Multi-Modal Clinical Documentation**: 3つのアプローチによるカルテ作成の効率化
   - 既存カルテからのテンプレート化・再利用
   - 紹介状・診療情報提供書からの自動構造化
   - 診療会話の音声認識とSOAP自動生成（Ambient Clinical Intelligence）
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
- **音声認識**: Cloud Speech-to-Text (医療用語特化モデル)
- **文書解析**: Gemini 1.5 Pro Document AI (紹介状・診療情報提供書の解析)
- **ストレージ**: Cloud Storage (音声、画像、PDF、バックアップ)
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
│   │       ├── prompts/  # プロンプトテンプレート
│   │       └── document_parser.go
│   ├── pkg/              # 外部パッケージで使用可能なコード
│   │   ├── auth/        # 認証ユーティリティ
│   │   ├── logger/      # ログ
│   │   ├── validator/   # バリデーション
│   │   ├── pdfgen/      # PDF生成ユーティリティ
│   │   └── excelgen/    # Excel生成ユーティリティ
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
│   ├── medical_record_templates/  # カルテテンプレート
│   │   ├── soap_template.json     # SOAP形式の標準テンプレート
│   │   ├── common_phrases.json    # よく使う定型文
│   │   └── specialty_templates/   # 診療科別テンプレート
│   │       ├── internal_medicine.json
│   │       ├── neurology.json
│   │       └── palliative_care.json
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

### Phase 1: MVP (Week 1-14 / 約3.5ヶ月) - **現在のフォーカス**

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

#### Sprint 6: 基本カルテ機能 (Week 13-14)
- **カルテデータモデル設計**: SOAP形式（Subjective, Objective, Assessment, Plan）をSpannerスキーマで定義
- **カルテCRUD API**: 手動入力によるカルテ作成・編集・削除
- **カルテテンプレート機能**: 既存カルテからの複製・再利用
  - 患者別の過去カルテ一覧表示
  - テンプレートとして保存（定型文、よく使う記載パターン）
  - 一括コピー・部分コピー機能
- **モバイルUI**: カルテ入力画面、履歴表示
- **実装理由**: AI機能の前に基本的なカルテ管理基盤を構築（最も簡単で即座に使える機能）

### Phase 2: AI Integration (Week 15-26 / 約3ヶ月)

#### Sprint 7: 法的書類AI生成（高優先度） (Week 15-18)
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

#### Sprint 8: 紹介状からのカルテ自動生成（中程度） (Week 19-22)
- **対象文書**: 他医療機関からの紹介状、診療情報提供書、退院サマリー
- **文書解析機能**:
  - PDFアップロード機能（Cloud Storage統合）
  - Gemini 1.5 ProのDocument AI機能による文書解析
  - 構造化されたテキストからの情報抽出（患者情報、既往歴、現病歴、処方薬等）
- **カルテ変換API**:
  - 抽出情報をSOAP形式に自動マッピング
  - 不足情報の検出とアラート表示
  - 手動補完が必要な項目のハイライト
- **レビューUI**:
  - 元文書と生成カルテの並列表示
  - 抽出精度の確認・修正
  - 承認ワークフロー
- **実装理由**: 医療文書は構造化されているため、音声よりも解析が容易。新規患者の初診カルテ作成を効率化

#### Sprint 9: 会話からのカルテ自動生成（複雑） (Week 23-26)
- **音声収録機能**:
  - モバイルアプリでの診療中音声録音
  - バックグラウンド録音対応
  - Cloud Storageへの安全なアップロード（暗号化）
- **Gemini 1.5 Pro統合**:
  - 音声認識（Speech-to-Text）
  - 会話のコンテキスト理解
  - 医療用語の正確な認識（日本語医療ドメイン特化）
- **SOAP自動生成**:
  - S（主観的情報）: 患者の訴え・症状の抽出
  - O（客観的情報）: バイタルサイン、検査所見の抽出
  - A（評価）: 病状の評価・診断の生成
  - P（計画）: 治療方針・処方の生成
- **AIサマリー確認・修正UI**:
  - 生成カルテのレビュー画面
  - 音声再生とテキストの同期表示
  - インライン編集機能
  - 根拠となった会話箇所へのリンク
- **プロンプトエンジニアリング最適化**:
  - Few-shot learningによる精度向上
  - 医師のフィードバックループ
  - ドメイン特化型ファインチューニング検討
- **実装理由**: 最も複雑だが、医師の記載業務を最大限削減。音声認識の精度とコンテキスト理解が課題

#### カルテ作成機能の実装優先順位まとめ
1. **Phase 1 Sprint 6**: 既存カルテからのテンプレート機能（簡単・即戦力）
2. **Phase 2 Sprint 8**: 紹介状からの自動生成（中程度・新規患者対応）
3. **Phase 2 Sprint 9**: 会話からの自動生成（複雑・最大の効率化）

### Phase 3: Optimization & Scale (Week 27+ / 3ヶ月以降)
- **Route Optimization API実装**: 複数患者への最適訪問順序の自動計算
- **FHIR API連携**: 他医療機関との標準規格データ交換
- **IoT/ウェアラブル連携**: 患者バイタルデータのリアルタイム収集
- **データ分析基盤（BigQuery）**: 訪問パターン分析、予測モデル構築
- **マルチテナント対応**: 複数医療機関での利用を想定した拡張

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
gcloud config set project stunning-grin-480914-n1

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
export GCP_PROJECT_ID=stunning-grin-480914-n1
export GCP_REGION=asia-northeast1
export SPANNER_INSTANCE=stunning-grin-480914-n1-instance
export SPANNER_DATABASE=stunning-grin-480914-n1-db
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
   gcloud spanner databases ddl update stunning-grin-480914-n1-db \
     --instance=stunning-grin-480914-n1-instance \
     --ddl="$(cat migrations/001_create_patients.sql)"
   ```
3. `internal/models/`のGoモデルを更新

### Flutterの新しい画面を追加

1. `lib/screens/`に画面ウィジェットを作成
2. `lib/models/`にデータモデルを定義
3. `lib/services/`にAPI通信ロジックを実装
4. `lib/providers/`に状態管理（Riverpod）を実装
5. ルーティング（GoRouter）を設定

### カルテ関連機能の追加

#### カルテテンプレートの追加
1. `backend/medical_record_templates/`にJSON形式でテンプレート定義を作成
   ```json
   {
     "template_id": "soap_standard_v1",
     "template_name": "SOAP標準テンプレート",
     "specialty": "general",
     "sections": {
       "subjective": {
         "label": "S (主観的情報)",
         "placeholder": "患者の訴え・症状を記載",
         "common_phrases": ["自覚症状なし", "疼痛あり", "食欲不振"]
       },
       "objective": {
         "label": "O (客観的情報)",
         "placeholder": "バイタルサイン・検査所見を記載",
         "fields": ["血圧", "脈拍", "体温", "SpO2"]
       },
       "assessment": {
         "label": "A (評価)",
         "placeholder": "病状の評価・診断を記載"
       },
       "plan": {
         "label": "P (計画)",
         "placeholder": "治療方針・処方を記載"
       }
     }
   }
   ```
2. `internal/services/medical_record_service.go`にテンプレート管理ロジックを実装
3. `internal/handlers/medical_records.go`にHTTPハンドラーを実装
4. API仕様を`docs/API_SPEC.md`に追加

#### 紹介状解析機能の追加
1. `internal/ai/document_parser.go`に文書解析ロジックを実装
   - Gemini 1.5 ProのDocument AI機能統合
   - PDFテキスト抽出
   - 情報の構造化（患者情報、既往歴、現病歴等）
2. `internal/services/referral_service.go`にカルテ変換ロジックを実装
3. `internal/handlers/referrals.go`にHTTPハンドラーを実装
4. Cloud Storageへのアップロード機能実装

#### 音声からのSOAP生成機能の追加
1. `internal/ai/speech_to_soap.go`に音声認識・SOAP生成ロジックを実装
   - Cloud Speech-to-Text統合
   - Gemini 1.5 Proによるコンテキスト理解
   - SOAP形式への構造化
2. `internal/services/voice_record_service.go`に音声データ管理を実装
3. モバイルアプリに録音機能を追加（`mobile/lib/services/audio_recorder.dart`）
4. プロンプトテンプレートを`internal/ai/prompts/soap_generation.txt`に作成

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

### 音声認識の精度問題
- **医療用語の誤認識**:
  - Cloud Speech-to-Textのカスタム語彙機能を使用
  - 専門用語リストを作成（病名、薬剤名、検査名等）
  - `speech_contexts`パラメータで医療用語をブースト
- **雑音・環境音の影響**:
  - ノイズキャンセリング機能付きマイクの使用を推奨
  - `enableAutomaticPunctuation`を有効化
  - 音声品質の事前チェック機能を実装
- **方言・訛りの認識**:
  - 地域方言モデルの活用を検討
  - ユーザーフィードバックによる継続的改善

### カルテ生成の品質問題
- **SOAP各項目の抽出漏れ**:
  - プロンプトに「必須項目チェックリスト」を含める
  - 不足項目を検出してアラート表示
  - 医師による手動補完をガイド
- **生成内容の不正確さ**:
  - グラウンディング機能で根拠データへのリンクを表示
  - 生成後の必須レビュープロセスを実装
  - 医師のフィードバックをプロンプト改善に活用
- **紹介状の解析エラー**:
  - PDF形式によってはテキスト抽出が困難な場合あり
  - OCR精度を確認（低い場合は手動入力をガイド）
  - 抽出データの信頼度スコアを表示

### Cloud Storageへのアップロード失敗
- **ネットワーク接続**:
  - オフライン時は自動リトライキューに追加
  - 再接続時にバックグラウンド同期
- **ファイルサイズ制限**:
  - 音声ファイルは最大100MB（約1時間の診療）
  - 必要に応じて圧縮（AAC形式推奨）
- **権限エラー**:
  - サービスアカウントに`storage.objects.create`権限を付与
  - Cloud Storageバケットのライフサイクルポリシーを確認

## 参考リソース

### 開発基盤
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Cloud Spanner Client](https://pkg.go.dev/cloud.google.com/go/spanner)
- [Flutter Medical App Best Practices](https://flutter.dev)

### 医療・コンプライアンス
- [3省2ガイドライン](https://www.mhlw.go.jp/)
- [GCP Healthcare Solutions](https://cloud.google.com/solutions/healthcare-life-sciences)
- [HL7 FHIR Specification](https://www.hl7.org/fhir/)

### AI/ML・文書解析
- [Gemini API Documentation](https://cloud.google.com/vertex-ai/docs/generative-ai/model-reference/gemini)
- [Cloud Speech-to-Text](https://cloud.google.com/speech-to-text/docs)
- [Gemini Document AI](https://cloud.google.com/vertex-ai/docs/generative-ai/multimodal/overview)
- [Prompt Engineering for Healthcare](https://cloud.google.com/vertex-ai/docs/generative-ai/learn/prompts/prompt-engineering)

## チーム連絡先

- **プロダクトオーナー**: [連絡先]
- **テックリード**: [連絡先]
- **医療監修**: [連絡先]

## ライセンス

[ライセンス情報]
