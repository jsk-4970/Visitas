# 患者マスタ リッチ化・セキュア化 設計書
**在宅医療プラットフォームVisitas - Phase 1 最優先実装**

---

## エグゼクティブサマリー

本文書は、Visitasにおける**患者マスタの完全なリッチ化とセキュア化**を定義する。患者データは在宅医療の全機能の起点であり、以下の3つの設計原則に基づいて再設計する：

1. **Spanner PostgreSQL Interface完全準拠**: インターリーブテーブルを使用せず、クラスター化インデックスとForeign Keyで親子関係を管理
2. **カラムレベル暗号化 + Row-Level Security**: 医療情報の多層防御
3. **FHIR R4概念モデル準拠**: 将来の相互運用性を担保

---

## 目次

1. [修正版テーブル構造の全体像](#1-修正版テーブル構造の全体像)
2. [患者基本情報テーブル (patients)](#2-患者基本情報テーブル-patients)
3. [患者識別子テーブル (patient_identifiers)](#3-患者識別子テーブル-patient_identifiers)
4. [患者社会的背景テーブル (patient_social_profiles)](#4-患者社会的背景テーブル-patient_social_profiles)
5. [保険情報テーブル (patient_coverages)](#5-保険情報テーブル-patient_coverages)
6. [病名・既往歴テーブル (medical_conditions)](#6-病名既往歴テーブル-medical_conditions)
7. [アレルギー情報テーブル (allergy_intolerances)](#7-アレルギー情報テーブル-allergy_intolerances)
8. [セキュリティ実装詳細](#8-セキュリティ実装詳細)
9. [Phase 1実装計画 (Week 1-4)](#9-phase-1実装計画-week-1-4)
10. [検証・テスト戦略](#10-検証テスト戦略)

---

## 1. 修正版テーブル構造の全体像

### 1.1 アーキテクチャ変更点

**変更前 (DATABASE_REQUIREMENTS.md):**
```
患者マスタ (patients) ← インターリーブの起点
  └─ INTERLEAVE IN PARENT による物理的隣接配置
      ├─ patient_identifiers
      ├─ patient_social_profiles
      ├─ patient_coverages
      └─ medical_conditions
```

**変更後 (本設計書):**
```
患者マスタ (patients) ← 論理的な親テーブル
  ├─ FOREIGN KEY + ON DELETE CASCADE
  ├─ クラスター化インデックス (patient_id基準)
  ├─ パーティショニング (作成日による水平分割)
  │
  ├─ patient_identifiers (1:N)
  ├─ patient_social_profiles (1:N, バージョン管理)
  ├─ patient_coverages (1:N, 有効期限管理)
  ├─ medical_conditions (1:N)
  └─ allergy_intolerances (1:N)
```

### 1.2 主要な設計決定

| 設計要素 | 採用技術 | 根拠 |
|---|---|---|
| **主キー型** | `varchar(36)` (UUID v4) | 分散環境での一意性保証、セキュリティ(推測不可能) |
| **JSONB使用** | 頻繁に変更される非定型データのみ | 検索性とスキーマ柔軟性のバランス |
| **Generated Columns** | 検索頻度が高く更新頻度が低いフィールドのみ | 書き込みコストと検索速度のトレードオフ |
| **論理削除** | 全テーブルに`is_deleted`, `deleted_at` | 医療法準拠(5年保存義務)、監査証跡 |
| **タイムゾーン** | `timestamptz` (UTC保存) | グローバル展開時の混乱回避 |
| **暗号化** | CMEK + アプリケーション層AEAD | 多層防御、機微情報の保護 |

---

## 2. 患者基本情報テーブル (patients)

### 2.1 設計思想

**最小限の不変情報のみカラム化:**
- `patient_id`, `birth_date`, `gender`は生涯変わらない（または変更が稀）
- 氏名、連絡先、住所は**変更履歴が必要**なため、JSONB配列で管理

**FHIR Patient Resourceマッピング:**
```
FHIR Patient.identifier → patient_identifiers テーブル
FHIR Patient.name → patients.name_history (JSONB配列)
FHIR Patient.address → patients.addresses (JSONB配列)
FHIR Patient.telecom → patients.contact_points (JSONB配列)
```

### 2.2 DDL定義

```sql
-- =====================================================
-- テーブル: patients (患者基本情報)
-- 説明: 在宅医療の全データの起点となるルートテーブル
-- FHIR: Patient Resource
-- セキュリティレベル: Level 3 (機密)
-- =====================================================

CREATE TABLE patients (
    -- 主キー (UUID v4)
    patient_id varchar(36) NOT NULL,

    -- 基本識別情報 (不変)
    birth_date date NOT NULL,
    gender varchar(20) NOT NULL,  -- FHIR AdministrativeGender: male | female | other | unknown
    blood_type varchar(5),  -- A+ | A- | B+ | B- | O+ | O- | AB+ | AB- | unknown

    -- 氏名 (変更履歴対応のJSONB配列)
    name_history jsonb NOT NULL DEFAULT '[]'::jsonb,

    -- 連絡先 (変更履歴対応のJSONB配列)
    contact_points jsonb DEFAULT '[]'::jsonb,

    -- 住所 (変更履歴対応のJSONB配列)
    addresses jsonb DEFAULT '[]'::jsonb,

    -- 写真 (本人確認用)
    photo_url text,  -- Cloud Storage URI (gs://visitas-photos/...)
    photo_uploaded_at timestamptz,

    -- Generated Columns (検索高速化)
    current_family_name text GENERATED ALWAYS AS (
        COALESCE(
            (name_history->-1->>'family'),
            ''
        )
    ) STORED,

    current_given_name text GENERATED ALWAYS AS (
        COALESCE(
            (name_history->-1->>'given'),
            ''
        )
    ) STORED,

    current_full_name_kana text GENERATED ALWAYS AS (
        COALESCE(
            (name_history->-1->>'kana'),
            ''
        )
    ) STORED,

    current_primary_phone text GENERATED ALWAYS AS (
        COALESCE(
            (contact_points->-1->>'value'),
            ''
        )
    ) STORED,

    current_postal_code varchar(8) GENERATED ALWAYS AS (
        COALESCE(
            (addresses->-1->>'postalCode'),
            ''
        )
    ) STORED,

    -- ステータス管理
    active boolean NOT NULL DEFAULT true,
    inactive_reason varchar(100),  -- "deceased" | "transferred" | "withdrawn_consent"
    inactive_date date,

    -- 論理削除 (物理削除禁止)
    is_deleted boolean NOT NULL DEFAULT false,
    deleted_at timestamptz,
    deleted_by varchar(36),  -- Firebase UID
    deletion_reason text,

    -- 同意管理
    consent_status varchar(20) NOT NULL DEFAULT 'pending',  -- pending | obtained | withdrawn | expired
    consent_obtained_at timestamptz,
    consent_document_url text,  -- 同意書PDF

    -- タイムスタンプ
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by varchar(36) NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by varchar(36),

    -- セキュリティメタデータ
    data_classification varchar(20) NOT NULL DEFAULT 'confidential',  -- Level 3
    encryption_key_version int NOT NULL DEFAULT 1,
    last_accessed_at timestamptz,
    last_accessed_by varchar(36),

    -- 制約
    PRIMARY KEY (patient_id),

    CONSTRAINT chk_gender CHECK (gender IN ('male', 'female', 'other', 'unknown')),
    CONSTRAINT chk_consent_status CHECK (consent_status IN ('pending', 'obtained', 'withdrawn', 'expired')),
    CONSTRAINT chk_birth_date CHECK (birth_date <= CURRENT_DATE),
    CONSTRAINT chk_inactive CHECK (
        (active = true AND inactive_reason IS NULL) OR
        (active = false AND inactive_reason IS NOT NULL)
    )
);

-- =====================================================
-- インデックス定義
-- =====================================================

-- 主検索インデックス: カナ氏名
CREATE INDEX idx_patients_name_kana
ON patients(current_full_name_kana text_pattern_ops)
WHERE is_deleted = false AND active = true;

-- 生年月日検索
CREATE INDEX idx_patients_birth_date
ON patients(birth_date)
WHERE is_deleted = false AND active = true;

-- 電話番号検索
CREATE INDEX idx_patients_phone
ON patients(current_primary_phone)
WHERE is_deleted = false AND active = true;

-- 郵便番号検索 (地域分析用)
CREATE INDEX idx_patients_postal_code
ON patients(current_postal_code)
WHERE is_deleted = false AND active = true;

-- 同意ステータス検索
CREATE INDEX idx_patients_consent_pending
ON patients(consent_status, created_at)
WHERE consent_status = 'pending';

-- アクセス監査用
CREATE INDEX idx_patients_last_accessed
ON patients(last_accessed_at DESC)
WHERE last_accessed_at IS NOT NULL;

-- クラスター化インデックス (範囲検索高速化)
CREATE INDEX idx_patients_clustered
ON patients(patient_id, created_at DESC)
STORING (current_family_name, current_given_name, birth_date, gender, active);

-- =====================================================
-- コメント
-- =====================================================

COMMENT ON TABLE patients IS '患者基本情報マスタ - 在宅医療システムのルートテーブル';
COMMENT ON COLUMN patients.patient_id IS 'UUID v4形式の患者ID (推測不可能性によるセキュリティ確保)';
COMMENT ON COLUMN patients.name_history IS '氏名変更履歴 (JSONB配列、最新が配列末尾)';
COMMENT ON COLUMN patients.consent_status IS '個人情報取扱同意ステータス (医療法・個人情報保護法対応)';
COMMENT ON COLUMN patients.is_deleted IS '論理削除フラグ (医療法24条: 5年保存義務により物理削除禁止)';
```

### 2.3 JSONB構造定義

#### 2.3.1 name_history (氏名履歴)

```json
[
  {
    "use": "official",
    "family": "山田",
    "given": "太郎",
    "kana": "ヤマダ タロウ",
    "validFrom": "1950-04-01",
    "validTo": null,
    "changeReason": null
  },
  {
    "use": "official",
    "family": "佐藤",
    "given": "太郎",
    "kana": "サトウ タロウ",
    "validFrom": "2020-05-01",
    "validTo": null,
    "changeReason": "marriage",
    "previousNameIndex": 0
  }
]
```

**フィールド説明:**
- `use`: FHIR NameUse (official | nickname | maiden)
- `validFrom`: この氏名の有効開始日
- `validTo`: 有効終了日 (現在の氏名はnull)
- `changeReason`: 変更理由 (marriage | divorce | legal_name_change)
- `previousNameIndex`: 前の氏名への参照 (配列インデックス)

#### 2.3.2 contact_points (連絡先履歴)

```json
[
  {
    "system": "phone",
    "value": "090-1234-5678",
    "use": "mobile",
    "rank": 1,
    "validFrom": "2023-01-01",
    "validTo": null,
    "verified": true,
    "verifiedAt": "2023-01-05T10:00:00Z"
  },
  {
    "system": "phone",
    "value": "03-1234-5678",
    "use": "home",
    "rank": 2,
    "validFrom": "2023-01-01",
    "validTo": null,
    "verified": false
  },
  {
    "system": "email",
    "value": "taro.yamada@example.com",
    "use": "home",
    "rank": 3,
    "validFrom": "2023-01-01",
    "validTo": null,
    "verified": true,
    "verifiedAt": "2023-01-05T10:05:00Z"
  }
]
```

**フィールド説明:**
- `system`: FHIR ContactPointSystem (phone | email | fax | sms)
- `use`: FHIR ContactPointUse (mobile | home | work)
- `rank`: 優先順位 (1が最優先)
- `verified`: 連絡先確認済みフラグ
- `verifiedAt`: 確認日時

#### 2.3.3 addresses (住所履歴)

```json
[
  {
    "use": "home",
    "type": "physical",
    "postalCode": "160-0023",
    "prefecture": "東京都",
    "city": "新宿区",
    "line": "西新宿1-2-3",
    "building": "西新宿マンション 501号室",
    "geolocation": {
      "latitude": 35.6895,
      "longitude": 139.6917,
      "geocodedAt": "2023-01-05T10:00:00Z",
      "geocodeSource": "google_maps_api"
    },
    "validFrom": "2023-01-01",
    "validTo": null,
    "isPrimary": true
  },
  {
    "use": "temporary",
    "type": "physical",
    "postalCode": "100-0001",
    "prefecture": "東京都",
    "city": "千代田区",
    "line": "千代田1-1-1",
    "building": "千代田病院 3階 301号室",
    "geolocation": {
      "latitude": 35.6762,
      "longitude": 139.7654,
      "geocodedAt": "2024-06-01T14:00:00Z",
      "geocodeSource": "manual_entry"
    },
    "validFrom": "2024-06-01",
    "validTo": "2024-06-15",
    "isPrimary": false,
    "notes": "入院中の一時住所"
  }
]
```

**フィールド説明:**
- `use`: FHIR AddressUse (home | temporary | billing)
- `type`: FHIR AddressType (physical | postal)
- `geolocation`: Google Maps APIでジオコーディングした座標
- `isPrimary`: 訪問看護のデフォルト住所

---

## 3. 患者識別子テーブル (patient_identifiers)

### 3.1 設計思想

患者は複数の識別子を持つため、FHIR Identifierパターンに従い分離テーブルで管理：
- マイナンバー
- 健康保険証番号
- 介護保険証番号
- 医療機関独自の患者ID (MRN: Medical Record Number)
- 他システムからの移行ID

**セキュリティ要件:**
- マイナンバーは**カラムレベル暗号化** (Cloud KMS AEAD)
- アクセス時は監査ログに記録

### 3.2 DDL定義

```sql
-- =====================================================
-- テーブル: patient_identifiers (患者識別子)
-- 説明: 患者の複数の識別子を管理
-- FHIR: Patient.identifier
-- セキュリティレベル: Level 3-4 (機密〜最高機密)
-- =====================================================

CREATE TABLE patient_identifiers (
    identifier_id varchar(36) NOT NULL,
    patient_id varchar(36) NOT NULL,

    -- 識別子体系
    system varchar(200) NOT NULL,  -- URN形式: "urn:oid:jpn-mynumber", "urn:oid:insurance-card"
    type_code varchar(50) NOT NULL,  -- "mynumber" | "insurance_number" | "care_insurance_number" | "mrn" | "legacy_id"
    type_display varchar(100) NOT NULL,

    -- 識別子の値
    value varchar(200) NOT NULL,

    -- 有効期限
    period_start date,
    period_end date,

    -- 発行者情報
    assigner_organization varchar(200),  -- 保険者名、医療機関名等
    assigner_organization_code varchar(50),

    -- 暗号化管理
    is_encrypted boolean NOT NULL DEFAULT false,
    encryption_key_id varchar(100),  -- Cloud KMS Key ID
    encrypted_value text,  -- 暗号化された値 (マイナンバー等)

    -- ステータス
    status varchar(20) NOT NULL DEFAULT 'active',  -- active | inactive | entered-in-error

    -- タイムスタンプ
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by varchar(36) NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- セキュリティメタデータ
    data_classification varchar(20) NOT NULL DEFAULT 'highly_confidential',  -- Level 4

    -- 制約
    PRIMARY KEY (identifier_id),
    CONSTRAINT fk_patient_identifiers_patient
        FOREIGN KEY (patient_id)
        REFERENCES patients(patient_id)
        ON DELETE CASCADE,

    -- 同じsystemとvalueの組み合わせは一意 (重複登録防止)
    CONSTRAINT uq_identifier_system_value
        UNIQUE (system, value),

    CONSTRAINT chk_identifier_status
        CHECK (status IN ('active', 'inactive', 'entered-in-error')),

    -- 暗号化されている場合はencrypted_valueが必須
    CONSTRAINT chk_encryption
        CHECK (
            (is_encrypted = false AND encrypted_value IS NULL) OR
            (is_encrypted = true AND encrypted_value IS NOT NULL)
        )
);

-- =====================================================
-- インデックス定義
-- =====================================================

-- 患者IDによる検索 (クラスター化)
CREATE INDEX idx_identifiers_patient
ON patient_identifiers(patient_id, type_code);

-- 識別子値による逆引き検索 (患者検索)
CREATE INDEX idx_identifiers_value
ON patient_identifiers(value)
WHERE status = 'active';

-- 識別子タイプによる検索
CREATE INDEX idx_identifiers_type
ON patient_identifiers(type_code, patient_id)
WHERE status = 'active';

-- 有効期限切れ検出用
CREATE INDEX idx_identifiers_expiration
ON patient_identifiers(period_end)
WHERE status = 'active' AND period_end IS NOT NULL AND period_end < CURRENT_DATE + INTERVAL '30 days';

-- =====================================================
-- コメント
-- =====================================================

COMMENT ON TABLE patient_identifiers IS '患者識別子マスタ (マイナンバー、保険証番号等)';
COMMENT ON COLUMN patient_identifiers.system IS 'URN形式の識別子体系 (例: urn:oid:1.2.392.100495.20.3.51.11311234567)';
COMMENT ON COLUMN patient_identifiers.encrypted_value IS 'Cloud KMS AEADで暗号化された識別子値 (マイナンバー等の機微情報)';
```

### 3.3 識別子体系の標準定義

| type_code | system (URN) | 説明 | 暗号化必須 |
|---|---|---|---|
| `mynumber` | `urn:oid:jpn-mynumber` | マイナンバー (12桁) | ✅ Yes |
| `insurance_number` | `urn:oid:jpn-health-insurance` | 健康保険証番号 | ❌ No |
| `care_insurance_number` | `urn:oid:jpn-care-insurance` | 介護保険証番号 | ❌ No |
| `mrn` | `urn:oid:organization-specific` | 医療機関独自患者ID | ❌ No |
| `legacy_id` | `urn:oid:legacy-system` | 旧システムからの移行ID | ❌ No |

---

## 4. 患者社会的背景テーブル (patient_social_profiles)

### 4.1 設計思想

**Subjective (S) ドメイン:**
- 患者の「語り」「生活文脈」を構造化
- 独居状況、介護者情報、経済的背景、宗教的配慮等
- FHIR Observation (Social History) / SDOH (Social Determinants of Health)

**バージョン管理:**
- 社会的状況は時間とともに変化するため、履歴管理が必須
- `profile_version`で変更履歴を追跡

### 4.2 DDL定義

```sql
-- =====================================================
-- テーブル: patient_social_profiles (患者社会的背景)
-- 説明: 患者の生活状況、家族関係、社会的背景
-- FHIR: Observation (Social History)
-- セキュリティレベル: Level 3 (機密)
-- =====================================================

CREATE TABLE patient_social_profiles (
    profile_id varchar(36) NOT NULL,
    patient_id varchar(36) NOT NULL,

    -- バージョン管理
    profile_version int NOT NULL DEFAULT 1,
    status varchar(20) NOT NULL DEFAULT 'active',  -- active | superseded | entered-in-error

    -- 記録メタデータ
    recorded_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    recorded_by varchar(36) NOT NULL,  -- 医師、看護師、MSW等のFirebase UID
    recorded_by_role varchar(50) NOT NULL,  -- "doctor" | "nurse" | "care_manager" | "social_worker"

    -- 社会的背景の本体 (JSONB)
    content jsonb NOT NULL,

    -- Generated Columns (検索高速化)
    living_status text GENERATED ALWAYS AS (
        content->'livingSituation'->>'status'
    ) STORED,

    is_living_alone boolean GENERATED ALWAYS AS (
        content->'livingSituation'->>'status' = 'living_alone'
    ) STORED,

    primary_caregiver_name text GENERATED ALWAYS AS (
        content->'keyPersons'->0->>'name'
    ) STORED,

    has_financial_concerns boolean GENERATED ALWAYS AS (
        content->'financialBackground' ? 'concerns'
    ) STORED,

    caregiver_burden_level varchar(20) GENERATED ALWAYS AS (
        content->'keyPersons'->0->'caregiverBurden'->>'burnoutRisk'
    ) STORED,

    -- タイムスタンプ
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- セキュリティメタデータ
    data_classification varchar(20) NOT NULL DEFAULT 'confidential',
    last_accessed_at timestamptz,
    last_accessed_by varchar(36),

    -- 制約
    PRIMARY KEY (profile_id),
    CONSTRAINT fk_social_profile_patient
        FOREIGN KEY (patient_id)
        REFERENCES patients(patient_id)
        ON DELETE CASCADE,

    CONSTRAINT chk_social_profile_status
        CHECK (status IN ('active', 'superseded', 'entered-in-error')),

    CONSTRAINT chk_social_profile_version
        CHECK (profile_version >= 1)
);

-- =====================================================
-- インデックス定義
-- =====================================================

-- 患者IDによる検索 (最新バージョン取得)
CREATE INDEX idx_social_profiles_patient
ON patient_social_profiles(patient_id, profile_version DESC)
WHERE status = 'active';

-- 独居高齢者の抽出
CREATE INDEX idx_social_profiles_living_alone
ON patient_social_profiles(patient_id, is_living_alone)
WHERE is_living_alone = true AND status = 'active';

-- 経済的不安を抱える患者の抽出
CREATE INDEX idx_social_profiles_financial_concerns
ON patient_social_profiles(patient_id, has_financial_concerns)
WHERE has_financial_concerns = true AND status = 'active';

-- 介護者負担が高い患者の抽出
CREATE INDEX idx_social_profiles_caregiver_burden
ON patient_social_profiles(patient_id, caregiver_burden_level)
WHERE caregiver_burden_level IN ('moderate', 'high', 'severe') AND status = 'active';

-- =====================================================
-- コメント
-- =====================================================

COMMENT ON TABLE patient_social_profiles IS '患者社会的背景マスタ (独居状況、家族関係、経済状況等)';
COMMENT ON COLUMN patient_social_profiles.profile_version IS 'バージョン番号 (1から開始、変更のたびにインクリメント)';
COMMENT ON COLUMN patient_social_profiles.content IS 'JSONB形式の社会的背景データ (livingSituation, keyPersons, financialBackground等)';
```

### 4.3 JSONB構造定義 (content)

**詳細は DATABASE_REQUIREMENTS.md 4.3節を参照。主要構造:**

```json
{
  "livingSituation": {
    "status": "living_alone | living_with_family | facility",
    "environment": {
      "housingType": "detached_house | apartment | group_home",
      "floor": 2,
      "hasElevator": false,
      "entryAccess": {
        "hasAutoLock": true,
        "unlockMethod": "key_box",
        "keyBoxLocation": "ガスメーター内",
        "unlockCode": "**ENCRYPTED**"
      }
    }
  },
  "keyPersons": [
    {
      "id": "kp-001",
      "name": "山田 花子",
      "relationship": "spouse",
      "priority": 1,
      "role": "primary_caregiver",
      "contactPoints": [...],
      "caregiverBurden": {
        "zaritScore": 45,
        "burnoutRisk": "moderate",
        "respiteNeeds": {...}
      }
    }
  ],
  "financialBackground": {
    "economicStatus": "pension_only",
    "monthlyIncome": "年金12万円",
    "concerns": "医療費・介護費の支払いに不安"
  }
}
```

---

## 5. 保険情報テーブル (patient_coverages)

### 5.1 設計思想

**日本の複雑な医療・介護保険制度への対応:**
- 医療保険 (国保、社保、後期高齢者)
- 介護保険 (要介護認定、支給限度額管理)
- 公費負担 (生活保護、難病医療費助成等)

**有効期限管理:**
- 保険証の有効期限切れを自動検出
- 要介護認定の更新時期を通知

### 5.2 DDL定義

```sql
-- =====================================================
-- テーブル: patient_coverages (患者保険情報)
-- 説明: 医療保険、介護保険、公費負担の管理
-- FHIR: Coverage Resource
-- セキュリティレベル: Level 3 (機密)
-- =====================================================

CREATE TABLE patient_coverages (
    coverage_id varchar(36) NOT NULL,
    patient_id varchar(36) NOT NULL,

    -- 保険種別
    insurance_type varchar(50) NOT NULL,  -- "medical" | "long_term_care" | "public_expense"
    insurance_subtype varchar(100),  -- "national_health" | "employee_health" | "latter_stage_elderly"

    -- 保険者情報
    payer_organization varchar(200),
    payer_organization_code varchar(20),  -- 保険者番号

    -- 被保険者情報
    insured_symbol varchar(50),  -- 被保険者記号
    insured_number varchar(50),  -- 被保険者番号

    -- 有効期限
    valid_from date NOT NULL,
    valid_to date,
    status varchar(20) NOT NULL DEFAULT 'active',  -- active | cancelled | draft | entered-in-error

    -- 保険証詳細 (JSONB)
    details jsonb NOT NULL,

    -- Generated Columns (介護保険の要介護度)
    care_level_code int GENERATED ALWAYS AS (
        CASE
            WHEN insurance_type = 'long_term_care'
            THEN (details->'careLevelCertification'->>'levelCode')::int
            ELSE NULL
        END
    ) STORED,

    care_level_display text GENERATED ALWAYS AS (
        CASE
            WHEN insurance_type = 'long_term_care'
            THEN details->'careLevelCertification'->>'levelDisplay'
            ELSE NULL
        END
    ) STORED,

    -- Generated Columns (負担割合)
    copayment_ratio numeric(3,2) GENERATED ALWAYS AS (
        (details->'copayment'->>'ratio')::numeric
    ) STORED,

    -- タイムスタンプ
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by varchar(36) NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by varchar(36),

    -- 保険証画像
    certificate_image_url text,  -- Cloud Storage URI
    certificate_uploaded_at timestamptz,

    -- 制約
    PRIMARY KEY (coverage_id),
    CONSTRAINT fk_coverage_patient
        FOREIGN KEY (patient_id)
        REFERENCES patients(patient_id)
        ON DELETE CASCADE,

    CONSTRAINT chk_coverage_insurance_type
        CHECK (insurance_type IN ('medical', 'long_term_care', 'public_expense')),

    CONSTRAINT chk_coverage_status
        CHECK (status IN ('active', 'cancelled', 'draft', 'entered-in-error')),

    CONSTRAINT chk_coverage_valid_period
        CHECK (valid_to IS NULL OR valid_to >= valid_from)
);

-- =====================================================
-- インデックス定義
-- =====================================================

-- 患者IDによる検索 (有効な保険証)
CREATE INDEX idx_coverages_patient
ON patient_coverages(patient_id, insurance_type, status)
WHERE status = 'active';

-- 要介護度による検索
CREATE INDEX idx_coverages_care_level
ON patient_coverages(care_level_code, patient_id)
WHERE insurance_type = 'long_term_care' AND status = 'active';

-- 有効期限切れ検出 (30日前通知)
CREATE INDEX idx_coverages_expiration
ON patient_coverages(valid_to, patient_id)
WHERE status = 'active'
  AND valid_to IS NOT NULL
  AND valid_to BETWEEN CURRENT_DATE AND CURRENT_DATE + INTERVAL '30 days';

-- =====================================================
-- コメント
-- =====================================================

COMMENT ON TABLE patient_coverages IS '患者保険情報マスタ (医療保険、介護保険、公費)';
COMMENT ON COLUMN patient_coverages.care_level_code IS '要介護度コード (1-5: 要介護1-5、11-12: 要支援1-2)';
```

### 5.3 JSONB構造定義 (details) - 介護保険の例

```json
{
  "insuranceType": "long_term_care",
  "insurerNumber": "123456",
  "insuredSymbol": "AB",
  "insuredNumber": "98765432",
  "certificateUrl": "gs://visitas-documents/patient-xxx/care-insurance-cert.pdf",

  "careLevelCertification": {
    "level": "care_level_3",
    "levelCode": 3,
    "levelDisplay": "要介護3",
    "certifiedDate": "2025-04-01",
    "validFrom": "2025-04-01",
    "validTo": "2027-03-31",
    "certificationStatus": "certified",
    "certificationAuthority": "新宿区",
    "serviceLimits": {
      "maxBenefitUnits": 27048,
      "usedUnitsThisMonth": 18500,
      "remainingUnits": 8548,
      "lastCalculated": "2025-12-01"
    }
  },

  "copayment": {
    "ratio": 0.1,
    "limitAmount": 15000,
    "limitType": "monthly_cap"
  }
}
```

---

## 6. 病名・既往歴テーブル (medical_conditions)

### 6.1 設計思想

**FHIR Condition Resource準拠:**
- 主病名、合併症、既往歴を統一的に管理
- ICD-10コード + 日本の病名マスターコードの併用
- 臨床ステータス (active/inactive/resolved)の管理

### 6.2 DDL定義

```sql
-- =====================================================
-- テーブル: medical_conditions (病名・既往歴)
-- 説明: 患者の診断名、既往歴、合併症
-- FHIR: Condition Resource
-- セキュリティレベル: Level 4 (最高機密)
-- =====================================================

CREATE TABLE medical_conditions (
    condition_id varchar(36) NOT NULL,
    patient_id varchar(36) NOT NULL,

    -- 臨床ステータス
    clinical_status varchar(20) NOT NULL,  -- active | recurrence | relapse | inactive | remission | resolved
    verification_status varchar(20) NOT NULL,  -- confirmed | provisional | differential | refuted

    -- 病名分類
    category varchar(50) NOT NULL,  -- "主病名" | "既往歴" | "合併症" | "疑い病名"
    severity varchar(20),  -- mild | moderate | severe

    -- 病名コード (ICD-10 + 病名マスター)
    code jsonb NOT NULL,

    -- 病名テキスト (検索用)
    display_name text NOT NULL,

    -- 発症日・寛解日
    onset_date date,
    abatement_date date,

    -- 病歴詳細
    notes text,

    -- 記録メタデータ
    recorded_date timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    recorded_by varchar(36) NOT NULL,  -- 医師のFirebase UID

    -- タイムスタンプ
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- セキュリティメタデータ
    data_classification varchar(20) NOT NULL DEFAULT 'highly_confidential',

    -- 制約
    PRIMARY KEY (condition_id),
    CONSTRAINT fk_condition_patient
        FOREIGN KEY (patient_id)
        REFERENCES patients(patient_id)
        ON DELETE CASCADE,

    CONSTRAINT chk_condition_clinical_status
        CHECK (clinical_status IN ('active', 'recurrence', 'relapse', 'inactive', 'remission', 'resolved')),

    CONSTRAINT chk_condition_verification
        CHECK (verification_status IN ('confirmed', 'provisional', 'differential', 'refuted')),

    CONSTRAINT chk_condition_category
        CHECK (category IN ('主病名', '既往歴', '合併症', '疑い病名'))
);

-- =====================================================
-- インデックス定義
-- =====================================================

-- 患者IDによる検索 (アクティブな病名)
CREATE INDEX idx_conditions_patient_active
ON medical_conditions(patient_id, category, clinical_status)
WHERE clinical_status IN ('active', 'recurrence', 'relapse');

-- 主病名検索
CREATE INDEX idx_conditions_primary
ON medical_conditions(patient_id, display_name)
WHERE category = '主病名' AND clinical_status = 'active';

-- 病名テキスト検索 (全文検索)
CREATE INDEX idx_conditions_display_name
ON medical_conditions USING gin(to_tsvector('japanese', display_name));

-- =====================================================
-- コメント
-- =====================================================

COMMENT ON TABLE medical_conditions IS '患者病名・既往歴マスタ (ICD-10準拠)';
COMMENT ON COLUMN medical_conditions.code IS 'JSONB形式の病名コード (ICD-10 + 日本病名マスター)';
```

### 6.3 JSONB構造定義 (code)

```json
{
  "coding": [
    {
      "system": "urn:oid:2.16.840.1.113883.6.3",
      "version": "2023",
      "code": "I50.9",
      "display": "心不全、詳細不明"
    },
    {
      "system": "urn:oid:jpn-disease-master",
      "code": "20059739",
      "display": "慢性心不全"
    }
  ],
  "text": "慢性心不全(NYHA class III)"
}
```

---

## 7. アレルギー情報テーブル (allergy_intolerances)

### 7.1 設計思想

**医療安全の最重要データ:**
- 薬剤アレルギー、食物アレルギー、環境アレルギー
- 重症度 (criticality) による処方システムへのアラート
- 反応内容 (蕁麻疹、アナフィラキシー等)の記録

### 7.2 DDL定義

```sql
-- =====================================================
-- テーブル: allergy_intolerances (アレルギー情報)
-- 説明: 患者のアレルギー・副作用歴
-- FHIR: AllergyIntolerance Resource
-- セキュリティレベル: Level 4 (最高機密)
-- =====================================================

CREATE TABLE allergy_intolerances (
    allergy_id varchar(36) NOT NULL,
    patient_id varchar(36) NOT NULL,

    -- 臨床ステータス
    clinical_status varchar(20) NOT NULL DEFAULT 'active',  -- active | inactive | resolved
    verification_status varchar(20) NOT NULL,  -- confirmed | unconfirmed | refuted

    -- アレルギー分類
    type varchar(20) NOT NULL,  -- allergy | intolerance
    category varchar(20) NOT NULL,  -- food | medication | environment | biologic
    criticality varchar(20) NOT NULL,  -- low | high | unable-to-assess

    -- アレルゲン (薬剤名、食物名等)
    substance jsonb NOT NULL,

    -- 反応内容 (症状、重症度)
    reaction jsonb,

    -- 発症日
    onset_date date,

    -- 記録メタデータ
    recorded_date timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    recorded_by varchar(36) NOT NULL,

    -- タイムスタンプ
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- セキュリティメタデータ
    data_classification varchar(20) NOT NULL DEFAULT 'highly_confidential',

    -- 制約
    PRIMARY KEY (allergy_id),
    CONSTRAINT fk_allergy_patient
        FOREIGN KEY (patient_id)
        REFERENCES patients(patient_id)
        ON DELETE CASCADE,

    CONSTRAINT chk_allergy_clinical_status
        CHECK (clinical_status IN ('active', 'inactive', 'resolved')),

    CONSTRAINT chk_allergy_verification
        CHECK (verification_status IN ('confirmed', 'unconfirmed', 'refuted')),

    CONSTRAINT chk_allergy_type
        CHECK (type IN ('allergy', 'intolerance')),

    CONSTRAINT chk_allergy_category
        CHECK (category IN ('food', 'medication', 'environment', 'biologic')),

    CONSTRAINT chk_allergy_criticality
        CHECK (criticality IN ('low', 'high', 'unable-to-assess'))
);

-- =====================================================
-- インデックス定義
-- =====================================================

-- 患者IDによる検索 (アクティブなアレルギー)
CREATE INDEX idx_allergies_patient_active
ON allergy_intolerances(patient_id, category, clinical_status)
WHERE clinical_status = 'active';

-- 高リスクアレルギーの即時検索 (処方チェック用)
CREATE INDEX idx_allergies_critical
ON allergy_intolerances(patient_id, criticality, category)
WHERE criticality = 'high' AND clinical_status = 'active';

-- =====================================================
-- コメント
-- =====================================================

COMMENT ON TABLE allergy_intolerances IS '患者アレルギー情報マスタ (医療安全の最重要データ)';
COMMENT ON COLUMN allergy_intolerances.criticality IS '重症度 (high: 処方時に強制アラート)';
```

---

## 8. セキュリティ実装詳細

### 8.1 カラムレベル暗号化 (Application-Level Encryption)

**対象フィールド:**
- `patient_identifiers.encrypted_value` (マイナンバー)
- `patient_social_profiles.content` の`unlockCode`

**実装方式: Cloud KMS AEAD (Authenticated Encryption with Associated Data)**

```go
// backend/pkg/encryption/kms_aead.go

package encryption

import (
    "context"
    "encoding/base64"

    kms "cloud.google.com/go/kms/apiv1"
    kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
)

type KMSEncryptor struct {
    client *kms.KeyManagementClient
    keyName string
}

func NewKMSEncryptor(ctx context.Context, projectID, locationID, keyRingID, keyID string) (*KMSEncryptor, error) {
    client, err := kms.NewKeyManagementClient(ctx)
    if err != nil {
        return nil, err
    }

    keyName := fmt.Sprintf(
        "projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s",
        projectID, locationID, keyRingID, keyID,
    )

    return &KMSEncryptor{
        client: client,
        keyName: keyName,
    }, nil
}

// EncryptMyNumber: マイナンバーを暗号化
func (e *KMSEncryptor) EncryptMyNumber(ctx context.Context, plaintext string) (string, error) {
    req := &kmspb.EncryptRequest{
        Name:      e.keyName,
        Plaintext: []byte(plaintext),
        AdditionalAuthenticatedData: []byte("mynumber"), // AADでコンテキストを追加
    }

    result, err := e.client.Encrypt(ctx, req)
    if err != nil {
        return "", fmt.Errorf("encryption failed: %w", err)
    }

    // Base64エンコードしてDB格納
    return base64.StdEncoding.EncodeToString(result.Ciphertext), nil
}

// DecryptMyNumber: マイナンバーを復号
func (e *KMSEncryptor) DecryptMyNumber(ctx context.Context, ciphertext string) (string, error) {
    // Base64デコード
    data, err := base64.StdEncoding.DecodeString(ciphertext)
    if err != nil {
        return "", fmt.Errorf("base64 decode failed: %w", err)
    }

    req := &kmspb.DecryptRequest{
        Name:       e.keyName,
        Ciphertext: data,
        AdditionalAuthenticatedData: []byte("mynumber"),
    }

    result, err := e.client.Decrypt(ctx, req)
    if err != nil {
        return "", fmt.Errorf("decryption failed: %w", err)
    }

    return string(result.Plaintext), nil
}
```

### 8.2 Row-Level Security (RLS) 実装

**要件:**
- 担当患者のみ閲覧可能
- 医師は全患者閲覧可能
- 看護師/ケアマネは担当患者のみ

**実装方式: セキュアビュー + スタッフ割当テーブル**

```sql
-- =====================================================
-- テーブル: staff_patient_assignments (スタッフ-患者割当)
-- =====================================================

CREATE TABLE staff_patient_assignments (
    assignment_id varchar(36) NOT NULL,
    staff_id varchar(36) NOT NULL,  -- Firebase UID
    patient_id varchar(36) NOT NULL,

    role varchar(50) NOT NULL,  -- "doctor" | "nurse" | "care_manager" | "physical_therapist"
    assignment_type varchar(50) NOT NULL,  -- "primary" | "backup" | "temporary"

    valid_from date NOT NULL,
    valid_to date,
    status varchar(20) NOT NULL DEFAULT 'active',

    PRIMARY KEY (assignment_id),
    CONSTRAINT fk_assignment_patient
        FOREIGN KEY (patient_id)
        REFERENCES patients(patient_id)
        ON DELETE CASCADE,

    CONSTRAINT uq_staff_patient_active
        UNIQUE (staff_id, patient_id, status)
);

CREATE INDEX idx_assignments_staff
ON staff_patient_assignments(staff_id, status)
WHERE status = 'active';

CREATE INDEX idx_assignments_patient
ON staff_patient_assignments(patient_id, status)
WHERE status = 'active';

-- =====================================================
-- ビュー: view_my_patients (担当患者のみ閲覧)
-- =====================================================

CREATE VIEW view_my_patients AS
SELECT
    p.patient_id,
    p.current_family_name,
    p.current_given_name,
    p.current_full_name_kana,
    p.birth_date,
    p.gender,
    p.current_primary_phone,
    p.current_postal_code,
    p.active,
    p.consent_status,
    spa.role AS my_role,
    spa.assignment_type
FROM patients p
INNER JOIN staff_patient_assignments spa
    ON p.patient_id = spa.patient_id
WHERE
    spa.staff_id = current_setting('app.current_user_id')
    AND spa.status = 'active'
    AND p.is_deleted = false
    AND p.active = true;

COMMENT ON VIEW view_my_patients IS 'RLS実装: ログインユーザーの担当患者のみ表示';
```

**アプリケーション層での使用:**

```go
// backend/internal/repository/spanner/patient_repository.go

func (r *PatientRepository) GetMyPatients(ctx context.Context, firebaseUID string) ([]Patient, error) {
    // セッション変数にFirebase UIDを設定
    _, err := r.client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
        stmt := spanner.Statement{
            SQL: "SET app.current_user_id = @uid",
            Params: map[string]interface{}{
                "uid": firebaseUID,
            },
        }
        _, err := txn.Update(ctx, stmt)
        return err
    })
    if err != nil {
        return nil, fmt.Errorf("failed to set session variable: %w", err)
    }

    // RLSビュー経由でクエリ
    stmt := spanner.Statement{
        SQL: `SELECT * FROM view_my_patients ORDER BY current_full_name_kana`,
    }

    iter := r.client.Single().Query(ctx, stmt)
    defer iter.Stop()

    // 結果をマッピング...
}
```

### 8.3 監査ログ実装

```sql
-- =====================================================
-- テーブル: audit_patient_access_logs (患者情報アクセス監査)
-- =====================================================

CREATE TABLE audit_patient_access_logs (
    log_id varchar(36) NOT NULL,
    event_time timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- アクター情報
    actor_id varchar(36) NOT NULL,  -- Firebase UID
    actor_role varchar(50) NOT NULL,
    actor_ip_address inet,

    -- アクション
    action varchar(50) NOT NULL,  -- "view" | "create" | "update" | "delete" | "export"
    resource_type varchar(50) NOT NULL,  -- "patient" | "patient_identifier" | "social_profile"
    resource_id varchar(36) NOT NULL,
    patient_id varchar(36) NOT NULL,

    -- 詳細
    accessed_fields jsonb,  -- 閲覧したカラム名
    change_details jsonb,  -- Before/After
    access_purpose varchar(200),  -- "診療" | "請求" | "統計分析"

    -- 結果
    success boolean NOT NULL,
    error_message text,

    PRIMARY KEY (log_id)
);

-- パーティショニング (月単位)
CREATE INDEX idx_audit_logs_time
ON audit_patient_access_logs(event_time DESC, patient_id)
PARTITION BY RANGE (event_time);

CREATE INDEX idx_audit_logs_actor
ON audit_patient_access_logs(actor_id, event_time DESC);

CREATE INDEX idx_audit_logs_patient
ON audit_patient_access_logs(patient_id, event_time DESC);
```

---

## 9. Phase 1実装計画 (Week 1-4)

### Week 1: インフラ構築とコアテーブル実装

#### Day 1-2: GCPセットアップ
- [ ] Spanner PostgreSQL Interfaceインスタンス作成
- [ ] Cloud KMS暗号鍵作成 (CMEK + アプリケーション層暗号化用)
- [ ] Firebase Authenticationプロジェクト設定
- [ ] Cloud Storageバケット作成 (患者写真、保険証画像)

#### Day 3-5: コアテーブルマイグレーション
- [ ] `patients` テーブル作成
- [ ] `patient_identifiers` テーブル作成
- [ ] Generated Columnsの動作検証
- [ ] インデックスパフォーマンステスト

**マイグレーションファイル配置:**
```
backend/migrations/
├── 001_create_patients.sql
├── 002_create_patient_identifiers.sql
├── 003_create_staff_patient_assignments.sql
└── 004_create_audit_logs.sql
```

#### Day 6-7: Go Repository層実装
- [ ] `internal/repository/spanner/patient_repository.go`
- [ ] CRUD操作の実装
- [ ] トランザクション管理
- [ ] ユニットテスト (カバレッジ80%以上)

### Week 2: セキュリティ機能実装

#### Day 8-10: 暗号化機能
- [ ] `pkg/encryption/kms_aead.go` 実装
- [ ] マイナンバー暗号化/復号のテスト
- [ ] 暗号化キーのローテーション戦略確立

#### Day 11-12: RLS (Row-Level Security)
- [ ] `staff_patient_assignments` テーブル作成
- [ ] `view_my_patients` ビュー作成
- [ ] セッション変数によるRLS実装
- [ ] 権限テスト (医師/看護師/ケアマネ)

#### Day 13-14: 監査ログ
- [ ] `audit_patient_access_logs` テーブル作成
- [ ] Go Middlewareでアクセスログ自動記録
- [ ] 監査ログクエリAPI実装 (管理者用)

### Week 3: リッチ化テーブル実装

#### Day 15-17: 社会的背景・保険情報
- [ ] `patient_social_profiles` テーブル作成
- [ ] `patient_coverages` テーブル作成
- [ ] JSONBバリデーション関数実装
- [ ] Generated Columnsの動作確認

#### Day 18-19: 医療情報テーブル
- [ ] `medical_conditions` テーブル作成
- [ ] `allergy_intolerances` テーブル作成
- [ ] ICD-10コードマスタ統合検討

#### Day 20-21: Service層実装
- [ ] `internal/services/patient_service.go`
- [ ] ビジネスロジック実装
- [ ] バリデーション (同意取得確認等)

### Week 4: API実装とテスト

#### Day 22-24: REST API実装
- [ ] `internal/handlers/patients.go`
- [ ] エンドポイント実装:
  - `POST /api/v1/patients` (患者登録)
  - `GET /api/v1/patients/:id` (患者詳細)
  - `PUT /api/v1/patients/:id` (患者更新)
  - `GET /api/v1/patients` (担当患者一覧)
- [ ] OpenAPI仕様書作成

#### Day 25-26: 統合テスト
- [ ] API E2Eテスト
- [ ] セキュリティテスト (RLS、暗号化)
- [ ] パフォーマンステスト (100患者で応答200ms以内)

#### Day 27-28: ドキュメント整備とレビュー
- [ ] API仕様書完成 (`docs/API_SPEC.md`)
- [ ] セキュリティチェックリスト確認
- [ ] コードレビュー
- [ ] Phase 1完了報告

---

## 10. 検証・テスト戦略

### 10.1 機能テスト

| テストケース | 期待結果 |
|---|---|
| **患者登録 (基本情報のみ)** | 201 Created、patient_idが返却される |
| **患者登録 (JSONB履歴あり)** | name_history配列に初期値が格納される |
| **マイナンバー暗号化登録** | patient_identifiers.encrypted_valueが暗号化される |
| **RLS: 担当患者のみ閲覧** | view_my_patients経由で自分の患者のみ取得 |
| **RLS: 他人の患者閲覧** | 403 Forbidden |
| **保険証有効期限切れ検出** | valid_to < 30日後の患者が抽出される |
| **Generated Column検索** | current_full_name_kanaでの部分一致検索が高速 |

### 10.2 セキュリティテスト

| テストケース | 検証方法 |
|---|---|
| **CMEK暗号化** | Spannerコンソールで暗号化設定確認 |
| **マイナンバー復号権限** | 権限のないユーザーでDecrypt API呼び出し → 403 |
| **監査ログ記録** | 患者情報閲覧後、audit_logsにレコードが存在 |
| **SQLインジェクション** | 準備済みステートメント使用、手動テスト |

### 10.3 パフォーマンステスト

**負荷シナリオ:**
```
患者数: 10,000件
同時接続: 50ユーザー
クエリ: 担当患者一覧取得 (10件/ページ)

目標:
- 平均応答時間: <200ms
- 95パーセンタイル: <500ms
- エラー率: <0.1%
```

**ツール:** k6 / Locust

---

## まとめ

本設計書は、Visitasの**患者マスタのリッチ化とセキュア化**を最優先事項として、以下を実現します:

### ✅ 達成事項

1. **Spanner PostgreSQL Interface完全準拠**
   - インターリーブテーブルを使用しない実装可能な設計
   - Foreign Key + クラスター化インデックスによるパフォーマンス最適化

2. **多層防御セキュリティ**
   - CMEK (Cloud KMS) による保存時暗号化
   - アプリケーション層AEADによるマイナンバー暗号化
   - Row-Level Security (RLS) による担当患者制限
   - 監査ログによる全アクセス記録

3. **医療情報のリッチな管理**
   - 氏名・住所・連絡先の変更履歴管理 (JSONB配列)
   - 社会的背景の詳細記録 (独居、介護者負担、経済状況)
   - 介護保険の要介護度・支給限度額管理
   - 病名・アレルギーのFHIR準拠管理

4. **実装可能性の担保**
   - Week 1-4の詳細実装計画
   - マイグレーションSQL、Go Repository実装、API実装の明確な順序
   - 検証・テスト戦略の具体化

### 次のステップ

- [ ] 本設計書のレビューと承認
- [ ] Week 1のGCPインフラ構築開始
- [ ] マイグレーションSQLファイルの作成
- [ ] Go Repository層の実装開始

---

**承認:**

| 役割 | 氏名 | 承認日 | 署名 |
|---|---|---|---|
| プロダクトオーナー | | | |
| テックリード | | | |
| セキュリティ責任者 | | |
