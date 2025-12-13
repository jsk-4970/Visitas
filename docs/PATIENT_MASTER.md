# 患者マスタ設計ドキュメント

## 概要

Visitasの患者マスタは、在宅医療における患者情報を包括的に管理するためのデータ構造です。FHIR R4標準に準拠し、日本の医療制度（介護保険、マイナンバー等）に対応しています。

## エンティティ関係図

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Patient (患者)                                  │
│  - 基本情報 (氏名履歴, 生年月日, 性別, 血液型)                                │
│  - 連絡先 (電話, メール, FAX)                                               │
│  - 住所 (位置情報含む)                                                       │
│  - 同意管理                                                                  │
└──────────────────────────────┬──────────────────────────────────────────────┘
                               │
       ┌───────────────────────┼───────────────────────────────┐
       │                       │                               │
       ▼                       ▼                               ▼
┌──────────────────┐  ┌──────────────────┐  ┌──────────────────────────────┐
│ PatientIdentifier│  │  SocialProfile   │  │       Coverage               │
│ (識別子)          │  │ (社会的背景)      │  │ (保険情報)                    │
│ - マイナンバー    │  │ - 生活状況        │  │ - 医療保険                    │
│ - 保険証番号      │  │ - キーパーソン    │  │ - 介護保険                    │
│ - 介護保険番号    │  │ - 経済状況        │  │ - 公費負担                    │
│ - MRN            │  │ - 社会的支援      │  │ - 自己負担率                  │
└──────────────────┘  └──────────────────┘  └──────────────────────────────┘
       │                       │                               │
       └───────────────────────┼───────────────────────────────┘
                               │
       ┌───────────────────────┴───────────────────────────────┐
       │                                                       │
       ▼                                                       ▼
