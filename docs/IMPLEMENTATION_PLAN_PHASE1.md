# Phase 1実装計画 - 患者マスタ最優先アプローチ
**Visitas 在宅医療プラットフォーム | Sprint 1-2 詳細計画**

---

## エグゼクティブサマリー

本計画書は、Visitas Phase 1において**患者マスタのリッチ化とセキュア化を最優先**で実装する戦略を定義する。

### 設計原則の変更点

| 項目 | 変更前 (DATABASE_REQUIREMENTS.md) | 変更後 (本計画) | 理由 |
|---|---|---|---|
| **Spannerインターフェース** | 明記なし | **PostgreSQL Interface** | JSONB型サポート、エコシステム充実 |
| **インターリーブテーブル** | 多用 | **使用しない** | PostgreSQL Interfaceでは非対応 |
| **親子関係管理** | INTERLEAVE IN PARENT | **Foreign Key + クラスター化インデックス** | 標準SQL構文、移植性向上 |
| **実装スケジュール** | Sprint 2(Week 3-4) | **Sprint 1前半(Week 1-2)** | 患者マスタがすべての起点 |

---

## Phase 1の目標

### Week 1-2: 患者マスタ構築 (本計画)
✅ **患者基本情報の完全なリッチ化**
✅ **多層防御セキュリティの実装**
✅ **CRUD APIの完成**

### Week 3-4: スケジュール管理 (CLAUDE.md Sprint 3)
- 訪問スケジュールCRUD
- 医師・患者の割り当て

### Week 5-6: モバイルアプリ基礎 (CLAUDE.md Sprint 4開始)
- Flutter雛形
- 患者一覧表示

---

## 実装する6つのコアテーブル

### 1. patients (患者基本情報)
**優先度: P0 (最高)**

**設計の特徴:**
- 氏名・住所・連絡先の**変更履歴管理** (JSONB配列)
- Generated Columnsによる高速検索
- 論理削除 (医療法5年保存義務)

**主要フィールド:**
```sql
- patient_id (UUID v4)
- birth_date, gender, blood_type
- name_history (JSONB配列)
- contact_points (JSONB配列)
- addresses (JSONB配列)
- consent_status (同意管理)
```

**実装期間:** Day 3-5

---

### 2. patient_identifiers (患者識別子)
**優先度: P0**

**設計の特徴:**
- マイナンバーの**カラムレベル暗号化** (Cloud KMS AEAD)
- 複数識別子対応 (保険証、介護保険証、MRN)

**暗号化実装:**
```go
// pkg/encryption/kms_aead.go
func (e *KMSEncryptor) EncryptMyNumber(ctx context.Context, plaintext string) (string, error)
func (e *KMSEncryptor) DecryptMyNumber(ctx context.Context, ciphertext string) (string, error)
```

**実装期間:** Day 3-5 (patients同時)

---

### 3. staff_patient_assignments (スタッフ-患者割当)
**優先度: P0**

**設計の特徴:**
- Row-Level Security (RLS)の基盤
- 担当患者のみ閲覧可能にする

**主要フィールド:**
```sql
- staff_id (Firebase UID)
- patient_id
- role ("doctor" | "nurse" | "care_manager")
- assignment_type ("primary" | "backup")
```

**実装期間:** Day 11-12

---

### 4. patient_social_profiles (患者社会的背景)
**優先度: P1 (高)**

**設計の特徴:**
- Subjective (S) ドメイン: 患者の「語り」を構造化
- 独居状況、介護者負担、経済的背景
- **バージョン管理** (profile_version)

**JSONB構造:**
```json
{
  "livingSituation": {...},
  "keyPersons": [
    {
      "caregiverBurden": {
        "zaritScore": 45,
        "burnoutRisk": "moderate"
      }
    }
  ],
  "financialBackground": {...}
}
```

**実装期間:** Day 15-17

---

### 5. patient_coverages (保険情報)
**優先度: P1**

**設計の特徴:**
- 医療保険・介護保険・公費の統一管理
- **要介護度のGenerated Column**
- 有効期限切れ自動検出

**主要フィールド:**
```sql
- insurance_type ("medical" | "long_term_care" | "public_expense")
- care_level_code (Generated Column)
- valid_from, valid_to
- details (JSONB: 保険証詳細)
```

**実装期間:** Day 15-17

---

### 6. medical_conditions & allergy_intolerances
**優先度: P1**

