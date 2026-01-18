package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/lib/pq"
	"github.com/zach-monroe/zetl/server/models"
)

// CreateUser inserts a new user into the database
func CreateUser(db *sql.DB, user *models.User) error {
	query := `
		INSERT INTO users (username, email, password_hash, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err := db.QueryRow(query, user.Username, user.Email, user.PasswordHash, user.IsActive).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		// Check for unique constraint violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			if pqErr.Constraint == "users_username_key" {
				return errors.New("username already exists")
			}
			if pqErr.Constraint == "users_email_key" {
				return errors.New("email already exists")
			}
		}
		return err
	}

	return nil
}

// GetUserByUsername retrieves a user by username
func GetUserByUsername(db *sql.DB, username string) (*models.User, error) {
	user := &models.User{}
	var privacySettingsJSON []byte
	query := `
		SELECT id, username, email, password_hash, COALESCE(bio, ''), COALESCE(privacy_settings::text, '{}')::bytea, created_at, updated_at, last_login, is_active
		FROM users
		WHERE username = $1
	`

	err := db.QueryRow(query, username).Scan(
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
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	user.PrivacySettings, _ = models.ParsePrivacySettings(privacySettingsJSON)
	return user, nil
}

// GetUserByEmail retrieves a user by email
func GetUserByEmail(db *sql.DB, email string) (*models.User, error) {
	user := &models.User{}
	var privacySettingsJSON []byte
	query := `
		SELECT id, username, email, password_hash, COALESCE(bio, ''), COALESCE(privacy_settings::text, '{}')::bytea, created_at, updated_at, last_login, is_active
		FROM users
		WHERE email = $1
	`

	err := db.QueryRow(query, email).Scan(
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
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	user.PrivacySettings, _ = models.ParsePrivacySettings(privacySettingsJSON)
	return user, nil
}

// GetUserByID retrieves a user by ID
func GetUserByID(db *sql.DB, id int) (*models.User, error) {
	user := &models.User{}
	var privacySettingsJSON []byte
	query := `
		SELECT id, username, email, password_hash, COALESCE(bio, ''), COALESCE(privacy_settings::text, '{}')::bytea, created_at, updated_at, last_login, is_active
		FROM users
		WHERE id = $1
	`

	err := db.QueryRow(query, id).Scan(
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
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	user.PrivacySettings, _ = models.ParsePrivacySettings(privacySettingsJSON)
	return user, nil
}

// UpdateLastLogin updates the last_login timestamp for a user
func UpdateLastLogin(db *sql.DB, userID int) error {
	query := `
		UPDATE users
		SET last_login = $1
		WHERE id = $2
	`

	_, err := db.Exec(query, time.Now(), userID)
	return err
}

// UpdateUserProfile updates user's username, email, and bio
func UpdateUserProfile(db *sql.DB, userID int, username, email, bio string) error {
	query := `
		UPDATE users
		SET username = $1, email = $2, bio = $3, updated_at = $4
		WHERE id = $5
	`

	_, err := db.Exec(query, username, email, bio, time.Now(), userID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			if pqErr.Constraint == "users_username_key" {
				return errors.New("username already exists")
			}
			if pqErr.Constraint == "users_email_key" {
				return errors.New("email already exists")
			}
		}
		return err
	}
	return nil
}

// UpdateUserPassword updates user's password hash
func UpdateUserPassword(db *sql.DB, userID int, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := db.Exec(query, passwordHash, time.Now(), userID)
	return err
}

// UpdateUserPrivacy updates user's privacy settings
func UpdateUserPrivacy(db *sql.DB, userID int, settings *models.PrivacySettings) error {
	query := `
		UPDATE users
		SET privacy_settings = $1, updated_at = $2
		WHERE id = $3
	`

	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	_, err = db.Exec(query, settingsJSON, time.Now(), userID)
	return err
}
