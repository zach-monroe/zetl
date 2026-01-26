package config

import "time"

const (
	// Session configuration
	SessionMaxAge = 86400 // 24 hours in seconds

	// Password hashing
	BcryptCost = 12

	// Password reset tokens
	PasswordResetExpiry = time.Hour
	TokenCleanupAge     = 24 * time.Hour

	// Limits
	MaxQuotesPerPrompt = 10

	// Validation
	MinPasswordLength = 8
	MinUsernameLength = 3
	MaxUsernameLength = 50
	MaxBioLength      = 500
)
