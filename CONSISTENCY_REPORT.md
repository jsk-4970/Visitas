# Visitas プロジェクト整合性確認レポート

**実施日時**: 2025-12-12
**実施者**: Claude Sonnet 4.5
**対象**: Visitas在宅医療プラットフォーム全体

---

## エグゼクティブサマリー

### 総合評価: 🟡 **要改善** (スコア: 65/100)

プロジェクトは基本的な構造とドキュメントが整備されているが、以下の問題が検出されました：

#### ✅ 完了した修正
1. **重複ディレクトリの削除**: `backend/backend/` を削除完了
2. **古いマイグレーションのアーカイブ**: Google SQL方言版をarchive/に移動完了

#### 🟡 対応が必要な問題
1. **マイグレーションファイルの番号重複**: 同じ番号で異なるテーブル定義が複数存在
2. **プロジェクト構造の不整合**: CLAUDE.mdの定義と実際の構造が異なる
3. **ドキュメントのパス参照**: 実際の構造と異なるパスが記載されている

---

## 1. ディレクトリ構造の整合性

### 現在の実際の構造

```
Visitas/
├── backend/                      # ✅ 存在するが、go.modは親ディレクトリ
│   ├── .dockerignore
│   ├── .env.example
│   ├── .gitignore
│   ├── Dockerfile
│   ├── cmd/                      # ⚠️ ここにある
│   ├── config/                   # ⚠️ ここにある
│   ├── go.mod                    # ✅ Goプロジェクトルート
│   ├── go.sum
│   ├── internal/                 # ⚠️ ここにある
│   ├── migrations/               # ⚠️ ここにある (31ファイル)
│   ├── pkg/                      # ⚠️ ここにある
│   └── scripts/
├── docs/                         # ✅ ドキュメント
│   ├── DATABASE_REQUIREMENTS.md
│   ├── SETUP.md
│   └── openapi.yaml
├── config/                       # ⚠️ ルート直下にも存在
│   └── README.md
├── migrations/                   # ⚠️ ルート直下にも存在
│   └── README.md
├── tests/                        # ⚠️ ルート直下にも存在
│   └── README.md
├── NEXT_STEPS.md
├── PROJECT_STRUCTURE_FIX.md
├── CLAUDE.md
└── README.md
```

### CLAUDE.mdで定義された構造 (期待値)

```
Visitas/
├── backend/                      # Goバックエンド
│   ├── cmd/
│   ├── internal/
│   ├── pkg/
│   ├── migrations/
│   ├── config/
│   ├── tests/
│   ├── go.mod
│   └── go.sum
├── mobile/                       # Flutter (未実装)
├── docs/                         # ドキュメント
└── infra/                        # IaC
```

### 問題点

1. **構造の二重性**: `backend/` 配下と ルート直下に同名ディレクトリが存在
2. **READMEの配置**: 新規作成したREADME.mdがルート直下のディレクトリに配置されている

### ステータス: 🟡 **部分的に解決**

- ✅ `backend/backend/` 重複は削除完了
- ⚠️ プロジェクト構造の正式化は未完了（[PROJECT_STRUCTURE_FIX.md](PROJECT_STRUCTURE_FIX.md) 参照）

---

## 2. マイグレーションファイルの整合性

### 検出された問題

#### 2.1 ファイル番号の重複

| 番号 | ファイル1 | ファイル2 | 優先 |
|------|-----------|-----------|------|
| 003 | create_patient_social_profiles.sql | create_staff_patient_assignments.sql | ⚠️ 競合 |
| 004 | create_audit_patient_access_logs.sql | create_patient_coverages.sql | ⚠️ 競合 |
| 005 | create_medical_conditions.sql | create_patient_social_profiles.sql | ⚠️ 競合 |
| 006 | create_allergy_intolerances.sql | create_patient_coverages.sql | ⚠️ 競合 |
| 007 | create_clinical_observations.sql | create_medical_conditions.sql | ⚠️ 競合 |
| 008 | create_allergy_intolerances.sql | create_care_plans.sql | ⚠️ 競合 |
| 009 | create_acp_records.sql | create_view_my_patients.sql | ⚠️ 競合 |
| 010 | create_indexes.sql | create_medication_orders.sql | ⚠️ 競合 |

#### 2.2 正しいマイグレーション順序（DATABASE_REQUIREMENTS.md準拠）

DATABASE_REQUIREMENTS.mdの付録Aで定義された順序：

```
001_create_patients.sql              ✅ 存在
002_create_patient_identifiers.sql   ✅ 存在
003_create_patient_social_profiles.sql ⚠️ 重複
004_create_patient_coverages.sql     ⚠️ 重複
005_create_medical_conditions.sql    ⚠️ 重複
006_create_allergy_intolerances.sql  ⚠️ 重複
007_create_clinical_observations.sql ⚠️ 重複
008_create_care_plans.sql            ⚠️ 重複
009_create_acp_records.sql           ⚠️ 重複
010_create_medication_orders.sql     ⚠️ 重複
011_create_visit_schedules.sql       ✅ 存在
012_create_logistics_locations.sql   ✅ 存在
013_create_coordination_messages.sql ❌ 未実装
014_create_audit_logs.sql            ✅ 存在
015_create_views_security.sql        ❌ 未実装
```