**設計の特徴:**
- FHIR Condition/AllergyIntolerance準拠
- 医療安全の最重要データ
- ICD-10コード対応

**実装期間:** Day 18-19

---

## セキュリティ実装 (3層防御)

### Layer 1: データ暗号化

#### 1.1 CMEK (Customer-Managed Encryption Keys)
**対象:** Spanner全体

**設定:**
```bash
# Cloud KMS暗号鍵作成
gcloud kms keyrings create visitas-keyring \
  --location=asia-northeast1

gcloud kms keys create spanner-cmek-key \
  --keyring=visitas-keyring \
  --location=asia-northeast1 \
  --purpose=encryption

# Spannerインスタンスに適用
gcloud spanner instances update stunning-grin-480914-n1-instance \
  --encryption-type=CUSTOMER_MANAGED_ENCRYPTION \
  --kms-key=projects/stunning-grin-480914-n1/locations/asia-northeast1/keyRings/visitas-keyring/cryptoKeys/spanner-cmek-key
```

**実装期間:** Day 1-2

#### 1.2 アプリケーション層暗号化 (AEAD)
**対象:** マイナンバー (patient_identifiers.encrypted_value)

**実装:**
- `pkg/encryption/kms_aead.go`
- Cloud KMS AEAD (Authenticated Encryption with Associated Data)
- AAD (追加認証データ): `"mynumber"`

**実装期間:** Day 8-10

---

### Layer 2: Row-Level Security (RLS)

**要件:**
- 医師: 全患者閲覧可能
- 看護師/ケアマネ: 担当患者のみ閲覧

**実装方式:**
```sql
CREATE VIEW view_my_patients AS
SELECT p.*
FROM patients p
INNER JOIN staff_patient_assignments spa
  ON p.patient_id = spa.patient_id
WHERE spa.staff_id = current_setting('app.current_user_id')
  AND spa.status = 'active';
```

**Go実装:**
```go
func (r *PatientRepository) GetMyPatients(ctx context.Context, firebaseUID string) ([]Patient, error) {
    // セッション変数設定
    _, err := r.client.Apply(ctx, []*spanner.Mutation{
        spanner.Insert("sessions", []string{"firebase_uid"}, []interface{}{firebaseUID}),
    })

    // RLSビュー経由でクエリ
    stmt := spanner.Statement{SQL: "SELECT * FROM view_my_patients"}
    // ...
}
```

**実装期間:** Day 11-12

---

### Layer 3: 監査ログ

**対象イベント:**
- 患者情報閲覧 (view)
- 患者情報作成 (create)
- 患者情報更新 (update)
- マイナンバー復号 (decrypt)

**テーブル:**
```sql
CREATE TABLE audit_patient_access_logs (
    log_id varchar(36) NOT NULL,
    event_time timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    actor_id varchar(36) NOT NULL,
    action varchar(50) NOT NULL,
    resource_id varchar(36) NOT NULL,
    patient_id varchar(36) NOT NULL,
    accessed_fields jsonb,
    success boolean NOT NULL,
    PRIMARY KEY (log_id)
);
```

**Go Middleware実装:**
```go
// internal/middleware/audit_logger.go
func AuditLogger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // リクエスト処理
        next.ServeHTTP(w, r)

        // 監査ログ記録
        auditLog := &AuditLog{
            EventTime: time.Now(),
            ActorID: r.Context().Value("firebase_uid").(string),
            Action: r.Method,
            ResourceID: chi.URLParam(r, "id"),
        }
        repo.InsertAuditLog(r.Context(), auditLog)
    })
}
```

**実装期間:** Day 13-14

---

## API設計

### エンドポイント一覧

| Method | Path | 説明 | 認証 | RLS |
|---|---|---|---|---|
| `POST` | `/api/v1/patients` | 患者登録 | Required | - |
| `GET` | `/api/v1/patients/:id` | 患者詳細取得 | Required | Yes |
| `PUT` | `/api/v1/patients/:id` | 患者情報更新 | Required | Yes |
| `DELETE` | `/api/v1/patients/:id` | 患者論理削除 | Required | Yes |
| `GET` | `/api/v1/patients` | 担当患者一覧 | Required | Yes |
| `POST` | `/api/v1/patients/:id/identifiers` | 識別子追加 (マイナンバー等) | Required | Yes |
| `GET` | `/api/v1/patients/:id/social-profile` | 社会的背景取得 | Required | Yes |
| `PUT` | `/api/v1/patients/:id/social-profile` | 社会的背景更新 | Required | Yes |
| `GET` | `/api/v1/patients/:id/coverages` | 保険情報一覧 | Required | Yes |
| `POST` | `/api/v1/patients/:id/coverages` | 保険情報追加 | Required | Yes |

