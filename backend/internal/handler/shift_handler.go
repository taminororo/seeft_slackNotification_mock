package handler

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"seeft-slack-notification/internal/model"
	"seeft-slack-notification/internal/repository"
	"seeft-slack-notification/internal/service"
)

type ShiftHandler struct {
	shiftService        *service.ShiftService
	slackService        *service.SlackService
	shiftRepo           *repository.ShiftRepository
	notificationRepo    *repository.NotificationRepository
	userRepo            *repository.UserRepository
}

func NewShiftHandler(
	shiftService *service.ShiftService,
	slackService *service.SlackService,
	shiftRepo *repository.ShiftRepository,
	notificationRepo *repository.NotificationRepository,
	userRepo *repository.UserRepository,
) *ShiftHandler {
	return &ShiftHandler{
		shiftService:     shiftService,
		slackService:     slackService,
		shiftRepo:        shiftRepo,
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
	}
}

// UpdateShifts GASからのPOSTリクエストを処理
func (h *ShiftHandler) UpdateShifts(c echo.Context) error {
	var req model.ShiftChangeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// シフト変更を処理し、差分を検知
	results, err := h.shiftService.ProcessShiftChanges(req.Changes)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// 差分があったレコードをDBに保存し、通知を送信
	notificationCount := 0
	for _, result := range results {
		if result.ChangeType == "no_change" {
			continue
		}

		var shift *model.Shift
		if result.ChangeType == "create" {
			// 新規作成
			shift, err = h.shiftRepo.Create(
				result.Change.YearID,
				result.Change.TimeID,
				result.Change.Date,
				result.Change.Weather,
				result.UserID,
				result.NewTaskName,
			)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": fmt.Sprintf("failed to create shift: %v", err),
				})
			}
		} else if result.ChangeType == "update" {
			// 更新
			shift, err = h.shiftRepo.Update(result.ExistingShiftID, result.NewTaskName)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": fmt.Sprintf("failed to update shift: %v", err),
				})
			}
		}

		// 通知を作成
		notification, err := h.notificationRepo.Create(
			result.UserID,
			shift.ID,
			result.Change.YearID,
			result.Change.TimeID,
			result.Change.Date,
			result.Change.Weather,
			result.OldTaskName,
			result.NewTaskName,
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("failed to create notification: %v", err),
			})
		}

		// ユーザー情報を取得（ユーザー名とSlackUserID）
		user, err := h.userRepo.GetByID(result.UserID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("failed to get user: %v", err),
			})
		}

		// Slack通知を送信
		if err := h.slackService.SendNotification(notification, user.Name, user.SlackUserID); err != nil {
			// 通知送信エラーはログに記録するが、処理は続行
			// 実際の実装ではロガーを使用
			_ = err
		} else {
			// 送信成功したらフラグを更新
			_ = h.notificationRepo.MarkDMSent(notification.ID)
			_ = h.notificationRepo.MarkChannelSent(notification.ID)
		}

		notificationCount++
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":                "success",
		"notifications_created": notificationCount,
	})
}

