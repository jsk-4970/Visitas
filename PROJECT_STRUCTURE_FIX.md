# プロジェクト構造の修正計画

## 現状の問題

### 発見された構造の不整合

**現在の実際の構造**:
```
Visitas/
├── backend/              # 空のディレクトリ（誤って作成）
│   ├── backend/          # さらに重複
│   ├── config/
│   ├── internal/
│   ├── migrations/
│   ├── pkg/
│   └── tests/
├── cmd/                  # ルート直下に配置（本来はbackend/内）
├── config/               # ルート直下に配置
├── internal/             # ルート直下に配置
├── migrations/           # ルート直下に配置
├── pkg/                  # ルート直下に配置
├── tests/                # ルート直下に配置
├── go.mod                # Goプロジェクトルート
├── go.sum
└── ...
```

**CLAUDE.mdで定義された正しい構造**:
```
Visitas/
├── backend/              # Goバックエンドサービス
│   ├── cmd/
│   │   └── api/
│   │       └── main.go
│   ├── internal/
│   │   ├── handlers/
│   │   ├── services/
│   │   ├── models/
│   │   └── repository/
│   ├── pkg/
│   ├── migrations/
│   ├── config/
│   ├── tests/
│   ├── go.mod
│   └── go.sum
├── mobile/               # Flutterアプリ
├── docs/                 # ドキュメント
└── infra/                # IaC
```

## 問題の原因

1. go.modが `Visitas/backend/` ではなく `Visitas/` 直下にある
2. それに伴い、cmd/, internal/, pkg/ 等がすべてルート直下に配置されている
3. 誤って `backend/` ディレクトリを作成し、さらに重複構造ができた

## 修正方針

### オプション1: go.modの配置を変更（推奨）

Goプロジェクトを `backend/` 配下に移動し、CLAUDE.mdの定義通りの構造にする。

**メリット**:
- CLAUDE.mdの定義と完全に一致
- monorepo構造として明確
- mobile/, docs/, infra/ との分離が明確

**デメリット**:
- go.modの移動が必要
- 既存のimport pathの変更が必要

**手順**:
```bash
cd /Users/yukinaribaba/Desktop/Visitas

# 1. 新しいbackendディレクトリを作成
mkdir -p backend_new

# 2. 必要なファイルを移動
mv go.mod go.sum backend_new/
mv cmd/ internal/ pkg/ migrations/ config/ tests/ scripts/ backend_new/
mv backend/.env.example backend/.gitignore backend/.dockerignore backend/Dockerfile backend_new/ 2>/dev/null

# 3. 古いbackendディレクトリを削除
rm -rf backend/

# 4. リネーム
mv backend_new/ backend/

# 5. go.modの確認
cd backend
go mod tidy
```

### オプション2: 現状を正式化（簡易）

CLAUDE.mdの定義を変更し、現在の構造を正式なものとする。

**メリット**:
- ファイル移動が不要
- すぐに実施可能

**デメリット**:
- monorepo構造として不明確
- 一般的なGoプロジェクト構造と異なる

**手順**:
```bash
# 1. 重複ディレクトリの削除のみ
rm -rf backend/

# 2. CLAUDE.mdの構造定義を更新
# 3. README.mdの構造説明を更新
```

## 推奨: オプション1を実施

理由:
- 将来的にmobile/, web/, infra/を追加する予定
- monorepoとして明確な構造が必要
- CLAUDE.mdの定義との整合性が重要

## 実施タイミング

**即時実施が可能**: まだコードが少なく、依存関係が単純な段階

## 実施後の確認事項

1. ✅ go.modが `backend/` 直下にある
2. ✅ `go mod tidy` が正常に動作
3. ✅ import pathが正しい
4. ✅ CLAUDE.mdの構造と一致
5. ✅ README.mdの構造説明が正しい
6. ✅ ドキュメント内のパス参照が正しい

## 修正コマンド（完全版）

```bash
#!/bin/bash
set -e

cd /Users/yukinaribaba/Desktop/Visitas

echo "=== プロジェクト構造の修正を開始 ==="

# バックアップ作成
echo "バックアップを作成中..."
tar -czf visitas_backup_$(date +%Y%m%d_%H%M%S).tar.gz . --exclude='.git' --exclude='node_modules'

# 新しいbackendディレクトリを作成
echo "新しいbackend構造を作成中..."
mkdir -p backend_new

# ファイルを移動
echo "ファイルを移動中..."
mv go.mod go.sum backend_new/ 2>/dev/null || true
mv cmd/ backend_new/ 2>/dev/null || true
mv internal/ backend_new/ 2>/dev/null || true
mv pkg/ backend_new/ 2>/dev/null || true
mv migrations/ backend_new/ 2>/dev/null || true
mv config/ backend_new/ 2>/dev/null || true
mv tests/ backend_new/ 2>/dev/null || true
mv scripts/ backend_new/ 2>/dev/null || true

# backend/から必要なファイルを移動
[ -f backend/.env.example ] && mv backend/.env.example backend_new/
[ -f backend/.gitignore ] && mv backend/.gitignore backend_new/
[ -f backend/.dockerignore ] && mv backend/.dockerignore backend_new/
[ -f backend/Dockerfile ] && mv backend/Dockerfile backend_new/

# 古いbackendディレクトリを削除
echo "古いbackendディレクトリを削除中..."
rm -rf backend/

# リネーム
echo "新しい構造を適用中..."
mv backend_new/ backend/

# 確認
echo "=== 構造の確認 ==="
ls -la backend/

echo "=== go mod tidy を実行 ==="
cd backend
go mod tidy || echo "⚠️ Goがインストールされていないため、go mod tidyをスキップしました"

echo "✅ プロジェクト構造の修正が完了しました"
echo "次のステップ: docs/SETUP.md の手順に従ってGoをインストールしてください"
```

## 手動での実施（Goがインストールされていない場合）

```bash
cd /Users/yukinaribaba/Desktop/Visitas

# 1. バックアップ（念のため）
tar -czf ~/visitas_backup_$(date +%Y%m%d_%H%M%S).tar.gz .

# 2. 重複ディレクトリの確認と削除
ls -la backend/backend/ && rm -rf backend/backend/

# 3. 当面は現状の構造を維持
# （Goインストール後にオプション1を実施）
```

## 次のステップ

1. この修正計画をレビュー
2. オプション1またはオプション2を選択
3. 修正を実施
4. 整合性を再確認
5. ドキュメントを更新

---

**作成日**: 2025-12-12
**ステータス**: 承認待ち
