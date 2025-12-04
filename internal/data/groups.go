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

// Permission represents a group permission
type Permission struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// GroupModel wraps database connection
type GroupModel struct {
	DB *sql.DB
}

// HasPermission checks if a user has a specific permission through their groups
func (m GroupModel) HasPermission(authUserID int64, resource, action string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM group_permissions gp
		INNER JOIN group_members gm ON gm.group_id = gp.group_id
		WHERE gm.auth_user_id = $1
		  AND gp.resource = $2
		  AND gp.action = $3`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int
	err := m.DB.QueryRowContext(ctx, query, authUserID, resource, action).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetUserPermissions fetches all permissions for a user
func (m GroupModel) GetUserPermissions(authUserID int64) ([]Permission, error) {
	query := `
		SELECT DISTINCT gp.resource, gp.action
		FROM group_permissions gp
		INNER JOIN group_members gm ON gm.group_id = gp.group_id
		WHERE gm.auth_user_id = $1
		ORDER BY gp.resource, gp.action`

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
		err := rows.Scan(&p.Resource, &p.Action)
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
