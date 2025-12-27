package repository

import (
	"database/sql"
	"fmt"

	"seeft-slack-notification/internal/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByName ユーザー名でユーザーを取得
func (r *UserRepository) GetByName(name string) (*model.User, error) {
	query := `SELECT id, name, slack_user_id, created_at, updated_at 
	          FROM users WHERE name = $1`
	
	var user model.User
	err := r.db.QueryRow(query, name).Scan(
		&user.ID,
		&user.Name,
		&user.SlackUserID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found: %s", name)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return &user, nil
}

// GetByID IDでユーザーを取得
func (r *UserRepository) GetByID(id int) (*model.User, error) {
	query := `SELECT id, name, slack_user_id, created_at, updated_at 
	          FROM users WHERE id = $1`
	
	var user model.User
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Name,
		&user.SlackUserID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found: id=%d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return &user, nil
}