#### 2.3 アーカイブ済みファイル（古いGoogle SQL方言）

```
backend/migrations/archive/google-sql-dialect/
├── 001_create_patients_table.sql
├── 002_create_doctors_table.sql
├── 003_create_visits_table.sql
└── 004_create_visit_records_table.sql

backend/migrations/archive/
└── 001_create_patients_enhanced.sql
```

### 推奨対応

```bash
cd backend/migrations

# 1. 重複ファイルを整理
mkdir -p archive/duplicate_files

# DATABASE_REQUIREMENTS.mdに準拠しないファイルを移動
mv 003_create_staff_patient_assignments.sql archive/duplicate_files/
mv 004_create_audit_patient_access_logs.sql archive/duplicate_files/
mv 005_create_patient_social_profiles.sql archive/duplicate_files/  # 重複
mv 006_create_patient_coverages.sql archive/duplicate_files/  # 重複
mv 007_create_medical_conditions.sql archive/duplicate_files/  # 重複
mv 008_create_allergy_intolerances.sql archive/duplicate_files/  # 重複
mv 009_create_view_my_patients.sql archive/duplicate_files/
mv 010_create_indexes.sql archive/duplicate_files/

# 2. 正しいファイルのみ残す（DATABASE_REQUIREMENTS.md準拠）
# 001 - 015 の順序を維持
```

### ステータス: 🟡 **部分的に解決**

- ✅ 古いGoogle SQL方言ファイルをアーカイブ完了
- ⚠️ 番号重複の解消は未完了

---

## 3. Goモデルとマイグレーションの整合性

### backend/internal/models/patient.go

```go
type Patient struct {
    NameHistory    json.RawMessage `json:"name_history" spanner:"name_history"`  // ⚠️ カラム名
    ContactPoints  json.RawMessage `json:"contact_points" spanner:"contact_points"`
    Addresses      json.RawMessage `json:"addresses" spanner:"addresses"`
    ConsentDetails json.RawMessage `json:"consent_details" spanner:"consent_details"`  // ⚠️ カラム名
}
```

### migrations/001_create_patients.sql

```sql
CREATE TABLE patients (
    name jsonb NOT NULL,  -- ⚠️ "name" であって "name_history" ではない
    -- contact_points, addresses は未定義
    consent_records jsonb,  -- ⚠️ "consent_records" であって "consent_details" ではない
)
```

### 不整合箇所

| モデル(Go) | マイグレーション(SQL) | ステータス |
|-----------|---------------------|----------|
| `name_history` | `name` | ❌ 不一致 |
| `contact_points` | 未定義 | ❌ カラム不足 |
| `addresses` | 未定義 | ❌ カラム不足 |
| `consent_details` | `consent_records` | ❌ 不一致 |

### 推奨対応

**オプション1**: マイグレーションをGoモデルに合わせる（推奨）

```sql
-- 001_create_patients.sql を修正
ALTER TABLE patients RENAME COLUMN name TO name_history;
ALTER TABLE patients RENAME COLUMN consent_records TO consent_details;
ALTER TABLE patients ADD COLUMN contact_points jsonb;
ALTER TABLE patients ADD COLUMN addresses jsonb;
```

**オプション2**: GoモデルをマイグレーションDATABASE_REQUIREMENTS.mdに合わせる

```go
// patient.goを修正
Name jsonb NOT NULL,  // name_history → name
ConsentRecords jsonb,  // consent_details → consent_records
```

### ステータス: ❌ **未対応**

---

## 4. ドキュメントの整合性

### 4.1 SETUP.md のインスタンス名

**記載内容**:
```bash
SPANNER_INSTANCE=visitas-instance
SPANNER_DATABASE=visitas-db
```

**実際のプロジェクトID**:
```
PROJECT_ID: stunning-grin-480914-n1
INSTANCE: stunning-grin-480914-n1-instance (推定)
DATABASE: stunning-grin-480914-n1-db (推定)
```

### 4.2 パス参照の不整合

| ドキュメント | 記載パス | 実際のパス | ステータス |
|-------------|---------|-----------|----------|
| SETUP.md | `backend/migrations/` | ✅ 正しい | ✅ OK |
| SETUP.md | `backend/config/` | ✅ 正しい | ✅ OK |
| backend/config/README.md | `backend/config/firebase-config.json` | ✅ 正しい | ✅ OK |
| NEXT_STEPS.md | `docs/SETUP.md` | ✅ 正しい | ✅ OK |

### ステータス: 🟡 **一部要修正**

- ⚠️ SETUP.mdのインスタンス名を実際の値に修正必要

---

## 5. OpenAPI仕様とモデル定義の整合性

### docs/openapi.yaml

