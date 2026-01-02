package model

import (
	"time"
)

// ShiftChangeRequest GASから送信されるリクエストボディ
type ShiftChangeRequest struct {
	Changes []ShiftChange `json:"changes"`
}

// ShiftChange 個別のシフト変更データ (GASとの通信用)
type ShiftChange struct {
	YearID   int    `json:"yearID"`
	TimeID   int    `json:"timeID"`
	Date     string `json:"date"`
	Weather  string `json:"weather"`
	UserName string `json:"userName"`
	TaskName string `json:"taskName"`
}

// Shift DB上のシフトデータ
type Shift struct {
	ID        int       `json:"id" db:"id"`
	YearID    int       `json:"year_id" db:"year_id"`
	TimeID    int       `json:"time_id" db:"time_id"`
	Date      string    `json:"date" db:"date"`
	Weather   string    `json:"weather" db:"weather"`
	UserID    int       `json:"user_id" db:"user_id"`
	TaskName  string    `json:"task_name" db:"task_name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	// 論理削除用：NULLを許容するために *time.Time か sql.NullTime を使用
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"`
}

// ShiftWithReadStatus アプリ表示用：シフト情報に既読ステータスを付与した構造体
type ShiftWithReadStatus struct {
	*Shift      // Shift構造体のフィールドをすべて継承（埋め込み）
	IsRead bool `json:"is_read" db:"is_read"` // 追加で欲しいデータ
}
