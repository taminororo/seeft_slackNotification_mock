package model

import "time"

type ShiftRead struct {
    ID        int       `json:"id" db:"id"`
    UserID    int       `json:"user_id" db:"user_id"`
    ShiftID   int       `json:"shift_id" db:"shift_id"`
    IsRead    bool      `json:"is_read" db:"is_read"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}