### リクエスト例: 患者登録

**POST /api/v1/patients**

```json
{
  "name": {
    "use": "official",
    "family": "山田",
    "given": "太郎",
    "kana": "ヤマダ タロウ"
  },
  "birthDate": "1950-04-01",
  "gender": "male",
  "bloodType": "A+",
  "contactPoints": [
    {
      "system": "phone",
      "value": "090-1234-5678",
      "use": "mobile",
      "rank": 1
    }
  ],
  "addresses": [
    {
      "use": "home",
      "postalCode": "160-0023",
      "prefecture": "東京都",
      "city": "新宿区",
      "line": "西新宿1-2-3",
      "geolocation": {
        "latitude": 35.6895,
        "longitude": 139.6917
      }
    }
  ],
  "consentStatus": "obtained",
  "consentObtainedAt": "2025-12-10T10:00:00+09:00"
}
```

**レスポンス:**
```json
{
  "patientId": "550e8400-e29b-41d4-a716-446655440000",
  "createdAt": "2025-12-12T14:30:00+09:00",
  "message": "患者情報を登録しました"
}
```

---

## 詳細スケジュール (28日間)

### Week 1: インフラとコアテーブル

| Day | タスク | 成果物 | 担当 |
|---|---|---|---|
| **1** | GCPプロジェクトセットアップ | Spannerインスタンス作成完了 | インフラ |
| **2** | CMEK暗号鍵作成、Firebase Auth設定 | 暗号化基盤完成 | セキュリティ |
| **3** | `patients`テーブルマイグレーション | DDL適用完了 | バックエンド |
| **4** | `patient_identifiers`テーブル作成 | DDL適用完了 | バックエンド |
| **5** | Generated Columnsのパフォーマンステスト | ベンチマーク結果 | バックエンド |
| **6** | Go Repository層実装 (患者CRUD) | `patient_repository.go`完成 | バックエンド |
| **7** | ユニットテスト作成 (カバレッジ80%以上) | テスト完了 | バックエンド |

**Week 1 完了条件:**
- [ ] Spannerインスタンスが稼働中
- [ ] `patients`, `patient_identifiers`テーブルが作成済み
- [ ] Go Repositoryのユニットテストが全てパス

---

### Week 2: セキュリティ実装

| Day | タスク | 成果物 | 担当 |
|---|---|---|---|
| **8** | KMS AEAD暗号化実装 | `pkg/encryption/kms_aead.go`完成 | バックエンド |
| **9** | マイナンバー暗号化テスト | 暗号化/復号テスト完了 | バックエンド |
| **10** | 暗号化キーローテーション戦略策定 | 運用手順書 | セキュリティ |
| **11** | `staff_patient_assignments`テーブル作成 | DDL適用完了 | バックエンド |
| **12** | RLSビュー (`view_my_patients`)作成 | ビュー作成完了 | バックエンド |
| **13** | 監査ログテーブル作成 | `audit_patient_access_logs`完成 | バックエンド |
| **14** | 監査ログMiddleware実装 | `audit_logger.go`完成 | バックエンド |

**Week 2 完了条件:**
- [ ] マイナンバーの暗号化/復号が正常に動作
- [ ] RLSビューで担当患者のみ取得できることを確認
- [ ] すべてのAPI呼び出しで監査ログが記録される

---

### Week 3: リッチ化テーブル実装

| Day | タスク | 成果物 | 担当 |
|---|---|---|---|
| **15** | `patient_social_profiles`テーブル作成 | DDL適用完了 | バックエンド |
| **16** | `patient_coverages`テーブル作成 | DDL適用完了 | バックエンド |
| **17** | JSONBバリデーション関数実装 | バリデーションロジック完成 | バックエンド |
| **18** | `medical_conditions`テーブル作成 | DDL適用完了 | バックエンド |
| **19** | `allergy_intolerances`テーブル作成 | DDL適用完了 | バックエンド |
| **20** | Service層実装 (ビジネスロジック) | `patient_service.go`完成 | バックエンド |
| **21** | Service層のユニットテスト | テスト完了 | バックエンド |

