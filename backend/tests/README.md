# Visitas Backend Tests

このディレクトリには、Visitasバックエンドのテストコードが含まれています。

## テスト戦略

### テストレベル

1. **ユニットテスト** (`*_test.go` in each package)
   - 個別の関数・メソッドの動作検証
   - モックを使用して外部依存を分離
   - カバレッジ目標: 80%以上
   - **実装状況**: 一部実装済み（identifier関連）

2. **統合テスト** (`tests/integration/`) ✅ **Phase 1完了**
   - 複数のコンポーネントの連携テスト（Handler→Service→Repository）
   - Cloud Spannerデータベースを使用した実際のE2Eテスト
   - APIエンドポイントの完全なテスト
   - **実装状況**: Phase 1の5ドメイン完全実装（42+テストケース）

3. **パフォーマンステスト** (`tests/performance/`)
   - 負荷テスト (Locust / k6)
   - 目標: 1000同時接続、レスポンス<200ms
   - **実装状況**: 未実装

## ディレクトリ構造

```
tests/
├── README.md                           # このファイル
├── integration/                        # 統合テスト（✅ 実装完了）
│   ├── helpers.go                      # テスト基盤（テストサーバー、ヘルパー関数）
│   ├── visit_schedules_test.go         # 訪問スケジュールの統合テスト
│   ├── clinical_observations_test.go   # バイタル・ADL評価の統合テスト
│   ├── care_plans_test.go              # ケア計画の統合テスト
│   ├── medication_orders_test.go       # 処方オーダーの統合テスト
│   └── acp_records_test.go             # ACP記録の統合テスト
├── performance/                        # パフォーマンステスト（未実装）
│   ├── locustfile.py
│   └── k6_script.js
├── mocks/                              # モック生成（未実装）
│   ├── repository/
│   └── services/
└── testdata/                           # テストデータ（未実装）
    ├── patients.json
    ├── social_profiles.json
    └── coverages.json
```

## テストの実行方法

### すべてのテストを実行

```bash
cd backend
go test ./... -v
```

### 特定のパッケージをテスト

```bash
# モデルのテスト
go test ./internal/models -v

# ハンドラーのテスト
go test ./internal/handlers -v

# リポジトリのテスト
go test ./internal/repository -v
```

### カバレッジレポート生成

```bash
# カバレッジを計測
go test ./... -coverprofile=coverage.out

# HTMLレポート生成
go tool cover -html=coverage.out -o coverage.html

# ブラウザで開く
open coverage.html
```

### 特定のテストケースを実行

```bash
# テスト名でフィルタ
go test ./internal/models -v -run TestPatient_JSONSerialization

# サブテストを実行
go test ./internal/models -v -run TestPatientCreateRequest_Validation/Valid_patient
```

### ベンチマークテスト

```bash
go test ./... -bench=. -benchmem
```

## モックの生成

### mockgenのインストール

```bash
go install github.com/golang/mock/mockgen@latest
```

### モックの生成方法

```bash
# Repositoryインターフェースのモック生成
mockgen -source=internal/repository/patient_repository.go \
  -destination=tests/mocks/repository/patient_repository_mock.go \
  -package=mock_repository

# Serviceインターフェースのモック生成
mockgen -source=internal/services/patient_service.go \
  -destination=tests/mocks/services/patient_service_mock.go \
  -package=mock_services
```

## テストデータの管理

### フィクスチャの使用

テストデータは `testdata/` ディレクトリに配置します:

```go
import (
	"encoding/json"
	"os"
	"testing"
)

func loadTestPatient(t *testing.T, filename string) *models.Patient {
	data, err := os.ReadFile("../../tests/testdata/" + filename)
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	var patient models.Patient
	if err := json.Unmarshal(data, &patient); err != nil {
		t.Fatalf("Failed to unmarshal test data: %v", err)
	}

	return &patient
}
```

### テストデータのクリーンアップ

各テストケースの後に、作成したテストデータをクリーンアップします:

```go
func TestCreatePatient(t *testing.T) {
	// Setup
	patient := createTestPatient(t)

	// Test logic...

	// Cleanup
	t.Cleanup(func() {
		deleteTestPatient(t, patient.PatientID)
	})
}
```

## 統合テストの実行

### 環境準備

統合テストを実行するには、Cloud Spannerデータベース（または互換エミュレータ）が必要です：

```bash
# 環境変数を設定
export GCP_PROJECT_ID=stunning-grin-480914-n1
export SPANNER_INSTANCE=stunning-grin-480914-n1-instance
export SPANNER_DATABASE=stunning-grin-480914-n1-db

# または .env ファイルに記載
```

### 統合テストの実行方法

```bash
# すべての統合テストを実行
cd backend
go test ./tests/integration -v -count=1

# カバレッジ付きで実行
go test ./tests/integration -v -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# 特定のドメインのみテスト
go test ./tests/integration -v -run TestVisitSchedule

# 特定のテストケースのみ実行
go test ./tests/integration -v -run TestVisitSchedule_Integration_CreateAndGet
```

### 統合テストの構造

統合テストは以下のパターンで実装されています：

