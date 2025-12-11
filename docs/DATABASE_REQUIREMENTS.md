# Visitas データベース要件定義書
**在宅医療特化型AIプラットフォームにおける患者データ管理アーキテクチャ**

---

## エグゼクティブサマリー

本文書は、在宅医療プラットフォーム「Visitas」におけるデータベース設計の完全仕様を定義する。本設計は、**患者情報のリッチな管理**と**多層防御型セキュリティ**の両立を最優先課題とし、以下の3つの設計原則に基づく：

1. **SOAP主導型ハイブリッドアーキテクチャ**: 医療記録の標準フレームワーク（Subjective/Objective/Assessment/Plan）に沿って、リレーショナルデータ管理とJSONB型による柔軟性を戦略的に使い分け
2. **3省2ガイドライン準拠のゼロトラストセキュリティ**: 暗号化、監査ログ、アクセス制御の多層防御による医療情報保護
3. **AI駆動型ワークフロー最適化**: Gemini/Vertex AIおよびGoogle Maps Route Optimization APIとのシームレスな統合

---

## 目次

1. [技術基盤と選定根拠](#1-技術基盤と選定根拠)
2. [データ分類とセキュリティレベル定義](#2-データ分類とセキュリティレベル定義)
3. [コアテーブル設計](#3-コアテーブル設計)
4. [Subjective (S) ドメイン: 社会的背景とナラティブ](#4-subjective-s-ドメイン-社会的背景とナラティブ)
5. [Objective (O) ドメイン: 基本情報・保険・医学的データ](#5-objective-o-ドメイン-基本情報保険医学的データ)
6. [Assessment (A) & Plan (P) ドメイン: 評価・計画・ACP](#6-assessment-a--plan-p-ドメイン-評価計画acp)
7. [訪問ロジスティクスとRoute Optimization統合](#7-訪問ロジスティクスとroute-optimization統合)
8. [Firestore: リアルタイムコラボレーション層](#8-firestore-リアルタイムコラボレーション層)
9. [セキュリティ実装要件](#9-セキュリティ実装要件)
10. [運用・監査・バックアップ戦略](#10-運用監査バックアップ戦略)
11. [パフォーマンス最適化戦略](#11-パフォーマンス最適化戦略)
12. [移行・拡張・災害復旧](#12-移行拡張災害復旧)

---

## 1. 技術基盤と選定根拠

### 1.1 データベース技術スタック

| コンポーネント | 技術 | 採用理由 | 用途 |
|---|---|---|---|
| **メインデータベース** | **Cloud Spanner (PostgreSQL Interface)** | 強整合性(External Consistency)、水平スケーラビリティ、99.999%可用性、JSONB型サポート | 患者マスター、医療記録(SOAP)、保険情報、スケジュール、訪問履歴 |
| **リアルタイムDB** | **Firestore** | リアルタイム同期、オフライン対応、モバイル最適化 | チャット、位置情報共有、通知、在庫ステータス |
| **認証基盤** | **Firebase Authentication + Identity Platform** | 多要素認証、監査ログ、企業向け機能(SAML/OIDC) | 医療従事者/患者/家族の認証・認可 |
| **暗号化管理** | **Cloud KMS (C
MEK)** | 暗号鍵のライフサイクル管理、HSM保護 | データベース暗号化、個人情報マスキング |
| **ストレージ** | **Cloud Storage** | ライフサイクルポリシー、オブジェクトバージョニング | 音声データ、画像(褥瘡写真)、PDF(同意書) |

### 1.2 Spanner PostgreSQLインターフェース採用の技術的根拠

**従来のSpanner Google SQL方言との比較:**

| 評価軸 | PostgreSQL Interface | Google SQL方言 | 判定 |
|---|---|---|---|
| **JSONB型サポート** | ✅ RFC 7159準拠、豊富な演算子(`@>`, `?`, `#>`) | ❌ JSON型のみ(文字列扱い) | **PG** |
| **エコシステム** | 豊富なツール(pgAdmin, Liquibase, TypeORM対応) | 限定的 | **PG** |
| **学習コスト** | 既存PostgreSQLスキル活用可 | 独自学習必要 | **PG** |
| **パフォーマンス** | 同等(Spannerエンジン共通) | 同等 | 引き分け |
| **新機能対応速度** | Googleの開発注力シフト | メンテナンスモード | **PG** |

**Spannerの核心的優位性:**
- **TrueTime API**: GPS/原子時計ベースの外部整合性により、地理的に分散した訪問看護ステーション間でもデータの因果関係を厳密に保証
- **インターリーブテーブル**: 親子関係のデータ(患者→訪問記録→バイタルサイン)を物理的に隣接配置し、JOINコストを最小化
- **ホットスポット自動分散**: 特定患者への集中アクセス(例:救急搬送時)でもパフォーマンス劣化なし

### 1.3 FHIR (HL7 Fast Healthcare Interoperability Resources) との関係

本システムは、FHIR R4を**概念モデル**として参照しつつ、実装上はSpannerの特性に最適化した独自スキーマを採用する。

**FHIR準拠方針:**

```
FHIR Resource → Spanner実装マッピング

Patient → patients テーブル(基本属性) + patient_social_profiles (Extension)
Observation → clinical_observations (バイタル、ADL評価)
MedicationRequest → medication_orders (JSONB内にFHIR DosageInstructionを格納)
CarePlan → care_plans + acp_records (JSONB内にFHIR Goal/Activityマッピング)
Coverage → patient_coverages (日本の保険制度拡張としてJSONB活用)
```

**FHIR Cloud Healthcare API連携(Phase 3想定):**
- データエクスポート時にSpannerスキーマ→FHIR JSON変換レイヤーを実装
- 他医療機関連携時のみCloud Healthcare APIを経由し、FHIR Bundle形式で送受信

---

## 2. データ分類とセキュリティレベル定義

### 2.1 3省2ガイドライン準拠の情報分類

厚生労働省「医療情報システムの安全管理に関するガイドライン 第6.0版」に基づく分類:

| レベル | データ分類 | 具体例 | 暗号化要件 | アクセス制限 |
|---|---|---|---|---|
| **Level 4: 最高機密** | 医療記録本体(診療録、看護記録) | SOAP記録、検査結果、病名、処方 | CMEK暗号化(鍵ローテーション90日) + カラムレベル暗号化 | 医療職種+患者本人のみ。監査ログ必須 |
| **Level 3: 機密** | 個人識別情報(PII) | 氏名、生年月日、住所、保険証番号 | CMEK暗号化 + マスキングビュー提供 | 医療職種+事務職。アクセスログ記録 |
| **Level 2: 限定公開** | ケア計画、スケジュール | 訪問予定、担当スタッフ割当 | CMEK暗号化 | 担当チーム内共有。家族への部分開示可 |
| **Level 1: 内部公開** | 統計データ、匿名化データ | 集計レポート、研究用データセット | 標準暗号化(Google-managed keys) | 研究者、経営陣 |

### 2.2 個人情報保護法・医療法対応

**要件:**
1. **利用目的の明示**: `patients` テーブルに `consent_records` JSONB列を追加し、同意取得履歴を記録
2. **開示請求対応**: 患者本人からの開示請求に対し、ビュー経由で全データをPDF出力可能な設計
3. **第三者提供記録**: `data_sharing_logs` テーブルで他医療機関へのデータ提供を記録(5年保存)

---

## 3. コアテーブル設計

### 3.1 テーブル構成の全体像

```
┌─────────────────────────────────────────────────┐
│         Root Table: patients                    │  ← インターリーブの起点
│  (patient_id, name, birth_date, ...)           │
└─────────────┬───────────────────────────────────┘
              │
              ├─ patient_identifiers (マイナンバー、保険証番号等)
              ├─ patient_social_profiles (Subjective: 社会的背景)
              ├─ patient_coverages (Objective: 保険情報)
              ├─ clinical_observations (Objective: バイタル、ADL)
              ├─ medical_conditions (Objective: 病名、既往歴)
              ├─ allergy_intolerances (Objective: アレルギー)
              ├─ care_plans (Plan: ケア計画)
              ├─ acp_records (Assessment/Plan: ACP)
              ├─ medication_orders (Plan: 処方)
              ├─ visit_schedules (Plan: 訪問スケジュール)
              │   └─ visit_records (訪問実施記録)
              │       └─ soap_notes (SOAP記録)
              │           └─ care_interventions (処置詳細)
              └─ coordination_messages (多職種連携)
```

### 3.2 命名規則とデータ型標準

| 要素 | 規則 | 例 |
|---|---|---|
| **テーブル名** | 複数形スネークケース | `patients`, `visit_schedules` |
| **カラム名** | 単数形スネークケース | `patient_id`, `birth_date` |
| **主キー** | `{entity}_id` 形式のUUID v4 | `patient_id varchar(36)` |
| **タイムスタンプ** | `timestamptz` 型、UTC保存 | `created_at timestamptz DEFAULT CURRENT_TIMESTAMP` |
| **論理削除** | `is_deleted boolean` + `deleted_at timestamptz` | 物理削除禁止(医療法準拠) |
| **JSONB命名** | キャメルケース(JavaScript互換性) | `{"keyPerson": {"name": "..."}}` |

---

## 4. Subjective (S) ドメイン: 社会的背景とナラティブ

### 4.1 設計思想

Subjectiveデータは患者の「語り」や生活文脈を含むため、**構造の変化が最も激しい**領域である。ここでは、FHIR `Observation` (Social History) およびSDOH (Social Determinants of Health) 拡張の概念を採用し、JSONBによる柔軟なスキーマを実現する。

**重要原則:**
- 固定スキーマでは表現困難な「家族関係の機微」「経済的不安」「宗教的配慮」等を構造化
- 検索可能性を担保するため、頻出フィールドはGenerated Columnで実体化

### 4.2 テーブル: `patient_social_profiles`

```sql
CREATE TABLE patient_social_profiles (
    profile_id varchar(36) NOT NULL,
    patient_id varchar(36) NOT NULL,
    recorded_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    author_id varchar(36) NOT NULL, -- 記録者(看護師/MSW)
    profile_version int NOT NULL DEFAULT 1, -- 変更履歴管理
    content jsonb NOT NULL, -- 社会的背景の本体

    -- Generated Columns (検索高速化用)
    living_status text GENERATED ALWAYS AS (content->'livingSituation'->>'status') STORED,
    primary_caregiver_name text GENERATED ALWAYS AS (content->'keyPersons'->0->>'name') STORED,
    has_financial_concerns boolean GENERATED ALWAYS AS (content->'financialBackground' ? 'concerns') STORED,

    -- セキュリティ・監査
    data_sensitivity varchar(20) NOT NULL DEFAULT 'confidential',
    last_accessed_at timestamptz,
    last_accessed_by varchar(36),

    PRIMARY KEY (patient_id, profile_id),
    CONSTRAINT fk_patient FOREIGN KEY (patient_id) REFERENCES patients(patient_id)
) INTERLEAVE IN PARENT patients ON DELETE CASCADE;

-- インデックス: 独居高齢者の抽出等に使用
CREATE INDEX idx_living_status ON patient_social_profiles(living_status);
CREATE INDEX idx_financial_concerns ON patient_social_profiles(has_financial_concerns)
    WHERE has_financial_concerns = true;
```

### 4.3 JSONB構造: `content` カラム詳細仕様

```json
{
  "livingSituation": {
    "status": "living_alone",  // 独居 | living_with_family | facility
    "statusDisplay": "独居",
    "description": "長男は県外在住。週1回電話連絡あり。",
    "lastUpdated": "2025-12-01",
    "environment": {
      "housingType": "detached_house",  // 戸建て | apartment | group_home
      "floor": 2,
      "hasElevator": false,
      "barrierFree": false,
      "stairs": {
        "steps": 15,
        "handrail": "right_side",
        "steepness": "steep"
      },
      "entryAccess": {
        "hasAutoLock": true,
        "unlockMethod": "key_box",
        "keyBoxLocation": "ガスメーター内",
        "unlockCode": "**ENCRYPTED**",  // アプリ層で暗号化した値を格納
        "specialInstructions": "玄関の呼び鈴は聞こえにくい。携帯に直接連絡を推奨"
      },
      "hazards": ["段差あり", "照明暗い", "猫2匹飼育中"],
      "pets": [
        {"type": "cat", "name": "たま", "allergyRisk": false}
      ]
    }
  },

  "keyPersons": [
    {
      "id": "kp-001",
      "name": "山田 花子",
      "relationship": "spouse",  // FHIR RelatedPerson準拠
      "relationshipDisplay": "配偶者",
      "priority": 1,  // 緊急連絡順位
      "role": "primary_caregiver",  // 主介護者 | decision_maker | emergency_contact
      "contactPoints": [
        {
          "system": "phone",
          "value": "090-1234-5678",
          "use": "mobile",
          "rank": 1
        },
        {
          "system": "email",
          "value": "hanako@example.com",
          "use": "home"
        }
      ],
      "address": {
        "use": "home",
        "postalCode": "123-4567",
        "prefecture": "東京都",
        "city": "新宿区",
        "line": "西新宿1-2-3",
        "geolocation": {"lat": 35.6895, "lng": 139.6917}
      },
      "availability": {
        "weekdayDaytime": "available",  // 平日日中対応可
        "weekdayNight": "limited",
        "weekend": "available",
        "notes": "火曜午後はパート勤務で不在"
      },
      "caregiverBurden": {
        "zaritScore": 45,  // Zarit介護負担尺度
        "burnoutRisk": "moderate",
        "lastAssessed": "2025-11-15",
        "respiteNeeds": {
          "desired": true,
          "frequency": "週1回",
          "preferredService": "ショートステイ"
        },
        "healthConcerns": "腰痛あり。整形外科通院中。",
        "sleepQuality": "poor",  // 睡眠障害の有無
        "mentalHealth": {
          "anxiety": "moderate",
          "depression": "mild",
          "supportServices": ["地域包括支援センター相談歴あり"]
        }
      },
      "decisionMakingAuthority": "full",  // 完全な意思決定権限 | partial | none
      "advanceDirectiveProxy": true  // 本人の意思決定代理人か
    }
  ],

  "emergencyContacts": [
    {
      "name": "山田 太郎",
      "relationship": "son",
      "phone": "080-9876-5432",
      "priority": 2,
      "notificationConditions": ["本人連絡不能時", "重大な病状変化"]
    }
  ],

  "socialHistory": {
    "occupation": {
      "current": "retired",
      "past": [
        {"job": "高校教師", "years": "1970-2010"}
      ]
    },
    "lifestyle": {
      "smoking": {
        "status": "former_smoker",  // LOINC 72166-2準拠
        "packsPerDay": 1,
        "yearsSmoked": 40,
        "quitDate": "2010-03-01",
        "quitReason": "医師勧告"
      },
      "alcohol": {
        "status": "never",  // LOINC 74013-4準拠
        "frequency": "none"
      },
      "diet": {
        "restrictions": ["減塩食", "糖質制限"],
        "preferences": ["和食好き", "魚中心"],
        "challenges": "食欲低下傾向"
      },
      "exercise": {
        "frequency": "週3回",
        "type": "近所を散歩",
        "duration": "30分",
        "limitations": "膝痛により長距離は困難"
      }
    },
    "hobbies": ["園芸", "俳句", "孫との電話"],
    "importantValues": "家族に迷惑をかけたくない。自宅で最期を迎えたい。"
  },

  "religiousCulturalNeeds": {
    "religion": "仏教",
    "sect": "浄土真宗",
    "practicesAffectingCare": [
      "朝夕のお勤め時間帯(7:00, 18:00)の訪問は避けてほしい"
    ],
    "endOfLifePreferences": {
      "rituals": "枕経を希望",
      "contactPerson": "菩提寺住職(連絡先: xxx-xxxx)"
    }
  },

  "financialBackground": {
    "economicStatus": "pension_only",
    "monthlyIncome": "年金12万円",
    "publicAssistance": false,
    "concerns": "医療費・介護費の支払いに不安。高額療養費制度の説明を実施済み。",
    "insuranceCoverage": {
      "lifeInsurance": false,
      "supplementalCareInsurance": false
    }
  },

  "cognitiveStatus": {
    "impairmentLevel": "mild",  // なし | mild | moderate | severe
    "mmseScore": 24,
    "lastAssessed": "2025-11-20",
    "decisionMakingCapacity": "retained_with_support",
    "communicationAbility": {
      "verbal": "clear",
      "comprehension": "good_with_repetition",
      "preferredLanguage": "日本語",
      "hearingAid": true,
      "visualAid": "reading_glasses"
    }
  },

  "safetyRisks": {
    "fallRisk": "high",  // Morse Fall Scale
    "wanderingRisk": false,
    "selfNeglect": false,
    "abuseOrNeglect": {
      "suspected": false,
      "assessedBy": "nurse-456",
      "assessmentDate": "2025-10-01"
    },
    "environmentalHazards": ["段差", "照明不足"]
  }
}
```

**ENCRYPTED**フィールドの実装:**
- オートロック解除コード等の機微情報は、Cloud KMSでアプリケーション層暗号化(AEAD)してからJSONB格納
- 復号はアクセス制御を通過した訪問スタッフのみ実行可能
- 監査ログに復号リクエストを記録(Decrypt API Call Logging)

### 4.4 介護者負担評価(Caregiver Burden Assessment)の重要性

**在宅医療における「隠れた患者」としての介護者:**

Visitasは、患者だけでなく**介護者の健康とウェルビーイング**もデータとして記録・評価する。介護者の燃え尽き(Burnout)は、在宅医療の継続可能性を左右する最大のリスク要因である。

**データ活用シナリオ:**
1. **レスパイトケアの自動提案**: Zarit Scoreが48以上(高負担)の介護者に対し、ショートステイ空き情報を自動通知
2. **訪問頻度の調整**: 介護者の健康悪化時は訪問看護頻度を増加
3. **多職種カンファレンス**: 医療ソーシャルワーカー(MSW)へ自動アラート送信

---

## 5. Objective (O) ドメイン: 基本情報・保険・医学的データ

### 5.1 テーブル: `patients` (Root Table)

**設計原則:**
- 最小限の不変情報のみカラム化(patient_id, 生年月日等)
- 変更可能な属性(住所、連絡先)はJSONB化し、履歴管理を実現

```sql
CREATE TABLE patients (
    patient_id varchar(36) NOT NULL,

    -- 基本識別情報(不変)
    birth_date date NOT NULL,
    gender varchar(20) NOT NULL,  -- FHIR AdministrativeGender: male | female | other | unknown
    blood_type varchar(5),  -- A | B | O | AB + Rh

    -- 氏名(変更可能性を考慮してJSONB)
    name jsonb NOT NULL,  -- {"family": "山田", "given": "太郎", "kana": "ヤマダ タロウ"}

    -- 連絡先・住所(履歴管理のためJSONB配列)
    contact_points jsonb,  -- [{"system": "phone", "value": "...", "period": {"start": "2025-01-01"}}]
    addresses jsonb,  -- 複数住所対応(自宅、施設、帰省先等)

    -- 写真(本人確認用)
    photo_url text,  -- Cloud Storage URI

    -- システム管理
    active boolean NOT NULL DEFAULT true,
    is_deleted boolean NOT NULL DEFAULT false,
    deleted_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- 同意管理
    consent_records jsonb,  -- 個人情報取扱同意、研究利用同意等

    -- セキュリティ
    data_classification varchar(20) NOT NULL DEFAULT 'confidential',
    encryption_key_version int NOT NULL DEFAULT 1,

    PRIMARY KEY (patient_id)
);

-- Generated Columns for 検索
ALTER TABLE patients
ADD COLUMN full_name text GENERATED ALWAYS AS (
    name->>'family' || ' ' || name->>'given'
) STORED;

ALTER TABLE patients
ADD COLUMN full_name_kana text GENERATED ALWAYS AS (name->>'kana') STORED;

ALTER TABLE patients
ADD COLUMN primary_phone text GENERATED ALWAYS AS (
    contact_points->0->>'value'
) STORED;

-- Indexes
CREATE INDEX idx_patient_name_kana ON patients(full_name_kana);
CREATE INDEX idx_patient_birth_date ON patients(birth_date);
CREATE INDEX idx_patient_active ON patients(active) WHERE active = true;
```

**`name` JSONB構造:**
```json
{
  "use": "official",
  "family": "山田",
  "given": "太郎",
  "kana": "ヤマダ タロウ",
  "previousNames": [
    {"family": "佐藤", "given": "太郎", "period": {"end": "2020-05-01"}, "changeReason": "婚姻"}
  ]
}
```

### 5.2 テーブル: `patient_identifiers` (複数ID管理)

患者は複数の識別子(マイナンバー、保険証番号、患者ID等)を持つため、FHIRの`Identifier`パターンに従い分離テーブルで管理。

```sql
CREATE TABLE patient_identifiers (
    identifier_id varchar(36) NOT NULL,
    patient_id varchar(36) NOT NULL,

    system varchar(100) NOT NULL,  -- 識別子体系: "urn:oid:jpn-mynumber", "urn:oid:insurance-number"
    value varchar(100) NOT NULL,
    type varchar(50) NOT NULL,  -- "mynumber" | "insurance_number" | "hospital_mrn"

    period_start date,
    period_end date,

    -- セキュリティ: マイナンバーは追加暗号化
    is_encrypted boolean NOT NULL DEFAULT false,

    PRIMARY KEY (patient_id, identifier_id),
    CONSTRAINT fk_patient FOREIGN KEY (patient_id) REFERENCES patients(patient_id),
    CONSTRAINT unique_identifier UNIQUE (system, value)
) INTERLEAVE IN PARENT patients ON DELETE CASCADE;

-- 検索インデックス
CREATE INDEX idx_identifier_value ON patient_identifiers(value);
```

### 5.3 テーブル: `patient_coverages` (保険情報)

**日本の医療・介護保険制度の複雑性への対応:**

```sql
CREATE TABLE patient_coverages (
    coverage_id varchar(36) NOT NULL,
    patient_id varchar(36) NOT NULL,

    insurance_type varchar(50) NOT NULL,  -- "medical" | "long_term_care" | "public_expense"
    payer_organization varchar(200),  -- 保険者名称

    valid_from date NOT NULL,
    valid_to date,
    status varchar(20) NOT NULL,  -- "active" | "cancelled" | "draft" | "entered-in-error"

    details jsonb NOT NULL,  -- 保険証詳細、負担割合、限度額等

    -- Generated Columns
    care_level int GENERATED ALWAYS AS (
        CASE WHEN insurance_type = 'long_term_care'
        THEN (details->'careLevelCertification'->>'levelCode')::int
        ELSE NULL END
    ) STORED,

    copayment_ratio numeric(3,2) GENERATED ALWAYS AS (
        (details->'copayment'->>'ratio')::numeric
    ) STORED,

    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (patient_id, coverage_id),
    CONSTRAINT fk_patient FOREIGN KEY (patient_id) REFERENCES patients(patient_id)
) INTERLEAVE IN PARENT patients ON DELETE CASCADE;

-- 有効期限切れ保険証の定期検出用
CREATE INDEX idx_coverage_expiration ON patient_coverages(valid_to)
    WHERE status = 'active' AND valid_to IS NOT NULL;
```

**`details` JSONB構造(介護保険の例):**

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
    },

    "previousLevels": [
      {
        "level": "care_level_2",
        "period": {"start": "2023-04-01", "end": "2025-03-31"}
      }
    ]
  },

  "copayment": {
    "ratio": 0.1,
    "limitAmount": 15000,
    "limitType": "monthly_cap",
    "reducedCopayment": false
  },

  "serviceRestrictions": {
    "hasRestriction": false,
    "restrictionPeriod": null,
    "reason": null
  },

  "publicExpenseSupplements": [
    {
      "programName": "高額介護サービス費",
      "recipientNumber": "xxx-yyy",
      "validFrom": "2025-01-01"
    }
  ]
}
```

### 5.4 テーブル: `medical_conditions` (病名・既往歴)

```sql
CREATE TABLE medical_conditions (
    condition_id varchar(36) NOT NULL,
    patient_id varchar(36) NOT NULL,

    clinical_status varchar(20) NOT NULL,  -- "active" | "recurrence" | "relapse" | "inactive" | "remission" | "resolved"
    verification_status varchar(20) NOT NULL,  -- "confirmed" | "provisional" | "differential" | "refuted"

    category varchar(50),  -- "主病名" | "既往歴" | "合併症"
    severity varchar(20),  -- "mild" | "moderate" | "severe"

    code jsonb NOT NULL,  -- ICD-10コード、病名マスターコード
    onset_date date,
    abatement_date date,

    notes text,  -- 病歴詳細

    recorded_date timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    recorded_by varchar(36),  -- 医師ID

    PRIMARY KEY (patient_id, condition_id),
    CONSTRAINT fk_patient FOREIGN KEY (patient_id) REFERENCES patients(patient_id)
) INTERLEAVE IN PARENT patients ON DELETE CASCADE;

-- 主病名検索
CREATE INDEX idx_condition_primary ON medical_conditions(patient_id, category)
    WHERE category = '主病名' AND clinical_status = 'active';
```

**`code` JSONB(ICD-10 + 病名マスター):**

```json
{
  "coding": [
    {
      "system": "urn:oid:2.16.840.1.113883.6.3",
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

### 5.5 テーブル: `allergy_intolerances` (アレルギー・副作用歴)

**医療安全の最重要データ:**

```sql
CREATE TABLE allergy_intolerances (
    allergy_id varchar(36) NOT NULL,
    patient_id varchar(36) NOT NULL,

    clinical_status varchar(20) NOT NULL DEFAULT 'active',  -- "active" | "inactive" | "resolved"
    verification_status varchar(20) NOT NULL,  -- "confirmed" | "unconfirmed" | "refuted"

    type varchar(20) NOT NULL,  -- "allergy" | "intolerance"
    category varchar(20) NOT NULL,  -- "food" | "medication" | "environment" | "biologic"
    criticality varchar(20),  -- "low" | "high" | "unable-to-assess"

    substance jsonb NOT NULL,  -- アレルゲン(薬剤名、食物名等)
    reaction jsonb,  -- 反応内容(症状、重症度)

    onset_date date,
    recorded_date timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    recorded_by varchar(36),

    -- セキュリティ
    alert_level varchar(20) NOT NULL DEFAULT 'warning',  -- 処方システムへのアラート強度

    PRIMARY KEY (patient_id, allergy_id),
    CONSTRAINT fk_patient FOREIGN KEY (patient_id) REFERENCES patients(patient_id)
) INTERLEAVE IN PARENT patients ON DELETE CASCADE;

-- 高リスクアレルギーの即時検索
CREATE INDEX idx_allergy_critical ON allergy_intolerances(patient_id)
    WHERE criticality = 'high' AND clinical_status = 'active';
```

**`substance` JSONB:**

```json
{
  "coding": [
    {
      "system": "urn:oid:jpn-yj-code",
      "code": "2171022F1",
      "display": "アモキシシリン"
    }
  ],
  "text": "ペニシリン系抗生物質"
}
```

**`reaction` JSONB:**

```json
{
  "manifestations": [
    {
      "code": {"text": "全身蕁麻疹"},
      "severity": "moderate"
    },
    {
      "code": {"text": "呼吸困難"},
      "severity": "severe"
    }
  ],
  "onset": "30分以内",
  "exposureRoute": "oral",
  "note": "2020年5月、歯科治療後に発症。エピペン処方歴あり。"
}
```

### 5.6 テーブル: `clinical_observations` (バイタルサイン・ADL評価)

**時系列データの効率的管理:**

```sql
CREATE TABLE clinical_observations (
    observation_id varchar(36) NOT NULL,
    patient_id varchar(36) NOT NULL,

    category varchar(50) NOT NULL,  -- "vital_signs" | "adl_assessment" | "cognitive_assessment" | "pain_scale"
    code jsonb NOT NULL,  -- LOINC/SNOMED CT準拠

    effective_datetime timestamptz NOT NULL,  -- 測定日時
    issued timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,  -- 記録日時

    value jsonb NOT NULL,  -- 測定値(数値、コード値、スコア等)
    interpretation varchar(20),  -- "normal" | "high" | "low" | "critical"

    performer_id varchar(36),  -- 測定者
    device_id varchar(36),  -- 測定機器ID(IoT連携想定)

    -- 参照元訪問記録
    visit_record_id varchar(36),

    PRIMARY KEY (patient_id, observation_id),
    CONSTRAINT fk_patient FOREIGN KEY (patient_id) REFERENCES patients(patient_id)
) INTERLEAVE IN PARENT patients ON DELETE CASCADE;

-- 時系列クエリ最適化
CREATE INDEX idx_observation_datetime ON clinical_observations(patient_id, effective_datetime DESC);
CREATE INDEX idx_observation_category ON clinical_observations(patient_id, category, effective_datetime DESC);
```

**バイタルサインのJSONB例:**

```json
{
  "code": {
    "coding": [{"system": "http://loinc.org", "code": "85354-9", "display": "Blood pressure panel"}]
  },
  "value": {
    "systolic": {"value": 135, "unit": "mmHg"},
    "diastolic": {"value": 82, "unit": "mmHg"},
    "heartRate": {"value": 78, "unit": "bpm"},
    "bodyTemperature": {"value": 36.5, "unit": "Cel"},
    "spo2": {"value": 96, "unit": "%"},
    "respiratoryRate": {"value": 18, "unit": "/min"}
  },
  "bodyPosition": "sitting",  // 測定体位
  "notes": "左上腕で測定。不整脈なし。"
}
```

**ADL評価(Barthel Index)のJSONB例:**

```json
{
  "code": {
    "coding": [{"system": "http://loinc.org", "code": "89210-8", "display": "Barthel Index"}]
  },
  "value": {
    "method": "Barthel Index",
    "totalScore": 45,
    "maxScore": 100,
    "components": {
      "feeding": {"score": 5, "maxScore": 10, "description": "部分介助"},
      "transfer": {"score": 5, "maxScore": 15, "description": "中等度介助"},
      "grooming": {"score": 0, "maxScore": 5, "description": "全介助"},
      "toiletUse": {"score": 5, "maxScore": 10, "description": "部分介助"},
      "bathing": {"score": 0, "maxScore": 5, "description": "全介助"},
      "mobility": {"score": 5, "maxScore": 15, "description": "車椅子自走可"},
      "stairs": {"score": 0, "maxScore": 10, "description": "不能"},
      "dressing": {"score": 5, "maxScore": 10, "description": "部分介助"},
      "bowels": {"score": 10, "maxScore": 10, "description": "自立"},
      "bladder": {"score": 5, "maxScore": 10, "description": "部分介助"}
    }
  },
  "assessorComment": "右麻痺により移乗動作に介助を要する。リハビリにより改善傾向。",
  "previousScore": {"totalScore": 40, "assessedDate": "2025-11-01"},
  "trend": "improving"
}
```

---

## 6. Assessment (A) & Plan (P) ドメイン: 評価・計画・ACP

### 6.1 テーブル: `care_plans` (ケア計画)

```sql
CREATE TABLE care_plans (
    plan_id varchar(36) NOT NULL,
    patient_id varchar(36) NOT NULL,

    status varchar(20) NOT NULL,  -- "draft" | "active" | "on-hold" | "revoked" | "completed"
    intent varchar(20) NOT NULL,  -- "proposal" | "plan" | "order"

    title varchar(200) NOT NULL,
    description text,

    period_start date NOT NULL,
    period_end date,

    goals jsonb,  -- 目標(SMART形式)
    activities jsonb,  -- 活動計画

    created_by varchar(36) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (patient_id, plan_id),
    CONSTRAINT fk_patient FOREIGN KEY (patient_id) REFERENCES patients(patient_id)
) INTERLEAVE IN PARENT patients ON DELETE CASCADE;

-- アクティブなケア計画の取得
CREATE INDEX idx_careplan_active ON care_plans(patient_id, status)
    WHERE status = 'active';
```

**`goals` JSONB(SMART目標の例):**

```json
{
  "goals": [
    {
      "id": "goal-001",
      "description": "3ヶ月以内に自力での車椅子移乗を可能にする",
      "category": "rehabilitation",
      "priority": "high",
      "startDate": "2025-12-01",
      "targetDate": "2026-03-01",
      "achievementStatus": "in_progress",
      "measurableOutcome": {
        "metric": "Barthel Index - Transfer",
        "baseline": 5,
        "target": 10,
        "current": 7,
        "lastMeasured": "2025-12-10"
      },
      "addresses": ["condition-id-123"]  // 対象病名への参照
    }
  ]
}
```

**`activities` JSONB:**

```json
{
  "activities": [
    {
      "id": "activity-001",
      "category": "physiotherapy",
      "description": "理学療法士による移乗訓練",
      "frequency": "週2回",
      "duration": "40分",
      "performerType": "physical_therapist",
      "status": "in_progress",
      "scheduledVisits": ["schedule-id-xxx", "schedule-id-yyy"]
    },
    {
      "id": "activity-002",
      "category": "medication",
      "description": "疼痛管理による訓練促進",
      "medicationRef": "med-order-789"
    }
  ]
}
```

### 6.2 テーブル: `acp_records` (Advance Care Planning)

**人生の最終段階における意思決定支援の記録:**

```sql
CREATE TABLE acp_records (
    acp_id varchar(36) NOT NULL,
    patient_id varchar(36) NOT NULL,

    recorded_date date NOT NULL,
    version int NOT NULL DEFAULT 1,  -- 意思変更の履歴管理
    status varchar(20) NOT NULL,  -- "draft" | "active" | "superseded"

    -- 意思決定の主体
    decision_maker varchar(20) NOT NULL,  -- "patient" | "proxy" | "guardian"
    proxy_person_id varchar(36),  -- key_personsへの参照

    -- ACP内容
    directives jsonb NOT NULL,  -- 具体的指示(DNAR, 人工呼吸器等)
    values_narrative text,  -- 価値観の記述

    -- 法的文書
    legal_documents jsonb,  -- 同意書、リビングウィル等のリンク

    -- ACPプロセス
    discussion_log jsonb,  -- 話し合いの履歴

    -- セキュリティ
    data_sensitivity varchar(20) NOT NULL DEFAULT 'highly_confidential',
    access_restricted_to jsonb,  -- アクセス許可リスト

    created_by varchar(36) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (patient_id, acp_id),
    CONSTRAINT fk_patient FOREIGN KEY (patient_id) REFERENCES patients(patient_id)
) INTERLEAVE IN PARENT patients ON DELETE CASCADE;

-- 最新有効なACPの取得
CREATE INDEX idx_acp_active ON acp_records(patient_id, version DESC)
    WHERE status = 'active';
```

**`directives` JSONB詳細:**

```json
{
  "cpr": {
    "decision": "do_not_attempt",
    "decisionDisplay": "心肺蘇生を希望しない(DNAR)",
    "decisionDate": "2025-10-15",
    "circumstances": "いかなる状況でも",
    "rationale": "自然な最期を迎えたい"
  },

  "artificialRespiration": {
    "decision": "refuse",
    "decisionDisplay": "人工呼吸器装着を拒否",
    "circumstances": "回復見込みがない場合",
    "trialPeriod": null
  },

  "artificialNutrition": {
    "peg": {
      "decision": "refuse",
      "decisionDisplay": "胃ろう造設を希望しない"
    },
    "ivFluids": {
      "decision": "short_term_only",
      "decisionDisplay": "点滴は短期間のみ許容",
      "maxDuration": "2週間",
      "purpose": "苦痛緩和目的のみ"
    },
    "nasogastricTube": {
      "decision": "refuse"
    }
  },

  "organDonation": {
    "decision": "not_registered",
    "futureConsideration": "検討中"
  },

  "preferredPlaceOfDeath": {
    "first": "home",
    "second": "hospice",
    "third": "hospital",
    "specificRequests": "自宅の和室で家族と過ごしたい"
  },

  "painManagement": {
    "priority": "symptom_relief_over_longevity",
    "acceptableSedation": "deep_sedation_if_needed",
    "notes": "苦痛の除去を最優先。意識レベルの低下は許容。"
  },

  "decisionMakerProxy": {
    "primary": {
      "keyPersonId": "kp-001",
      "name": "山田 花子",
      "relationship": "spouse",
      "authority": "full"
    },
    "alternate": {
      "keyPersonId": "kp-002",
      "name": "山田 太郎",
      "relationship": "son"
    }
  }
}
```

**`discussion_log` JSONB:**

```json
{
  "discussions": [
    {
      "date": "2025-10-15",
      "participants": [
        {"role": "patient", "name": "山田 三郎"},
        {"role": "family", "name": "山田 花子", "relationship": "spouse"},
        {"role": "physician", "id": "doctor-456", "name": "田中医師"},
        {"role": "nurse", "id": "nurse-789", "name": "佐藤看護師"}
      ],
      "location": "patient_home",
      "duration_minutes": 60,
      "topics_discussed": [
        "DNAR",
        "人工栄養の選択",
        "療養場所の希望"
      ],
      "patient_understanding": "good",
      "patient_emotional_state": "calm_and_accepting",
      "family_agreement": "unanimous",
      "notes": "本人の意思は明確。家族も本人の希望を尊重することで合意。次回3ヶ月後に再確認予定。",
      "next_review_date": "2026-01-15"
    }
  ]
}
```

**`legal_documents` JSONB:**

```json
{
  "documents": [
    {
      "type": "living_will",
      "title": "リビングウィル",
      "signed_date": "2025-10-15",
      "file_url": "gs://visitas-legal-docs/patient-xxx/living-will-signed.pdf",
      "file_hash": "sha256:abcdef123456...",
      "witnesses": [
        {"name": "田中医師", "id": "doctor-456"},
        {"name": "佐藤看護師", "id": "nurse-789"}
      ]
    },
    {
      "type": "power_of_attorney",
      "title": "医療代理権委任状",
      "proxy_name": "山田 花子",
      "signed_date": "2025-10-15",
      "file_url": "gs://visitas-legal-docs/patient-xxx/proxy-authorization.pdf"
    }
  ]
}
```

### 6.3 テーブル: `medication_orders` (処方オーダー)

```sql
CREATE TABLE medication_orders (
    order_id varchar(36) NOT NULL,
    patient_id varchar(36) NOT NULL,

    status varchar(20) NOT NULL,  -- "active" | "on-hold" | "cancelled" | "completed" | "entered-in-error"
    intent varchar(20) NOT NULL,  -- "order" | "plan"

    medication jsonb NOT NULL,  -- 薬剤情報(YJコード、一般名、商品名)
    dosage_instruction jsonb NOT NULL,  -- 用法用量(FHIR DosageInstruction準拠)

    prescribed_date date NOT NULL,
    prescribed_by varchar(36) NOT NULL,  -- 医師ID

    dispense_pharmacy jsonb,  -- 調剤薬局情報

    reason_reference varchar(36),  -- 処方理由(病名IDへの参照)

    PRIMARY KEY (patient_id, order_id),
    CONSTRAINT fk_patient FOREIGN KEY (patient_id) REFERENCES patients(patient_id)
) INTERLEAVE IN PARENT patients ON DELETE CASCADE;

-- 有効処方の検索
CREATE INDEX idx_medication_active ON medication_orders(patient_id, status)
    WHERE status = 'active';
```

**`medication` JSONB:**

```json
{
  "coding": [
    {
      "system": "urn:oid:jpn-yj-code",
      "code": "2171022F1024",
      "display": "アモキシシリンカプセル250mg"
    }
  ],
  "genericName": "アモキシシリン",
  "brandName": "サワシリン",
  "form": "capsule",
  "strength": "250mg"
}
```

**`dosage_instruction` JSONB(FHIR準拠):**

```json
{
  "text": "1回1カプセル(250mg) 1日3回 朝昼夕食後 7日間",
  "timing": {
    "repeat": {
      "frequency": 3,
      "period": 1,
      "periodUnit": "d",
      "when": ["ACM", "ACD", "ACV"]  // 朝食後、昼食後、夕食後
    }
  },
  "route": {
    "coding": [{"code": "PO", "display": "経口"}]
  },
  "doseAndRate": [{
    "doseQuantity": {"value": 1, "unit": "capsule"}
  }],
  "duration": {"value": 7, "unit": "d"},
  "totalDailyDose": {"value": 750, "unit": "mg"},
  "additionalInstruction": "空腹時服用は避ける",
  "patientInstruction": "必ず7日間飲み切ってください"
}
```

---

## 7. 訪問ロジスティクスとRoute Optimization統合

### 7.1 設計思想: データベース駆動型ルート最適化

Google Maps Route Optimization APIは、訪問看護における「移動の無駄」を数学的に最小化する強力なツールだが、その真価を発揮するには**データベースがAPI仕様に合わせた構造を持つ**必要がある。

**統合アーキテクチャ:**

```
Spanner DB (Source of Truth)
  ↓
[visit_schedules + logistics_locations]
  ↓ データ変換レイヤー
ShipmentModel JSON生成
  ↓
Route Optimization API Request
  ↓ 最適化結果
visit_schedules.result_metrics 更新(到着予定時刻、走行距離等)
  ↓
モバイルアプリへPush通知
```

### 7.2 テーブル: `visit_schedules` (訪問スケジュール)

```sql
CREATE TABLE visit_schedules (
    schedule_id varchar(36) NOT NULL,
    patient_id varchar(36) NOT NULL,

    visit_date date NOT NULL,
    visit_type varchar(50) NOT NULL,  -- "regular" | "emergency" | "initial_assessment" | "terminal_care"

    -- 訪問時間枠
    time_window_start time,
    time_window_end time,
    estimated_duration_minutes int NOT NULL,  -- サービス提供時間

    -- スタッフ割当
    assigned_staff_id varchar(36),
    assigned_vehicle_id varchar(36),

    -- ステータス
    status varchar(20) NOT NULL,  -- "draft" | "optimized" | "assigned" | "in_progress" | "completed" | "cancelled"

    -- Route Optimization統合
    priority_score int NOT NULL DEFAULT 5,  -- 1(低)-10(高)
    constraints jsonb,  -- API ShipmentオブジェクトのVisitRequest相当
    optimization_result jsonb,  -- API Response格納

    -- ケア計画との紐付け
    care_plan_ref varchar(36),
    activity_ref varchar(36),

    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (patient_id, schedule_id),
    CONSTRAINT fk_patient FOREIGN KEY (patient_id) REFERENCES patients(patient_id)
) INTERLEAVE IN PARENT patients ON DELETE CASCADE;

-- 日付・ステータス別検索
CREATE INDEX idx_schedule_date_status ON visit_schedules(visit_date, status);
CREATE INDEX idx_schedule_staff ON visit_schedules(assigned_staff_id, visit_date);
```

**`constraints` JSONB(Route Optimization API `VisitRequest`マッピング):**

```json
{
  "arrivalLocation": {
    "latitude": 35.6895,
    "longitude": 139.6917
  },
  "duration": "45m",  // ISO 8601 duration

  "timeWindows": [
    {
      "startTime": "2025-12-12T10:00:00+09:00",
      "endTime": "2025-12-12T12:00:00+09:00"
    }
  ],

  "loadDemands": {
    "nursing_skill_level": {"amount": "3"},  // 正看護師レベル必要
    "physical_load": {"amount": "2"},  // 入浴介助等の身体負荷
    "equipment_weight_kg": {"amount": "5"}
  },

  "visitTypes": ["pickup_and_delivery"],  // 薬剤受取+訪問のケース

  "label": "患者: 山田三郎様 / 訪問看護(入浴介助含む)",

  "tags": ["requires_parking_permit", "wheelchair_access_needed"],

  "cost": 100,  // ベースコスト
  "penaltyCost": 1000  // 未訪問時のペナルティ(高優先度患者)
}
```

**`optimization_result` JSONB(API Responseのメトリクス格納):**

```json
{
  "optimized_at": "2025-12-11T20:00:00Z",
  "route_id": "route-abc-123",
  "vehicle_index": 2,
  "visit_index": 3,

  "timing": {
    "arrival_time": "2025-12-12T10:25:00+09:00",
    "departure_time": "2025-12-12T11:10:00+09:00",
    "waiting_time_minutes": 5,
    "service_time_minutes": 45
  },

  "travel_from_previous": {
    "distance_meters": 3200,
    "duration_seconds": 480,
    "mode": "car"
  },

  "metrics": {
    "total_route_distance_km": 45.3,
    "total_route_duration_hours": 6.5,
    "visits_completed": 8,
    "optimization_score": 92
  }
}
```

### 7.3 テーブル: `logistics_locations` (訪問先物理情報)

患者宅だけでなく、ステーション、薬局、駐車場、検査機関などロジスティクスに関わる全ての場所を管理。

```sql
CREATE TABLE logistics_locations (
    location_id varchar(36) NOT NULL,
    location_type varchar(50) NOT NULL,  -- "patient_home" | "station" | "pharmacy" | "parking" | "lab"

    -- 地理座標
    geolocation geography(point) NOT NULL,  -- PostGISスタイルの地理型
    latitude numeric(10, 7) NOT NULL,
    longitude numeric(10, 7) NOT NULL,

    -- 住所
    address jsonb NOT NULL,

    -- 関連エンティティ
    related_entity_type varchar(50),  -- "patient" | "organization"
    related_entity_id varchar(36),

    -- ロジスティクス詳細
    logistics_meta jsonb NOT NULL,  -- 駐車情報、入館方法、TransitionAttributes等

    active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (location_id)
);

-- 地理検索インデックス
CREATE INDEX idx_location_geo ON logistics_locations USING gist(geolocation);
CREATE INDEX idx_location_type ON logistics_locations(location_type);
```

**`logistics_meta` JSONB(患者宅の例):**

```json
{
  "entryInstructions": {
    "hasAutoLock": true,
    "unlockMethod": "key_box",
    "keyBoxLocation": "ガスメーター内",
    "unlockCode": "**ENCRYPTED**",
    "specialNotes": "玄関呼び鈴は聞こえにくい。到着5分前に携帯へ連絡推奨。",
    "emergencyKeyHolder": {"name": "山田花子", "phone": "090-xxxx-xxxx"}
  },

  "parkingInstructions": {
    "hasPrivateParking": false,
    "nearbyParkingSpots": [
      {
        "id": "parking-001",
        "name": "タイムズ新宿西口",
        "latitude": 35.6890,
        "longitude": 139.6915,
        "walking_distance_meters": 150,
        "walking_time_minutes": 3,
        "cost_per_hour_jpy": 400,
        "accepts_medical_discount": true,
        "spaces_available_typical": 20
      }
    ],
    "streetParkingAllowed": false,
    "loadingZone": {"available": false}
  },

  "buildingAccess": {
    "type": "apartment",
    "floor": 5,
    "hasElevator": true,
    "elevatorWaitTimeAvgMinutes": 3,
    "stairsOnly": false,
    "wheelchairAccessible": true,
    "narrowHallways": false
  },

  "transitionDelaySeconds": 300,  // 到着→サービス開始の標準ロス時間

  "safetyNotes": ["ペット(猫2匹)が玄関に出てくる可能性", "段差注意"],

  "photos": [
    {"type": "building_entrance", "url": "gs://locations/loc-xxx/entrance.jpg"},
    {"type": "parking_nearby", "url": "gs://locations/loc-xxx/parking.jpg"}
  ]
}
```

### 7.4 テーブル: `route_optimization_jobs` (最適化実行履歴)

```sql
CREATE TABLE route_optimization_jobs (
    job_id varchar(36) NOT NULL,
    execution_date date NOT NULL,

    status varchar(20) NOT NULL,  -- "pending" | "running" | "completed" | "failed"

    -- API Request
    request_payload jsonb NOT NULL,  -- 送信したShipmentModelの完全なコピー

    -- API Response
    response_payload jsonb,
    optimization_metrics jsonb,  -- 総走行距離、時間、コスト削減率等

    execution_started_at timestamptz,
    execution_completed_at timestamptz,
    execution_duration_seconds int,

    -- エラーハンドリング
    error_message text,
    retry_count int DEFAULT 0,

    created_by varchar(36),  -- システムまたは管理者ID

    PRIMARY KEY (job_id)
);

CREATE INDEX idx_optimization_date ON route_optimization_jobs(execution_date DESC);
```

---

## 8. Firestore: リアルタイムコラボレーション層

### 8.1 Spannerとの役割分担

| データ種別 | 保存先 | 理由 |
|---|---|---|
| **マスターデータ**(患者、保険、病名) | **Spanner** | 強整合性、複雑なクエリ、監査要件 |
| **医療記録**(SOAP、処方) | **Spanner** | 法的保存義務、ACID特性必須 |
| **訪問スケジュール** | **Spanner** | ルート最適化の計算基盤 |
| **チャットメッセージ** | **Firestore** | リアルタイム同期、オフライン対応 |
| **位置情報共有** | **Firestore** | 高頻度更新、低レイテンシ |
| **在庫ステータス**(医療材料) | **Firestore** | リアルタイム在庫確認 |
| **通知キュー** | **Firestore** | Push通知トリガー |

### 8.2 Firestoreコレクション設計

#### コレクション: `coordination_chats`

```
/coordination_chats/{chatRoomId}
  ├─ metadata (Document)
  │   ├─ patientId: "patient-123"
  │   ├─ participants: ["doctor-456", "nurse-789", "caremgr-012"]
  │   ├─ createdAt: Timestamp
  │   └─ lastMessageAt: Timestamp
  │
  └─ messages (Subcollection)
      └─ {messageId} (Document)
          ├─ senderId: "nurse-789"
          ├─ senderName: "佐藤看護師"
          ├─ content: "本日訪問時、SpO2 92%に低下。酸素流量2L/分へ変更。"
          ├─ type: "text" | "vital_alert" | "image" | "file"
          ├─ timestamp: Timestamp
          ├─ readBy: {"doctor-456": Timestamp, "nurse-789": Timestamp}
          ├─ priority: "routine" | "urgent" | "stat"
          ├─ attachments: [
          │   {"type": "vital_record", "spannerRef": "obs-id-xxx"}
          │ ]
          └─ reactions: {"👍": ["doctor-456"], "確認": ["caremgr-012"]}
```

**セキュリティルール(例):**

```javascript
match /coordination_chats/{chatRoomId}/messages/{messageId} {
  allow read: if request.auth != null &&
    request.auth.uid in get(/databases/$(database)/documents/coordination_chats/$(chatRoomId)/metadata).data.participants;

  allow create: if request.auth != null &&
    request.resource.data.senderId == request.auth.uid;
}
```

#### コレクション: `staff_locations` (スタッフ位置情報)

```
/staff_locations/{staffId} (Document)
  ├─ currentLocation: GeoPoint(latitude, longitude)
  ├─ accuracy: 10  // meters
  ├─ timestamp: Timestamp
  ├─ status: "on_route" | "at_patient" | "at_station" | "off_duty"
  ├─ currentVisitId: "schedule-xxx"  // Spannerへの参照
  └─ nextVisitETA: Timestamp
```

**位置情報更新頻度:**
- 移動中: 30秒ごと
- 訪問中: 更新停止(プライバシー保護)
- 緊急時: 10秒ごと

---

## 9. セキュリティ実装要件

### 9.1 多層防御アーキテクチャ

```
レイヤー1: ネットワーク境界
  ├─ Cloud Armor (DDoS防御、WAF)
  └─ Identity-Aware Proxy (IAP) - ゼロトラストアクセス

レイヤー2: 認証・認可
  ├─ Firebase Authentication (MFA必須)
  ├─ カスタムクレーム(role: "doctor" | "nurse" | "admin")
  └─ Spanner IAM (roles/spanner.databaseUser)

レイヤー3: データ暗号化
  ├─ 転送時: TLS 1.3 (強制)
  ├─ 保存時: CMEK (Cloud KMS管理)
  └─ アプリケーション層: 機微フィールド個別暗号化(AEAD)

レイヤー4: アクセス制御
  ├─ Row-Level Security (RLS) - 患者単位
  ├─ Column-Level Access (VIEWによるマスキング)
  └─ Firestore Security Rules

レイヤー5: 監査・検知
  ├─ Cloud Audit Logs (5年保存)
  ├─ Security Command Center (異常検知)
  └─ アプリケーションログ (誰が何を見たか)
```

### 9.2 Row-Level Security (RLS) 実装

**患者データへのアクセス制限:**

```sql
-- 担当患者のみ閲覧可能にするビュー
CREATE VIEW view_my_patients SQL SECURITY INVOKER AS
SELECT p.*
FROM patients p
INNER JOIN staff_patient_assignments spa
  ON p.patient_id = spa.patient_id
WHERE spa.staff_id = CURRENT_USER()  -- Firebase UIDをCURRENT_USER()として渡す
  AND spa.assignment_status = 'active'
  AND p.is_deleted = false;
```

**アプリケーション層での実装:**

```go
// Firebase ID Tokenからユーザー情報取得
func (db *SpannerDB) GetMyPatients(ctx context.Context, firebaseUID string) ([]Patient, error) {
    // SET SESSION AUTHORIZATION でSpannerのセッションユーザーを設定
    _, err := db.client.Apply(ctx, []*spanner.Mutation{
        spanner.Insert("sessions",
            []string{"session_id", "firebase_uid"},
            []interface{}{uuid.New().String(), firebaseUID}),
    })

    // ビュー経由でクエリ
    stmt := spanner.Statement{
        SQL: "SELECT * FROM view_my_patients ORDER BY full_name_kana",
    }
    iter := db.client.Single().Query(ctx, stmt)
    defer iter.Stop()

    // 結果処理...
}
```

### 9.3 Column-Level Masking (個人情報マスキング)

**開発者・アナリスト向けの匿名化ビュー:**

```sql
CREATE VIEW view_patients_anonymized SQL SECURITY DEFINER AS
SELECT
    patient_id,
    'MASKED' AS full_name,
    'MASKED' AS full_name_kana,
    DATE_TRUNC('month', birth_date) AS birth_month,  -- 日付は月単位に丸める
    gender,
    (addresses #- '{0,line}' #- '{0,postalCode}') AS address_summary,  -- 番地削除
    NULL AS contact_points,
    created_at
FROM patients
WHERE is_deleted = false;

-- 付与権限
GRANT SELECT ON view_patients_anonymized TO ROLE 'data_analyst';
```

### 9.4 監査ログ要件

**記録必須イベント:**

| イベント | ログ出力先 | 保存期間 | 内容 |
|---|---|---|---|
| **患者情報閲覧** | Spanner `audit_access_logs` | 5年 | user_id, patient_id, accessed_fields, timestamp, IP |
| **データ変更** | Cloud Audit Logs | 10年 | Before/After値、変更理由 |
| **ACP閲覧・変更** | 専用テーブル `acp_audit_logs` | 永久 | 閲覧者、変更内容、患者同意の有無 |
| **暗号鍵使用** | Cloud KMS Audit Logs | 10年 | 鍵ID、操作(encrypt/decrypt)、呼び出し元 |
| **データエクスポート** | `data_export_logs` | 10年 | エクスポート範囲、受領者、目的 |

**監査テーブル例:**

```sql
CREATE TABLE audit_access_logs (
    log_id varchar(36) NOT NULL,
    event_time timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,

    actor_id varchar(36) NOT NULL,  -- Firebase UID
    actor_role varchar(50) NOT NULL,
    actor_ip_address varchar(45),

    action varchar(50) NOT NULL,  -- "view" | "update" | "delete" | "export"
    resource_type varchar(50) NOT NULL,  -- "patient" | "soap_note" | "acp_record"
    resource_id varchar(36) NOT NULL,

    accessed_fields jsonb,  -- 閲覧したカラム名のリスト
    change_details jsonb,  -- Before/After

    purpose varchar(200),  -- アクセス目的(診療、請求、研究等)

    PRIMARY KEY (log_id)
);

-- 時系列検索インデックス
CREATE INDEX idx_audit_time ON audit_access_logs(event_time DESC);
CREATE INDEX idx_audit_actor ON audit_access_logs(actor_id, event_time DESC);
CREATE INDEX idx_audit_resource ON audit_access_logs(resource_id, event_time DESC);
```

---

## 10. 運用・監査・バックアップ戦略

### 10.1 バックアップポリシー

| データ種別 | RPO (目標復旧時点) | RTO (目標復旧時間) | バックアップ方式 |
|---|---|---|---|
| **Spanner** | 1時間 | 4時間 | 自動バックアップ(24時間ごと) + PITR(Point-in-Time Recovery) |
| **Firestore** | 24時間 | 8時間 | 日次エクスポート → Cloud Storage |
| **Cloud Storage** | 0 (バージョニング) | 即時 | オブジェクトバージョニング有効化 |

**Spanner PITRによる誤操作復旧シナリオ:**

```bash
# 2025-12-12 10:30に誤って患者データを削除した場合
gcloud spanner databases restore \
  --source-database=visitas-prod-db \
  --destination-database=visitas-restored-db \
  --destination-instance=visitas-instance \
  --point-in-time="2025-12-12T10:25:00Z"

# 復旧データを検証後、本番DBへマージ
```

### 10.2 災害復旧(DR)計画

**リージョン戦略:**

```
Primary: asia-northeast1 (東京)
Secondary: asia-northeast2 (大阪)

Spannerマルチリージョン構成: nam-eur-asia3 (グローバルレプリケーション)
```

**DRテスト頻度:**
- フルスイッチオーバー訓練: 年2回
- 部分復旧訓練: 四半期ごと

### 10.3 データ保持ポリシー

| データ種別 | 法的要件 | Visitas保持期間 | 削除方式 |
|---|---|---|---|
| **診療録(SOAP)** | 5年(医師法24条) | 10年 | 論理削除→物理削除(10年後) |
| **処方箋** | 3年(薬剤師法28条) | 5年 | 論理削除→物理削除 |
| **監査ログ** | 5年(ガイドライン) | 10年 | 自動アーカイブ(Coldline Storage) |
| **チャットログ** | - | 3年 | 物理削除(患者同意に基づく) |

---

## 11. パフォーマンス最適化戦略

### 11.1 Generated Columnsの戦略的活用

**頻出検索パターンの実体化:**

```sql
-- 要介護度での検索を高速化
ALTER TABLE patient_coverages
ADD COLUMN care_level_code int GENERATED ALWAYS AS (
    CASE WHEN insurance_type = 'long_term_care'
    THEN (details->'careLevelCertification'->>'levelCode')::int
    ELSE NULL END
) STORED;

CREATE INDEX idx_care_level ON patient_coverages(care_level_code)
WHERE insurance_type = 'long_term_care';

-- 独居患者の抽出を高速化
ALTER TABLE patient_social_profiles
ADD COLUMN is_living_alone boolean GENERATED ALWAYS AS (
    content->'livingSituation'->>'status' = 'living_alone'
) STORED;

CREATE INDEX idx_living_alone ON patient_social_profiles(is_living_alone)
WHERE is_living_alone = true;
```

### 11.2 クエリパフォーマンスのベストプラクティス

**❌ 悪い例(フルスキャン):**

```sql
-- JSONB内のキーをフィルタ条件に使用→インデックス未使用
SELECT * FROM patient_coverages
WHERE details->'careLevelCertification'->>'levelCode' = '3';
```

**✅ 良い例(Generated Column利用):**

```sql
-- Generated Columnのインデックスを利用
SELECT * FROM patient_coverages
WHERE care_level_code = 3;
```

### 11.3 インターリーブによるJOIN最適化

**患者+保険情報+最新バイタルを一括取得:**

```sql
-- インターリーブテーブルは物理的に近接配置されるため高速
SELECT
    p.patient_id,
    p.name,
    c.details AS insurance_details,
    o.value AS latest_vitals
FROM patients p
INNER JOIN patient_coverages c
    ON p.patient_id = c.patient_id
INNER JOIN LATERAL (
    SELECT value FROM clinical_observations
    WHERE patient_id = p.patient_id
      AND category = 'vital_signs'
    ORDER BY effective_datetime DESC
    LIMIT 1
) o ON true
WHERE p.active = true;
```

---

## 12. 移行・拡張・災害復旧

### 12.1 既存システムからの移行戦略

**フェーズド移行アプローチ:**

```
Phase 1: 患者マスター移行(patients, patient_identifiers)
  ├─ ETLパイプライン構築(Cloud Data Fusion)
  ├─ データクレンジング(重複排除、正規化)
  └─ 並行稼働期間(2週間)

Phase 2: 保険・医療記録移行
  ├─ 過去3年分の診療録をSpannerへ
  └─ JSONB構造への変換(カスタムスクリプト)

Phase 3: 新機能(ACP, Route Optimization)の段階的展開
  └─ パイロットステーション(3拠点)→全国展開
```

### 12.2 スケーラビリティ設計

**想定負荷:**

| 指標 | Year 1 | Year 3 | Year 5 |
|---|---|---|---|
| **登録患者数** | 10,000 | 100,000 | 500,000 |
| **月間訪問記録** | 30,000 | 300,000 | 1,500,000 |
| **同時接続スタッフ** | 200 | 2,000 | 10,000 |
| **Spanner Node数** | 1 | 3 | 10 |

**自動スケーリング戦略:**

```yaml
# Cloud Spanner Autoscaler (Terraform例)
resource "google_spanner_instance" "main" {
  config       = "regional-asia-northeast1"
  display_name = "visitas-production"

  autoscaling {
    min_nodes = 1
    max_nodes = 10

    autoscaling_limits {
      max_processing_units = 10000
    }

    autoscaling_targets {
      high_priority_cpu_utilization_percent = 65
      storage_utilization_percent           = 75
    }
  }
}
```

### 12.3 将来拡張ポイント

**Phase 3以降の機能拡張:**

1. **IoT/ウェアラブル連携**
   - `clinical_observations` にデバイスIDフィールド追加
   - Firestoreでリアルタイムバイタル受信→Spannerへバッチ同期

2. **AI予測モデル統合**
   - BigQueryへのデータエクスポート(日次)
   - Vertex AI Pipelinesで予測モデル訓練
   - 予測結果を `patient_risk_scores` テーブルへ格納

3. **FHIR API公開**
   - Cloud Healthcare APIとの統合
   - `patients` → FHIR `Patient` 変換レイヤー
   - OAuth 2.0によるサードパーティアプリ連携

---

## 付録A: SQL DDL全集

**完全なテーブル定義は以下のディレクトリに配置:**

```
/backend/migrations/
  ├─ 001_create_patients.sql
  ├─ 002_create_patient_identifiers.sql
  ├─ 003_create_patient_social_profiles.sql
  ├─ 004_create_patient_coverages.sql
  ├─ 005_create_medical_conditions.sql
  ├─ 006_create_allergy_intolerances.sql
  ├─ 007_create_clinical_observations.sql
  ├─ 008_create_care_plans.sql
  ├─ 009_create_acp_records.sql
  ├─ 010_create_medication_orders.sql
  ├─ 011_create_visit_schedules.sql
  ├─ 012_create_logistics_locations.sql
  ├─ 013_create_coordination_messages.sql
  ├─ 014_create_audit_logs.sql
  └─ 015_create_views_security.sql
```

---

## 付録B: JSONB検索クエリ集

**よく使うJSONB演算子:**

```sql
-- 1. 特定キーの存在確認
SELECT * FROM patient_social_profiles
WHERE content ? 'financialBackground';

-- 2. ネストしたキーの値検索
SELECT * FROM patient_social_profiles
WHERE content->'keyPersons'->0->>'role' = 'primary_caregiver';

-- 3. 配列要素の検索
SELECT * FROM patient_social_profiles
WHERE content->'socialHistory'->'lifestyle'->'smoking'->>'status' = 'former_smoker';

-- 4. JSONBオブジェクトの包含関係(@>演算子)
SELECT * FROM acp_records
WHERE directives @> '{"cpr": {"decision": "do_not_attempt"}}'::jsonb;

-- 5. 配列の長さ取得
SELECT patient_id, jsonb_array_length(content->'keyPersons') AS key_person_count
FROM patient_social_profiles;

-- 6. JSONB → テーブル形式への展開(jsonb_to_recordset)
SELECT patient_id, kp.*
FROM patient_social_profiles,
LATERAL jsonb_to_recordset(content->'keyPersons') AS kp(
    name text,
    relationship text,
    priority int
)
WHERE kp.priority = 1;
```

---

## 付録C: セキュリティチェックリスト

**本番環境リリース前の必須確認項目:**

- [ ] CMEKによるSpanner暗号化が有効化されている
- [ ] すべてのテーブルに`is_deleted`カラムが存在し、物理削除が無効化されている
- [ ] 監査ログテーブル(`audit_access_logs`)への書き込みがすべてのCRUD操作で実行される
- [ ] Firebase Authenticationで多要素認証(MFA)が強制されている
- [ ] IAMロールが最小権限原則に従っている(開発者に本番DBへの直接アクセス権がない)
- [ ] Cloud Armorが有効化され、DDoS保護ルールが設定されている
- [ ] Identity-Aware Proxy(IAP)がWeb管理画面に適用されている
- [ ] Firestore Security Rulesが本番デプロイされている
- [ ] PITRバックアップが正常に機能することが確認されている
- [ ] 災害復旧訓練が実施され、RTOが達成可能であることが検証されている
- [ ] 個人情報取扱同意の記録が`patients.consent_records`に格納されている
- [ ] データ保持ポリシーに基づく自動削除ジョブが稼働している

---

## まとめ

本データベース要件定義書は、在宅医療プラットフォーム「Visitas」における**患者データのリッチな管理とセキュアな運用**を実現するための包括的設計指針である。

**核心的な設計判断:**

1. **SOAP主導型ハイブリッドアーキテクチャ**: 医療記録の性質に合わせて、リレーショナルとJSONBを戦略的に使い分け
2. **多層防御セキュリティ**: ネットワーク、認証、暗号化、アクセス制御、監査の5層で医療情報を保護
3. **API統合最優先設計**: Route Optimization APIとの深い統合により、データベースがロジスティクスエンジンとして機能
4. **制度変更への耐性**: 日本の複雑な保険制度をJSONBでカプセル化し、スキーマ変更コストを最小化
5. **介護者ウェルビーイングの可視化**: 患者だけでなく介護者の負担も構造化データとして記録・評価

このアーキテクチャは、単なるデータ保存基盤ではなく、**AIが理解可能な構造化された医療ナラティブ**として機能し、Geminiによるカルテ自動生成、法的書類のAI下書き作成、予測的ケアプランニングといった次世代機能の基盤となる。

---

**承認:**

| 役割 | 氏名 | 承認日 | 署名 |
|---|---|---|---|
| プロダクトオーナー | | | |
| テックリード | | | |
| 医療監修 | | | |
| セキュリティ責任者 | | | |

**改訂履歴:**

| 版 | 日付 | 変更内容 | 変更者 |
|---|---|---|---|
| 1.0 | 2025-12-12 | 初版作成 | |
