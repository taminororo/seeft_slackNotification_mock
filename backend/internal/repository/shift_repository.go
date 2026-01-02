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
func (r *ShiftRepository) Create(tx *sql.Tx, shift *model.Shift) error {
	query := `INSERT INTO shifts (year_id, time_id, date, weather, user_id, task_name) 
	          VALUES ($1, $2, $3, $4, $5, $6) 
	          RETURNING id, created_at, updated_at`

	err := tx.QueryRow(query, shift.YearID, shift.TimeID, shift.Date, shift.Weather, shift.UserID, shift.TaskName).Scan(
		&shift.ID,
		&shift.CreatedAt,
		&shift.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create shift: %w", err)
	}

	return nil
}

// Update シフトを更新
func (r *ShiftRepository) Update(tx *sql.Tx, shift *model.Shift) error {
	query := `UPDATE shifts 
	          SET task_name = $1, weather = $2, updated_at = CURRENT_TIMESTAMP 
	          WHERE id = $3`

	_, err := tx.Exec(query, shift.TaskName, shift.Weather, shift.ID)
	if err != nil {
		return fmt.Errorf("failed to update shift: %w", err)
	}

	return nil
}

// GetAll(): deleted_at IS NULLでフィルタリング
// 有効なシフトだけを取得
func (r *ShiftRepository) GetAll() ([]*model.Shift, error) {
	query := `SELECT id, year_id, time_id, date, weather, user_id, task_name, created_at, updated_at, deleted_at
	FROM shifts
	WHERE deleted_at IS NULL`

	// 複数のシフトレコードを取得
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get shifts: %w", err)
	}
	defer rows.Close()

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

// Delete 論理削除（deleted_atを設定）
func (r *ShiftRepository) Delete(tx *sql.Tx, shiftID int) error {
	query := `UPDATE shifts 
	          SET deleted_at = CURRENT_TIMESTAMP 
	          WHERE id = $1`

	_, err := tx.Exec(query, shiftID)
	if err != nil {
		return fmt.Errorf("failed to delete shift ID %d: %w", shiftID, err)
	}

	return nil
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
