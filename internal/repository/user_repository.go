package repository

import (
	"database/sql"
	"time"

	"github.com/auth-service/internal/models"
	"github.com/google/uuid"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, display_name, is_active, is_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(query, user.ID, user.Email, user.PasswordHash, user.DisplayName,
		user.IsActive, user.IsVerified, user.CreatedAt, user.UpdatedAt)
	return err
}

func (r *UserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, display_name, is_active, is_verified, 
			   created_at, updated_at, last_login_at, failed_login_count, locked_until
		FROM users WHERE id = $1
	`
	user := &models.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName,
		&user.IsActive, &user.IsVerified, &user.CreatedAt, &user.UpdatedAt,
		&user.LastLoginAt, &user.FailedLoginCount, &user.LockedUntil,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, display_name, is_active, is_verified, 
			   created_at, updated_at, last_login_at, failed_login_count, locked_until
		FROM users WHERE email = $1
	`
	user := &models.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName,
		&user.IsActive, &user.IsVerified, &user.CreatedAt, &user.UpdatedAt,
		&user.LastLoginAt, &user.FailedLoginCount, &user.LockedUntil,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) Update(user *models.User) error {
	query := `
		UPDATE users SET email = $1, display_name = $2, is_active = $3, is_verified = $4,
			   updated_at = $5, last_login_at = $6, failed_login_count = $7, locked_until = $8
		WHERE id = $9
	`
	_, err := r.db.Exec(query, user.Email, user.DisplayName, user.IsActive, user.IsVerified,
		time.Now(), user.LastLoginAt, user.FailedLoginCount, user.LockedUntil, user.ID)
	return err
}

func (r *UserRepository) UpdatePassword(userID uuid.UUID, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Exec(query, passwordHash, time.Now(), userID)
	return err
}

func (r *UserRepository) Delete(id uuid.UUID) error {
	_, err := r.db.Exec("DELETE FROM users WHERE id = $1", id)
	return err
}

func (r *UserRepository) List(page, perPage int, search string) ([]models.User, int64, error) {
	offset := (page - 1) * perPage
	var total int64

	countQuery := "SELECT COUNT(*) FROM users"
	if search != "" {
		countQuery += " WHERE email ILIKE $1 OR display_name ILIKE $1"
		r.db.QueryRow(countQuery, "%"+search+"%").Scan(&total)
	} else {
		r.db.QueryRow(countQuery).Scan(&total)
	}

	query := `
		SELECT id, email, password_hash, display_name, is_active, is_verified, 
			   created_at, updated_at, last_login_at, failed_login_count, locked_until
		FROM users
	`
	var rows *sql.Rows
	var err error

	if search != "" {
		query += " WHERE email ILIKE $1 OR display_name ILIKE $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3"
		rows, err = r.db.Query(query, "%"+search+"%", perPage, offset)
	} else {
		query += " ORDER BY created_at DESC LIMIT $1 OFFSET $2"
		rows, err = r.db.Query(query, perPage, offset)
	}

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(
			&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName,
			&user.IsActive, &user.IsVerified, &user.CreatedAt, &user.UpdatedAt,
			&user.LastLoginAt, &user.FailedLoginCount, &user.LockedUntil,
		); err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	return users, total, nil
}

func (r *UserRepository) IncrementFailedLogin(userID uuid.UUID) error {
	query := `UPDATE users SET failed_login_count = failed_login_count + 1 WHERE id = $1`
	_, err := r.db.Exec(query, userID)
	return err
}

func (r *UserRepository) LockAccount(userID uuid.UUID, until time.Time) error {
	query := `UPDATE users SET locked_until = $1 WHERE id = $2`
	_, err := r.db.Exec(query, until, userID)
	return err
}

func (r *UserRepository) ResetFailedLogin(userID uuid.UUID) error {
	query := `UPDATE users SET failed_login_count = 0, locked_until = NULL, last_login_at = $1 WHERE id = $2`
	_, err := r.db.Exec(query, time.Now(), userID)
	return err
}

func (r *UserRepository) SetVerified(userID uuid.UUID) error {
	query := `UPDATE users SET is_verified = true, updated_at = $1 WHERE id = $2`
	_, err := r.db.Exec(query, time.Now(), userID)
	return err
}