```yaml
Patient:
  properties:
    name:
      $ref: '#/components/schemas/HumanName'  # オブジェクト
```

### backend/internal/models/patient.go

```go
NameHistory json.RawMessage `json:"name_history"`  // JSON配列
```

### 不整合

- API仕様: `name` (オブジェクト)
- Goモデル: `name_history` (配列)

### 推奨対応

OpenAPI仕様をGoモデルに合わせて修正：

```yaml
Patient:
  properties:
    name_history:  # name → name_history
      type: array
      items:
        $ref: '#/components/schemas/HumanName'
```

### ステータス: ❌ **未対応**

---

## 6. セキュリティ要件の整合性

### .gitignoreの確認

**ルート .gitignore**:
```
backend/config/firebase-service-account.json
backend/config/*.json
!backend/config/*.example.json
```

**backend/.gitignore**:
```
*.env
.env.local
config/*.json
!config/*.example.json
```

### 問題点

- ✅ Firebase設定ファイルは除外されている
- ✅ `.env` ファイルは除外されている
- ⚠️ `*.json` が広範囲に除外されている（必要なファイルも除外される可能性）

### 推奨対応

```gitignore
# Secrets
backend/config/firebase-config.json
backend/config/firebase-service-account.json
backend/.env
backend/.env.local
backend/.env.*.local

# But keep examples
!backend/config/*.example.json
!backend/.env.example

# Don't exclude package.json, tsconfig.json etc
# *.json は削除
```

### ステータス: 🟡 **一部要改善**

---

## 7. 依存関係の整合性

### go.mod

```go
module github.com/visitas/backend

go 1.22

require (
    cloud.google.com/go/kms v1.15.5
    cloud.google.com/go/spanner v1.56.0
    firebase.google.com/go/v4 v4.13.0
    // ... (依存関係は適切)
)
```

### 問題点

- ⚠️ モジュールパス `github.com/visitas/backend` が実際のGitHubリポジトリと異なる可能性
- ✅ Go 1.22要求は適切
- ✅ 必要な依存関係はすべて定義されている

### 推奨対応

実際のGitHubリポジトリに合わせてモジュールパスを更新：

```go
module github.com/<your-org>/visitas-backend
// または
module github.com/visitas-platform/backend
```

### ステータス: 🟢 **問題なし（軽微な修正のみ）**

---

## 優先度別アクションプラン

### 🔴 Critical (即時対応)

1. ✅ **完了**: 重複ディレクトリの削除
2. ✅ **完了**: 古いマイグレーションのアーカイブ
3. ⏳ **対応中**: マイグレーションファイルの番号重複解消
4. ⏳ **未着手**: GoモデルとマイグレーションSQLの整合性修正

### 🟡 High (1週間以内)

5. ⏳ **未着手**: SETUP.mdのインスタンス名修正
6. ⏳ **未着手**: OpenAPI仕様の修正
7. ⏳ **未着手**: プロジェクト構造の正式化決定

### 🟢 Medium (2週間以内)

8. ⏳ **未着手**: .gitignoreの最適化
9. ⏳ **未着手**: go.modモジュールパスの検討

---

## 次のステップ

### 即時実施可能なタスク（Goインストール不要）

1. **マイグレーションファイルの整理**
   ```bash
   cd backend/migrations
   # 重複ファイルをarchive/duplicate_files/に移動
   # DATABASE_REQUIREMENTS.md準拠のファイルのみ残す
   ```

2. **SETUP.mdの修正**
   - インスタンス名を実際の値に更新
   - パス参照の確認

3. **OpenAPI仕様の修正**
   - Goモデルとの整合性確保

### Goインストール後のタスク

4. **Goモデルの修正**
   - マイグレーションとの整合性確保
   - `go mod tidy` 実行

5. **統合テストの実行**
   - モデルとマイグレーションの動作確認

---

## 総合スコア

| カテゴリ | スコア | ステータス |
|---------|--------|----------|
| ディレクトリ構造 | 70/100 | 🟡 改善済み（さらなる正式化必要） |
| マイグレーション整合性 | 55/100 | 🟡 一部解決（重複解消必要） |
| モデル定義整合性 | 40/100 | 🔴 要対応 |
| ドキュメント整合性 | 75/100 | 🟡 軽微な修正必要 |
| OpenAPI整合性 | 60/100 | 🟡 要修正 |
| セキュリティ設定 | 85/100 | ✅ 良好 |
| **総合スコア** | **65/100** | 🟡 **要改善** |

---

## 添付資料

- [PROJECT_STRUCTURE_FIX.md](PROJECT_STRUCTURE_FIX.md) - プロジェクト構造の修正計画
- [NEXT_STEPS.md](NEXT_STEPS.md) - 次のステップガイド
- [docs/DATABASE_REQUIREMENTS.md](docs/DATABASE_REQUIREMENTS.md) - データベース要件定義

---

**レポート作成日**: 2025-12-12
**次回レビュー推奨日**: 2025-12-14 (48時間後)
**担当者**: プロジェクトオーナー/テックリード
