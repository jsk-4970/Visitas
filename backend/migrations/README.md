# Database Migrations

このディレクトリには、Visitasプロジェクトのデータベーススキーマ定義が含まれています。

## ✅ 実装済みマイグレーションファイル

### Phase 1 Core Tables (2025-12-12 完成)

1. **`001_create_patients.sql`** - 患者マスターテーブル (Root Table)
   - JSONB: `name_history`, `contact_points`, `addresses`, `consent_details`
   - Generated Columns: `current_family_name`, `current_given_name`, `primary_phone`, `current_prefecture`, `current_city`
   - 論理削除、同意管理、監査フィールド完備

2. **`002_create_social_profiles.sql`** - 社会的背景テーブル (Subjective - FHIR SDOH)
   - JSONB: `content` (生活状況、キーパーソン、経済状況、社会的支援)
   - Generated Columns: `lives_alone`, `requires_caregiver_support`
   - バージョニング、有効期間管理

3. **`003_create_coverages.sql`** - 保険情報テーブル
   - JSONB: `details` (保険種別ごとの詳細情報)
   - Generated Columns: `care_level_code`, `copay_rate`
   - 優先順位管理、検証ステータス

4. **`004_create_medical_conditions.sql`** - 病名・既往歴テーブル (FHIR Condition準拠)
   - 臨床ステータス、検証ステータス (FHIR準拠)
   - ICD-10/SNOMED-CT コード管理
   - 発症・寛解情報追跡

5. **`005_create_allergy_intolerances.sql`** - アレルギー・副作用歴テーブル (FHIR AllergyIntolerance準拠)
   - JSONB: `reactions` (反応イベント配列)
   - Generated Column: `max_severity` (最大重症度の自動計算)
   - クリティカリティ評価、薬剤アレルギー特化検索

## 📋 計画中のマイグレーション (Phase 2-3)

6. `006_create_clinical_observations.sql` - バイタルサイン・ADL評価テーブル (未実装)
7. `007_create_care_plans.sql` - ケア計画テーブル (未実装)
8. `008_create_acp_records.sql` - ACP (Advance Care Planning) テーブル (未実装)
9. `009_create_medication_orders.sql` - 処方オーダーテーブル (未実装)
10. `010_create_visit_schedules.sql` - 訪問スケジュールテーブル (未実装)
11. `011_create_logistics_locations.sql` - ロジスティクス拠点テーブル (未実装)
12. `012_create_route_optimization_jobs.sql` - ルート最適化ジョブ履歴テーブル (未実装)
13. `013_create_audit_logs.sql` - 監査ログテーブル (未実装)
14. `014_create_staff_tables.sql` - スタッフ・車両管理テーブル (未実装)

## 適用方法

### ローカル開発環境 (PostgreSQL)

```bash
# PostgreSQLの場合、順番に実行
for file in backend/migrations/*.sql; do
  psql -U visitas_user -d visitas_dev -f "$file"
done
```

### Cloud Spanner (本番環境)

**重要**: Cloud Spannerに適用する際は、以下の修正が必要です:

1. **INTERLEAVE構文の追加**:
   - `patient_*` テーブルには `) INTERLEAVE IN PARENT patients ON DELETE CASCADE;` を追加

2. **Generated Columns**:
   - Spanner PostgreSQLインターフェースでのサポート状況を確認
   - サポートされていない場合は、アプリケーション層で処理

3. **地理データ型**:
   - PostGIS `geography` 型は非対応のため、`latitude`/`longitude` numeric型で管理

### 適用コマンド (Spanner)

```bash
# 個別ファイルの適用例
gcloud spanner databases ddl update visitas-db \
  --instance=visitas-instance \
  --ddl="$(cat backend/migrations/001_create_patients.sql)"

# または、まとめて適用
gcloud spanner databases ddl update visitas-db \
  --instance=visitas-instance \
  --ddl-file=backend/migrations/all_migrations.sql
```

## データベース設計の原則

- **SOAP主導型ハイブリッドアーキテクチャ**: リレーショナルとJSONBの戦略的使い分け
- **FHIR準拠**: 概念モデルとしてFHIR R4を参照
- **3省2ガイドライン準拠**: 医療情報システムの安全管理要件に準拠
- **論理削除**: 全テーブルに `is_deleted` カラムを設置し、物理削除を禁止

## セキュリティ要件

- **暗号化**: CMEK (Customer-Managed Encryption Keys) 必須
- **監査ログ**: 全データアクセスを `audit_access_logs` に記録 (5年保存)
- **アクセス制御**: Row-Level Security (RLS) による担当患者のみ閲覧可能
- **データ分類**: `data_classification` カラムでLevel 1-4を管理

## 参考資料

- [DATABASE_REQUIREMENTS.md](/docs/DATABASE_REQUIREMENTS.md) - 完全な要件定義書
- [Cloud Spanner PostgreSQL](https://cloud.google.com/spanner/docs/postgresql-interface)
- [FHIR R4 Specification](https://www.hl7.org/fhir/)
