package data

import (
	"context"
	"database/sql"
	"time"
)

// FID константы — идентификаторы функций API
// Соответствуют значениям fid в таблице system_rights
const (
	FIDAccountsRead  int64 = 1 // Чтение аккаунтов
	FIDTariffsRead   int64 = 2 // Чтение тарифов
	FIDTariffsUpdate int64 = 3 // Обновление тарифов
)

// PermissionModel обрабатывает операции с правами
type PermissionModel struct {
	DB *sql.DB
}

// HasPermission проверяет, есть ли у пользователя право (fid) через его группы
func (m PermissionModel) HasPermission(userID, fid int64) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM system_groups sg
			INNER JOIN system_rights sr ON sg.group_id = sr.group_id
			WHERE sg.user_id = $1 AND sr.fid = $2
		)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var hasPermission bool
	err := m.DB.QueryRowContext(ctx, query, userID, fid).Scan(&hasPermission)
	if err != nil {
		return false, err
	}

	return hasPermission, nil
}

// GetUserPermissions возвращает все FID пользователя
func (m PermissionModel) GetUserPermissions(userID int64) ([]int64, error) {
	query := `
		SELECT DISTINCT sr.fid
		FROM system_groups sg
		INNER JOIN system_rights sr ON sg.group_id = sr.group_id
		WHERE sg.user_id = $1
		ORDER BY sr.fid`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []int64
	for rows.Next() {
		var fid int64
		if err := rows.Scan(&fid); err != nil {
			return nil, err
		}
		permissions = append(permissions, fid)
	}

	return permissions, rows.Err()
}
