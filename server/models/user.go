package models

import (
	"encoding/json"
	"time"
)

type PrivacySettings struct {
	ProfilePublic bool `json:"profile_public"`
	QuotesPublic  bool `json:"quotes_public"`
}

type User struct {
	ID              int              `json:"id"`
	Username        string           `json:"username"`
	Email           string           `json:"email"`
	PasswordHash    string           `json:"-"`
	Bio             string           `json:"bio"`
	PrivacySettings *PrivacySettings `json:"privacy_settings"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	LastLogin       *time.Time       `json:"last_login,omitempty"`
	IsActive        bool             `json:"is_active"`
}

// DefaultPrivacySettings returns the default privacy settings
func DefaultPrivacySettings() *PrivacySettings {
	return &PrivacySettings{
		ProfilePublic: true,
		QuotesPublic:  true,
	}
}

// ParsePrivacySettings parses JSON into PrivacySettings
func ParsePrivacySettings(data []byte) (*PrivacySettings, error) {
	if len(data) == 0 {
		return DefaultPrivacySettings(), nil
	}
	var ps PrivacySettings
	if err := json.Unmarshal(data, &ps); err != nil {
		return DefaultPrivacySettings(), nil
	}
	return &ps, nil
}

type SignupRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	UsernameOrEmail string `json:"username_or_email" binding:"required"`
	Password        string `json:"password" binding:"required"`
}

func (u *User) ToResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":               u.ID,
		"username":         u.Username,
		"email":            u.Email,
		"bio":              u.Bio,
		"privacy_settings": u.PrivacySettings,
		"created_at":       u.CreatedAt,
	}
}

// Request types for settings updates
type UpdateProfileRequest struct {
	Username string `json:"username" binding:"omitempty,min=3,max=50"`
	Email    string `json:"email" binding:"omitempty,email"`
	Bio      string `json:"bio" binding:"max=500"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

type UpdatePrivacyRequest struct {
	ProfilePublic *bool `json:"profile_public"`
	QuotesPublic  *bool `json:"quotes_public"`
}

// Password reset types
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}
