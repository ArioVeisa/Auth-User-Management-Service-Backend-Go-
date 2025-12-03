package repository

import (
	"database/sql"
	"time"

	"github.com/auth-service/internal/models"
	"github.com/google/uuid"
)

type TokenRepository struct {
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) CreateRefreshToken(token *models.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, issued_at, expires_at, revoked, user_agent, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(query, token.ID, token.UserID, token.TokenHash, token.IssuedAt,
		token.ExpiresAt, token.Revoked, token.UserAgent, token.IPAddress)
	return err
}

func (r *TokenRepository) GetRefreshTokenByHash(hash string) (*models.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, issued_at, expires_at, revoked, user_agent, ip_address
		FROM refresh_tokens WHERE token_hash = $1
	`
	token := &models.RefreshToken{}
	err := r.db.QueryRow(query, hash).Scan(
		&token.ID, &token.UserID, &token.TokenHash, &token.IssuedAt,
		&token.ExpiresAt, &token.Revoked, &token.UserAgent, &token.IPAddress,
	)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *TokenRepository) RevokeRefreshToken(id uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *TokenRepository) RevokeAllUserTokens(userID uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE user_id = $1`
	_, err := r.db.Exec(query, userID)
	return err
}

func (r *TokenRepository) CreateEmailToken(token *models.EmailToken) error {
	query := `
		INSERT INTO email_tokens (id, user_id, token_hash, type, expires_at, used)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(query, token.ID, token.UserID, token.TokenHash,
		token.Type, token.ExpiresAt, token.Used)
	return err
}

func (r *TokenRepository) GetEmailTokenByHash(hash string, tokenType models.EmailTokenType) (*models.EmailToken, error) {
	query := `
		SELECT id, user_id, token_hash, type, expires_at, used
		FROM email_tokens WHERE token_hash = $1 AND type = $2
	`
	token := &models.EmailToken{}
	err := r.db.QueryRow(query, hash, tokenType).Scan(
		&token.ID, &token.UserID, &token.TokenHash, &token.Type,
		&token.ExpiresAt, &token.Used,
	)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *TokenRepository) MarkEmailTokenUsed(id uuid.UUID) error {
	query := `UPDATE email_tokens SET used = true WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *TokenRepository) CleanupExpiredTokens() error {
	now := time.Now()
	_, err := r.db.Exec("DELETE FROM refresh_tokens WHERE expires_at < $1", now)
	if err != nil {
		return err
	}
	_, err = r.db.Exec("DELETE FROM email_tokens WHERE expires_at < $1", now)
	return err
}
