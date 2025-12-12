# Database Migrations

このディレクトリには、Visitasプロジェクトのデータベーススキーマ定義が含まれています。

## ✅ 実装済みマイグレーションファイル (14/14 完成 🎉)

**最終更新**: 2025-12-12
**ステータス**: 全テーブルマイグレーション完了

### Phase 1 Core Tables (患者基本情報・医療記録)

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

6. **`006_create_visit_schedules.sql`** - 訪問スケジュールテーブル
   - JSONB: `visit_purpose`, `required_equipment`, `recurrence_rule`, `visit_address`
   - Generated Columns: `duration_minutes`, `visit_prefecture`, `visit_city`
   - ルート最適化連携、繰り返しスケジュール対応、位置情報管理

7. **`007_create_clinical_observations.sql`** - バイタルサイン・ADL評価テーブル (FHIR Observation準拠)
   - JSONB: `value_structured`, `adl_details`
   - Generated Columns: `systolic_bp`, `diastolic_bp`
   - LOINC/SNOMED-CT コード管理、多様な観察値型対応 (quantity, string, boolean)

8. **`008_create_medication_orders.sql`** - 処方オーダーテーブル (FHIR MedicationRequest準拠)
   - JSONB: `dosage_instruction`, `dispense_request`, `check_warnings`
   - Generated Columns: `dose_quantity`, `dose_unit`, `frequency`
   - 薬価基準(YJ)コード、安全性チェック、リフィル処方対応

9. **`009_create_care_plans.sql`** - ケア計画テーブル (FHIR CarePlan準拠)
   - JSONB: `care_team`, `goals`, `activities`, `subject_condition_references`
   - Generated Column: `plan_duration_days`
   - 在宅医療・看護・介護計画の統合管理

10. **`010_create_staff_tables.sql`** - スタッフ・車両管理テーブル (2テーブル)
    - `staff_members`: 医療従事者マスター
      - JSONB: `specialties`, `certifications`, `work_schedule`
      - Generated Column: `full_name`
      - リアルタイム位置情報、免許・資格期限管理
    - `vehicles`: 訪問車両管理
      - JSONB: `medical_equipment`
      - GPS追跡、保険・車検期限管理

11. **`011_create_acp_records.sql`** - ACP (Advance Care Planning) 記録テーブル
    - JSONB: `participants`, `care_preferences`, `treatment_preferences`, `spiritual_preferences`
    - DNAR/POLST管理、終末期医療意思決定記録
    - バージョン管理、見直し頻度追跡、代理人情報管理

12. **`012_create_logistics_locations.sql`** - ロジスティクス拠点テーブル
    - JSONB: `operating_hours`, `service_area`, `facilities`
    - Generated Column: `full_address`
    - Google Maps連携 (Place ID)、ルート最適化起点/終点管理

13. **`013_create_route_optimization_jobs.sql`** - ルート最適化ジョブ履歴テーブル
    - JSONB: `optimization_params`, `google_api_request_payload`, `google_api_response_payload`, `optimized_route`
    - Generated Columns: `total_visits_count`, `total_distance_km`, `total_duration_hours`, `execution_duration_seconds`
    - Google Maps Route Optimization API連携記録、コスト削減効果追跡

14. **`014_create_audit_access_logs.sql`** - 監査ログテーブル (3省2ガイドライン準拠)
    - JSONB: `accessed_fields`, `modified_fields`, `previous_values`, `new_values`, `geolocation`
    - Generated Column: `retention_expires_at`
    - 5年保存、全アクセス記録、失敗ログ、機密データ追跡

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
