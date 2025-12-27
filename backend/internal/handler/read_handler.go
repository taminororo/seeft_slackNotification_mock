package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"seeft-slack-notification/internal/repository"
)

type ReadHandler struct {
	notificationRepo *repository.NotificationRepository
}

func NewReadHandler(notificationRepo *repository.NotificationRepository) *ReadHandler {
	return &ReadHandler{
		notificationRepo: notificationRepo,
	}
}

// MarkAsRead 通知を既読にする
func (h *ReadHandler) MarkAsRead(c echo.Context) error {
	// URLパラメータからnotification_idを取得
	notificationIDStr := c.Param("id")
	notificationID, err := strconv.Atoi(notificationIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid notification id",
		})
	}

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

	if err := h.notificationRepo.MarkAsRead(notificationID, userID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "success",
	})
}

