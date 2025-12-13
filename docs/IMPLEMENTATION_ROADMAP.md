# Visitas 実装ロードマップ

## 現在の実装状況

### Phase 1: MVP - 完了済み機能

| Sprint | 機能 | ステータス | 備考 |
|--------|------|----------|------|
| Sprint 1 | GCPプロジェクトセットアップ | ✅ 完了 | Cloud Spanner, Firebase Auth 統合済み |
| Sprint 1 | CI/CDパイプライン基礎 | ✅ 完了 | GitHub Actions設定済み |
| Sprint 2 | 患者CRUD API | ✅ 完了 | 14エンティティ実装済み |
| Sprint 2 | バリデーション | ✅ 完了 | サービス層で必須フィールド検証 |
| Sprint 3 | スケジュール管理 | ✅ 完了 | VisitSchedule CRUD実装済み |
| Sprint 6 | カルテ基本機能 | ✅ 完了 | MedicalRecord, Template実装済み |

### バックエンド実装済みコンポーネント

```
internal/
├── handlers/     14ファイル (3,753行)
├── services/     14ファイル (5,727行)
├── repository/   18ファイル (6,387行)
├── models/       13ファイル (1,921行)
├── middleware/   2ファイル (認証, 監査ログ)
└── config/       設定管理
```

### データベース

- **マイグレーション**: 19ファイル実装済み
- **テーブル**: patients, identifiers, social_profiles, coverages, conditions, allergies, visit_schedules, clinical_observations, care_plans, medication_orders, acp_records, medical_records, medical_record_templates, staff_patient_assignments, audit_logs

---

## Phase 1 残タスク

### Sprint 4: モバイルアプリ基本機能

| タスク | 優先度 | 説明 |
|--------|--------|------|
| Flutter プロジェクト作成 | 高 | mobile/ ディレクトリ配下に雛形作成 |
| ログイン・認証画面 | 高 | Firebase Auth統合、ログインUI |
| 患者一覧表示 | 高 | GET /api/v1/patients 連携 |
| スケジュール表示 | 高 | タイムライン形式での訪問予定表示 |
| オフラインDB統合 | 中 | Isar/sqflite によるローカルキャッシュ |

**実装方針**:
```
mobile/
├── lib/
│   ├── main.dart
│   ├── screens/
│   │   ├── login_screen.dart
│   │   ├── patient_list_screen.dart
│   │   ├── patient_detail_screen.dart
│   │   └── schedule_screen.dart
│   ├── models/           # API レスポンスモデル
│   ├── services/         # API 通信層
│   ├── providers/        # Riverpod 状態管理
│   └── widgets/          # 再利用可能なUI部品
├── test/
└── pubspec.yaml
```

### Sprint 5: 地図連携とルート表示

| タスク | 優先度 | 説明 |
|--------|--------|------|
| Google Maps API統合 | 高 | google_maps_flutter パッケージ |
| 患者住所の地図表示 | 高 | Addresses JSONB の geolocation 使用 |
| Googleマップアプリ連携 | 高 | ディープリンクでナビ起動 |
| 移動距離計算 | 中 | Distance Matrix API |

**実装方針**:
```dart
// 患者住所から地図表示
class PatientMapScreen extends StatelessWidget {
  final Patient patient;

  // addresses[0].geolocation から LatLng を取得
  LatLng get patientLocation => LatLng(
    patient.addresses[0].geolocation.latitude,
    patient.addresses[0].geolocation.longitude,
  );

  // Googleマップアプリで経路案内を起動
  void launchNavigation() {
    final url = 'google.navigation:q=${patientLocation.latitude},${patientLocation.longitude}';
    launchUrl(Uri.parse(url));
  }
}
```

---

## Phase 2: AI Integration

### Sprint 7: 法的書類AI生成（高優先度）

**対象書類**:
1. 訪問看護指示書
2. 居宅療養指導書
3. 特別訪問看護指示書
4. 訪問薬剤管理指導書
5. 点滴指示書

**実装計画**:

```
backend/
├── internal/
│   ├── ai/
│   │   ├── gemini_client.go       # Vertex AI Gemini クライアント
│   │   ├── prompts/
│   │   │   ├── nursing_instruction.txt    # 訪問看護指示書プロンプト
│   │   │   └── care_instruction.txt       # 居宅療養指導書プロンプト
│   │   └── document_generator.go  # 書類生成ロジック
│   ├── handlers/
│   │   └── documents.go           # 書類生成API
│   └── services/
│       └── document_service.go    # 書類生成サービス
├── pkg/
│   ├── pdfgen/                    # gofpdf ラッパー
│   └── excelgen/                  # excelize ラッパー
└── document_templates/            # 書類テンプレート定義
```

