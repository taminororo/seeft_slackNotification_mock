package repository

import (
    "database/sql"
    "encoding/json"
)

type ActionLogRepository struct {
    db *sql.DB
}

func NewActionLogRepository(db *sql.DB) *ActionLogRepository {
    return &ActionLogRepository{db: db}
}

func (r *ActionLogRepository) Create(tx *sql.Tx, shiftID int, actionType string, diffPayload interface{}) error {
    payloadBytes, err := json.Marshal(diffPayload)
    if err != nil {
        return err
    }
    
    query := `INSERT INTO action_log (shift_id, action_type, diff_payload) VALUES ($1, $2, $3)`
    // トランザクション(tx)を使用
    _, err = tx.Exec(query, shiftID, actionType, payloadBytes)
    return err
}