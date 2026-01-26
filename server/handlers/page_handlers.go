package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zach-monroe/zetl/server/database"
)

// LoginPageHandler renders the login page
func LoginPageHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Redirect if already logged in
		if GetUserFromSession(c, db) != nil {
			c.Redirect(http.StatusFound, "/")
			return
		}
		c.HTML(http.StatusOK, "login.html", gin.H{
			"title": "Login",
		})
	}
}

// SignupPageHandler renders the signup page
func SignupPageHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Redirect if already logged in
		if GetUserFromSession(c, db) != nil {
			c.Redirect(http.StatusFound, "/")
			return
		}
		c.HTML(http.StatusOK, "signup.html", gin.H{
			"title": "Sign Up",
		})
	}
}

// ForgotPasswordPageHandler renders the forgot password page
func ForgotPasswordPageHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "forgot-password.html", gin.H{
			"title": "Forgot Password",
		})
	}
}

// ResetPasswordPageHandler renders the reset password page
func ResetPasswordPageHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")

		// Validate token exists and is valid
		if token == "" {
			c.HTML(http.StatusOK, "reset-password.html", gin.H{
				"title":         "Reset Password",
				"invalid_token": true,
			})
			return
		}

		// Check if token is valid
		_, err := database.GetPasswordResetToken(c.Request.Context(), db, token)
		if err != nil {
			c.HTML(http.StatusOK, "reset-password.html", gin.H{
				"title":         "Reset Password",
				"invalid_token": true,
			})
			return
		}

		c.HTML(http.StatusOK, "reset-password.html", gin.H{
			"title": "Reset Password",
			"token": token,
		})
	}
}

// SettingsPageHandler renders the settings page
func SettingsPageHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Redirect(http.StatusFound, "/login")
			return
		}

		user, err := database.GetUserByID(c.Request.Context(), db, userID.(int))
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			return
		}

		c.HTML(http.StatusOK, "settings.html", gin.H{
			"title": "Settings",
			"user":  user,
		})
	}
}

// ProfilePageHandler renders the profile page
func ProfilePageHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Redirect(http.StatusFound, "/login")
			return
		}

		ctx := c.Request.Context()
		user, err := database.GetUserByID(ctx, db, userID.(int))
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			return
		}

		// Fetch user's quotes
		quotes, err := database.FetchQuotesByUserID(ctx, db, userID.(int))
		if err != nil {
			quotes = nil
		}

		c.HTML(http.StatusOK, "profile.html", gin.H{
			"title":          "My Profile",
			"user":           user.ToResponse(),
			"profile_user":   user,
			"items":          quotes,
			"is_own_profile": true,
		})
	}
}
