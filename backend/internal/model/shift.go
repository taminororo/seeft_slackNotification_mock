package model

// ShiftChangeRequest GASから送信されるリクエストボディの構造体
type ShiftChangeRequest struct {
	Changes []ShiftChange `json:"changes"`
}

// ShiftChange 個別のシフト変更データ
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
	ID        int    `json:"id"`
	YearID    int    `json:"year_id"`
	TimeID    int    `json:"time_id"`
	Date      string `json:"date"`
	Weather   string `json:"weather"`
	UserID    int    `json:"user_id"`
	TaskName  string `json:"task_name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

