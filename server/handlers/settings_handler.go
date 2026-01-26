package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zach-monroe/zetl/server/database"
	"github.com/zach-monroe/zetl/server/models"
	"github.com/zach-monroe/zetl/server/services"
)

// UpdateProfileHandler updates user's profile information
func UpdateProfileHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}

		var req models.UpdateProfileRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx := c.Request.Context()

		// Get current user
		user, err := database.GetUserByID(ctx, db, userID.(int))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		// Use existing values if not provided
		username := req.Username
		if username == "" {
			username = user.Username
		}

		email := req.Email
		if email == "" {
			email = user.Email
		}

		// Validate username if changed
		if username != user.Username {
			if err := services.ValidateUsername(username); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}

		// Validate email if changed
		if email != user.Email {
			if err := services.ValidateEmail(email); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}

		// Update profile
		err = database.UpdateUserProfile(ctx, db, userID.(int), username, email, req.Bio)
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
	}
}

// UpdatePasswordHandler updates user's password
func UpdatePasswordHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}

		var req models.UpdatePasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx := c.Request.Context()

		// Get current user
		user, err := database.GetUserByID(ctx, db, userID.(int))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		// Verify current password
		if err := services.VerifyPassword(user.PasswordHash, req.CurrentPassword); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
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
		err = database.UpdateUserPassword(ctx, db, userID.(int), hashedPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
	}
}

// UpdatePrivacyHandler updates user's privacy settings
func UpdatePrivacyHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}

		var req models.UpdatePrivacyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx := c.Request.Context()

		// Get current user privacy settings
		user, err := database.GetUserByID(ctx, db, userID.(int))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		// Create updated privacy settings
		settings := user.PrivacySettings
		if settings == nil {
			settings = models.DefaultPrivacySettings()
		}

		if req.ProfilePublic != nil {
			settings.ProfilePublic = *req.ProfilePublic
		}
		if req.QuotesPublic != nil {
			settings.QuotesPublic = *req.QuotesPublic
		}

		// Update privacy settings
		err = database.UpdateUserPrivacy(ctx, db, userID.(int), settings)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update privacy settings"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Privacy settings updated successfully"})
	}
}