┌──────────────────────────────┐  ┌──────────────────────────────────────────┐
│     MedicalCondition         │  │         AllergyIntolerance               │
│ (疾患・病歴)                  │  │ (アレルギー・不耐性)                      │
│ - ICD-10/SNOMED-CT           │  │ - アレルゲン                              │
│ - 臨床ステータス              │  │ - 反応歴                                  │
│ - 発症日・治癒日              │  │ - 重症度                                  │
│ - 重症度                      │  │ - 臨床ステータス                          │
└──────────────────────────────┘  └──────────────────────────────────────────┘
```

## 1. Patient（患者）

### 基本フィールド

| フィールド | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| `patient_id` | VARCHAR(36) | ✓ | UUID形式の主キー |
| `birth_date` | DATE | ✓ | 生年月日 |
| `gender` | VARCHAR(10) | ✓ | 性別 ("male", "female", "other", "unknown") |
| `blood_type` | VARCHAR(10) | - | 血液型 (A, B, O, AB, 不明) |

### JSONB フィールド

#### name_history（氏名履歴）
婚姻・離婚などによる氏名変更履歴を保持します。

```json
[
  {
    "family": "田中",
    "given": "太郎",
    "kana": "タナカ タロウ",
    "valid_from": "1990-01-15T00:00:00Z",
    "valid_to": null,
    "change_reason": null
  },
  {
    "family": "山田",
    "given": "太郎",
    "kana": "ヤマダ タロウ",
    "valid_from": "1960-01-01T00:00:00Z",
    "valid_to": "1990-01-14T00:00:00Z",
    "change_reason": "marriage"
  }
]
```

#### contact_points（連絡先）
優先順位付きの連絡先情報を管理します。

```json
[
  {
    "system": "phone",
    "value": "090-1234-5678",
    "use": "mobile",
    "rank": 1,
    "verified_at": "2024-12-01T00:00:00Z"
  },
  {
    "system": "email",
    "value": "tanaka@example.com",
    "use": "home",
    "rank": 2,
    "verified_at": null
  }
]
```

#### addresses（住所）
訪問診療に必要な位置情報とアクセス情報を含みます。

```json
[
  {
    "use": "home",
    "postal_code": "100-0001",
    "prefecture": "東京都",
    "city": "千代田区",
    "line": "丸の内1-1-1",
    "building": "マンションA 101号室",
    "geolocation": {
      "latitude": 35.6762,
      "longitude": 139.7674
    },
    "access_instructions": "2階の北側入口からお入りください。インターホン101",
    "valid_from": "2020-04-01T00:00:00Z",
    "valid_to": null
  }
]
```

### 生成カラム（検索用）

JSONB から自動抽出される読み取り専用カラム：

| カラム | 抽出元 | 説明 |
|--------|--------|------|
| `current_family_name` | name_history[0].family | 現在の姓 |
| `current_given_name` | name_history[0].given | 現在の名 |
| `primary_phone` | contact_points[rank=1].value | 主連絡先電話番号 |
| `current_prefecture` | addresses[use=home].prefecture | 現住所の都道府県 |
| `current_city` | addresses[use=home].city | 現住所の市区町村 |

### 同意管理

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `consent_status` | VARCHAR(20) | "obtained", "not_obtained", "conditional" |
| `consent_obtained_at` | TIMESTAMPTZ | 同意取得日時 |
| `consent_withdrawn_at` | TIMESTAMPTZ | 同意撤回日時 |
| `consent_details` | JSONB | 同意内容の詳細 |

### 監査フィールド

全エンティティ共通：

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `created_at` | TIMESTAMPTZ | 作成日時 |
| `created_by` | VARCHAR(36) | 作成者ID |
| `updated_at` | TIMESTAMPTZ | 更新日時 |
| `updated_by` | VARCHAR(36) | 更新者ID |
| `deleted` | BOOLEAN | 論理削除フラグ |
| `deleted_at` | TIMESTAMPTZ | 削除日時 |
| `deleted_reason` | TEXT | 削除理由 |

---

## 2. PatientIdentifier（患者識別子）

### 識別子タイプ

| タイプ | 説明 | 暗号化 |
|--------|------|--------|
| `my_number` | マイナンバー（12桁） | ✓ KMS暗号化 |
| `insurance_id` | 健康保険証番号 | - |
| `care_insurance_id` | 介護保険被保険者番号 | - |
| `mrn` | 医療記録番号 | - |
| `other` | その他の識別子 | - |

### 基本フィールド

| フィールド | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| `identifier_id` | VARCHAR(36) | ✓ | 主キー |
| `patient_id` | VARCHAR(36) | ✓ | 患者への外部キー |
| `identifier_type` | VARCHAR(30) | ✓ | 識別子タイプ |
| `identifier_value` | TEXT | ✓ | 識別子の値（マイナンバーは暗号化） |
| `is_primary` | BOOLEAN | - | 同タイプ内での主識別子フラグ |

### 有効期間・発行者

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `valid_from` | DATE | 有効期間開始 |
| `valid_to` | DATE | 有効期間終了 |
| `issuer_name` | VARCHAR(200) | 発行者名（保険者名等） |
| `issuer_code` | VARCHAR(20) | 発行者コード |

### 確認ステータス

| ステータス | 説明 |
|-----------|------|
| `verified` | 確認済み |
| `unverified` | 未確認 |
| `expired` | 期限切れ |
| `invalid` | 無効 |

### セキュリティ対策

マイナンバーの取り扱い：
1. **暗号化**: Cloud KMSによる暗号化（保存時暗号化）
2. **アクセス制御**: 閲覧権限を持つユーザーのみ復号可能
3. **監査ログ**: 全アクセスを`audit_logs`テーブルに記録
4. **復号時の認可チェック**: 毎回`CheckStaffAccess`で権限確認

---

## 3. SocialProfile（社会的背景）

バージョン管理により、変更履歴を保持します。

### 基本フィールド

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `profile_id` | VARCHAR(36) | 主キー |
| `patient_id` | VARCHAR(36) | 患者への外部キー |
| `profile_version` | INT64 | バージョン番号 |
| `content` | JSONB | 詳細情報（下記構造） |

### content JSONB 構造

```json
{
  "livingSituation": {
    "livingAlone": false,
    "requiresCaregiverSupport": true,
    "housingType": "apartment",
    "accessibility": {
      "wheelchairAccessible": true,
      "elevatorAvailable": true,
      "stairsCount": 0
    }
  },
  "keyPersons": [
    {
      "relationship": "spouse",
      "name": "田中花子",
      "age": 68,
      "contactInfo": {
        "phone": "090-9876-5432"
      },
      "livesWith": true,
      "isPrimaryCaregiver": true,
      "caregiverBurden": {
        "zaritScore": 35,
        "burnoutRisk": "moderate",
        "assessedAt": "2024-12-01T00:00:00Z",
        "notes": "介護疲れの兆候あり"
      }
    }
  ],
  "financialBackground": {
    "incomeLevel": "middle",
    "publicAssistance": false,
    "insuranceCoverage": {
      "medicalInsurance": true,
      "longTermCareInsurance": true,
      "privateInsurance": true
    }
  },
  "socialSupport": {
    "communityServices": ["meal_delivery", "daycare"],
    "neighborSupport": "moderate",
    "religiousAffiliation": null
  }
}
```

### 生成カラム

| カラム | 説明 |
|--------|------|
| `lives_alone` | 独居かどうか |
| `requires_caregiver_support` | 介護者支援が必要か |

---

## 4. Coverage（保険情報）

### 保険タイプ

| タイプ | 説明 |
|--------|------|
| `medical` | 医療保険（健康保険、国民健康保険等） |
| `long_term_care` | 介護保険 |
| `public_expense` | 公費負担医療（特定疾患、生活保護等） |

### 基本フィールド

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `coverage_id` | VARCHAR(36) | 主キー |
| `patient_id` | VARCHAR(36) | 患者への外部キー |
| `insurance_type` | VARCHAR(30) | 保険タイプ |
| `status` | VARCHAR(20) | "active", "expired", "suspended", "terminated" |
| `priority` | INT | 優先順位（1=主保険） |
| `details` | JSONB | タイプ別詳細（下記） |

### details JSONB（医療保険）

```json
{
  "insurerName": "全国健康保険協会",
  "insurerNumber": "01234567",
  "certificateNumber": "12345678",
  "certificateSymbol": "001",
  "copayRate": 30,
  "insuredPersonCategory": "本人",
  "employerName": "株式会社サンプル"
}
```

### details JSONB（介護保険）

```json
{
  "insurerName": "〇〇市",
  "insurerNumber": "131001",
  "certificateNumber": "0123456789",
  "careLevelCode": "要介護3",
  "careLevelAssessedAt": "2024-10-01T00:00:00Z",
  "copayRate": 10,
  "monthlyServiceLimit": 270480,
  "certificationValidFrom": "2024-10-01",
  "certificationValidTo": "2025-09-30"
}
```

### 生成カラム

| カラム | 抽出元 | 説明 |
|--------|--------|------|
| `care_level_code` | details.careLevelCode | 要介護度 |
| `copay_rate` | details.copayRate | 自己負担率 |

---

## 5. MedicalCondition（疾患・病歴）

FHIR R4 Conditionリソースに準拠しています。

### 臨床ステータス

| ステータス | 説明 |
|-----------|------|
| `active` | 現在治療中 |
| `recurrence` | 再発 |
| `relapse` | 再燃 |
| `inactive` | 非活動性 |
| `remission` | 寛解 |
| `resolved` | 治癒 |

### 診断コード

| フィールド | 説明 |
|-----------|------|
| `code_system` | "ICD-10" または "SNOMED-CT" |
| `code` | 疾患コード（例: "E11.9"） |
| `display_name` | 表示名（例: "2型糖尿病"） |

### 重症度

| レベル | 説明 |
|--------|------|
| `mild` | 軽度 |
| `moderate` | 中等度 |
| `severe` | 重度 |
| `life-threatening` | 生命を脅かす |

---

## 6. AllergyIntolerance（アレルギー・不耐性）

FHIR R4 AllergyIntoleranceリソースに準拠しています。

### タイプと分類

| フィールド | 値 | 説明 |
|-----------|-----|------|
| `type` | allergy | 免疫反応によるアレルギー |
| `type` | intolerance | 非免疫性の不耐性 |
| `category` | food | 食物 |
| `category` | medication | 薬剤 |
| `category` | environment | 環境（花粉、ダニ等） |
| `category` | biologic | 生物製剤 |

### 反応履歴（reactions JSONB）

```json
[
  {
    "substance": "ペニシリン",
    "manifestation": ["発疹", "呼吸困難"],
    "severity": "severe",
    "onset": "2020-05-15T14:30:00Z",
    "duration": "2時間",
    "exposureRoute": "injection",
    "note": "アナフィラキシーショックに近い状態"
  }
]
```

### 重篤度（Criticality）

| レベル | 説明 |
|--------|------|
| `low` | 低リスク |
| `high` | 高リスク（生命を脅かす可能性） |
| `unable-to-assess` | 評価不能 |

---

## API エンドポイント

### 患者管理

```
POST   /api/v1/patients                              # 患者作成
GET    /api/v1/patients                              # 担当患者一覧
GET    /api/v1/patients/{id}                         # 患者詳細取得
PUT    /api/v1/patients/{id}                         # 患者更新
DELETE /api/v1/patients/{id}                         # 患者削除（論理削除）
POST   /api/v1/patients/{id}/assign                  # スタッフ割り当て
```

### 識別子管理

```
POST   /api/v1/patients/{patient_id}/identifiers     # 識別子追加
GET    /api/v1/patients/{patient_id}/identifiers     # 識別子一覧
GET    /api/v1/patients/{patient_id}/identifiers/{id}?decrypt=true  # 復号取得
PUT    /api/v1/patients/{patient_id}/identifiers/{id}# 識別子更新
DELETE /api/v1/patients/{patient_id}/identifiers/{id}# 識別子削除
```

### 社会的背景

```
POST   /api/v1/patients/{patient_id}/social-profiles # プロファイル作成
GET    /api/v1/patients/{patient_id}/social-profiles # プロファイル一覧
GET    /api/v1/patients/{patient_id}/social-profiles/current  # 現在有効なプロファイル
```

### 保険情報

```
POST   /api/v1/patients/{patient_id}/coverages       # 保険情報追加
GET    /api/v1/patients/{patient_id}/coverages       # 保険情報一覧
GET    /api/v1/patients/{patient_id}/coverages/active # 有効な保険のみ
PUT    /api/v1/patients/{patient_id}/coverages/{id}  # 保険情報更新
POST   /api/v1/patients/{patient_id}/coverages/{id}/verify # 確認処理
```

### 疾患・アレルギー

```
POST   /api/v1/patients/{patient_id}/conditions      # 疾患登録
GET    /api/v1/patients/{patient_id}/conditions      # 疾患一覧
GET    /api/v1/patients/{patient_id}/conditions/active # 現在治療中のみ

