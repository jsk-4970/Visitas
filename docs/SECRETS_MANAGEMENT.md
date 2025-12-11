# Secret Manager による秘密情報管理

## 概要

Visitasプロジェクトでは、すべての秘密情報（API キー、認証情報等）を**Google Cloud Secret Manager**で管理しています。これにより、以下を実現しています：

1. **セキュリティ**: 秘密情報をGitリポジトリに含めない
2. **監査**: すべてのアクセスをCloud Audit Logsで記録
3. **アクセス制御**: IAMベースの細かいアクセス管理
4. **バージョン管理**: 秘密情報の履歴を保持
5. **暗号化**: 自動的にKMSで暗号化

## 管理されている秘密情報

### 開発環境 (dev)

| Secret名 | 説明 | 必須 |
|---------|------|-----|
| `firebase-service-account-dev` | Firebase Admin SDK認証情報 | ✅ |
| `google-maps-api-key-dev` | Google Maps API キー | ✅ |
| `gemini-api-key-dev` | Gemini API キー（AI機能用） | ✅ |
| `cors-allowed-origins-dev` | CORS許可オリジン | ✅ |

### ステージング環境 (staging)

| Secret名 | 説明 | 必須 |
|---------|------|-----|
| `firebase-service-account-staging` | Firebase Admin SDK認証情報 | ✅ |
| `google-maps-api-key-staging` | Google Maps API キー | ✅ |
| `gemini-api-key-staging` | Gemini API キー（AI機能用） | ✅ |
| `cors-allowed-origins-staging` | CORS許可オリジン | ✅ |

### 本番環境 (prod)

| Secret名 | 説明 | 必須 |
|---------|------|-----|
| `firebase-service-account-prod` | Firebase Admin SDK認証情報 | ✅ |
| `google-maps-api-key-prod` | Google Maps API キー | ✅ |
| `gemini-api-key-prod` | Gemini API キー（AI機能用） | ✅ |
| `cors-allowed-origins-prod` | CORS許可オリジン | ✅ |

## 秘密情報管理スクリプト

プロジェクトには、秘密情報管理を簡単にする専用スクリプトが用意されています：

```bash
./scripts/manage-secrets.sh [command] [options]
```

### 使用可能なコマンド

#### 1. 秘密情報の一覧表示

```bash
# すべての秘密情報を表示
./scripts/manage-secrets.sh list

# 特定の環境のみ表示
./scripts/manage-secrets.sh list dev
./scripts/manage-secrets.sh list staging
./scripts/manage-secrets.sh list prod
```

#### 2. 秘密情報の値を表示

```bash
# 最新バージョンの値を表示
./scripts/manage-secrets.sh view google-maps-api-key-dev
./scripts/manage-secrets.sh view gemini-api-key-dev
```

#### 3. 秘密情報の更新

```bash
# インタラクティブに値を更新
./scripts/manage-secrets.sh update google-maps-api-key-dev

# 実行例:
# Enter new value: AIzaSyD...（実際のAPIキーを入力）
```

#### 4. 全秘密情報の作成

```bash
# 環境の全秘密情報を作成（初回セットアップ時）
./scripts/manage-secrets.sh create dev
./scripts/manage-secrets.sh create staging
./scripts/manage-secrets.sh create prod
```

#### 5. アクセス権限の付与

```bash
# Cloud Runサービスアカウントにアクセス権限を付与
./scripts/manage-secrets.sh grant dev
./scripts/manage-secrets.sh grant staging
./scripts/manage-secrets.sh grant prod
```

#### 6. ローカル開発環境への同期

```bash
# Secret Managerから秘密情報を取得してローカル.envファイルに同期
./scripts/manage-secrets.sh sync-local dev

# これにより以下が実行されます:
# 1. backend/.envファイルを作成/上書き
# 2. backend/config/firebase-service-account.jsonをダウンロード
```

**⚠️ 警告**: このコマンドは既存の`.env`ファイルを上書きします。

