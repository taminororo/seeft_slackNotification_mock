package main

import (
	"fmt"
	"log"

	"seeft-slack-notification/internal/config"
	"seeft-slack-notification/internal/database"
	"seeft-slack-notification/internal/handler"
	"seeft-slack-notification/internal/repository"
	"seeft-slack-notification/internal/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// 設定を読み込む
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// データベース接続
	db, err := database.NewDBConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 1. リポジトリの初期化
	// ActionLogRepository が必要になったので追加します
	userRepo := repository.NewUserRepository(db)
	shiftRepo := repository.NewShiftRepository(db)
	//notificationRepo := repository.NewNotificationRepository(db)
	actionLogRepo := repository.NewActionLogRepository(db) // ★追加
	shiftReadRepo := repository.NewShiftReadRepository(db) // ★追加

	// 2. サービスの初期化
	// SlackServiceを先に作ります
	slackService := service.NewSlackService(cfg)

	// ShiftServiceには、DB(トランザクション用)と、ログRepo、SlackServiceなど全てを渡します
	shiftService := service.NewShiftService(
		db,
		shiftRepo,
		userRepo,
		actionLogRepo,
		slackService,
		shiftReadRepo,
	)

	// 3. ハンドラーの初期化
	// ShiftHandlerは Service だけを受け取るシンプルな形になりました
	shiftHandler := handler.NewShiftHandler(shiftService)

	// 他のハンドラー（変更なし）
	//notificationHandler := handler.NewNotificationHandler(notificationRepo)
	//readHandler := handler.NewReadHandler(notificationRepo)

	// Echoインスタンスの作成
	e := echo.New()

	// ミドルウェア
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// CORS設定
	corsConfig := middleware.CORSConfig{
		AllowOrigins: cfg.CORSAllowOrigins,
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}
	e.Use(middleware.CORSWithConfig(corsConfig))

	// ルーティング
	api := e.Group("/api")
	api.POST("/update_shifts", shiftHandler.UpdateShifts)
	//api.GET("/notifications", notificationHandler.GetNotifications)
	//api.POST("/notifications/:id/read", readHandler.MarkAsRead)

	// サーバー起動
	port := fmt.Sprintf(":%s", cfg.APIPort)
	if err := e.Start(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
