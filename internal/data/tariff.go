package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// AccountTariffLink represents a tariff assignment to an account
type AccountTariffLink struct {
	ID        int64     `json:"id"`
	AccountID int64     `json:"account_id"`
	TariffID  int64     `json:"tariff_id"`
	Version   int64     `json:"version"`
	UpdatedAt time.Time `json:"updated_at"`
	UpdatedBy *int64    `json:"updated_by,omitempty"`
}

// UpdatedByUser contains info about who made the last change
type UpdatedByUser struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
}

// AccountTariffLinkWithUser includes the user who made last update
type AccountTariffLinkWithUser struct {
	AccountTariffLink
	UpdatedByUser *UpdatedByUser `json:"updated_by_user,omitempty"`
}

// AccountTariffLinkModel handles database operations for tariff links
type AccountTariffLinkModel struct {
	DB *sql.DB
}

// Get retrieves an account tariff link by ID with updater info
func (m AccountTariffLinkModel) Get(id int64) (*AccountTariffLinkWithUser, error) {
	query := `
		SELECT 
			atl.id, atl.account_id, atl.tariff_id, 
			atl.version, atl.updated_at, atl.updated_by,
			au.id, au.login
		FROM account_tariff_link atl
		LEFT JOIN system_accounts au ON atl.updated_by = au.id
		WHERE atl.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var link AccountTariffLinkWithUser
	var updatedByID sql.NullInt64
	var userID sql.NullInt64
	var userLogin sql.NullString

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&link.ID,
		&link.AccountID,
		&link.TariffID,
		&link.Version,
		&link.UpdatedAt,
		&updatedByID,
		&userID,
		&userLogin,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	if updatedByID.Valid {
		link.UpdatedBy = &updatedByID.Int64
	}

	if userID.Valid && userLogin.Valid {
		link.UpdatedByUser = &UpdatedByUser{
			ID:    userID.Int64,
			Login: userLogin.String,
		}
	}

	return &link, nil
}

// GetByAccountID retrieves tariff link by account ID
func (m AccountTariffLinkModel) GetByAccountID(accountID int64) (*AccountTariffLinkWithUser, error) {
	query := `
		SELECT 
			atl.id, atl.account_id, atl.tariff_id, 
			atl.version, atl.updated_at, atl.updated_by,
			au.id, au.login
		FROM account_tariff_link atl
		LEFT JOIN system_accounts au ON atl.updated_by = au.id
		WHERE atl.account_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var link AccountTariffLinkWithUser
	var updatedByID sql.NullInt64
	var userID sql.NullInt64
	var userLogin sql.NullString

	err := m.DB.QueryRowContext(ctx, query, accountID).Scan(
		&link.ID,
		&link.AccountID,
		&link.TariffID,
		&link.Version,
		&link.UpdatedAt,
		&updatedByID,
		&userID,
		&userLogin,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	if updatedByID.Valid {
		link.UpdatedBy = &updatedByID.Int64
	}

	if userID.Valid && userLogin.Valid {
		link.UpdatedByUser = &UpdatedByUser{
			ID:    userID.Int64,
			Login: userLogin.String,
		}
	}

	return &link, nil
}

// Update changes the tariff with optimistic locking
// Returns ErrEditConflict if version doesn't match
func (m AccountTariffLinkModel) Update(link *AccountTariffLink) error {
	query := `
		UPDATE account_tariff_link
		SET 
			tariff_id = $1,
			version = version + 1,
			updated_at = NOW(),
			updated_by = $2
		WHERE id = $3 AND version = $4
		RETURNING version, updated_at`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query,
		link.TariffID,
		link.UpdatedBy,
		link.ID,
		link.Version,
	).Scan(&link.Version, &link.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrEditConflict
		}
		return err
	}

	return nil
}
