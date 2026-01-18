package database

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"
)

type PasswordResetToken struct {
	ID        int
	UserID    int
	Token     string
	ExpiresAt time.Time
	Used      bool
	CreatedAt time.Time
}

// GenerateToken creates a secure random token
func GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreatePasswordResetToken creates a new password reset token for a user
func CreatePasswordResetToken(db *sql.DB, userID int) (*PasswordResetToken, error) {
	token, err := GenerateToken()
	if err != nil {
		return nil, err
	}

	// Token expires in 1 hour
	expiresAt := time.Now().Add(time.Hour)

	query := `
		INSERT INTO password_reset_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	prt := &PasswordResetToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		Used:      false,
	}

	err = db.QueryRow(query, userID, token, expiresAt).Scan(&prt.ID, &prt.CreatedAt)
	if err != nil {
		return nil, err
	}

	return prt, nil
}

// GetPasswordResetToken retrieves a valid, unused token
func GetPasswordResetToken(db *sql.DB, token string) (*PasswordResetToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, used, created_at
		FROM password_reset_tokens
		WHERE token = $1 AND used = false AND expires_at > $2
	`

	prt := &PasswordResetToken{}
	err := db.QueryRow(query, token, time.Now()).Scan(
		&prt.ID,
		&prt.UserID,
		&prt.Token,
		&prt.ExpiresAt,
		&prt.Used,
		&prt.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("invalid or expired token")
	}
	if err != nil {
		return nil, err
	}

	return prt, nil
}

// MarkTokenAsUsed marks a token as used
func MarkTokenAsUsed(db *sql.DB, tokenID int) error {
	query := `
		UPDATE password_reset_tokens
		SET used = true
		WHERE id = $1
	`

	_, err := db.Exec(query, tokenID)
	return err
}

// InvalidateUserTokens invalidates all pending tokens for a user
func InvalidateUserTokens(db *sql.DB, userID int) error {
	query := `
		UPDATE password_reset_tokens
		SET used = true
		WHERE user_id = $1 AND used = false
	`

	_, err := db.Exec(query, userID)
	return err
}

// CleanupExpiredTokens removes old expired tokens
func CleanupExpiredTokens(db *sql.DB) error {
	query := `
		DELETE FROM password_reset_tokens
		WHERE expires_at < $1 OR used = true
	`

	_, err := db.Exec(query, time.Now().Add(-24*time.Hour))
	return err
}
