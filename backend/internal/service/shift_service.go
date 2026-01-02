package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"seeft-slack-notification/internal/model"
	"seeft-slack-notification/internal/repository"
)

type ShiftService struct {
	db            *sql.DB
	shiftRepo     *repository.ShiftRepository
	userRepo      *repository.UserRepository
	actionLogRepo *repository.ActionLogRepository
	slackService  *SlackService // ★追加: Slack通知用サービス
	shiftReadRepo *repository.ShiftReadRepository
}

// NewShiftService コンストラクタ
func NewShiftService(
	db *sql.DB,
	shiftRepo *repository.ShiftRepository,
	userRepo *repository.UserRepository,
	logRepo *repository.ActionLogRepository,
	slackService *SlackService, // ★引数に追加
	shiftReadRepo *repository.ShiftReadRepository,
) *ShiftService {
	return &ShiftService{
		db:            db,
		shiftRepo:     shiftRepo,
		userRepo:      userRepo,
		actionLogRepo: logRepo,
		slackService:  slackService,
		shiftReadRepo: shiftReadRepo,
	}
}

// SyncShifts GASからのデータを元に、DBを完全同期（作成・更新・削除）する
func (s *ShiftService) SyncShifts(gasChanges []model.ShiftChange) error {
	// 1. トランザクション開始
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 2. 準備: ユーザー情報を全取得してマップ化 (名前 -> User構造体)
	// 通知用にSlackUserIDも必要なので、IDだけではなくUserごと取得します
	nameToUserMap, err := s.preloadUserMap()
	if err != nil {
		return err
	}

	// 削除通知用に ID -> User のマップも作っておく
	idToUserMap := make(map[int]*model.User)
	for _, u := range nameToUserMap {
		idToUserMap[u.ID] = u
	}

	// 3. 準備: 現在の有効なシフトを全取得してマップ化 (Key -> Shift)
	currentShifts, err := s.shiftRepo.GetAll()
	if err != nil {
		return fmt.Errorf("failed to get all shifts: %w", err)
	}

	// Key: "YearID-TimeID-Date-UserID"
	currentShiftMap := make(map[string]*model.Shift)
	for _, shift := range currentShifts {
		key := makeKey(shift.YearID, shift.TimeID, shift.Date, shift.UserID)
		currentShiftMap[key] = shift
	}

	// 4. GASデータ(gasChanges)をループして「新規」か「更新」を処理
	for _, change := range gasChanges {
		// ユーザー名からUser情報を取得
		user, ok := nameToUserMap[change.UserName]
		if !ok {
			log.Printf("Warning: User not found: %s", change.UserName)
			continue // 知らないユーザーはスキップ
		}

		key := makeKey(change.YearID, change.TimeID, change.Date, user.ID)

		if oldShift, exists := currentShiftMap[key]; exists {
			// --- 【更新パターン】DBに既に存在する ---

			// 差分があるかチェック
			if oldShift.TaskName != change.TaskName || oldShift.Weather != change.Weather {
				newShift := *oldShift // コピーを作成
				newShift.TaskName = change.TaskName
				newShift.Weather = change.Weather

				// DB更新
				if err := s.shiftRepo.Update(tx, &newShift); err != nil {
					return fmt.Errorf("failed to update shift: %w", err)
				}

				// ログ保存 & Slack通知
				if err := s.logAction(tx, oldShift.ID, "UPDATE", oldShift, &newShift, user); err != nil {
					return err
				}
			}

			// 処理済みとしてマップから消す
			delete(currentShiftMap, key)

		} else {
			// --- 【新規パターン】DBに存在しない ---

			newShift := &model.Shift{
				YearID:   change.YearID,
				TimeID:   change.TimeID,
				Date:     change.Date,
				Weather:  change.Weather,
				UserID:   user.ID,
				TaskName: change.TaskName,
			}

			// DB作成 (Create内でnewShift.IDがセットされる想定)
			if err := s.shiftRepo.Create(tx, newShift); err != nil {
				return fmt.Errorf("failed to create shift: %w", err)
			}

			// ★追加: 既読レコードを「未読(false)」で作成
			if err := s.shiftReadRepo.Upsert(tx, newShift.ID, user.ID, false); err != nil {
				return err
			}

			// ログ保存 & Slack通知
			if err := s.logAction(tx, newShift.ID, "CREATE", nil, newShift, user); err != nil {
				return err
			}
		}
	}

	// 5. 【削除パターン】マップに残っているデータ = GASには無かったデータ
	for _, deletedShift := range currentShiftMap {
		// 論理削除
		if err := s.shiftRepo.Delete(tx, deletedShift.ID); err != nil {
			return fmt.Errorf("failed to delete shift ID %d: %w", deletedShift.ID, err)
		}

		// 削除対象のユーザー情報を取得
		user, ok := idToUserMap[deletedShift.UserID]
		if !ok {
			// ユーザーが見つからなくても削除は進めるが、User情報は空のダミーを入れるかログを吐く
			log.Printf("Info: Orphan shift deleted (User ID %d not found). Skipping notification.", deletedShift.UserID)
			continue
		}

		// ログ保存 & Slack通知
		if err := s.logAction(tx, deletedShift.ID, "DELETE", deletedShift, nil, user); err != nil {
			return fmt.Errorf("failed to log delete action: %w", err)
		}
	}

	// 6. 全ての処理が成功したので、コミット（保存確定）
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// --- 以下、ヘルパー関数 ---

// makeKey 比較用の一意なキーを生成
func makeKey(year, time int, date string, user int) string {
	return fmt.Sprintf("%d-%d-%s-%d", year, time, date, user)
}

// preloadUserMap 全ユーザーを取得して 名前->User構造体 のマップを作る
func (s *ShiftService) preloadUserMap() (map[string]*model.User, error) {
	// さっき作った GetAll をここで呼ぶ！
	users, err := s.userRepo.GetAll()
	if err != nil {
		return nil, err
	}

	// 使いやすい辞書（マップ）に変換する
	// "田中" -> {ID: 1, Name: "田中", SlackID: "U12345..."}
	m := make(map[string]*model.User)
	for _, u := range users {
		m[u.Name] = u
	}

	return m, nil
}

// logAction 変更履歴を保存し、Slack通知キューに追加する
// 引数に user (*model.User) を追加しました
func (s *ShiftService) logAction(tx *sql.Tx, shiftID int, actionType string, oldVal, newVal *model.Shift, user *model.User) error {
	// 1. DB用: 差分Payloadの作成
	diff := map[string]interface{}{}

	if actionType == "UPDATE" {
		diff["changes"] = []map[string]string{
			{"field": "task_name", "old": oldVal.TaskName, "new": newVal.TaskName},
		}
	} else if actionType == "CREATE" {
		diff["new_task"] = newVal.TaskName
	} else if actionType == "DELETE" {
		diff["deleted_task"] = oldVal.TaskName
	}

	payloadJSON, err := json.Marshal(diff)
	if err != nil {
		return fmt.Errorf("failed to marshal diff: %w", err)
	}

	// DBにログ保存
	if err := s.actionLogRepo.Create(tx, shiftID, actionType, payloadJSON); err != nil {
		return err
	}

	// 2. Slack通知用: データの準備
	// DELETEの場合は newVal が nil なので、oldVal から情報を取る必要がある
	var targetShift *model.Shift
	var taskName, oldTaskName string

	if newVal != nil {
		targetShift = newVal
		taskName = newVal.TaskName
	} else {
		// DELETEの場合
		targetShift = oldVal
	}

	if oldVal != nil {
		oldTaskName = oldVal.TaskName
	}

	notificationPayload := NotificationPayload{
		ActionType:  actionType,
		UserName:    user.Name,
		SlackUserID: user.SlackUserID, // ここでSlackIDをセット
		Date:        targetShift.Date,
		TimeID:      targetShift.TimeID,
		Weather:     targetShift.Weather,
		TaskName:    taskName,
		OldTaskName: oldTaskName,
	}

	// Slack通知キューに放り込む (非同期)
	s.slackService.EnqueueNotification(notificationPayload)

	return nil
}
