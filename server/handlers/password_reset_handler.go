package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zach-monroe/zetl/server/database"
	"github.com/zach-monroe/zetl/server/models"
	"github.com/zach-monroe/zetl/server/services"
)

// ForgotPasswordHandler handles password reset requests
func ForgotPasswordHandler(db *sql.DB, emailService *services.EmailService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.ForgotPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Always return success to prevent email enumeration
		successMessage := "If an account exists with this email, you will receive a password reset link."

		ctx := c.Request.Context()

		// Look up user by email
		user, err := database.GetUserByEmail(ctx, db, req.Email)
		if err != nil {
			// User not found - still return success message
			c.JSON(http.StatusOK, gin.H{"message": successMessage})
			return
		}

		// Invalidate any existing tokens for this user
		database.InvalidateUserTokens(ctx, db, user.ID)

		// Create new reset token
		token, err := database.CreatePasswordResetToken(ctx, db, user.ID)
		if err != nil {
			log.Printf("Failed to create password reset token: %v", err)
			c.JSON(http.StatusOK, gin.H{"message": successMessage})
			return
		}

		// Send reset email
		if emailService.IsConfigured() {
			err = emailService.SendPasswordResetEmail(user.Email, token.Token)
			if err != nil {
				log.Printf("Failed to send password reset email: %v", err)
			}
		} else {
			// For development: log the reset link
			log.Printf("Password reset token for %s: %s", user.Email, token.Token)
		}

		c.JSON(http.StatusOK, gin.H{"message": successMessage})
	}
}

// ResetPasswordHandler handles setting a new password
func ResetPasswordHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.ResetPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx := c.Request.Context()

		// Verify token
		token, err := database.GetPasswordResetToken(ctx, db, req.Token)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired reset token"})
			return
		}

		// Validate new password
		if err := services.ValidatePassword(req.NewPassword); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Hash new password
		hashedPassword, err := services.HashPassword(req.NewPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
			return
		}

		// Update password
		err = database.UpdateUserPassword(ctx, db, token.UserID, hashedPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
			return
		}

		// Mark token as used
		database.MarkTokenAsUsed(ctx, db, token.ID)

		// Invalidate all other tokens for this user
		database.InvalidateUserTokens(ctx, db, token.UserID)

		// Auto-login: create session for the user
		if err := CreateUserSession(c, token.UserID); err != nil {
			log.Printf("[PasswordReset] Failed to create session: %v", err)
			// Still return success - password was reset, just login failed
			c.JSON(http.StatusOK, gin.H{"message": "Password reset successful"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Password reset successful", "redirect": "/"})
	}
}
