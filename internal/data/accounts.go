package data

import (
	"context"
	"database/sql"
	"time"
)

// Account represents a user account (ЛС)
type Account struct {
	ID int64 `json:"id"`
}

// AccountModel wraps database connection
type AccountModel struct {
	DB *sql.DB
}

// GetByUserID fetches all accounts for a specific user
func (m AccountModel) GetByUserID(userID int64) ([]*Account, error) {
	query := `
		SELECT a.id
		FROM accounts a
		INNER JOIN users_accounts ua ON ua.account_id = a.id
		WHERE ua.uid = $1
		ORDER BY a.id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := []*Account{}

	for rows.Next() {
		var account Account

		err := rows.Scan(
			&account.ID,
		)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, &account)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return accounts, nil
}
