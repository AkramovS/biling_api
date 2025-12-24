package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateLogin     = errors.New("duplicate login")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// AuthUser represents a system authentication user
type AuthUser struct {
	ID        int64     `json:"id"`
	Login     string    `json:"login"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

// AuthUserModel wraps database connection
type AuthUserModel struct {
	DB *sql.DB
}

// Insert creates a new auth user
func (m AuthUserModel) Insert(login, password string) (*AuthUser, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO system_accounts (login, password)
		VALUES ($1, $2)
		RETURNING id, login`

	args := []interface{}{login, hash}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user AuthUser

	err = m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.Login,
	)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "system_accounts_login_key"`:
			return nil, ErrDuplicateLogin
		default:
			return nil, err
		}
	}

	return &user, nil
}

// GetByLogin fetches an auth user by login
func (m AuthUserModel) GetByLogin(login string) (*AuthUser, error) {
	query := `
		SELECT id, login, password
		FROM system_accounts
		WHERE login = $1 AND is_deleted = 0`

	var user AuthUser

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

// Authenticate verifies login and password
func (m AuthUserModel) Authenticate(login, password string) (*AuthUser, error) {
	user, err := m.GetByLogin(login)
	if err != nil {
		// Dummy bcrypt для константного времени ответа
		bcrypt.CompareHashAndPassword(
			[]byte("$2a$12$000000000000000000000000000000000000000000000000000000"),
			[]byte(password),
		)
		if errors.Is(err, ErrRecordNotFound) {
			return nil, ErrInvalidCredentials // Тот же ответ, что и при неверном пароле
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	return user, nil
}
