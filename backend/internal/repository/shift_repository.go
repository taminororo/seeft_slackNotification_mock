package repository

import (
	"database/sql"
	"fmt"

	"seeft-slack-notification/internal/model"
)

type ShiftRepository struct {
	db *sql.DB
}

func NewShiftRepository(db *sql.DB) *ShiftRepository {
	return &ShiftRepository{db: db}
}

// GetByUniqueKey ユニークキーでシフトを取得
func (r *ShiftRepository) GetByUniqueKey(yearID, timeID int, date, weather string, userID int) (*model.Shift, error) {
	query := `SELECT id, year_id, time_id, date, weather, user_id, task_name, created_at, updated_at 
	          FROM shifts 
	          WHERE year_id = $1 AND time_id = $2 AND date = $3 AND weather = $4 AND user_id = $5`
	
	var shift model.Shift
	err := r.db.QueryRow(query, yearID, timeID, date, weather, userID).Scan(
		&shift.ID,
		&shift.YearID,
		&shift.TimeID,
		&shift.Date,
		&shift.Weather,
		&shift.UserID,
		&shift.TaskName,
		&shift.CreatedAt,
		&shift.UpdatedAt,
		&shift.DeletedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil // レコードが存在しない場合はnilを返す
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get shift: %w", err)
	}
	
	return &shift, nil
}

// Create 新規シフトを作成
func (r *ShiftRepository) Create(yearID, timeID int, date, weather string, userID int, taskName string) (*model.Shift, error) {
	query := `INSERT INTO shifts (year_id, time_id, date, weather, user_id, task_name) 
	          VALUES ($1, $2, $3, $4, $5, $6) 
	          RETURNING id, year_id, time_id, date, weather, user_id, task_name, created_at, updated_at`
	
	var shift model.Shift
	err := r.db.QueryRow(query, yearID, timeID, date, weather, userID, taskName).Scan(
		&shift.ID,
		&shift.YearID,
		&shift.TimeID,
		&shift.Date,
		&shift.Weather,
		&shift.UserID,
		&shift.TaskName,
		&shift.CreatedAt,
		&shift.UpdatedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create shift: %w", err)
	}
	
	return &shift, nil
}

// Update シフトを更新
func (r *ShiftRepository) Update(shiftID int, taskName string) (*model.Shift, error) {
	query := `UPDATE shifts 
	          SET task_name = $1, updated_at = CURRENT_TIMESTAMP 
	          WHERE id = $2 
	          RETURNING id, year_id, time_id, date, weather, user_id, task_name, created_at, updated_at`
	
	var shift model.Shift
	err := r.db.QueryRow(query, taskName, shiftID).Scan(
		&shift.ID,
		&shift.YearID,
		&shift.TimeID,
		&shift.Date,
		&shift.Weather,
		&shift.UserID,
		&shift.TaskName,
		&shift.CreatedAt,
		&shift.UpdatedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to update shift: %w", err)
	}
	
	return &shift, nil
}

// GetAll(): deleted_at IS NULLでフィルタリング
// 有効なシフトだけを取得
func (r *ShiftRepository) GetAll() ([]*model.Shift, error) {
	query :=  `SELECT id, year_id, time_id, date, weather, user_id, task_name, created_at, deleted_at
	FROM shifts
	WHERE deleted_at IS NULL`

	// 複数のシフトレコードを取得
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get shifts: %w", err)
	}
	defer rows.Close()

	// 
	var shifts []*model.Shift
	for rows.Next() {
		var shift model.Shift
		err := rows.Scan(
			&shift.ID,
			&shift.YearID,
			&shift.TimeID,
			&shift.Date,
			&shift.Weather,
			&shift.UserID,
			&shift.TaskName,
			&shift.CreatedAt,
			&shift.UpdatedAt,
			&shift.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan shift: %w", err)
		}
		shifts = append(shifts, &shift)
	}

	// 4. ループの途中でエラーが起きていなかったか確認
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return shifts, nil
}
//Delete(shiftID int): 論理削除（deleted_atを設定）
func (s *ShiftService) DeleteShift(id int) error {
    // 1. トランザクション開始（仮書き込みモードスタート！）
    // db ではなく tx という新しい「管理者」が生まれます
    tx, err := s.db.Begin()
    if err != nil {
        return err
    }

    // ★重要：万が一パニック（強制終了）などが起きても、必ず「破棄」するように予約しておく
    defer tx.Rollback()

    // 2. シフト削除（tx を渡すことで「仮書き込み」にする）
    if err := s.shiftRepo.Delete(tx, id); err != nil {
        return err // ここで関数が終わると、deferしたRollbackが動いて元通り！
    }

    // 3. ログ保存（tx を渡すことで、↑と同じグループの処理にする）
    if err := s.logRepo.Create(tx, id, "DELETE", nil); err != nil {
        return err // ここで失敗しても、シフト削除ごと無かったことになる！
    }

    // 4. 全部うまくいったので、最後に「送信（確定）」！
    return tx.Commit()
}

// GetByUserID 指定したユーザーの有効なシフトを、既読状態付きで取得する
func (r *ShiftRepository) GetByUserID(userID int) ([]*model.ShiftWithReadStatus, error) {
    // 1. SQLの組み立て
    // s.* : シフトの全カラム
    // COALESCE(sr.is_read, false) : 既読レコードがない場合は false (未読) とする
    query := `
        SELECT 
            s.id, s.year_id, s.time_id, s.date, s.weather, s.user_id, s.task_name, 
            s.created_at, s.updated_at, s.deleted_at,
            COALESCE(sr.is_read, false) as is_read
        FROM shifts s
        LEFT JOIN shift_reads sr 
            ON s.id = sr.shift_id AND sr.user_id = $1
        WHERE 
            s.user_id = $1 
            AND s.deleted_at IS NULL
        ORDER BY s.date ASC, s.time_id ASC`

    // 2. クエリ実行
    rows, err := r.db.Query(query, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to query shifts for user %d: %w", userID, err)
    }
    defer rows.Close()

    // 3. データの詰め替え
    var shifts []*model.ShiftWithReadStatus

    for rows.Next() {
        // Shift構造体を初期化
        s := &model.Shift{}
        var isRead bool

        // Scan: DBの列の並び順通りに変数を指定する
        err := rows.Scan(
            &s.ID,
            &s.YearID,
            &s.TimeID,
            &s.Date,
            &s.Weather,
            &s.UserID,
            &s.TaskName,
            &s.CreatedAt,
            &s.UpdatedAt,
            &s.DeletedAt,
            &isRead, // 追加したカラム
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan shift: %w", err)
        }

        // 拡張した箱に入れてリストに追加
        shifts = append(shifts, &model.ShiftWithReadStatus{
            Shift:  s,
            IsRead: isRead,
        })
    }

    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("rows error: %w", err)
    }

    return shifts, nil
}