package handlers

import (
	"database/sql"
	"log"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/zach-monroe/zetl/server/database"
)

// GetUserFromSession retrieves the user from session if logged in.
// Returns nil if the user is not logged in or cannot be retrieved.
func GetUserFromSession(c *gin.Context, db *sql.DB) map[string]interface{} {
	session := sessions.Default(c)
	userID := session.Get("user_id")

	if userID == nil {
		return nil
	}

	// Handle different integer types that the session store might return
	var userIDInt int
	switch v := userID.(type) {
	case int:
		userIDInt = v
	case int64:
		userIDInt = int(v)
	case float64:
		userIDInt = int(v)
	default:
		log.Printf("[Session] Unexpected user_id type: %T", userID)
		return nil
	}

	user, err := database.GetUserByID(c.Request.Context(), db, userIDInt)
	if err != nil {
		log.Printf("[Session] Failed to get user by ID %d: %v", userIDInt, err)
		return nil
	}

	return user.ToResponse()
}

// CreateUserSession creates a new session for the given user ID.
func CreateUserSession(c *gin.Context, userID int) error {
	session := sessions.Default(c)
	session.Set("user_id", userID)
	return session.Save()
}
