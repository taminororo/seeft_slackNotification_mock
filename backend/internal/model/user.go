package model

// User ユーザーデータ
type User struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	SlackUserID string `json:"slack_user_id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

