# Visitas Backend Tests

このディレクトリには、Visitasバックエンドのテストコードが含まれています。

## テスト戦略

### テストレベル

1. **ユニットテスト** (`*_test.go` in each package)
   - 個別の関数・メソッドの動作検証
   - モックを使用して外部依存を分離
   - カバレッジ目標: 80%以上

2. **統合テスト** (`tests/integration/`)
   - 複数のコンポーネントの連携テスト
   - Testcontainersを使用したローカルDB
   - APIエンドポイントのE2Eテスト

3. **パフォーマンステスト** (`tests/performance/`)
   - 負荷テスト (Locust / k6)
   - 目標: 1000同時接続、レスポンス<200ms

## ディレクトリ構造

```
tests/
├── README.md                  # このファイル
├── integration/               # 統合テスト
│   ├── api_test.go
│   ├── db_test.go
│   └── fixtures/
├── performance/               # パフォーマンステスト
│   ├── locustfile.py
│   └── k6_script.js
├── mocks/                     # モック生成
│   ├── repository/
│   └── services/
└── testdata/                  # テストデータ
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

### Testcontainersを使用したローカルDB

```go
import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) *testcontainers.Container {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "visitas_test",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	return &container
}
```

### APIエンドポイントのテスト

```go
func TestPatientAPI_CreatePatient(t *testing.T) {
	// Setup test server
	router := setupTestRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	// Create request
	reqBody := models.PatientCreateRequest{
		BirthDate: "1950-01-15",
		Gender:    "male",
		Name: models.NameRecord{
			Family: "山田",
			Given:  "太郎",
			Kana:   "ヤマダ タロウ",
		},
		// ...
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", server.URL+"/v1/patients", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+getTestToken())

	// Execute
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var patient models.Patient
	json.NewDecoder(resp.Body).Decode(&patient)
	assert.NotEmpty(t, patient.PatientID)
}
```

## パフォーマンステスト

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
