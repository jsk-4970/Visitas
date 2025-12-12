# 楽観的ロック実装完了レポート

## 実装日時
2025-12-12

## 実装概要
CarePlansとMedicationOrdersドメインに楽観的ロック（Optimistic Locking）を追加しました。これにより、複数のユーザーが同時に同じレコードを編集しようとした際の競合を防ぎます。

## 変更内容

### 1. データベース層（Migration）
**ファイル**: `migrations/017_add_optimistic_locking.sql`

- `care_plans` テーブルに `version INT NOT NULL DEFAULT 1` カラムを追加
- `medication_orders` テーブルに `version INT NOT NULL DEFAULT 1` カラムを追加

### 2. モデル層（Models）
**ファイル**: 
- `internal/models/care_plan.go`
- `internal/models/medication_order.go`

**変更点**:
- `CarePlan` 構造体に `Version int` フィールドを追加
- `MedicationOrder` 構造体に `Version int` フィールドを追加
- `CarePlanUpdateRequest` に `ExpectedVersion *int` フィールドを追加
- `MedicationOrderUpdateRequest` に `ExpectedVersion *int` フィールドを追加

### 3. Repository層
**ファイル**:
- `internal/repository/care_plan_repository.go`
- `internal/repository/medication_order_repository.go`

**変更点**:
- `Create` メソッド: version を 1 に初期化
- すべての `SELECT` クエリに `version` カラムを追加
- `Update` メソッド: version を自動インクリメント
- `UpdateWithVersion` メソッドを新規追加（バージョンチェック付き）
- `scan*` メソッド: version フィールドをスキャン

### 4. Service層
**ファイル**:
- `internal/services/care_plan_service.go`
- `internal/services/medication_order_service.go`

**変更点**:
- `Update*` メソッドに楽観的ロックチェックを追加
- バージョン不一致時に `CONFLICT` エラーを返す
- `ExpectedVersion` が提供された場合は `UpdateWithVersion` を呼び出し
- 提供されない場合は通常の `Update` を呼び出し（後方互換性維持）

### 5. API仕様（OpenAPI）
**ファイル**: `openapi.yaml`

**変更点**:
- `CarePlan` スキーマに以下を追加:
  - `version` (integer): 楽観的ロックバージョンカウンター
  - `created_by`, `created_at`, `updated_at` フィールド
- `MedicationOrder` スキーマに `version` フィールドを追加

## 実装パターン

### 楽観的ロックの動作フロー

1. **レコード取得時**: クライアントは現在の `version` を取得
2. **更新リクエスト時**: リクエストボディに `expected_version` を含める
3. **バージョンチェック**: 
   - DB内の現在バージョンと `expected_version` を比較
   - 一致する場合: 更新を実行し、`version` をインクリメント
   - 不一致の場合: 409 Conflict エラーを返す

### エラーレスポンス例

```json
{
  "error": "CONFLICT: Care plan was modified by another user. Please refresh and try again. Expected version 3 but found 5"
}
```

## HTTPステータスコード

- **200 OK**: 更新成功
- **409 Conflict**: バージョン不一致（別ユーザーによる同時編集）
- **403 Forbidden**: アクセス権限なし
- **404 Not Found**: レコードが存在しない

## 使用例

### CarePlanの更新（楽観的ロック使用）

```json
PUT /api/v1/patients/{patient_id}/care-plans/{plan_id}
Content-Type: application/json

{
  "title": "Updated Care Plan",
  "status": "active",
  "expected_version": 3
}
```

### MedicationOrderの更新（楽観的ロック使用）

```json
PUT /api/v1/patients/{patient_id}/medication-orders/{order_id}
Content-Type: application/json

{
  "status": "completed",
  "expected_version": 2
}
```

## 後方互換性

`expected_version` を含めない場合、従来通りの動作（バージョンチェックなし）で更新されます。これにより、既存のクライアントとの互換性が維持されます。

## 実装参考

MedicalRecordsドメインの既存実装を参考にしました：
- `internal/models/medical_record.go`
- `internal/repository/medical_record_repository.go`  
- `internal/services/medical_record_service.go`

## テスト状況

- ✅ ビルド成功（`go build ./internal/...`）
- ⚠️ 統合テストは既存のDBスキーマ問題により一部FAIL（別途対応必要）

## 次のステップ

1. ✅ 楽観的ロックの拡大（完了）
2. ⚠️ テストカバレッジ60% → 80%への向上（未着手）
3. ✅ openapi.yaml同期確認と更新（完了）

## コミット情報

- マイグレーション: `017_add_optimistic_locking.sql`
- モデル層: 2ファイル更新
- Repository層: 2ファイル更新
- Service層: 2ファイル更新
- API仕様: 1ファイル更新

