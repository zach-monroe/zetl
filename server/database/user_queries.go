package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/lib/pq"
	"github.com/zach-monroe/zetl/server/models"
)

// CreateUser inserts a new user into the database
func CreateUser(ctx context.Context, db *sql.DB, user *models.User) error {
	query := `
		INSERT INTO users (username, email, password_hash, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err := db.QueryRowContext(ctx, query, user.Username, user.Email, user.PasswordHash, user.IsActive).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		// Check for unique constraint violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			if pqErr.Constraint == "users_username_key" {
				return ErrUsernameExists
			}
			if pqErr.Constraint == "users_email_key" {
				return ErrEmailExists
			}
		}
		return err
	}

	return nil
}

// getUserBy is a generic helper to retrieve a user by a specified column
func getUserBy(ctx context.Context, db *sql.DB, whereClause string, arg interface{}) (*models.User, error) {
	user := &models.User{}
	var privacySettingsJSON []byte

	query := `
		SELECT id, username, email, password_hash, COALESCE(bio, ''),
		       COALESCE(privacy_settings::text, '{}')::bytea,
		       created_at, updated_at, last_login, is_active
		FROM users
		WHERE ` + whereClause

	err := db.QueryRowContext(ctx, query, arg).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Bio,
		&privacySettingsJSON,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLogin,
		&user.IsActive,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	user.PrivacySettings, _ = models.ParsePrivacySettings(privacySettingsJSON)
	return user, nil
}

// GetUserByUsername retrieves a user by username
func GetUserByUsername(ctx context.Context, db *sql.DB, username string) (*models.User, error) {
	return getUserBy(ctx, db, "username = $1", username)
}

// GetUserByEmail retrieves a user by email
func GetUserByEmail(ctx context.Context, db *sql.DB, email string) (*models.User, error) {
	return getUserBy(ctx, db, "email = $1", email)
}

// GetUserByID retrieves a user by ID
func GetUserByID(ctx context.Context, db *sql.DB, id int) (*models.User, error) {
	return getUserBy(ctx, db, "id = $1", id)
}

// UpdateLastLogin updates the last_login timestamp for a user
func UpdateLastLogin(ctx context.Context, db *sql.DB, userID int) error {
	query := `
		UPDATE users
		SET last_login = $1
		WHERE id = $2
	`

	_, err := db.ExecContext(ctx, query, time.Now(), userID)
	return err
}

// UpdateUserProfile updates user's username, email, and bio
func UpdateUserProfile(ctx context.Context, db *sql.DB, userID int, username, email, bio string) error {
	query := `
		UPDATE users
		SET username = $1, email = $2, bio = $3, updated_at = $4
		WHERE id = $5
	`

	_, err := db.ExecContext(ctx, query, username, email, bio, time.Now(), userID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			if pqErr.Constraint == "users_username_key" {
				return ErrUsernameExists
			}
			if pqErr.Constraint == "users_email_key" {
				return ErrEmailExists
			}
		}
		return err
	}
	return nil
}

// UpdateUserPassword updates user's password hash
func UpdateUserPassword(ctx context.Context, db *sql.DB, userID int, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := db.ExecContext(ctx, query, passwordHash, time.Now(), userID)
	return err
}

// UpdateUserPrivacy updates user's privacy settings
func UpdateUserPrivacy(ctx context.Context, db *sql.DB, userID int, settings *models.PrivacySettings) error {
	query := `
		UPDATE users
		SET privacy_settings = $1, updated_at = $2
		WHERE id = $3
	`

	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, query, settingsJSON, time.Now(), userID)
	return err
}
