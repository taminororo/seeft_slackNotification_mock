package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	SlackBotToken   string
	SlackChannelID  string
	APIPort         string
	CORSAllowOrigins []string
}

func LoadConfig() (*Config, error) {
	// .envファイルを読み込む（存在しない場合は無視）
	_ = godotenv.Load()
	_ = godotenv.Load("../.env") // プロジェクトルートの.envも試す

	// CORS許可オリジンをパース
	corsOriginsStr := getEnv("CORS_ALLOW_ORIGINS", "http://localhost:3000,http://localhost:8080")
	corsOrigins := []string{}
	if corsOriginsStr != "" {
		// カンマ区切りで分割
		parts := strings.Split(corsOriginsStr, ",")
		for _, origin := range parts {
			if trimmed := strings.TrimSpace(origin); trimmed != "" {
				corsOrigins = append(corsOrigins, trimmed)
			}
		}
	}

	config := &Config{
		DBHost:          getEnv("DB_HOST", "localhost"),
		DBPort:          getEnv("DB_PORT", "5432"),
		DBUser:          getEnv("DB_USER", "postgres"),
		DBPassword:      getEnv("DB_PASSWORD", "postgres"),
		DBName:          getEnv("DB_NAME", "seeft_shift"),
		SlackBotToken:   getEnv("SLACK_BOT_TOKEN", ""),
		SlackChannelID:  getEnv("SLACK_CHANNEL_ID", ""),
		APIPort:         getEnv("API_PORT", "8080"),
		CORSAllowOrigins: corsOrigins,
	}

	// 必須項目のチェック
	if config.SlackBotToken == "" {
		return nil, fmt.Errorf("SLACK_BOT_TOKEN is required")
	}
	if config.SlackChannelID == "" {
		return nil, fmt.Errorf("SLACK_CHANNEL_ID is required")
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