**APIエンドポイント**:
```
POST /api/v1/patients/{patient_id}/documents/generate
{
  "document_type": "nursing_instruction",
  "date_range": {
    "from": "2024-11-01",
    "to": "2024-11-30"
  }
}

Response:
{
  "document_id": "...",
  "content": {
    "patient_info": {...},
    "condition_summary": "AI生成テキスト...",
    "nursing_plan": "AI生成テキスト..."
  },
  "sources": [
    {"type": "visit_record", "id": "...", "date": "..."},
    {"type": "clinical_observation", "id": "...", "date": "..."}
  ],
  "download_urls": {
    "pdf": "https://...",
    "excel": "https://..."
  }
}
```

**Gemini プロンプト例**（訪問看護指示書）:
```
以下の患者情報と過去1ヶ月の訪問記録に基づき、訪問看護指示書の「病状・治療状態」欄の下書きを作成してください。

【患者情報】
- 氏名: {{patient.name}}
- 生年月日: {{patient.birth_date}}
- 要介護度: {{coverage.care_level_code}}
- 主たる疾患: {{conditions}}

【過去1ヶ月の訪問記録】
{{visit_records}}

【バイタルサイン推移】
{{clinical_observations}}

【作成条件】
- 医療専門職向けの簡潔な文体
- 200文字以内
- 断定的表現は避け、「〜の傾向がある」「〜と考えられる」等を使用
- 生成根拠となった記録の日付を明示
```

### Sprint 8: 紹介状からのカルテ自動生成

**実装計画**:

```
backend/
├── internal/
│   ├── ai/
│   │   └── document_parser.go     # PDF解析 → 構造化データ
│   ├── handlers/
│   │   └── referrals.go           # 紹介状アップロードAPI
│   └── services/
│       └── referral_service.go    # 紹介状処理サービス
```

**処理フロー**:
```
1. PDFアップロード → Cloud Storage
2. Gemini 1.5 Pro Document AI で解析
3. 抽出情報:
   - 患者基本情報
   - 既往歴
   - 現病歴
   - 処方薬
   - 紹介目的
4. MedicalRecord (SOAP形式) に変換
5. レビューUI表示 → 医師承認
```

### Sprint 9: 会話からのカルテ自動生成（ACI）

**Ambient Clinical Intelligence 実装**:

```
mobile/
├── lib/
│   ├── services/
│   │   └── audio_recorder.dart    # 音声録音サービス
│   └── screens/
│       └── visit_recording_screen.dart

backend/
├── internal/
│   ├── ai/
│   │   ├── speech_to_text.go      # Cloud Speech-to-Text
│   │   └── soap_generator.go      # 会話 → SOAP 変換
│   ├── handlers/
│   │   └── voice_records.go
│   └── services/
│       └── voice_record_service.go
```

**処理フロー**:
```
1. モバイルで診療中音声録音
2. 録音終了 → Cloud Storage にアップロード（暗号化）
3. Cloud Speech-to-Text で文字起こし
4. Gemini 1.5 Pro で SOAP 形式に構造化:
   - S (主観的情報): 患者の訴え抽出
   - O (客観的情報): バイタル・所見抽出
   - A (評価): 診断・アセスメント生成
   - P (計画): 治療方針生成
5. 医師がレビュー・編集・承認
```

---

## Phase 3: Optimization & Scale

### Route Optimization API 実装

**目的**: 複数患者への最適訪問順序の自動計算

```
backend/
├── internal/
│   ├── services/
│   │   └── route_optimization_service.go
│   └── handlers/
│       └── routes.go
```

**API**:
```
POST /api/v1/routes/optimize
{
  "date": "2024-12-15",
  "staff_id": "...",
  "constraints": {
    "start_location": {...},
    "end_location": {...},
    "time_windows": true,
    "max_travel_time": 480  // 8時間
  }
}

Response:
{
  "optimized_route": [
    {"order": 1, "patient_id": "...", "arrival_time": "09:00", "departure_time": "09:45"},
    {"order": 2, "patient_id": "...", "arrival_time": "10:15", "departure_time": "11:00"},
    ...
  ],
  "total_distance_km": 45.2,
  "total_travel_time_minutes": 120,
  "savings_vs_original": {
    "distance_km": 12.5,
    "time_minutes": 35
  }
}
```

### FHIR API 連携

**目的**: 他医療機関との標準規格データ交換

```
backend/
├── internal/
│   ├── fhir/
│   │   ├── client.go           # Cloud Healthcare API クライアント
│   │   ├── patient_mapper.go   # Patient ⇔ FHIR Patient 変換
│   │   ├── condition_mapper.go # MedicalCondition ⇔ FHIR Condition
│   │   └── allergy_mapper.go   # AllergyIntolerance ⇔ FHIR
│   └── handlers/
│       └── fhir.go
```

