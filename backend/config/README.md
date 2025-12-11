# Configuration Files

このディレクトリには、アプリケーションの設定ファイルを配置します。

## Firebase設定ファイル

### 1. Firebase Consoleから秘密鍵を取得

1. [Firebase Console](https://console.firebase.google.com/)にアクセス
2. プロジェクト「stunning-grin-480914-n1」を選択
3. ⚙️ (設定) → プロジェクトの設定
4. 「サービスアカウント」タブを選択
5. 「新しい秘密鍵の生成」ボタンをクリック
6. JSONファイルがダウンロードされます

### 2. ファイルの配置

ダウンロードしたJSONファイルを `firebase-config.json` という名前で、このディレクトリに配置します:

```bash
# macOSの場合
mv ~/Downloads/stunning-grin-480914-n1-firebase-adminsdk-*.json \
   backend/config/firebase-config.json

# ファイルの権限を設定(セキュリティ)
chmod 600 backend/config/firebase-config.json
```

### 3. .gitignoreの確認

**重要**: `firebase-config.json` はGitリポジトリにコミットしないでください。

`.gitignore` に以下が含まれていることを確認:

```
backend/config/firebase-config.json
backend/config/*.json
!backend/config/*.example.json
```

### 4. 環境変数での指定

`.env` ファイルで、Firebase設定ファイルのパスを指定します:

```bash
FIREBASE_CONFIG_PATH=./config/firebase-config.json
```

## その他の設定ファイル

### API Keys設定

機微な情報は環境変数で管理し、このディレクトリには配置しません:

- `GEMINI_API_KEY`: Gemini APIキー
- `GOOGLE_MAPS_API_KEY`: Google Maps APIキー
- `CLOUD_KMS_KEY_NAME`: Cloud KMS暗号鍵名

## セキュリティガイドライン

1. **秘密鍵は絶対にコミットしない**
   - firebase-config.json
   - API keys
   - データベース認証情報

2. **ファイル権限の設定**
   ```bash
   chmod 600 backend/config/firebase-config.json
   ```

3. **本番環境では環境変数を使用**
   - Cloud Runの場合: Secret Managerと統合
   - ローカル開発の場合: .envファイル

4. **定期的なキーローテーション**
   - Firebase秘密鍵: 90日ごと
   - API keys: 180日ごと

## トラブルシューティング

### firebase-config.jsonが見つからない

```bash
# ファイルの存在を確認
ls -la backend/config/firebase-config.json

# ない場合は、firebase-config.example.jsonをテンプレートとして使用
cp backend/config/firebase-config.example.json \
   backend/config/firebase-config.json

# 実際の認証情報を入力
nano backend/config/firebase-config.json
```

### JSONの形式エラー

```bash
# JSONの形式を検証
cat backend/config/firebase-config.json | python3 -m json.tool

# または
cat backend/config/firebase-config.json | jq .
```

### 権限エラー

```bash
# サービスアカウントに必要なロールを付与
gcloud projects add-iam-policy-binding stunning-grin-480914-n1 \
  --member="serviceAccount:firebase-adminsdk-xxxxx@stunning-grin-480914-n1.iam.gserviceaccount.com" \
  --role="roles/firebase.admin"
```
