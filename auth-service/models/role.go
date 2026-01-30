package models

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Role struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type RoleModel struct {
	db *pgxpool.Pool
}

func NewRoleModel(db *pgxpool.Pool) *RoleModel {
	return &RoleModel{db: db}
}

// GetRoleByName retrieves a role by name
func (m *RoleModel) GetRoleByName(ctx context.Context, name string) (*Role, error) {
	var role Role
	query := `
		SELECT id, name, description, created_at
		FROM roles
		WHERE name = $1
	`

	err := m.db.QueryRow(ctx, query, name).
		Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &role, nil
}

// GetUserRoles retrieves all roles for a user
func (m *RoleModel) GetUserRoles(ctx context.Context, userID int64) ([]Role, error) {
	query := `
		SELECT r.id, r.name, r.description, r.created_at
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
	`

	rows, err := m.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// AssignRoleToUser assigns a role to a user
func (m *RoleModel) AssignRoleToUser(ctx context.Context, userID, roleID int64) error {
	query := `
		INSERT INTO user_roles (user_id, role_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`

	_, err := m.db.Exec(ctx, query, userID, roleID)
	return err
}

// RemoveRoleFromUser removes a role from a user
func (m *RoleModel) RemoveRoleFromUser(ctx context.Context, userID, roleID int64) error {
	query := `
		DELETE FROM user_roles
		WHERE user_id = $1 AND role_id = $2
	`

	_, err := m.db.Exec(ctx, query, userID, roleID)
	return err
}

// HasRole checks if a user has a specific role
func (m *RoleModel) HasRole(ctx context.Context, userID int64, roleName string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM user_roles ur
			INNER JOIN roles r ON ur.role_id = r.id
			WHERE ur.user_id = $1 AND r.name = $2
		)
	`

	var exists bool
	err := m.db.QueryRow(ctx, query, userID, roleName).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
