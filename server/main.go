package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/postgres"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/zach-monroe/zetl/server/database"
	"github.com/zach-monroe/zetl/server/handlers"
	"github.com/zach-monroe/zetl/server/middleware"
	"github.com/zach-monroe/zetl/server/models"
)

//type Quote struct {
//	QuoteID int      `json:"quote_id"`
//	UserID  int      `json:"user_id"`
//	Quote   string   `json:"quote"`
//	Author  string   `json:"author"`
//	Book    string   `json:"book"`
//	Tags    []string `json:"tags"`
//}
//

func UnmarshalQuotes(data []byte) (models.Quotes, error) {
	var q models.Quotes
	err := json.Unmarshal(data, &q)
	return q, err
}

// getUserFromSession retrieves the user from session if logged in
func getUserFromSession(c *gin.Context, db *database.DBConnection) map[string]interface{} {
	session := sessions.Default(c)
	userID := session.Get("user_id")

	if userID == nil {
		return nil
	}

	user, err := database.GetUserByID(db.DB, userID.(int))
	if err != nil {
		return nil
	}

	return user.ToResponse()
}

func setupRouter(dbConn *database.DBConnection) *gin.Engine {
	r := gin.Default()

	// Set up PostgreSQL session store
	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		panic("SESSION_SECRET environment variable is not set")
	}

	store, err := postgres.NewStore(dbConn.DB, []byte(sessionSecret))
	if err != nil {
		panic(fmt.Sprintf("Failed to create session store: %v", err))
	}

	// Configure session options
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400, // 24 hours
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	r.Use(sessions.Sessions("zetl_session", store))

	// Custom template functions
	r.SetFuncMap(template.FuncMap{
		"join": strings.Join,
	})

	// HTML templates - load from templates directory
	r.LoadHTMLGlob("../client/templates/*.html")
	r.Static("/css", "../client/css")

	// Public page routes
	r.GET("/", func(c *gin.Context) {
		// Fetch quotes from database as JSON and unmarshal
		quotesJSON := database.FetchQuotesAsJson(dbConn.DB)
		quotes, err := UnmarshalQuotes([]byte(quotesJSON))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load quotes"})
			return
		}
		user := getUserFromSession(c, dbConn)
		c.HTML(http.StatusOK, "index.html", gin.H{"items": quotes, "user": user})
	})

	r.GET("/login", handlers.LoginPageHandler(dbConn.DB))
	r.GET("/signup", handlers.SignupPageHandler(dbConn.DB))
	r.GET("/forgot-password", handlers.ForgotPasswordPageHandler(dbConn.DB))
	r.GET("/reset-password", handlers.ResetPasswordPageHandler(dbConn.DB))

	// Public API routes
	r.GET("/user/:id/quotes", handlers.GetUserQuotesHandler(dbConn.DB))

	// Authentication routes
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/signup", handlers.SignupHandler(dbConn.DB))
		authGroup.POST("/login", handlers.LoginHandler(dbConn.DB))
		authGroup.POST("/logout", handlers.LogoutHandler())
		authGroup.POST("/forgot-password", handlers.ForgotPasswordHandler(dbConn.DB))
		authGroup.POST("/reset-password", handlers.ResetPasswordHandler(dbConn.DB))
	}

	// Protected page routes - require authentication
	r.GET("/settings", middleware.AuthRequired(), handlers.SettingsPageHandler(dbConn.DB))
	r.GET("/profile", middleware.AuthRequired(), handlers.ProfilePageHandler(dbConn.DB))

	// Protected API routes - require authentication
	apiGroup := r.Group("/api")
	apiGroup.Use(middleware.AuthRequired())
	{
		// Get current user
		apiGroup.GET("/user", handlers.GetCurrentUserHandler(dbConn.DB))

		// User settings
		apiGroup.PUT("/user/profile", handlers.UpdateProfileHandler(dbConn.DB))
		apiGroup.PUT("/user/password", handlers.UpdatePasswordHandler(dbConn.DB))
		apiGroup.PUT("/user/privacy", handlers.UpdatePrivacyHandler(dbConn.DB))

		// Quote creation
		apiGroup.POST("/quote", handlers.CreateQuoteHandler(dbConn.DB))

		// Quote modification (requires ownership)
		apiGroup.PUT("/quote/:id", middleware.QuoteOwnershipRequired(dbConn.DB), handlers.UpdateQuoteHandler(dbConn.DB))
		apiGroup.DELETE("/quote/:id", middleware.QuoteOwnershipRequired(dbConn.DB), handlers.DeleteQuoteHandler(dbConn.DB))
	}

	return r
}

func main() {
	dbConn := database.StartDatabase()
	defer dbConn.DB.Close()

	r := setupRouter(dbConn)
	r.Run(":8080")
}