**Week 3 完了条件:**
- [ ] 6つのテーブルすべてが作成済み
- [ ] JSONBバリデーションが動作
- [ ] Service層のテストが全てパス

---

### Week 4: API実装と総合テスト

| Day | タスク | 成果物 | 担当 |
|---|---|---|---|
| **22** | REST APIハンドラー実装 (患者CRUD) | `patients_handler.go`完成 | バックエンド |
| **23** | 識別子・社会的背景APIハンドラー実装 | `identifiers_handler.go`等完成 | バックエンド |
| **24** | OpenAPI仕様書作成 | `openapi.yaml`完成 | バックエンド |
| **25** | API統合テスト (E2E) | Postmanコレクション完成 | QA |
| **26** | セキュリティテスト (RLS、暗号化) | テスト完了 | セキュリティ |
| **27** | パフォーマンステスト (100患者) | ベンチマーク結果 | バックエンド |
| **28** | ドキュメント整備とレビュー | Phase 1完了報告 | 全員 |

**Week 4 完了条件:**
- [ ] 全APIエンドポイントが正常に動作
- [ ] RLSテストが全てパス (他人の患者は閲覧不可)
- [ ] 応答時間 <200ms (95パーセンタイル <500ms)
- [ ] OpenAPI仕様書が完成

---

## マイグレーションファイル構成

```
backend/migrations/
├── 001_create_patients.sql
├── 002_create_patient_identifiers.sql
├── 003_create_staff_patient_assignments.sql
├── 004_create_audit_patient_access_logs.sql
├── 005_create_patient_social_profiles.sql
├── 006_create_patient_coverages.sql
├── 007_create_medical_conditions.sql
├── 008_create_allergy_intolerances.sql
├── 009_create_view_my_patients.sql
└── 010_create_indexes.sql
```

**適用方法:**
```bash
# PostgreSQL Interface用のマイグレーション適用
for file in backend/migrations/*.sql; do
  echo "Applying $file..."
  gcloud spanner databases execute-sql stunning-grin-480914-n1-db \
    --instance=stunning-grin-480914-n1-instance \
    --sql="$(cat $file)"
done
```

---

## Go プロジェクト構成

```
backend/
├── cmd/
│   └── api/
│       └── main.go                    # エントリーポイント
├── internal/
│   ├── handlers/
│   │   ├── patients.go                # 患者CRUD APIハンドラー
│   │   ├── identifiers.go             # 識別子APIハンドラー
│   │   ├── social_profiles.go         # 社会的背景APIハンドラー
│   │   └── coverages.go               # 保険情報APIハンドラー
│   ├── services/
│   │   └── patient_service.go         # ビジネスロジック
│   ├── repository/
│   │   └── spanner/
│   │       ├── patient_repository.go  # 患者データアクセス
│   │       ├── identifier_repository.go
│   │       └── audit_repository.go
│   ├── models/
│   │   ├── patient.go                 # 患者データモデル
│   │   ├── identifier.go
│   │   └── social_profile.go
│   └── middleware/
│       ├── auth.go                    # Firebase認証
│       └── audit_logger.go            # 監査ログ記録
├── pkg/
│   ├── encryption/
│   │   └── kms_aead.go                # KMS暗号化ユーティリティ
│   ├── validator/
│   │   └── patient_validator.go       # バリデーション
│   └── logger/
│       └── logger.go                   # 構造化ログ
├── tests/
│   ├── integration/
│   │   └── patient_api_test.go        # API統合テスト
│   └── security/
│       └── rls_test.go                 # RLSテスト
├── go.mod
└── go.sum
```

---

## テスト戦略

### ユニットテスト (Day 7, 21)
**カバレッジ目標: 80%以上**

```bash
go test ./... -v -cover -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### 統合テスト (Day 25)
**ツール: Postman / Newman**

```json
{
  "name": "Visitas Phase 1 API Tests",
  "tests": [
    {
      "name": "患者登録 (正常系)",
      "request": "POST /api/v1/patients",
      "assertions": ["status == 201", "response.patientId != null"]
    },
    {
      "name": "RLS: 他人の患者閲覧 (異常系)",
      "request": "GET /api/v1/patients/:other_patient_id",
      "assertions": ["status == 403"]
    }
  ]
}
```

### セキュリティテスト (Day 26)

| テストケース | 期待結果 |
|---|---|
| **マイナンバー暗号化確認** | DB上で暗号化された値が格納されている |
| **RLS: 担当患者のみ閲覧** | view_my_patients経由で自分の患者のみ取得 |
| **RLS: 他人の患者閲覧拒否** | 403 Forbidden |
| **監査ログ記録** | audit_logsにレコードが存在 |

### パフォーマンステスト (Day 27)
**ツール: k6**

```javascript
import http from 'k6/http';
import { check } from 'k6';

