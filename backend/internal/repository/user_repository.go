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

// GetAll 全ユーザーを取得する
func (r *UserRepository) GetAll() ([]*model.User, error) {
	// 1. 全ユーザーを取得するシンプルなクエリ
	query := `SELECT id, name, slack_user_id, created_at, updated_at FROM users`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all users: %w", err)
	}
	defer rows.Close()

	var users []*model.User

	// 2. 1行ずつ取り出してリストに追加
	for rows.Next() {
		var u model.User
		if err := rows.Scan(
			&u.ID,
			&u.Name,
			&u.SlackUserID,
			&u.CreatedAt,
			&u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &u)
	}

	// 3. エラーチェック (ループ終了後の確認)
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return users, nil
}
