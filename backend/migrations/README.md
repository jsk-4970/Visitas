# Database Migrations

このディレクトリには、Visitasプロジェクトのデータベーススキーマ定義が含まれています。

## マイグレーションファイル一覧

1. `001_create_patients.sql` - 患者マスターテーブル (Root Table)
2. `002_create_patient_identifiers.sql` - 患者識別子テーブル (マイナンバー、保険証番号等)
3. `003_create_patient_social_profiles.sql` - 社会的背景テーブル (Subjective)
4. `004_create_patient_coverages.sql` - 保険情報テーブル
5. `005_create_medical_conditions.sql` - 病名・既往歴テーブル
6. `006_create_allergy_intolerances.sql` - アレルギー・副作用歴テーブル
7. `007_create_clinical_observations.sql` - バイタルサイン・ADL評価テーブル
8. `008_create_care_plans.sql` - ケア計画テーブル
9. `009_create_acp_records.sql` - ACP (Advance Care Planning) テーブル
10. `010_create_medication_orders.sql` - 処方オーダーテーブル
11. `011_create_visit_schedules.sql` - 訪問スケジュールテーブル
12. `012_create_logistics_locations.sql` - ロジスティクス拠点テーブル
13. `013_create_route_optimization_jobs.sql` - ルート最適化ジョブ履歴テーブル
14. `014_create_audit_logs.sql` - 監査ログテーブル
15. `015_create_staff_tables.sql` - スタッフ・車両管理テーブル

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
