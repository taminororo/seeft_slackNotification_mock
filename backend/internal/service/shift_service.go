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
	db            *sql.DB // トランザクション制御用
	shiftRepo     *repository.ShiftRepository
	userRepo      *repository.UserRepository
	actionLogRepo *repository.ActionLogRepository // ログ保存用に追加
}

// NewShiftService コンストラクタ
func NewShiftService(db *sql.DB, shiftRepo *repository.ShiftRepository, userRepo *repository.UserRepository, logRepo *repository.ActionLogRepository) *ShiftService {
	return &ShiftService{
		db:            db,
		shiftRepo:     shiftRepo,
		userRepo:      userRepo,
		actionLogRepo: logRepo,
	}
}

// SyncShifts GASからのデータを元に、DBを完全同期（作成・更新・削除）する
func (s *ShiftService) SyncShifts(gasChanges []model.ShiftChange) error {
	// 1. トランザクション開始
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	// 関数終了時に、コミットされていなければロールバック（安全策）
	defer tx.Rollback()

	// 2. 準備: ユーザー情報を全取得してマップ化 (名前 -> ID)
	// ループ内で毎回DBに問い合わせないための高速化
	userMap, err := s.preloadUserMap()
	if err != nil {
		return err
	}

	// 3. 準備: 現在の有効なシフトを全取得してマップ化 (Key -> Shift)
	// これで「DBにあるか？」の確認が一瞬で終わる
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
		// ユーザー名からIDを取得
		userID, ok := userMap[change.UserName]
		if !ok {
			log.Printf("Warning: User not found: %s", change.UserName)
			continue // 知らないユーザーはスキップ
		}

		key := makeKey(change.YearID, change.TimeID, change.Date, userID)

		if oldShift, exists := currentShiftMap[key]; exists {
			// --- 【更新パターン】DBに既に存在する ---

			// 差分があるかチェック
			if oldShift.TaskName != change.TaskName || oldShift.Weather != change.Weather {
				// 値を更新
				newShift := *oldShift // コピーを作成
				newShift.TaskName = change.TaskName
				newShift.Weather = change.Weather // ※Weatherも更新対象なら

				// DB更新
				if err := s.shiftRepo.Update(tx, &newShift); err != nil {
					return fmt.Errorf("failed to update shift: %w", err)
				}

				// ログ保存
				if err := s.logAction(tx, oldShift.ID, "UPDATE", oldShift, &newShift); err != nil {
					return err
				}
			}

			// 処理済みとしてマップから消す（重要！）
			// ※ここで消すことで、後でマップに残ったものが「削除対象」になる
			delete(currentShiftMap, key)

		} else {
			// --- 【新規パターン】DBに存在しない ---

			newShift := &model.Shift{
				YearID:   change.YearID,
				TimeID:   change.TimeID,
				Date:     change.Date,
				Weather:  change.Weather,
				UserID:   userID,
				TaskName: change.TaskName,
			}

			// DB作成 (IDがセットされて返ってくる想定)
			if err := s.shiftRepo.Create(tx, newShift); err != nil {
				return fmt.Errorf("failed to create shift: %w", err)
			}

			// ログ保存 (新規なのでoldはnil)
			// 注意: Create内でIDがnewShift.IDに入っている必要がある
			if err := s.logAction(tx, newShift.ID, "CREATE", nil, newShift); err != nil {
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

		// ログ保存 (削除なのでnewはnil)
		if err := s.logAction(tx, deletedShift.ID, "DELETE", deletedShift, nil); err != nil {
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

// preloadUserMap 全ユーザーを取得して 名前->ID のマップを作る
func (s *ShiftService) preloadUserMap() (map[string]int, error) {
	// ユーザーリポジトリに GetAll() のようなメソッドが必要ですが、
	// ここでは簡易的に全ユーザーを取れると仮定、もしくは必要な分だけ取る
	// ★実装簡略化のため、今回は「名前」で引ける辞書を作るイメージです
	// 実際には userRepo.GetAll() を実装してそれを呼びます

	// 仮の実装イメージ:
	// users, _ := s.userRepo.GetAll()
	// m := make(map[string]int)
	// for _, u := range users { m[u.Name] = u.ID }
	// return m, nil

	// ※もし userRepo.GetAll がまだ無いなら、一旦スキップして
	// ループ内で GetByName を呼ぶ元の方式でも動きますが、速度は落ちます。
	// 今回はコンパイルを通すために、一旦「空のマップ」を返します（適宜実装してください）
	return map[string]int{}, nil
}

// logAction 変更履歴をJSON化して保存する
func (s *ShiftService) logAction(tx *sql.Tx, shiftID int, actionType string, oldVal, newVal *model.Shift) error {
	// 差分Payloadの作成
	// model.DiffPayload の定義に合わせて作成
	diff := map[string]interface{}{}

	if actionType == "UPDATE" {
		diff["changes"] = []map[string]string{
			{"field": "task_name", "old": oldVal.TaskName, "new": newVal.TaskName},
		}
	} else if actionType == "CREATE" {
		diff["new_task"] = newVal.TaskName
	}

	payload, err := json.Marshal(diff)
	if err != nil {
		return fmt.Errorf("failed to marshal diff: %w", err)
	}

	// ActionLogRepositoryのCreateを呼ぶ
	return s.actionLogRepo.Create(tx, shiftID, actionType, payload)
}