export let options = {
  stages: [
    { duration: '1m', target: 50 },  // 50同時接続まで増加
    { duration: '3m', target: 50 },  // 3分間維持
    { duration: '1m', target: 0 },   // 0まで減少
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],  // 95%が500ms以内
    http_req_failed: ['rate<0.01'],    // エラー率 1%未満
  },
};

export default function() {
  let res = http.get('http://localhost:8080/api/v1/patients');
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 200ms': (r) => r.timings.duration < 200,
  });
}
```

---

## リスク管理

### 高リスク項目

| リスク | 影響度 | 対策 | 担当 |
|---|---|---|---|
| **Spanner PostgreSQL InterfaceのJSONB関数未サポート** | 高 | 事前検証 (Day 5)、代替実装準備 | バックエンド |
| **Generated Columnsの書き込みパフォーマンス劣化** | 中 | ベンチマーク実施 (Day 5)、必要に応じて削除 | バックエンド |
| **RLS実装の複雑性** | 中 | シンプルなビューベース実装を採用 | バックエンド |
| **暗号化キーのアクセス権限エラー** | 高 | IAMロールの事前確認 (Day 2) | セキュリティ |

---

## 成功基準 (Definition of Done)

### Phase 1完了の定義

- [ ] 6つのコアテーブルがすべて作成され、データ投入可能
- [ ] 患者CRUD APIが全て実装され、200 OKを返す
- [ ] マイナンバー暗号化/復号が正常に動作
- [ ] RLSテストが全てパス (担当患者のみ閲覧可能)
- [ ] 監査ログがすべてのAPI呼び出しで記録される
- [ ] ユニットテストカバレッジ 80%以上
- [ ] API統合テストが全てパス
- [ ] パフォーマンステストで応答時間 <200ms (平均)
- [ ] OpenAPI仕様書が完成し、レビュー済み
- [ ] セキュリティチェックリストが全て✅

---

## 次のフェーズへの引き継ぎ

### Phase 1完了後、Phase 2 (Sprint 3)で実装すべき項目

1. **訪問スケジュールテーブル** (`visit_schedules`)
   - 患者マスタとのForeign Key設定
   - Google Maps Route Optimization API統合準備

2. **訪問記録テーブル** (`visit_records`)
   - 実施記録の管理

3. **SOAPノートテーブル** (`soap_notes`)
   - カルテデータモデルの実装

4. **バイタルサイン観測テーブル** (`clinical_observations`)
   - 時系列データ管理

### Phase 2への技術的前提条件

- [ ] 患者マスタAPIが安定稼働
- [ ] セキュリティ基盤 (CMEK, RLS, 監査ログ)が確立
- [ ] JSONBデータ構造の設計パターンが確立
- [ ] Go プロジェクト構成が標準化

---

## まとめ

本実装計画は、**患者マスタのリッチ化とセキュア化を最優先**とし、以下を達成します:

### ✅ 技術的達成事項

1. **Spanner PostgreSQL Interface完全対応**
   - インターリーブテーブルを使わない実装可能な設計
   - JSONB型による柔軟なスキーマ管理

2. **多層防御セキュリティ**
   - CMEK + アプリケーション層暗号化
   - Row-Level Security (RLS)
   - 監査ログ完全記録

3. **医療情報のリッチな管理**
   - 氏名・住所の変更履歴
   - 社会的背景の構造化
   - 保険情報の完全管理

### 📅 28日間で実現可能な理由

- 明確なマイルストーン (Week 1-4)
- 具体的なタスク配分 (Day 1-28)
- テスト戦略の明確化
- リスク管理の徹底

**この計画に従うことで、Phase 1を28日間で完了し、Phase 2以降の機能実装の強固な基盤を構築します。**

---

**改訂履歴:**

| 版 | 日付 | 変更内容 | 変更者 |
|---|---|---|---|
| 1.0 | 2025-12-12 | 初版作成 | |
