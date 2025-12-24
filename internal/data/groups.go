package data

import (
	"context"
	"database/sql"
	"time"
)

// Group represents an access control group
type Group struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Permission represents a group permission (FID from system_rights)
type Permission struct {
	FID int64 `json:"fid"` // Feature ID from system_rights
}

// GroupModel wraps database connection
type GroupModel struct {
	DB *sql.DB
}

// HasPermission checks if a user has a specific permission (fid) through their groups
// Uses system_rights table with fid (feature ID)
func (m GroupModel) HasPermission(authUserID int64, fid int64) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM system_rights sr
		INNER JOIN system_groups sg ON sg.group_id = sr.group_id
		WHERE sg.user_id = $1
		  AND sr.fid = $2
		  AND sg.user_id > 0`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int
	err := m.DB.QueryRowContext(ctx, query, authUserID, fid).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetUserPermissions fetches all permissions (fids) for a user through their groups
func (m GroupModel) GetUserPermissions(authUserID int64) ([]Permission, error) {
	query := `
		SELECT DISTINCT sr.fid
		FROM system_rights sr
		INNER JOIN system_groups sg ON sg.group_id = sr.group_id
		WHERE sg.user_id = $1
		  AND sg.user_id > 0
		ORDER BY sr.fid`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, authUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	permissions := []Permission{}

	for rows.Next() {
		var p Permission
		err := rows.Scan(&p.FID)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

// GetUserGroups fetches all groups that a user belongs to
func (m GroupModel) GetUserGroups(authUserID int64) ([]Group, error) {
	query := `
		SELECT sgi.id, sgi.name, sgi.description
		FROM system_group_info sgi
		INNER JOIN system_groups sg ON sg.group_id = sgi.id
		WHERE sg.user_id = $1
		  AND sg.user_id > 0
		ORDER BY sgi.id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, authUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := []Group{}

	for rows.Next() {
		var g Group
		err := rows.Scan(&g.ID, &g.Name, &g.Description)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}
