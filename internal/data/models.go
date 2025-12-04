package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

// Models wraps all data models
type Models struct {
	Users     UserModel
	Accounts  AccountModel
	AuthUsers AuthUserModel
	Groups    GroupModel
	Tokens    TokenModel
}

// NewModels creates a new Models instance
func NewModels(db *sql.DB) Models {
	return Models{
		Users:     UserModel{DB: db},
		Accounts:  AccountModel{DB: db},
		AuthUsers: AuthUserModel{DB: db},
		Groups:    GroupModel{DB: db},
		Tokens:    TokenModel{},
	}
}
