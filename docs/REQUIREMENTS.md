# Visitas - 本番要件定義書

## ドキュメント管理情報

- **作成日**: 2025-12-11
- **対象フェーズ**: Phase 1 MVP (1-3ヶ月)
- **最終更新**: 2025-12-11
- **承認者**: [承認者名]

---

## 目次

1. [プロジェクト概要](#1-プロジェクト概要)
2. [Phase 1 MVP の目的と範囲](#2-phase-1-mvp-の目的と範囲)
3. [機能要件](#3-機能要件)
4. [非機能要件](#4-非機能要件)
5. [技術要件](#5-技術要件)
6. [セキュリティ要件](#6-セキュリティ要件)
7. [実装スプリント計画](#7-実装スプリント計画)
8. [成功基準](#8-成功基準)
9. [制約事項とリスク](#9-制約事項とリスク)

---

## 1. プロジェクト概要

### 1.1 背景

日本の在宅医療現場では、以下の課題が顕在化している：

- 医師の移動時間が1日の業務時間の30-40%を占める
- 帰院後のカルテ記載業務（「パジャマタイム」）による過重労働
- 患者情報の分散と、緊急時の情報アクセス困難
- 非効率な訪問ルート計画による経営効率の低下

### 1.2 プロダクトビジョン

Visitasは、AIとクラウド技術により在宅医療の生産性を向上させ、医師が患者と向き合う時間を最大化する。

### 1.3 Phase 1 MVP の位置づけ

Phase 1では、**AI機能を除外した基本的な業務支援システム**を構築する。これにより：
- 技術基盤の確立とリスクの低減
- 実ユーザーからのフィードバック収集
- Phase 2でのAI統合のための土台構築

---

## 2. Phase 1 MVP の目的と範囲

### 2.1 達成目標

- 患者情報のデジタル管理（紙カルテからの脱却）
- 訪問スケジュールの可視化と管理
- モバイルからの患者情報アクセス（オフライン対応）
- Google Maps連携による移動支援

### 2.2 スコープ

#### 含まれる機能（In Scope）
- 患者管理（CRUD）
- 訪問スケジュール管理
- モバイルアプリ（医師向け）
- Web管理画面（事務向け）
- Google Maps表示・ナビゲーション
- オフライン基本対応

#### 含まれない機能（Out of Scope）
- AI音声認識・自動カルテ生成 → Phase 2
- Route Optimization API → Phase 3
- FHIR連携 → Phase 3
- IoT/ウェアラブル連携 → Phase 3
- データ分析・BI → Phase 3

---

## 3. 機能要件

### 3.1 ユーザー管理・認証

#### FR-AUTH-001: ログイン機能
- **優先度**: 必須
- **説明**: 医師・看護師・事務スタッフが、メールアドレスとパスワードでログインできる
- **技術**: Firebase Authentication
- **受け入れ基準**:
  - [ ] メール・パスワードでログイン成功
  - [ ] ログイン失敗時にエラーメッセージ表示
  - [ ] ログイン状態の永続化（セッション管理）

#### FR-AUTH-002: 多要素認証（MFA）
- **優先度**: 推奨（Phase 1後半）
- **説明**: SMSまたは認証アプリによる2段階認証
- **受け入れ基準**:
  - [ ] MFA有効化機能
  - [ ] ログイン時のワンタイムコード入力

#### FR-AUTH-003: 生体認証（モバイル）
- **優先度**: 必須
- **説明**: Face ID / Touch IDによるアプリロック解除
- **技術**: Flutter `local_auth`
- **受け入れ基準**:
  - [ ] アプリ起動時に生体認証要求
  - [ ] バックグラウンドから復帰時も認証
  - [ ] 生体認証失敗時はパスワード入力へフォールバック

---

### 3.2 患者管理

#### FR-PATIENT-001: 患者登録
- **優先度**: 必須
- **説明**: 新規患者の基本情報を登録できる
- **入力項目**:
  - 患者ID（自動採番）
  - 氏名（姓・名）
  - 生年月日
  - 性別
  - 住所（郵便番号、都道府県、市区町村、番地、建物名）
  - 電話番号
  - 緊急連絡先
  - 保険情報（保険者番号、被保険者番号、負担割合）
  - 主病名
  - アレルギー情報
  - 特記事項（任意）
- **受け入れ基準**:
  - [ ] 全必須項目入力でDB保存成功
  - [ ] 郵便番号から住所自動入力（外部API使用）
  - [ ] バリデーションエラー時の適切なエラー表示
  - [ ] 重複チェック（同姓同名・生年月日）

#### FR-PATIENT-002: 患者一覧表示
- **優先度**: 必須
- **説明**: 登録された患者の一覧を表示し、検索・フィルタリングできる
- **表示項目**: 患者ID、氏名、生年月日、住所（市区町村まで）、最終訪問日
- **機能**:
  - 氏名による部分一致検索
  - 地域（市区町村）によるフィルタ
  - 最終訪問日によるソート
  - ページネーション（50件/ページ）
- **受け入れ基準**:
  - [ ] 一覧表示が1秒以内に完了
  - [ ] 検索結果がリアルタイムで更新
  - [ ] モバイル・Webの両方で動作

#### FR-PATIENT-003: 患者詳細表示
- **優先度**: 必須
- **説明**: 患者の詳細情報を表示する
- **表示内容**:
  - 基本情報（FR-PATIENT-001の全項目）
  - 訪問履歴（最新10件）
  - 処方履歴（最新10件）
  - バイタルサイン推移（グラフ）
  - 地図上の住所表示（Google Maps）
- **受け入れ基準**:
  - [ ] 全情報が正しく表示される
  - [ ] 地図が正確な位置を示す
  - [ ] タップで電話発信（モバイル）

#### FR-PATIENT-004: 患者情報編集
- **優先度**: 必須
- **説明**: 既存患者の情報を更新できる
- **制約**: 患者IDは変更不可
- **受け入れ基準**:
  - [ ] 編集内容がDB保存される
  - [ ] 変更履歴がログに記録される（誰が・いつ・何を変更）

#### FR-PATIENT-005: 患者削除（論理削除）
- **優先度**: 必須
- **説明**: 患者データを削除（アーカイブ）する
- **制約**: 物理削除は行わず、削除フラグを立てる
- **受け入れ基準**:
  - [ ] 削除確認ダイアログ表示
  - [ ] 削除後は一覧に表示されない
  - [ ] 管理者は削除済みデータを閲覧可能

---

### 3.3 訪問スケジュール管理

#### FR-SCHEDULE-001: スケジュール登録
- **優先度**: 必須
- **説明**: 患者への訪問予定を登録できる
- **入力項目**:
  - 患者ID（ドロップダウン選択）
  - 訪問日時
  - 担当医師（ドロップダウン選択）
  - 訪問目的（定期訪問 / 緊急往診 / 初診）
  - 予定所要時間（15分 / 30分 / 60分）
  - メモ（任意）
- **受け入れ基準**:
  - [ ] カレンダーUIから日時選択
  - [ ] 既存スケジュールとの重複チェック
  - [ ] 医師の勤務時間外の場合は警告表示

#### FR-SCHEDULE-002: デイリー・タイムライン表示（モバイル）
- **優先度**: 必須
- **説明**: 医師が当日の訪問予定を時系列で確認できる
- **表示内容**:
  - 訪問予定時刻
  - 患者名
  - 住所
  - 移動所要時間（前の訪問先からの推定時間）
  - 重要タグ（「独居」「認知症」など）
- **受け入れ基準**:
  - [ ] 時刻順にソートされたリスト表示
  - [ ] カードタップで患者詳細へ遷移
  - [ ] 「ナビ開始」ボタンでGoogleマップアプリ起動

#### FR-SCHEDULE-003: スケジュール一覧（Web管理画面）
- **優先度**: 必須
- **説明**: 事務スタッフが全医師のスケジュールを管理できる
- **表示形式**: ガントチャート / カレンダービュー切り替え
- **機能**:
  - 医師別フィルタ
  - 週・月・日表示切り替え
  - ドラッグ＆ドロップでスケジュール変更
  - 未割当患者リストの表示
- **受け入れ基準**:
  - [ ] 複数医師のスケジュールを同時表示
  - [ ] スケジュール変更が即座に反映
  - [ ] 印刷機能（PDF出力）

#### FR-SCHEDULE-004: スケジュール編集・削除
- **優先度**: 必須
- **説明**: 既存スケジュールの変更・キャンセルができる
- **制約**: 訪問完了後のスケジュールは削除不可
- **受け入れ基準**:
  - [ ] 訪問日時・担当医師の変更が可能
  - [ ] キャンセル理由の入力（キャンセル時）
  - [ ] 変更通知（担当医師へのプッシュ通知）

---

### 3.4 診療記録（簡易版）

#### FR-RECORD-001: 訪問記録の手動入力
- **優先度**: 必須
- **説明**: 医師が訪問後に診療内容を記録できる
- **入力項目**:
  - 訪問日時（自動入力）
  - バイタルサイン（血圧、脈拍、体温、SpO2）
  - 主訴（自由記述）
  - 所見（自由記述）
  - 処方内容（薬品名、用量、日数）
  - 次回訪問予定日
- **受け入れ基準**:
  - [ ] モバイルアプリから入力可能
  - [ ] 下書き保存機能（オフライン対応）
  - [ ] 確定後はDB保存

#### FR-RECORD-002: 写真撮影・添付
- **優先度**: 推奨
- **説明**: 患部や褥瘡の写真を撮影し、記録に添付できる
- **制約**: 写真は暗号化してCloud Storageに保存
- **受け入れ基準**:
  - [ ] カメラ起動・撮影
  - [ ] 複数枚添付可能（最大10枚）
  - [ ] サムネイル表示

#### FR-RECORD-003: 訪問記録の閲覧
- **優先度**: 必須
- **説明**: 過去の訪問記録を時系列で閲覧できる
- **表示形式**: タイムライン形式
- **受け入れ基準**:
  - [ ] 最新から過去へスクロール
  - [ ] 日付・担当医師でフィルタ
  - [ ] 記録をタップして詳細表示

---

### 3.5 地図・ナビゲーション連携

#### FR-MAP-001: 患者住所の地図表示
- **優先度**: 必須
- **説明**: 患者詳細画面で住所を地図上に表示する
- **技術**: Google Maps JavaScript API（Web）/ Google Maps SDK（Mobile）
- **受け入れ基準**:
  - [ ] 正確な緯度経度への変換（Geocoding API）
  - [ ] マーカー表示
  - [ ] ズーム・パン操作可能

#### FR-MAP-002: Googleマップアプリ連携
- **優先度**: 必須
- **説明**: 「ナビ開始」ボタンでGoogleマップアプリを起動し、ナビゲーションを開始する
- **プラットフォーム**: iOS / Android
- **受け入れ基準**:
  - [ ] ディープリンクでGoogleマップ起動
  - [ ] 目的地が正しく設定される
  - [ ] Googleマップ未インストール時はブラウザ版を開く

#### FR-MAP-003: 訪問ルート表示
- **優先度**: 推奨
- **説明**: 当日の訪問予定地点を地図上に表示し、ルートを可視化する
- **表示内容**:
  - クリニック（出発地点）
  - 訪問予定地点（1, 2, 3...の順番付き）
  - 推定ルート（Directions API）
- **受け入れ基準**:
  - [ ] 全訪問地点がマーカー表示
  - [ ] ルートが線で結ばれる
  - [ ] 総移動距離・所要時間の表示

---

### 3.6 オフライン対応

#### FR-OFFLINE-001: オフラインデータ保存
- **優先度**: 必須
- **説明**: モバイルアプリがオフライン時もデータ閲覧・編集できる
- **技術**: Isar（Flutter ローカルDB）
- **対象データ**:
  - 当日の訪問スケジュール
  - 訪問予定患者の詳細情報
  - 過去7日間の訪問記録
- **受け入れ基準**:
  - [ ] オンライン時に自動プリフェッチ
  - [ ] オフライン時もデータ閲覧可能
  - [ ] オフライン編集はローカルに保存

#### FR-OFFLINE-002: 自動同期
- **優先度**: 必須
- **説明**: オンライン復帰時にローカル変更をサーバーに同期する
- **同期方式**: 楽観的UI + バックグラウンド同期
- **受け入れ基準**:
  - [ ] 接続復帰を検知
  - [ ] ローカル変更をサーバーに送信
  - [ ] 同期成功/失敗の通知
  - [ ] コンフリクト検出（タイムスタンプ比較）

#### FR-OFFLINE-003: コンフリクト解決
- **優先度**: 必須
- **説明**: サーバーとローカルのデータが競合した場合、ユーザーに選択させる
- **解決方法**:
  - 医療記録: LWW（Last Write Wins）またはマージUI
  - スケジュール: サーバー優先
- **受け入れ基準**:
  - [ ] コンフリクト検出時にダイアログ表示
  - [ ] ユーザーが「サーバー版を採用」「ローカル版を採用」を選択
  - [ ] 選択結果を反映して再同期

---

## 4. 非機能要件

### 4.1 パフォーマンス

| 要件ID | 項目 | 目標値 | 測定方法 |
|--------|------|--------|----------|
| NFR-PERF-001 | API応答時間 | 平均200ms以下、P95で500ms以下 | Cloud Monitoringで監視 |
| NFR-PERF-002 | ページ読み込み時間 | 初回3秒以内、2回目以降1秒以内 | Lighthouse |
| NFR-PERF-003 | モバイルアプリ起動時間 | 2秒以内（冷起動） | Firebase Performance Monitoring |
| NFR-PERF-004 | データ同期時間 | 1日分のデータを10秒以内 | 実機テスト |

### 4.2 可用性

| 要件ID | 項目 | 目標値 | 対策 |
|--------|------|--------|------|
| NFR-AVAIL-001 | システム稼働率 | 99.9%（月間43分以内のダウンタイム） | Cloud Runの自動スケーリング、ヘルスチェック |
| NFR-AVAIL-002 | データベース可用性 | 99.99% | Cloud Spanner Regional構成 |
| NFR-AVAIL-003 | 計画停止 | 月1回以内、深夜帯のみ | メンテナンスウィンドウを事前通知 |

### 4.3 拡張性

| 要件ID | 項目 | 目標値 |
|--------|------|--------|
| NFR-SCALE-001 | 同時接続ユーザー数 | 1000ユーザー |
| NFR-SCALE-002 | 患者データ | 100,000レコード |
| NFR-SCALE-003 | 訪問記録 | 1,000,000レコード（年間） |

### 4.4 互換性

| 要件ID | 項目 | 対応バージョン |
|--------|------|----------------|
| NFR-COMPAT-001 | iOS | iOS 15以上 |
| NFR-COMPAT-002 | Android | Android 10以上 |
| NFR-COMPAT-003 | Webブラウザ | Chrome, Safari, Edge（最新版） |

---

## 5. 技術要件

### 5.1 開発環境

| カテゴリ | 技術 | バージョン |
|----------|------|------------|
| バックエンド言語 | Go | 1.22+ |
| フロントエンド（Mobile） | Flutter | 3.19+ |
| フロントエンド（Web） | Flutter Web / React | - |
| データベース | Cloud Spanner | Regional |
| 認証 | Firebase Authentication | - |
| ホスティング | Cloud Run | - |
| CI/CD | Cloud Build / GitHub Actions | - |

### 5.2 インフラ構成

#### 5.2.1 GCPリソース

| リソース | 用途 | 構成 |
|----------|------|------|
| Cloud Run | APIサーバー | asia-northeast1、min-instances=1 |
| Cloud Spanner | メインDB | Regional（asia-northeast1）、1ノード |
| Firestore | リアルタイムDB | asia-northeast1 |
| Cloud Storage | 画像・音声 | asia-northeast1、Standard |
| Firebase Auth | 認証 | - |
| Cloud Armor | WAF | DDoS保護有効 |
| Identity-Aware Proxy | アクセス制御 | 管理画面用 |

#### 5.2.2 データモデル（Spanner）

```sql
-- 患者テーブル
CREATE TABLE Patients (
  patient_id STRING(36) NOT NULL,
  name_last STRING(50) NOT NULL,
  name_first STRING(50) NOT NULL,
  birth_date DATE NOT NULL,
  gender STRING(10),
  postal_code STRING(8),
  address_prefecture STRING(10),
  address_city STRING(50),
  address_street STRING(100),
  address_building STRING(100),
  phone STRING(15),
  emergency_contact STRING(15),
  insurance_number STRING(20),
  insurance_symbol STRING(20),
  copay_rate INT64,
  primary_diagnosis STRING(200),
  allergies STRING(500),
  notes STRING(2000),
  deleted BOOL NOT NULL DEFAULT (false),
  created_at TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
  updated_at TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
) PRIMARY KEY (patient_id);

-- 訪問スケジュールテーブル
CREATE TABLE Visits (
  visit_id STRING(36) NOT NULL,
  patient_id STRING(36) NOT NULL,
  doctor_id STRING(36) NOT NULL,
  scheduled_at TIMESTAMP NOT NULL,
  duration_minutes INT64 NOT NULL,
  visit_type STRING(20) NOT NULL, -- 'regular', 'emergency', 'initial'
  status STRING(20) NOT NULL, -- 'scheduled', 'completed', 'canceled'
  notes STRING(1000),
  created_at TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
  updated_at TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
) PRIMARY KEY (visit_id),
  INTERLEAVE IN PARENT Patients ON DELETE CASCADE;

-- 訪問記録テーブル
CREATE TABLE VisitRecords (
  record_id STRING(36) NOT NULL,
  visit_id STRING(36) NOT NULL,
  blood_pressure_systolic INT64,
  blood_pressure_diastolic INT64,
  pulse INT64,
  temperature FLOAT64,
  spo2 INT64,
  chief_complaint STRING(1000),
  findings STRING(2000),
  prescription STRING(1000),
  next_visit_date DATE,
  created_by STRING(36) NOT NULL,
  created_at TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
) PRIMARY KEY (record_id);

-- 医師テーブル
CREATE TABLE Doctors (
  doctor_id STRING(36) NOT NULL,
  name STRING(100) NOT NULL,
  email STRING(100) NOT NULL,
  phone STRING(15),
  specialization STRING(50),
  license_number STRING(20),
  created_at TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
) PRIMARY KEY (doctor_id);
```

---

## 6. セキュリティ要件

### 6.1 3省2ガイドライン準拠

| ガイドライン | 対応項目 | 実装内容 |
|--------------|----------|----------|
| 医療情報システムの安全管理に関するガイドライン（厚労省） | データレジデンシー | 全データを日本国内リージョンに保存 |
| 〃 | アクセスログ | Cloud Audit Logsで5年間保存 |
| 〃 | 暗号化 | TLS 1.3（転送時）、CMEK（保存時） |
| 医療情報を取り扱う情報システム・サービスの提供事業者における安全管理ガイドライン（経産省） | 責任分界点 | SLAで明記 |
| 〃 | バックアップ | 日次バックアップ、30日間保持 |

### 6.2 認証・認可

| 要件ID | 項目 | 実装内容 |
|--------|------|----------|
| SEC-AUTH-001 | パスワード強度 | 8文字以上、英数字記号混在 |
| SEC-AUTH-002 | セッションタイムアウト | 30分間操作がない場合は再認証 |
| SEC-AUTH-003 | ロールベースアクセス制御（RBAC） | 医師・看護師・事務の3ロール |
| SEC-AUTH-004 | 生体認証（モバイル） | Face ID / Touch ID必須 |

### 6.3 データ保護

| 要件ID | 項目 | 実装内容 |
|--------|------|----------|
| SEC-DATA-001 | PHI暗号化 | AES-256で暗号化 |
| SEC-DATA-002 | 個人情報マスキング | ログには患者名・ID等を出力しない |
| SEC-DATA-003 | バックアップ暗号化 | Cloud Storageへの暗号化バックアップ |
| SEC-DATA-004 | 削除ポリシー | 論理削除、物理削除は管理者のみ |

---

## 7. 実装スプリント計画

### Sprint 1: 基本インフラとDB設計（Week 1-2）

#### 目標
GCPプロジェクトのセットアップとバックエンド基盤の構築

#### タスク
- [ ] GCPプロジェクト作成（dev/staging/prod）
- [ ] Organization Policyでリージョン制限設定
- [ ] Cloud Spanner インスタンス作成（dev: 100 processing units）
- [ ] Spannerスキーマ作成（Patients, Visits, VisitRecords, Doctors）
- [ ] Firebase プロジェクト作成・Authentication有効化
- [ ] Go API サーバー雛形作成
  - Chi ルーター設定
  - Spannerクライアント初期化
  - Firebase Admin SDK統合
  - ヘルスチェックエンドポイント `/health`
- [ ] Dockerfile作成
- [ ] Cloud Runへの初回デプロイ
- [ ] Cloud Buildトリガー設定（main ブランチへのpush時）

#### 成果物
- 動作するAPIサーバー（https://api-dev.visitas.example.com/health）
- Spannerデータベース
- CI/CDパイプライン

---

### Sprint 2: 患者管理API（Week 3-4）

#### 目標
患者CRUDのバックエンドAPI実装

#### タスク
- [ ] `internal/models/patient.go` 作成
- [ ] `internal/repository/patient_repository.go` 作成
  - CreatePatient
  - GetPatientByID
  - ListPatients（ページネーション）
  - UpdatePatient
  - DeletePatient（論理削除）
- [ ] `internal/handlers/patient_handler.go` 作成
  - POST /api/v1/patients
  - GET /api/v1/patients/:id
  - GET /api/v1/patients（検索・フィルタ対応）
  - PUT /api/v1/patients/:id
  - DELETE /api/v1/patients/:id
- [ ] バリデーション実装（go-playground/validator）
- [ ] ユニットテスト作成（80%カバレッジ）
- [ ] 統合テスト（Testcontainers + Spanner Emulator）

#### 成果物
- 患者管理API
- API仕様書（OpenAPI 3.0形式）

---

### Sprint 3: スケジュール管理API（Week 5-6）

#### 目標
訪問スケジュールCRUDのバックエンドAPI実装

#### タスク
- [ ] `internal/models/visit.go` 作成
- [ ] `internal/repository/visit_repository.go` 作成
  - CreateVisit
  - GetVisitByID
  - ListVisitsByDate（日付範囲指定）
  - ListVisitsByDoctor（医師IDで絞り込み）
  - UpdateVisit
  - DeleteVisit
- [ ] `internal/handlers/visit_handler.go` 作成
- [ ] スケジュール重複チェックロジック
- [ ] ユニットテスト・統合テスト
- [ ] 医師マスタAPI（CRUD）

#### 成果物
- スケジュール管理API
- 医師管理API

---

### Sprint 4: モバイルアプリ基本機能（Week 7-9）

#### 目標
Flutterモバイルアプリの基本UI実装

#### タスク
- [ ] Flutter プロジェクト作成
- [ ] ディレクトリ構造構築
- [ ] Firebase Authenticationプラグイン統合
- [ ] ログイン画面実装
  - メール・パスワード入力
  - ログインボタン
  - エラーハンドリング
- [ ] 生体認証実装（local_auth）
- [ ] APIクライアント実装（Dio + Retrofit）
  - 認証トークン自動付与（Interceptor）
  - リトライロジック
- [ ] 患者一覧画面
  - リスト表示
  - Pull-to-refresh
  - 検索バー
- [ ] 患者詳細画面
  - 基本情報表示
  - タブ（概要・訪問履歴・バイタル）
- [ ] デイリー・タイムライン画面
  - 当日スケジュール表示（カード形式）
  - スケジュールカードタップで患者詳細へ
- [ ] 状態管理（Riverpod）
- [ ] ローカルDB（Isar）基本設定

#### 成果物
- iOS/Androidアプリ（TestFlight/内部テスト配布）

---

### Sprint 5: 地図連携とオフライン対応（Week 10-12）

#### 目標
Google Maps統合とオフラインファーストの実装

#### タスク
- [ ] Google Maps API有効化・APIキー取得
- [ ] Geocoding API統合（住所→緯度経度変換）
- [ ] 患者詳細画面に地図表示
- [ ] 「ナビ開始」ボタン実装
  - `url_launcher`でGoogleマップ起動
  - iOS: `comgooglemaps://` URLスキーム
  - Android: `google.navigation:` Intent
- [ ] 訪問ルート表示画面
  - Directions API呼び出し
  - ポリライン描画
  - 総移動距離・時間表示
- [ ] オフライン対応
  - Isarスキーマ定義
  - 同期ロジック（SyncService）
  - 接続状態監視（connectivity_plus）
  - バックグラウンド同期（Workmanager）
- [ ] オフライン時のUI表示（バナー）
- [ ] コンフリクト解決UI

#### 成果物
- 地図連携機能
- オフライン対応完了

---

### Sprint 6: Web管理画面（Week 13-15）

#### 目標
事務スタッフ向けWeb管理画面の実装

#### タスク
- [ ] Flutter Webプロジェクトセットアップ
- [ ] ログイン画面
- [ ] ダッシュボード
  - 今日の訪問件数
  - 医師別稼働状況
- [ ] 患者管理画面
  - 一覧・検索
  - 新規登録フォーム
  - 編集・削除
- [ ] スケジュール管理画面
  - カレンダービュー（fullcalendar.js的なUI）
  - ドラッグ＆ドロップでスケジュール変更
  - 未割当患者リスト
- [ ] 地図表示（Google Maps JavaScript API）
- [ ] Firebase Hostingへのデプロイ

#### 成果物
- Web管理画面（https://admin-dev.visitas.example.com）

---

### Sprint 7: 訪問記録機能（Week 16-18）

#### 目標
診療記録の手動入力機能

#### タスク
- [ ] VisitRecordsテーブル作成（Spanner）
- [ ] 訪問記録CRUD API
- [ ] モバイル：診療記録入力画面
  - バイタルサイン入力（数値キーボード）
  - 主訴・所見（テキストエリア）
  - 処方入力
  - 下書き保存機能
- [ ] 写真撮影・添付機能
  - `image_picker`で撮影
  - Cloud Storageへアップロード
  - Signed URL生成
- [ ] 訪問記録タイムライン表示
- [ ] Web：訪問記録閲覧画面

#### 成果物
- 訪問記録機能

---

### Sprint 8: 統合テスト・本番準備（Week 19-21）

#### 目標
テスト・セキュリティ強化・本番環境構築

#### タスク
- [ ] E2Eテスト（Patrol / Maestro）
  - ログイン→スケジュール確認→訪問記録入力のフロー
- [ ] 負荷テスト（k6）
  - 100同時接続でのAPI応答時間測定
- [ ] セキュリティ監査
  - Cloud Armorルール設定
  - IAPの有効化（管理画面）
  - CMEKでの暗号化設定
- [ ] 本番環境構築
  - Terraform for Production
  - Cloud Spannerノード数調整
  - Cloud Runオートスケール設定
- [ ] 監視・アラート設定
  - Cloud Monitoring ダッシュボード
  - Uptimeチェック
  - エラー率アラート（>5%でSlack通知）
- [ ] ドキュメント整備
  - 運用マニュアル
  - 障害対応手順書

#### 成果物
- 本番環境
- 運用ドキュメント

---

## 8. 成功基準

### 8.1 MVP成功の定義

以下の基準を**全て**満たすことでPhase 1 MVPの成功とする：

#### 機能面
- [ ] 患者50名以上を登録し、管理できる
- [ ] 月間200件以上の訪問スケジュールを管理できる
- [ ] モバイルアプリから訪問記録を入力し、保存できる
- [ ] オフライン→オンライン復帰時に自動同期が動作する
- [ ] Googleマップと連携し、ナビゲーションが開始できる

#### 品質面
- [ ] API応答時間P95が500ms以内
- [ ] システム稼働率99.9%以上
- [ ] ユニットテストカバレッジ80%以上
- [ ] クリティカルなバグ0件、メジャーバグ5件以内

#### ユーザー満足度
- [ ] パイロット医療機関（1-2施設）での実運用開始
- [ ] 医師からのNPS（Net Promoter Score）が+20以上
- [ ] 「カルテ記載時間が30%短縮された」との定量評価

### 8.2 Phase 2への移行判断基準

以下の条件を満たした場合、Phase 2（AI統合）へ移行する：

- [ ] Phase 1の全機能が安定稼働（1ヶ月間の大きな障害なし）
- [ ] パイロットユーザーから「AI音声入力を試したい」との要望が3件以上
- [ ] 予算・リソースの確保

---

## 9. 制約事項とリスク

### 9.1 制約事項

| 制約 | 内容 | 対応策 |
|------|------|--------|
| 予算 | Phase 1予算は月額30万円以内（GCP利用料） | Cloud Spanner 100 PU、Cloud Run min-instances=1に抑える |
| 開発期間 | 3ヶ月以内 | MVPスコープを厳格に管理、AI機能は後回し |
| 医療従事者の多忙さ | フィードバック収集が困難 | 週次30分のオンラインMTGを必須化 |

### 9.2 リスクと対策

| リスクID | リスク内容 | 発生確率 | 影響度 | 対策 |
|----------|------------|----------|--------|------|
| RISK-001 | Spanner利用料が予算超過 | 中 | 高 | 利用状況を週次モニタリング、閾値アラート設定 |
| RISK-002 | オフライン同期のバグ | 高 | 高 | 早期に統合テスト実施、パイロット運用で検証 |
| RISK-003 | Googleマップ連携の不具合（住所の誤認識） | 中 | 中 | Geocodingの精度検証、手動補正機能を用意 |
| RISK-004 | 医療従事者のITリテラシー不足 | 中 | 中 | 直感的UIの追求、オンボーディング動画作成 |
| RISK-005 | 法規制の変更 | 低 | 高 | 厚労省ガイドライン改訂を継続監視 |

---

## 10. 付録

### 10.1 用語集

| 用語 | 説明 |
|------|------|
| SOAP | Subjective（主訴）, Objective（所見）, Assessment（評価）, Plan（計画）の医療記録形式 |
| PHI | Protected Health Information（保護すべき医療情報） |
| 3省2ガイドライン | 厚労省・経産省・総務省による医療情報システムに関する2つのガイドライン |
| CMEK | Customer-Managed Encryption Keys（顧客管理の暗号化鍵） |
| LWW | Last Write Wins（最終書き込み優先） |

### 10.2 参考資料

- [医療情報システムの安全管理に関するガイドライン 第6.0版](https://www.mhlw.go.jp/stf/shingi/0000516275.html)
- [Google Cloud Healthcare Solutions](https://cloud.google.com/solutions/healthcare-life-sciences)
- [Cloud Spanner Best Practices](https://cloud.google.com/spanner/docs/best-practice-list)
- [Flutter Offline-First Architecture](https://docs.flutter.dev/cookbook/persistence)

---

## 変更履歴

| バージョン | 日付 | 変更者 | 変更内容 |
|------------|------|--------|----------|
| 1.0 | 2025-12-11 | [作成者] | 初版作成 |

---

**承認欄**

| 役職 | 氏名 | 承認日 | 署名 |
|------|------|--------|------|
| プロダクトオーナー | | | |
| テックリード | | | |
| 医療監修 | | | |
