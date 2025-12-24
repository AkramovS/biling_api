package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Users              UserModel
	Accounts           AccountModel
	AuthUsers          AuthUserModel
	Groups             GroupModel
	Permissions        PermissionModel
	Tokens             TokenModel
	AccountTariffLinks AccountTariffLinkModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Users:              UserModel{DB: db},
		Accounts:           AccountModel{DB: db},
		AuthUsers:          AuthUserModel{DB: db},
		Groups:             GroupModel{DB: db},
		Permissions:        PermissionModel{DB: db},
		Tokens:             TokenModel{},
		AccountTariffLinks: AccountTariffLinkModel{DB: db},
	}
}
