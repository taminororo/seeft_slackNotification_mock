package model

// Notification 通知データ
type Notification struct {
	ID              int    `json:"id"`
	UserID          int    `json:"user_id"`
	ShiftID         int    `json:"shift_id"`
	YearID          int    `json:"year_id"`
	TimeID          int    `json:"time_id"`
	Date            string `json:"date"`
	Weather         string `json:"weather"`
	OldTaskName     string `json:"old_task_name"`
	NewTaskName     string `json:"new_task_name"`
	IsRead          bool   `json:"is_read"`
	SlackDMSent     bool   `json:"slack_dm_sent"`
	SlackChannelSent bool   `json:"slack_channel_sent"`
	CreatedAt       string `json:"created_at"`
}

// NotificationResponse APIレスポンス用
type NotificationResponse struct {
	ID          int    `json:"id"`
	UserName    string `json:"user_name"`
	YearID      int    `json:"year_id"`
	TimeID      int    `json:"time_id"`
	Date        string `json:"date"`
	Weather     string `json:"weather"`
	OldTaskName string `json:"old_task_name"`
	NewTaskName string `json:"new_task_name"`
	IsRead      bool   `json:"is_read"`
	CreatedAt   string `json:"created_at"`
}