POST   /api/v1/patients/{patient_id}/allergies       # アレルギー登録
GET    /api/v1/patients/{patient_id}/allergies       # アレルギー一覧
GET    /api/v1/patients/{patient_id}/allergies/active # 活動性のみ
```

---

## アクセス制御

### スタッフ-患者割り当て

`staff_patient_assignments`テーブルで管理：

```sql
CREATE TABLE staff_patient_assignments (
    assignment_id VARCHAR(36) PRIMARY KEY,
    staff_id VARCHAR(36) NOT NULL,
    patient_id VARCHAR(36) NOT NULL REFERENCES patients(patient_id),
    role VARCHAR(50) NOT NULL,  -- 'primary_physician', 'nurse', 'care_manager'
    assigned_at TIMESTAMPTZ NOT NULL,
    assigned_by VARCHAR(36) NOT NULL,
    valid_from DATE,
    valid_to DATE,
    is_active BOOLEAN DEFAULT true
);
```

### アクセス制御フロー

1. 全APIリクエストで`CheckStaffAccess(staffID, patientID)`を実行
2. `staff_patient_assignments`テーブルで割り当て確認
3. 割り当てがない場合は`403 Forbidden`を返却

---

## 監査ログ

`audit_logs`テーブルで全アクセスを記録：

| フィールド | 説明 |
|-----------|------|
| `event_time` | イベント発生日時 |
| `actor_id` | 操作者ID |
| `action` | view, create, update, delete, decrypt |
| `resource_id` | 対象リソースID |
| `patient_id` | 患者ID |
| `accessed_fields` | アクセスしたフィールド一覧 |
| `success` | 成功/失敗 |
| `ip_address` | クライアントIPアドレス |
| `user_agent` | ユーザーエージェント |

マイナンバー復号時は必ず`action=decrypt`で記録されます。

---

## マイグレーションファイル

| ファイル | 内容 |
|----------|------|
| `001_create_patients_clean.sql` | patients テーブル |
| `002_create_social_profiles.sql` | social_profiles テーブル |
| `003_create_coverages.sql` | patient_coverages テーブル |
| `004_create_medical_conditions.sql` | medical_conditions テーブル |
| `005_create_allergy_intolerances.sql` | allergy_intolerances テーブル |
| `014_create_staff_patient_assignments.sql` | スタッフ-患者割り当て |
| `015_create_patient_identifiers.sql` | patient_identifiers テーブル |
| `016_create_audit_logs.sql` | audit_logs テーブル |
