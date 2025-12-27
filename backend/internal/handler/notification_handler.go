package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"seeft-slack-notification/internal/model"
	"seeft-slack-notification/internal/repository"
)

type NotificationHandler struct {
	notificationRepo *repository.NotificationRepository
}

func NewNotificationHandler(notificationRepo *repository.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{
		notificationRepo: notificationRepo,
	}
}

// GetNotifications 未読通知一覧を取得
func (h *NotificationHandler) GetNotifications(c echo.Context) error {
	// クエリパラメータからuser_idを取得（実際の実装では認証から取得）
	userIDStr := c.QueryParam("user_id")
	if userIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "user_id is required",
		})
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid user_id",
		})
	}

	notifications, err := h.notificationRepo.GetUnreadByUserID(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// nilの場合は空のスライスに変換して、必ず空の配列[]を返すようにする
	if notifications == nil {
		notifications = []model.NotificationResponse{}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"notifications": notifications,
	})
}

