package middleware

import (
	"database/sql"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/zach-monroe/zetl/server/database"
)

// AuthRequired verifies that a user is logged in
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")

		if userID == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Store user_id in context for handlers to use
		c.Set("user_id", userID)
		c.Next()
	}
}

// OptionalAuth attaches user_id to context if logged in, but doesn't require it
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")

		if userID != nil {
			c.Set("user_id", userID)
		}

		c.Next()
	}
}

// APITokenRequired validates a Bearer token against the API_TOKEN env var
// and sets user_id from API_TOKEN_USER_ID env var
func APITokenRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiToken := os.Getenv("API_TOKEN")
		apiTokenUserID := os.Getenv("API_TOKEN_USER_ID")

		if apiToken == "" || apiTokenUserID == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "API token not configured"})
			c.Abort()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token != apiToken {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		userID, err := strconv.Atoi(apiTokenUserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid API_TOKEN_USER_ID configuration"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

// QuoteOwnershipRequired verifies that the authenticated user owns the quote
func QuoteOwnershipRequired(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user_id from context (set by AuthRequired middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Get quote_id from URL parameter
		quoteIDStr := c.Param("id")
		quoteID, err := strconv.Atoi(quoteIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quote ID"})
			c.Abort()
			return
		}

		// Verify ownership
		isOwner, err := database.VerifyQuoteOwnership(c.Request.Context(), db, quoteID, userID.(int))
		if err != nil {
			if errors.Is(err, database.ErrQuoteNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Quote not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify ownership"})
			}
			c.Abort()
			return
		}

		if !isOwner {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to modify this quote"})
			c.Abort()
			return
		}

		// Store quote_id in context for handlers
		c.Set("quote_id", quoteID)
		c.Next()
	}
}
