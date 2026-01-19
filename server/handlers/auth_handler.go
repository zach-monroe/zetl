package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/zach-monroe/zetl/server/database"
	"github.com/zach-monroe/zetl/server/models"
	"github.com/zach-monroe/zetl/server/services"
)

// SignupHandler handles user registration
func SignupHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.SignupRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate username
		if err := services.ValidateUsername(req.Username); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate email
		if err := services.ValidateEmail(req.Email); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate password
		if err := services.ValidatePassword(req.Password); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Hash password
		hashedPassword, err := services.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
			return
		}

		// Create user
		user := &models.User{
			Username:     req.Username,
			Email:        req.Email,
			PasswordHash: hashedPassword,
			IsActive:     true,
		}

		if err := database.CreateUser(db, user); err != nil {
			if strings.Contains(err.Error(), "already exists") {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			}
			return
		}

		// Create session
		session := sessions.Default(c)
		session.Set("user_id", user.ID)
		log.Printf("[Signup Debug] Setting session user_id to: %d (type: %T)", user.ID, user.ID)
		if err := session.Save(); err != nil {
			log.Printf("[Signup Debug] Failed to save session: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
			return
		}
		log.Printf("[Signup Debug] Session saved successfully for user: %s (ID: %d)", user.Username, user.ID)

		c.JSON(http.StatusCreated, gin.H{
			"message": "User created successfully",
			"user":    user.ToResponse(),
		})
	}
}

// LoginHandler handles user authentication
func LoginHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Try to find user by username or email
		var user *models.User
		var err error

		// Check if it's an email (contains @)
		if strings.Contains(req.UsernameOrEmail, "@") {
			user, err = database.GetUserByEmail(db, req.UsernameOrEmail)
		} else {
			user, err = database.GetUserByUsername(db, req.UsernameOrEmail)
		}

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// Check if user is active
		if !user.IsActive {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is inactive"})
			return
		}

		// Verify password
		if err := services.VerifyPassword(user.PasswordHash, req.Password); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// Update last login
		if err := database.UpdateLastLogin(db, user.ID); err != nil {
			log.Printf("[Login Debug] Failed to update last_login: %v", err)
		}

		// Create session
		session := sessions.Default(c)
		session.Set("user_id", user.ID)
		log.Printf("[Login Debug] Setting session user_id to: %d (type: %T)", user.ID, user.ID)
		if err := session.Save(); err != nil {
			log.Printf("[Login Debug] Failed to save session: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
			return
		}
		log.Printf("[Login Debug] Session saved successfully for user: %s (ID: %d)", user.Username, user.ID)

		c.JSON(http.StatusOK, gin.H{
			"message": "Login successful",
			"user":    user.ToResponse(),
		})
	}
}

// LogoutHandler handles user logout
func LogoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		if err := session.Save(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
	}
}

// GetCurrentUserHandler returns the currently logged in user
func GetCurrentUserHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not logged in"})
			return
		}

		user, err := database.GetUserByID(db, userID.(int))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"user": user.ToResponse()})
	}
}
