package repository

import (
	"database/sql"
	"fmt"
)

type ActionLogRepository struct {
	db *sql.DB
}

func NewActionLogRepository(db *sql.DB) *ActionLogRepository {
	return &ActionLogRepository{db: db}
}

func (r *ActionLogRepository) Create(tx *sql.Tx, shiftID int, actionType string, diffPayload interface{}) error {
	// payloadBytes, err := json.Marshal(diffPayload)
	// if err != nil {
	//     return err
	// }

	query := `
		INSERT INTO action_log (shift_id, action_type, payload)
		VALUES ($1, $2, $3)`

	// トランザクション(tx)を使用
	_, err := tx.Exec(query, shiftID, actionType, diffPayload)
	if err != nil {
		return fmt.Errorf("failed to create action log: %w", err)
	}

	return err
}
