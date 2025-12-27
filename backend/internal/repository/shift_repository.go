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

