package repository

import (
	"database/sql"

	"github.com/auth-service/internal/models"
	"github.com/google/uuid"
)

type RoleRepository struct {
	db *sql.DB
}

func NewRoleRepository(db *sql.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) Create(role *models.Role) error {
	query := `INSERT INTO roles (name, description) VALUES ($1, $2) RETURNING id`
	return r.db.QueryRow(query, role.Name, role.Description).Scan(&role.ID)
}

func (r *RoleRepository) GetByID(id int) (*models.Role, error) {
	query := `SELECT id, name, description FROM roles WHERE id = $1`
	role := &models.Role{}
	err := r.db.QueryRow(query, id).Scan(&role.ID, &role.Name, &role.Description)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *RoleRepository) GetByName(name string) (*models.Role, error) {
	query := `SELECT id, name, description FROM roles WHERE name = $1`
	role := &models.Role{}
	err := r.db.QueryRow(query, name).Scan(&role.ID, &role.Name, &role.Description)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *RoleRepository) Update(role *models.Role) error {
	query := `UPDATE roles SET name = $1, description = $2 WHERE id = $3`
	_, err := r.db.Exec(query, role.Name, role.Description, role.ID)
	return err
}

func (r *RoleRepository) Delete(id int) error {
	_, err := r.db.Exec("DELETE FROM roles WHERE id = $1", id)
	return err
}

func (r *RoleRepository) List() ([]models.Role, error) {
	query := `SELECT id, name, description FROM roles ORDER BY id`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *RoleRepository) AssignRoleToUser(userID uuid.UUID, roleID int, assignedBy uuid.UUID) error {
	query := `
		INSERT INTO user_roles (user_id, role_id, assigned_by, assigned_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id, role_id) DO NOTHING
	`
	_, err := r.db.Exec(query, userID, roleID, assignedBy)
	return err
}

func (r *RoleRepository) UnassignRoleFromUser(userID uuid.UUID, roleID int) error {
	query := `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`
	_, err := r.db.Exec(query, userID, roleID)
	return err
}

func (r *RoleRepository) GetUserRoles(userID uuid.UUID) ([]models.Role, error) {
	query := `
		SELECT r.id, r.name, r.description
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *RoleRepository) UserHasRole(userID uuid.UUID, roleName string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM user_roles ur
			INNER JOIN roles r ON ur.role_id = r.id
			WHERE ur.user_id = $1 AND r.name = $2
		)
	`
	var exists bool
	err := r.db.QueryRow(query, userID, roleName).Scan(&exists)
	return exists, err
}
