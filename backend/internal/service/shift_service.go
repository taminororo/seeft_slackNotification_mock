package service

import (
	"fmt"

	"seeft-slack-notification/internal/model"
	"seeft-slack-notification/internal/repository"
)

type ShiftService struct {
	shiftRepo *repository.ShiftRepository
	userRepo  *repository.UserRepository
}

func NewShiftService(shiftRepo *repository.ShiftRepository, userRepo *repository.UserRepository) *ShiftService {
	return &ShiftService{
		shiftRepo: shiftRepo,
		userRepo:  userRepo,
	}
}

// ProcessShiftChanges シフト変更を処理し、差分を検知
func (s *ShiftService) ProcessShiftChanges(changes []model.ShiftChange) ([]ShiftChangeResult, error) {
	var results []ShiftChangeResult

	for _, change := range changes {
		// ユーザー名でユーザーIDを取得
		user, err := s.userRepo.GetByName(change.UserName)
		if err != nil {
			// ユーザーが見つからない場合はスキップ（ログに記録）
			continue
		}

		// 既存のシフトを検索
		existingShift, err := s.shiftRepo.GetByUniqueKey(
			change.YearID,
			change.TimeID,
			change.Date,
			change.Weather,
			user.ID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get existing shift: %w", err)
		}

		var result ShiftChangeResult
		result.Change = change
		result.UserID = user.ID
		result.SlackUserID = user.SlackUserID

		if existingShift == nil {
			// 新規作成
			result.ChangeType = "create"
			result.OldTaskName = ""
			result.NewTaskName = change.TaskName
		} else if existingShift.TaskName != change.TaskName {
			// 変更検知
			result.ChangeType = "update"
			result.OldTaskName = existingShift.TaskName
			result.NewTaskName = change.TaskName
			result.ExistingShiftID = existingShift.ID
		} else {
			// 変更なし
			result.ChangeType = "no_change"
			continue
		}

		results = append(results, result)
	}

	return results, nil
}

// ShiftChangeResult シフト変更処理の結果
type ShiftChangeResult struct {
	Change          model.ShiftChange
	UserID          int
	SlackUserID     string
	ChangeType      string // "create", "update", "no_change"
	OldTaskName     string
	NewTaskName     string
	ExistingShiftID int // updateの場合のみ
}

