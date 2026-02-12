package repository

import (
	"auth-service/internal/entity"
	"context"
	"database/sql"
	"errors"
	"time"
)

// RefreshTokenRepository defines refresh token repository interface
type RefreshTokenRepository interface {
	SaveRefreshToken(ctx context.Context, token *entity.RefreshToken) error
	GetRefreshToken(ctx context.Context, token string) (*entity.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, token string) error
	RevokeAllRefreshTokensByUserID(ctx context.Context, userID string) error
}

// RefreshTokenRepositoryImpl implements RefreshTokenRepository
type RefreshTokenRepositoryImpl struct {
	db *sql.DB
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(db *sql.DB) RefreshTokenRepository {
	return &RefreshTokenRepositoryImpl{db: db}
}

// SaveRefreshToken saves a refresh token
func (r *RefreshTokenRepositoryImpl) SaveRefreshToken(ctx context.Context, token *entity.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (token, user_id, tenant_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(ctx, query,
		token.Token, token.UserID, token.TenantID, token.ExpiresAt, token.CreatedAt)
	return err
}

// GetRefreshToken gets a refresh token
func (r *RefreshTokenRepositoryImpl) GetRefreshToken(ctx context.Context, token string) (*entity.RefreshToken, error) {
	query := `
		SELECT token, user_id, tenant_id, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE token = $1
	`

	row := r.db.QueryRowContext(ctx, query, token)
	var rt entity.RefreshToken
	err := row.Scan(&rt.Token, &rt.UserID, &rt.TenantID, &rt.ExpiresAt, &rt.RevokedAt, &rt.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("refresh token not found")
		}
		return nil, err
	}

	return &rt, nil
}

// RevokeRefreshToken revokes a refresh token
func (r *RefreshTokenRepositoryImpl) RevokeRefreshToken(ctx context.Context, token string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = $1
		WHERE token = $2
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), token)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("refresh token not found")
	}

	return nil
}

// RevokeAllRefreshTokensByUserID revokes all refresh tokens for a user
func (r *RefreshTokenRepositoryImpl) RevokeAllRefreshTokensByUserID(ctx context.Context, userID string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = $1
		WHERE user_id = $2 AND revoked_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	return err
}