### IoT/ウェアラブル連携

**目的**: 患者バイタルデータのリアルタイム収集

```
backend/
├── internal/
│   ├── iot/
│   │   ├── device_manager.go   # デバイス管理
│   │   └── data_ingestion.go   # データ取り込み (Pub/Sub)
│   └── handlers/
│       └── devices.go
```

**対応デバイス候補**:
- 血圧計（Omron Connect 対応機種）
- パルスオキシメーター
- 体重計
- 活動量計

### データ分析基盤

**BigQuery 連携**:
```
infra/terraform/
├── bigquery.tf              # データセット・テーブル定義
└── dataflow.tf              # Spanner → BigQuery パイプライン

backend/
├── internal/
│   └── analytics/
│       ├── export_service.go    # 定期エクスポート
│       └── dashboard_queries.go # ダッシュボード用クエリ
```

**分析項目**:
- 訪問パターン分析
- 患者状態悪化予測
- スタッフ稼働率分析
- ルート効率分析

---

## 技術的負債・改善タスク

### 高優先度

| タスク | 説明 | 影響範囲 |
|--------|------|---------|
| ユニットテスト拡充 | handlers, services のカバレッジ向上 | 全体 |
| 統合テスト自動化 | CI での Spanner エミュレータテスト | CI/CD |
| エラーハンドリング統一 | カスタムエラー型の導入 | 全サービス |
| レスポンス形式統一 | 標準 API レスポンス構造体 | 全ハンドラー |

### 中優先度

| タスク | 説明 | 影響範囲 |
|--------|------|---------|
| キャッシュ層追加 | Redis/Memorystore 導入 | 読み取り性能 |
| レート制限 | Cloud Armor + ミドルウェア | API 全体 |
| OpenAPI 仕様書更新 | docs/openapi.yaml の自動生成 | ドキュメント |
| ログ構造化 | JSON ログ形式への統一 | 運用監視 |

### 低優先度

| タスク | 説明 | 影響範囲 |
|--------|------|---------|
| gRPC 対応 | 内部マイクロサービス間通信 | 将来拡張 |
| GraphQL 検討 | 複雑なクエリ対応 | モバイル連携 |
| イベントソーシング | 監査履歴の完全性向上 | 監査要件 |

---

## セキュリティ対応タスク

| タスク | 優先度 | 説明 |
|--------|--------|------|
| ペネトレーションテスト | 高 | 本番前に外部診断実施 |
| WAF ルール設定 | 高 | Cloud Armor OWASP ルール適用 |
| シークレットローテーション | 中 | Secret Manager 自動ローテーション |
| 脆弱性スキャン | 中 | Container Analysis 統合 |
| SOC2 準拠確認 | 中 | 監査要件の文書化 |

---

## 推奨実装順序

```
Phase 1 残り (4週間)
├── Week 1-2: Flutter プロジェクト作成 + 認証画面
├── Week 2-3: 患者一覧 + スケジュール表示
└── Week 3-4: 地図連携 + オフライン基盤

Phase 2 (12週間)
├── Week 1-4: 法的書類AI生成
│   ├── Gemini クライアント実装
│   ├── プロンプトエンジニアリング
│   ├── PDF/Excel 生成
│   └── レビューUI (Web)
├── Week 5-8: 紹介状からのカルテ生成
│   ├── Document AI 統合
│   ├── 構造化抽出ロジック
│   └── レビュー・承認フロー
└── Week 9-12: 会話からのSOAP生成
    ├── 音声録音機能 (モバイル)
    ├── Speech-to-Text 統合
    ├── SOAP 変換ロジック
    └── 編集・承認UI

Phase 3 (継続的)
├── Route Optimization
├── FHIR 連携
├── IoT 統合
└── データ分析基盤
```

---

## KPI・成功指標

### Phase 1 完了基準

- [ ] モバイルアプリで患者一覧表示可能
- [ ] 訪問スケジュールがタイムライン表示される
- [ ] 患者住所から地図ナビ起動可能
- [ ] オフライン時もデータ閲覧可能

### Phase 2 完了基準

- [ ] 訪問看護指示書の下書きが30秒以内に生成される
- [ ] 紹介状PDFから初診カルテの80%以上が自動入力される
- [ ] 診療会話からSOAPカルテの下書きが生成される
- [ ] 医師のカルテ記載時間が50%削減される

### Phase 3 完了基準

- [ ] 最適ルートにより移動時間が20%削減される
- [ ] 他医療機関とFHIRでデータ交換可能
- [ ] 患者バイタルがリアルタイムで収集される
- [ ] ダッシュボードで運営分析が可能
