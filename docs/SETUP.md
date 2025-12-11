# Visitas 開発環境セットアップガイド

このドキュメントでは、Visitasプロジェクトの開発環境をセットアップする手順を説明します。

## 前提条件

- macOS 26.1以降
- インターネット接続

## 1. Homebrewのインストール

Homebrewは、macOS用のパッケージマネージャーです。

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

インストール後、パスを通します:

```bash
# Apple Silicon (M1/M2/M3)の場合
echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zprofile
eval "$(/opt/homebrew/bin/brew shellenv)"

# Intelプロセッサの場合
echo 'eval "$(/usr/local/bin/brew shellenv)"' >> ~/.zprofile
eval "$(/usr/local/bin/brew shellenv)"
```

## 2. Goのインストール

プロジェクトはGo 1.22以降を要求します。

```bash
# Homebrewを使用してGoをインストール
brew install go@1.22

# または最新版をインストール
brew install go

# インストールの確認
go version
# 出力例: go version go1.22.x darwin/arm64
```

### 環境変数の設定

```bash
# ~/.zshrc または ~/.bash_profile に追加
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
export PATH=$PATH:/usr/local/go/bin

# 設定を反映
source ~/.zshrc  # zshの場合
# または
source ~/.bash_profile  # bashの場合
```

## 3. GCP CLIツールのインストール

```bash
# gcloud CLIのインストール
brew install --cask google-cloud-sdk

# 認証
gcloud auth login

# プロジェクトの設定
gcloud config set project stunning-grin-480914-n1

# Application Default Credentials (ADC)の設定
gcloud auth application-default login
```

## 4. Goの依存関係のインストール

```bash
cd backend

# 依存関係をダウンロード
go mod download

# go.modとgo.sumを整理
go mod tidy

# ベンダーディレクトリの作成(オプション)
go mod vendor
```

## 5. 開発ツールのインストール

### ホットリロードツール (Air)

```bash
go install github.com/cosmtrek/air@latest
```

### Lintツール

```bash
# golangci-lintのインストール
brew install golangci-lint

# 使用方法
cd backend
golangci-lint run
```

### その他の推奨ツール

```bash
# goimportsのインストール(コードフォーマット)
go install golang.org/x/tools/cmd/goimports@latest

# mockgenのインストール(モック生成)
go install github.com/golang/mock/mockgen@latest
```

## 6. 環境変数の設定

`.env.example`をコピーして`.env`ファイルを作成します:

```bash
cd backend
cp .env.example .env
```

`.env`ファイルを編集して、必要な環境変数を設定します:

```bash
# GCP設定
GCP_PROJECT_ID=stunning-grin-480914-n1
GCP_REGION=asia-northeast1
SPANNER_INSTANCE=visitas-instance
SPANNER_DATABASE=visitas-db

# Firebase設定
FIREBASE_CONFIG_PATH=./config/firebase-config.json

# API Keys
GEMINI_API_KEY=your_gemini_api_key_here
GOOGLE_MAPS_API_KEY=your_google_maps_api_key_here

# サーバー設定
PORT=8080
LOG_LEVEL=debug
ENV=development
```

## 7. Firebase設定ファイルの配置

1. [Firebase Console](https://console.firebase.google.com/)にアクセス
2. プロジェクトを選択
3. プロジェクト設定 → サービスアカウント
4. 「新しい秘密鍵の生成」をクリック
5. ダウンロードしたJSONファイルを `backend/config/firebase-config.json` に配置

```bash
mkdir -p backend/config
# ダウンロードしたファイルを配置
mv ~/Downloads/stunning-grin-480914-n1-firebase-adminsdk-xxxxx.json \
   backend/config/firebase-config.json
```

## 8. データベースのセットアップ

### ローカル開発用 (PostgreSQL)

```bash
# PostgreSQLのインストール
brew install postgresql@15

# PostgreSQLの起動
brew services start postgresql@15

# データベースの作成
createdb visitas_dev

# マイグレーションの適用
cd backend/migrations
for file in *.sql; do
  psql -U $(whoami) -d visitas_dev -f "$file"
done
```

### Cloud Spanner (本番環境)

```bash
# Spannerインスタンスの作成(初回のみ)
gcloud spanner instances create visitas-instance \
  --config=regional-asia-northeast1 \
  --description="Visitas Production Instance" \
  --nodes=1

# データベースの作成
gcloud spanner databases create visitas-db \
  --instance=visitas-instance \
  --database-dialect=POSTGRESQL

# マイグレーションの適用
cd backend/migrations
gcloud spanner databases ddl update visitas-db \
  --instance=visitas-instance \
  --ddl="$(cat 001_create_patients.sql)"
```

## 9. アプリケーションの起動

### 通常起動

```bash
cd backend
go run cmd/api/main.go
```

### ホットリロード起動(開発時推奨)

```bash
cd backend
air
```

アプリケーションは `http://localhost:8080` で起動します。

## 10. テストの実行

```bash
# すべてのテストを実行
cd backend
go test ./... -v

# カバレッジレポート生成
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
open coverage.html
```

## トラブルシューティング

### Goがインストールされているか確認できない

```bash
# Goのパスを確認
which go

# パスが表示されない場合、PATHを確認
echo $PATH

# Homebrew経由でインストールしたGoのパス
# Apple Silicon: /opt/homebrew/bin/go
# Intel: /usr/local/bin/go
```

### go mod tidyでエラーが発生する

```bash
# キャッシュをクリア
go clean -modcache

# 再度実行
go mod download
go mod tidy
```

### Cloud Spannerへの接続エラー

```bash
# Application Default Credentialsを再設定
gcloud auth application-default login

# IAMロールを確認
gcloud projects get-iam-policy stunning-grin-480914-n1 \
  --flatten="bindings[].members" \
  --format="table(bindings.role)" \
  --filter="bindings.members:$(gcloud config get-value account)"

# 必要なロール:
# - roles/spanner.databaseUser
# - roles/spanner.databaseReader
```

### Firebaseへの接続エラー

```bash
# firebase-config.jsonのパスを確認
ls -la backend/config/firebase-config.json

# 環境変数を確認
echo $FIREBASE_CONFIG_PATH

# JSONファイルの形式を検証
cat backend/config/firebase-config.json | python3 -m json.tool
```

## 次のステップ

1. [API仕様書](./API_SPEC.md)を確認
2. [アーキテクチャ設計](./ARCHITECTURE.md)を読む
3. [データベース要件定義書](./DATABASE_REQUIREMENTS.md)を理解
4. [デプロイ手順](./DEPLOYMENT.md)を確認

## 参考リソース

- [Go公式ドキュメント](https://golang.org/doc/)
- [Cloud Spanner Go Client](https://pkg.go.dev/cloud.google.com/go/spanner)
- [Firebase Admin SDK Go](https://firebase.google.com/docs/admin/setup#go)
- [Chi Router](https://github.com/go-chi/chi)
