# シフト変更通知システム

Google Sheets + GASから送信されるシフト変更データをGoバックエンドで受信し、PostgreSQLに保存。差分検知後にSlack通知（DM + チャンネル）を送信し、Flutterアプリで変更を表示・既読管理するシステムです。

## プロジェクト構成

```
seeft_slackNotification_mock/
├── backend/          # Goバックエンド（Echo）
├── frontend/         # Flutterアプリ
├── .env.example      # 環境変数テンプレート
└── README.md
```

## セットアップ

### Docker環境でのセットアップ（推奨）

#### 1. 環境変数の設定

プロジェクトルートに`.env`ファイルを作成し、以下の環境変数を設定してください。

```bash
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

**注意**: Docker環境では、`DB_HOST=db`（Dockerサービス名）を使用します。

#### 2. Docker Composeで全サービスを起動

```bash
# 全サービスを起動
make up
# または
docker-compose up -d

# ログを確認
make logs
# または
docker-compose logs -f
```

#### 3. データベースのマイグレーション

```bash
# マイグレーションを実行（golang-migrateを使用する場合）
make migrate-up

# または、手動でSQLファイルを実行
docker-compose exec db psql -U postgres -d seeft_shift -f /docker-entrypoint-initdb.d/001_create_users_table.up.sql
docker-compose exec db psql -U postgres -d seeft_shift -f /docker-entrypoint-initdb.d/002_create_shifts_table.up.sql
docker-compose exec db psql -U postgres -d seeft_shift -f /docker-entrypoint-initdb.d/003_create_notifications_table.up.sql
```

#### 4. ユーザーデータの投入

```bash
docker-compose exec db psql -U postgres -d seeft_shift -c "INSERT INTO users (name, slack_user_id) VALUES ('山田太郎', 'U1234567890'), ('佐藤花子', 'U0987654321');"
```

#### 5. アクセス

- **Go API**: http://localhost:8080
- **Flutter Web**: http://localhost:3000

### ローカル環境でのセットアップ

#### 1. 環境変数の設定

`.env.example`をコピーして`.env`を作成し、必要な値を設定してください。

```bash
cp .env.example .env
```

**注意**: ローカル環境では、`DB_HOST=localhost`を使用します。

#### 2. PostgreSQLのセットアップ

PostgreSQLを起動し、データベースを作成します。

```sql
CREATE DATABASE seeft_shift;
```

マイグレーションファイルを実行してテーブルを作成します。

```bash
# マイグレーションツールを使用する場合（例: golang-migrate）
migrate -path backend/database/migrations -database "postgres://user:password@localhost:5432/seeft_shift?sslmode=disable" up
```

または、各マイグレーションファイルを手動で実行してください。

#### 3. ユーザーデータの投入

`users`テーブルにユーザー情報を登録してください。

```sql
INSERT INTO users (name, slack_user_id) VALUES 
  ('山田太郎', 'U1234567890'),
  ('佐藤花子', 'U0987654321');
```

#### 4. Goバックエンドの起動

```bash
cd backend
go mod download
go run cmd/server/main.go
```

#### 5. Flutterアプリの起動

```bash
# Docker環境を使用する場合
make frontend-local

# または、直接実行
cd frontend
flutter pub get
flutter run -d chrome --web-port=3000
```

## Dockerコマンド一覧

```bash
# 全サービスを起動
make up

# 全サービスを停止
make down

# 全サービスをビルド
make build

# 全サービスを再起動
make restart

# ログを表示
make logs
make logs-go      # Goサービスのみ
make logs-flutter # Flutterサービスのみ
make logs-db      # DBサービスのみ

# 全コンテナとボリュームを削除（データも削除される）
make clean

# マイグレーション
make migrate-up   # マイグレーション実行
make migrate-down # マイグレーションロールバック

# フロントエンドのみローカルで実行
make frontend-local
```

## APIエンドポイント

### POST /api/update_shifts

GASからシフト変更データを受信します。

**リクエスト例:**
```json
{
  "changes": [
    {
      "yearID": 43,
      "timeID": 25,
      "date": "1日目",
      "weather": "晴れ",
      "userName": "山田太郎",
      "taskName": "受付"
    }
  ]
}
```

**レスポンス例:**
```json
{
  "status": "success",
  "notifications_created": 1
}
```

### GET /api/notifications?user_id={user_id}

未読通知一覧を取得します。

**レスポンス例:**
```json
{
  "notifications": [
    {
      "id": 1,
      "user_name": "山田太郎",
      "year_id": 43,
      "time_id": 25,
      "date": "1日目",
      "weather": "晴れ",
      "old_task_name": "受付",
      "new_task_name": "案内",
      "is_read": false,
      "created_at": "2024-01-01T12:00:00Z"
    }
  ]
}
```

### POST /api/notifications/:id/read?user_id={user_id}

通知を既読にします。

**レスポンス例:**
```json
{
  "status": "success"
}
```

## 技術スタック

- **Go**: 1.21+
- **Echo**: v4.x
- **PostgreSQL**: 14+
- **Slack SDK**: github.com/slack-go/slack
- **Flutter**: 3.x