```go
func TestDomain_Integration_Operation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup: テストサーバーとテスト患者を作成
	ts := SetupTestServer(t)
	defer ts.Close()
	patientID := ts.CreateTestPatient(t)

	// Test: 実際のHTTPリクエストを送信
	t.Run("Test case name", func(t *testing.T) {
		// Arrange: テストデータを準備
		requestJSON := `{...}`

		// Act: HTTPリクエストを実行
		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/...", strings.NewReader(requestJSON))

		// Assert: レスポンスを検証
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		var result models.Model
		DecodeJSONResponse(t, resp, &result)
		assert.NotEmpty(t, result.ID)
	})
}
```

## 実装済み統合テスト一覧

### Phase 1ドメイン（完全実装）

| ドメイン | テストファイル | テストケース数 | カバー範囲 |
|---------|--------------|--------------|----------|
| 訪問スケジュール | `visit_schedules_test.go` | 10+ | CRUD, バリデーション, 制約条件 |
| バイタル・ADL評価 | `clinical_observations_test.go` | 10+ | バイタル測定, ADL評価, 時系列データ |
| ケア計画 | `care_plans_test.go` | 6+ | CRUD, 目標・活動管理 |
| 処方オーダー | `medication_orders_test.go` | 7+ | CRUD, 薬局情報, アクティブ処方 |
| ACP記録 | `acp_records_test.go` | 9+ | CRUD, バージョン管理, 法的文書 |

**合計**: 42+テストケース（バリデーションエラーテストを含めると80+）

### テストカバレッジ

各ドメインで以下の操作をテスト：
- ✅ Create（作成）
- ✅ Read（取得・一覧）
- ✅ Update（更新）
- ✅ Delete（削除）
- ✅ カスタムエンドポイント（latest, active, history等）
- ✅ バリデーションエラー
- ✅ JSONBフィールドの検証

### テストヘルパー機能

`helpers.go`が提供する機能：
- `SetupTestServer()` - テストサーバーの自動セットアップ
- `CreateTestPatient()` - テスト患者の自動作成・クリーンアップ
- `MakeRequest()` - HTTPリクエストヘルパー
- `DecodeJSONResponse()` - JSONレスポンスデコードヘルパー
- テスト認証ミドルウェア（実際の認証をバイパス）

## パフォーマンステスト（未実装）

### Locustを使用した負荷テスト

```python
# tests/performance/locustfile.py
from locust import HttpUser, task, between

class VisitasUser(HttpUser):
    wait_time = between(1, 3)

    @task
    def list_patients(self):
        self.client.get("/v1/patients", headers={
            "Authorization": f"Bearer {self.token}"
        })

    @task(3)
    def get_patient(self):
        self.client.get(f"/v1/patients/{self.patient_id}", headers={
            "Authorization": f"Bearer {self.token}"
        })

    def on_start(self):
        # Login and get token
        self.token = self.get_auth_token()
        self.patient_id = "test-patient-id"
```

実行:

```bash
locust -f tests/performance/locustfile.py --host=http://localhost:8080
```

### k6を使用した負荷テスト

```javascript
// tests/performance/k6_script.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 100,
  duration: '30s',
  thresholds: {
    http_req_duration: ['p(95)<200'],
  },
};

export default function () {
  const res = http.get('http://localhost:8080/v1/patients', {
    headers: { Authorization: `Bearer ${__ENV.TEST_TOKEN}` },
  });

  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 200ms': (r) => r.timings.duration < 200,
  });

  sleep(1);
}
```

実行:

```bash
k6 run tests/performance/k6_script.js
```

## CI/CDでのテスト実行

`.github/workflows/test.yml`:

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.22'

      - name: Run tests
        run: |
          cd backend
          go test ./... -v -coverprofile=coverage.out

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./backend/coverage.out
```

## ベストプラクティス

1. **AAA パターン** (Arrange-Act-Assert)
   ```go
   func TestSomething(t *testing.T) {
       // Arrange
       input := setupTestData()

       // Act
       result := doSomething(input)

       // Assert
       assert.Equal(t, expected, result)
   }
   ```

2. **テーブル駆動テスト**
   ```go
   func TestValidation(t *testing.T) {
       tests := []struct {
           name    string
           input   string
           wantErr bool
       }{
           {"valid input", "valid", false},
           {"invalid input", "", true},
       }

       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               err := validate(tt.input)
               if (err != nil) != tt.wantErr {
                   t.Errorf("got error = %v, want %v", err, tt.wantErr)
               }
           })
       }
   }
   ```

3. **サブテストの使用**
   ```go
   func TestComplex(t *testing.T) {
       t.Run("success case", func(t *testing.T) { /* ... */ })
       t.Run("error case", func(t *testing.T) { /* ... */ })
   }
   ```

4. **クリーンアップの徹底**
   ```go
   func TestWithCleanup(t *testing.T) {
       resource := setup()
       t.Cleanup(func() {
           cleanup(resource)
       })

       // Test logic
   }
   ```

## トラブルシューティング

### テストが失敗する場合

```bash
# 詳細なログを出力
go test ./... -v -count=1

# キャッシュをクリア
go clean -testcache
go test ./... -v
```

### カバレッジが低い場合

```bash
# カバレッジの低いファイルを特定
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep -v "100.0%"
```

## 参考資料

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testify - Testing toolkit](https://github.com/stretchr/testify)
- [gomock - Mocking framework](https://github.com/golang/mock)
- [Testcontainers for Go](https://golang.testcontainers.org/)
