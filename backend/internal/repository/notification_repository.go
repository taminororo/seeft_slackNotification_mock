package repository

import (
	"database/sql"
	"fmt"

	"seeft-slack-notification/internal/model"
)

type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create 通知を作成
func (r *NotificationRepository) Create(userID, shiftID, yearID, timeID int, date, weather, oldTaskName, newTaskName string) (*model.Notification, error) {
	query := `INSERT INTO notifications 
	          (user_id, shift_id, year_id, time_id, date, weather, old_task_name, new_task_name) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
	          RETURNING id, user_id, shift_id, year_id, time_id, date, weather, old_task_name, new_task_name, is_read, slack_dm_sent, slack_channel_sent, created_at`
	
	var notification model.Notification
	err := r.db.QueryRow(query, userID, shiftID, yearID, timeID, date, weather, oldTaskName, newTaskName).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.ShiftID,
		&notification.YearID,
		&notification.TimeID,
		&notification.Date,
		&notification.Weather,
		&notification.OldTaskName,
		&notification.NewTaskName,
		&notification.IsRead,
		&notification.SlackDMSent,
		&notification.SlackChannelSent,
		&notification.CreatedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}
	
	return &notification, nil
}

// MarkDMSent DM送信済みフラグを更新
func (r *NotificationRepository) MarkDMSent(notificationID int) error {
	query := `UPDATE notifications SET slack_dm_sent = TRUE WHERE id = $1`
	_, err := r.db.Exec(query, notificationID)
	if err != nil {
		return fmt.Errorf("failed to mark dm sent: %w", err)
	}
	return nil
}

// MarkChannelSent チャンネル送信済みフラグを更新
func (r *NotificationRepository) MarkChannelSent(notificationID int) error {
	query := `UPDATE notifications SET slack_channel_sent = TRUE WHERE id = $1`
	_, err := r.db.Exec(query, notificationID)
	if err != nil {
		return fmt.Errorf("failed to mark channel sent: %w", err)
	}
	return nil
}

// GetUnreadByUserID ユーザーの未読通知一覧を取得
func (r *NotificationRepository) GetUnreadByUserID(userID int) ([]model.NotificationResponse, error) {
	query := `SELECT n.id, u.name, n.year_id, n.time_id, n.date, n.weather, n.old_task_name, n.new_task_name, n.is_read, n.created_at
	          FROM notifications n
	          JOIN users u ON n.user_id = u.id
	          WHERE n.user_id = $1 AND n.is_read = FALSE
	          ORDER BY n.created_at DESC`
	
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unread notifications: %w", err)
	}
	defer rows.Close()
	
	// makeで初期化することで、nilではなく空のスライスを返すようにする
	notifications := make([]model.NotificationResponse, 0)
	for rows.Next() {
		var n model.NotificationResponse
		err := rows.Scan(
			&n.ID,
			&n.UserName,
			&n.YearID,
			&n.TimeID,
			&n.Date,
			&n.Weather,
			&n.OldTaskName,
			&n.NewTaskName,
			&n.IsRead,
			&n.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		notifications = append(notifications, n)
	}
	
	return notifications, nil
}

// MarkAsRead 通知を既読にする
func (r *NotificationRepository) MarkAsRead(notificationID, userID int) error {
	query := `UPDATE notifications SET is_read = TRUE WHERE id = $1 AND user_id = $2`
	result, err := r.db.Exec(query, notificationID, userID)
	if err != nil {
		return fmt.Errorf("failed to mark as read: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("notification not found or unauthorized")
	}
	
	return nil
}

