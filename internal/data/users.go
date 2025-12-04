package data

import (
	"context"
	"database/sql"
	"time"
)

// User represents a business user
type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// UserModel wraps database connection
type UserModel struct {
	DB *sql.DB
}

// Get fetches a user by ID
func (m UserModel) Get(id int64) (*User, error) {
	query := `
		SELECT id, name
		FROM users
		WHERE id = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
	)

	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}
