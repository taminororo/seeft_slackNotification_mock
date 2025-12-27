.PHONY: up down build restart logs clean

# Docker Composeで全サービスを起動
up:
	docker-compose up -d

# Docker Composeで全サービスを停止
down:
	docker-compose down

# Docker Composeで全サービスをビルド
build:
	docker-compose build

# Docker Composeで全サービスを再起動
restart:
	docker-compose restart

# ログを表示
logs:
	docker-compose logs -f

# Goサービスのログのみ表示
logs-go:
	docker-compose logs -f go

# Flutterサービスのログのみ表示
logs-flutter:
	docker-compose logs -f flutter

# DBサービスのログのみ表示
logs-db:
	docker-compose logs -f db

# 全コンテナとボリュームを削除（データも削除される）
clean:
	docker-compose down -v

# マイグレーションを実行（golang-migrateを使用する場合）
migrate-up:
	docker-compose exec go migrate -path /app/database/migrations -database "postgres://postgres:postgres@db:5432/seeft_shift?sslmode=disable" up

migrate-down:
	docker-compose exec go migrate -path /app/database/migrations -database "postgres://postgres:postgres@db:5432/seeft_shift?sslmode=disable" down

# フロントエンドのみローカルで実行（Dockerを使わない）
frontend-local:
	cd frontend && flutter run -d chrome --web-port=3000

