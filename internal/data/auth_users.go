package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail     = errors.New("duplicate email")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// AuthUser represents a system authentication user
type AuthUser struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// AuthUserModel wraps database connection
type AuthUserModel struct {
	DB *sql.DB
}

// Insert creates a new auth user
func (m AuthUserModel) Insert(email, password string) (*AuthUser, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO auth_users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email, created_at`

	args := []interface{}{email, hash}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user AuthUser
	user.PasswordHash = string(hash)

	err = m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.Email,
		&user.CreatedAt,
	)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "auth_users_email_key"`:
			return nil, ErrDuplicateEmail
		default:
			return nil, err
		}
	}

	return &user, nil
}

// GetByEmail fetches an auth user by email
func (m AuthUserModel) GetByEmail(email string) (*AuthUser, error) {
	query := `
		SELECT id, email, password_hash, created_at
		FROM auth_users
		WHERE email = $1`

	var user AuthUser

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
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

// Authenticate verifies email and password
func (m AuthUserModel) Authenticate(email, password string) (*AuthUser, error) {
	user, err := m.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	return user, nil
}