#### 7. 秘密情報の削除

```bash
# 秘密情報を完全に削除（要確認）
./scripts/manage-secrets.sh delete google-maps-api-key-old
```

**⚠️ 警告**: この操作は取り消せません。削除前に確認プロンプトが表示されます。

## 手動での秘密情報管理

スクリプトを使用せずに、`gcloud`コマンドで直接管理することもできます：

### 秘密情報の作成

```bash
# 新しい秘密情報を作成
gcloud secrets create SECRET_NAME \
  --project=stunning-grin-480914-n1 \
  --replication-policy="user-managed" \
  --locations="asia-northeast1" \
  --labels="environment=dev,project=visitas"

# 初期値を設定
echo 'YOUR_SECRET_VALUE' | gcloud secrets versions add SECRET_NAME --data-file=-
```

### 秘密情報の更新

```bash
# 新しいバージョンを追加（既存の値は保持されます）
echo 'NEW_SECRET_VALUE' | gcloud secrets versions add SECRET_NAME --data-file=-
```

### 秘密情報の表示

```bash
# 最新バージョンの値を表示
gcloud secrets versions access latest --secret=SECRET_NAME

# 特定バージョンの値を表示
gcloud secrets versions access VERSION_NUMBER --secret=SECRET_NAME
```

### アクセス権限の付与

```bash
# Cloud Runサービスアカウントに読み取り権限を付与
gcloud secrets add-iam-policy-binding SECRET_NAME \
  --member="serviceAccount:visitas-dev-run@stunning-grin-480914-n1.iam.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"
```

### 秘密情報の削除

```bash
# 秘密情報を完全に削除
gcloud secrets delete SECRET_NAME
```

## ローカル開発環境でのセットアップ

### 初回セットアップ

1. **`.env.example`をコピー**:
   ```bash
   cd backend
   cp .env.example .env
   ```

2. **Firebase認証情報をダウンロード**:
   ```bash
   mkdir -p backend/config
   gcloud secrets versions access latest \
     --secret="firebase-service-account-dev" \
     > backend/config/firebase-service-account.json
   ```

3. **API キーを設定** (`.env`ファイルを編集):
   ```bash
   # Option 1: Secret Managerから取得して設定
   ./scripts/manage-secrets.sh sync-local dev

   # Option 2: 手動で.envファイルを編集
   vim backend/.env
   # GOOGLE_MAPS_API_KEY=your_actual_api_key
   # GEMINI_API_KEY=your_actual_api_key
   ```

### 開発時の注意事項

1. **`.env`ファイルはGitにコミットしない**
   - `.gitignore`に追加済み（自動的に除外されます）

2. **Firebase認証情報ファイルもコミットしない**
   - `backend/config/*.json`は`.gitignore`で除外済み

3. **本番環境の秘密情報は使用しない**
   - 開発環境専用の秘密情報を使用してください

## Cloud Runでの秘密情報の使用

### 自動インジェクション

Cloud Runへのデプロイ時、秘密情報は自動的に環境変数としてインジェクションされます：

```bash
# deploy.shの設定（抜粋）
gcloud run deploy visitas-api-dev \
  --set-secrets "GOOGLE_MAPS_API_KEY=google-maps-api-key-dev:latest" \
  --set-secrets "GEMINI_API_KEY=gemini-api-key-dev:latest" \
  --set-secrets "ALLOWED_ORIGINS=cors-allowed-origins-dev:latest" \
  ...
```

### ファイルとしてのマウント

Firebase認証情報は、ファイルとしてマウントされます：

```bash
--set-env-vars "FIREBASE_CONFIG_PATH=/secrets/firebase.json" \
--set-secrets "/secrets/firebase.json=firebase-service-account-dev:latest"
```

これにより、アプリケーションは `/secrets/firebase.json` からファイルを読み込めます。

## セキュリティベストプラクティス

### ✅ すべき事

