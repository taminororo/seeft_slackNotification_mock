# Docker環境セットアップガイド

## 前提条件

- Docker Desktop または Docker Engine がインストールされていること
- Docker Compose が利用可能であること

## クイックスタート

### 1. 環境変数の設定

プロジェクトルートに`.env`ファイルを作成し、以下の内容を設定してください：

```env
# Database Configuration
DB_HOST=db
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=seeft_shift

# Slack Configuration
SLACK_BOT_TOKEN=xoxb-your-bot-token-here
SLACK_CHANNEL_ID=C1234567890

# Server Configuration
API_PORT=8080

# CORS Configuration
CORS_ALLOW_ORIGINS=http://localhost:3000,http://localhost:8080

# Flutter Configuration
FLUTTER_WEB_PORT=3000
FLUTTER_API_BASE_URL=http://localhost:8080
```

### 2. サービス起動

```bash
# 全サービスを起動
make up

# または
docker-compose up -d
```

### 3. データベースマイグレーション

```bash
# マイグレーションを実行
make migrate-up
```

### 4. アクセス

- **Go API**: http://localhost:8080
- **Flutter Web**: http://localhost:3000

## サービス構成

### Go バックエンド (`go`)

- **ポート**: 8080
- **ホットリロード**: Airを使用（コード変更時に自動再起動）
- **データベース接続**: Dockerネットワーク内の`db`サービスに接続

### PostgreSQL (`db`)

- **ポート**: 5432
- **データ永続化**: Dockerボリューム `postgres_data`
- **初期化**: `backend/database/migrations` ディレクトリのSQLファイルを自動実行

### Flutter Web (`flutter`)

- **ポート**: 3000
- **ホットリロード**: Flutterの開発モードで実行
- **API接続**: 環境変数 `FLUTTER_API_BASE_URL` で設定（デフォルト: `http://localhost:8080`）

## 開発時の注意点

### Goバックエンドのホットリロード

`backend/.air.toml`でAirの設定を管理しています。コード変更時に自動で再ビルド・再起動されます。

### Flutter Webの開発

Flutterコンテナが重い場合は、フロントエンドのみローカルで実行できます：

```bash
make frontend-local
```

この場合、`.env`の`FLUTTER_API_BASE_URL`を`http://localhost:8080`に設定してください。

### CORS設定

Flutter WebからGo APIへのリクエストを許可するため、`CORS_ALLOW_ORIGINS`環境変数にFlutterのURLを追加してください。

例：
```env
CORS_ALLOW_ORIGINS=http://localhost:3000,http://localhost:8080,http://127.0.0.1:3000
```

## トラブルシューティング

### データベース接続エラー

- `DB_HOST=db`（Dockerサービス名）が設定されているか確認
- データベースコンテナが起動しているか確認: `docker-compose ps`

### CORSエラー

- `CORS_ALLOW_ORIGINS`にFlutterのURLが含まれているか確認
- ブラウザの開発者ツールでエラーメッセージを確認

### Flutterコンテナが起動しない

- Flutter SDKのダウンロードに時間がかかる場合があります（初回ビルド時）
- ログを確認: `make logs-flutter`

### ポート競合

- 既に使用されているポートがある場合、`.env`でポート番号を変更してください

## よく使うコマンド

```bash
# 全サービスを停止
make down

# ログを確認
make logs

# 特定のサービスのログ
make logs-go
make logs-flutter
make logs-db

# 全サービスを再起動
make restart

# 全コンテナとボリュームを削除（データも削除）
make clean
```

