package repository

import "database/sql"

type ShiftReadRepository struct {
	db *sql.DB
}

func NewShiftReadRepository(db *sql.DB) *ShiftReadRepository {
	return &ShiftReadRepository{db: db}
}

// Upsert 指定したシフトの既読状態を保存・更新する
// 新規なら作成、既存なら is_read を指定した値(false)で上書きする
func (r *ShiftReadRepository) Upsert(tx *sql.Tx, shiftID, userID int, isRead bool) error {
	query := `
        INSERT INTO shift_reads (shift_id, user_id, is_read)
        VALUES ($1, $2, $3)
        ON CONFLICT (shift_id, user_id) 
        DO UPDATE SET is_read = EXCLUDED.is_read, updated_at = CURRENT_TIMESTAMP`

	_, err := tx.Exec(query, shiftID, userID, isRead)
	return err
}