1. **環境ごとに異なる秘密情報を使用**
   - 開発、ステージング、本番で別々のAPI キーを使用

2. **定期的なローテーション**
   - 重要な秘密情報は定期的に更新（3-6ヶ月ごと）

3. **最小権限の原則**
   - 必要なサービスアカウントにのみアクセス権限を付与

4. **監査ログの確認**
   - Cloud Audit Logsで秘密情報へのアクセスを定期的に確認

5. **バージョン管理の活用**
   - 古いバージョンを無効化する前に、新しいバージョンの動作を確認

### ❌ してはいけない事

1. **秘密情報をGitにコミットしない**
   - `.env`, `*.json`, `*.key`等はすべて`.gitignore`で除外

2. **秘密情報をログに出力しない**
   - デバッグ時も秘密情報を標準出力に表示しない

3. **秘密情報をハードコードしない**
   - ソースコード内に直接API キーを記述しない

4. **本番環境の秘密情報を開発で使用しない**
   - 必ず環境ごとに分離された秘密情報を使用

5. **秘密情報を平文でシェアしない**
   - Slack、メール、ドキュメント等で秘密情報を共有しない

## トラブルシューティング

### 秘密情報にアクセスできない

**症状**: Cloud Runで秘密情報が読み込めない

**解決策**:
```bash
# IAMポリシーを確認
gcloud secrets get-iam-policy SECRET_NAME

# アクセス権限を再付与
./scripts/manage-secrets.sh grant dev
```

### ローカルで秘密情報が見つからない

**症状**: ローカル開発で秘密情報が読み込めない

**解決策**:
```bash
# .envファイルが存在するか確認
ls -la backend/.env

# Secret Managerから同期
./scripts/manage-secrets.sh sync-local dev
```

### Firebase認証エラー

**症状**: Firebase Admin SDKの初期化に失敗

**解決策**:
```bash
# 認証情報ファイルが存在するか確認
ls -la backend/config/firebase-service-account.json

# 再ダウンロード
gcloud secrets versions access latest \
  --secret="firebase-service-account-dev" \
  > backend/config/firebase-service-account.json
```

### 秘密情報の値が古い

**症状**: 更新したはずの秘密情報が反映されていない

**解決策**:
```bash
# 最新バージョンを確認
gcloud secrets versions list SECRET_NAME

# Cloud Runを再デプロイ（最新バージョンを取得）
./scripts/deploy.sh dev
```

## API キーの取得方法

### Google Maps API キー

1. [Google Cloud Console](https://console.cloud.google.com/)にアクセス
2. プロジェクトを選択: `stunning-grin-480914-n1`
3. 「APIとサービス」→「認証情報」
4. 「認証情報を作成」→「APIキー」
5. APIキーを制限（推奨）:
   - HTTPリファラー制限（Webアプリ用）
   - IPアドレス制限（サーバー用）
   - API制限（Maps JavaScript API, Routes API等）

6. Secret Managerに保存:
   ```bash
   echo 'AIzaSyD...' | gcloud secrets versions add google-maps-api-key-dev --data-file=-
   ```

### Gemini API キー

1. [Google AI Studio](https://makersuite.google.com/)にアクセス
2. 「Get API Key」をクリック
3. プロジェクトを選択または新規作成
4. APIキーをコピー

5. Secret Managerに保存:
   ```bash
   echo 'YOUR_GEMINI_KEY' | gcloud secrets versions add gemini-api-key-dev --data-file=-
   ```

## 参考リンク

- [Google Cloud Secret Manager ドキュメント](https://cloud.google.com/secret-manager/docs)
- [Cloud Run での Secret の使用](https://cloud.google.com/run/docs/configuring/secrets)
- [IAM ロールとアクセス権限](https://cloud.google.com/secret-manager/docs/access-control)
- [監査ログの表示](https://cloud.google.com/secret-manager/docs/audit-logging)

---

**最終更新**: 2025-12-12
**担当者**: Claude Sonnet 4.5
