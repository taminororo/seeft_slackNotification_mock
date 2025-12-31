package model

import (
    "encoding/json"
    "time"
)

type ActionLog struct {
    ID          int             `json:"id" db:"id"`
    ShiftID     int             `json:"shift_id" db:"shift_id"`
    ActionType  string          `json:"action_type" db:"action_type"`
    DiffPayload json.RawMessage `json:"diff_payload" db:"diff_payload"` // JSONB対応
    CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

type DiffPayload struct {
	User string `json:"user"`
	Date string `json:"date"`
	Slot string `json:"slot"`
	Changes []ChangeItem `json:"changes"`
}

type ChangeItem struct {
	Field string `json:"field"`
	Old string `json:"old"`
	New string `json:"new"`
}