package handler

import (
	"net/http"

	"seeft-slack-notification/internal/model"
	"seeft-slack-notification/internal/service"

	"github.com/labstack/echo/v4"
)

type ShiftHandler struct {
	shiftService *service.ShiftService
	// SlackServiceなども、Handlerで直接使わないなら削除してOKです
	// ただし、別のAPIで使うかもしれないので残しておいても害はありません
}

// NewShiftHandler コンストラクタもシンプルになります
func NewShiftHandler(shiftService *service.ShiftService) *ShiftHandler {
	return &ShiftHandler{
		shiftService: shiftService,
	}
}

// UpdateShifts GASからのPOSTリクエストを処理
func (h *ShiftHandler) UpdateShifts(c echo.Context) error {
	// 1. JSONを受け取る
	var req model.ShiftChangeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// 2. サービスに「同期」を依頼する (ここでDB更新もログ保存も通知予約も全部やる！)
	// ※ SyncShiftsの引数が []model.ShiftChange である前提です
	if err := h.shiftService.SyncShifts(req.Changes); err != nil {
		// エラーなら500を返す
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// 3. 成功レスポンスを返す
	// 通知の件数などは非同期処理になったため、即座には分かりません（「受け付けました」というスタンス）
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Shift sync started",
	})
}
