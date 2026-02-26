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
	"github.com/zach-monroe/zetl/server/config"
	"github.com/zach-monroe/zetl/server/database"
	"github.com/zach-monroe/zetl/server/handlers"
	"github.com/zach-monroe/zetl/server/middleware"
	"github.com/zach-monroe/zetl/server/models"
	"github.com/zach-monroe/zetl/server/services"
)

func UnmarshalQuotes(data []byte) (models.Quotes, error) {
	var q models.Quotes
	err := json.Unmarshal(data, &q)
	return q, err
}

func setupRouter(dbConn *database.DBConnection, emailService *services.EmailService, geminiService *services.GeminiService) *gin.Engine {
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
		MaxAge:   config.SessionMaxAge,
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
	r.Static("/js", "../client/js")

	// Health check endpoint (used by k8s probes)
	r.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// Public page routes
	r.GET("/", func(c *gin.Context) {
		// Fetch quotes from database as JSON and unmarshal
		quotesJSON := database.FetchQuotesAsJson(dbConn.DB)
		quotes, err := UnmarshalQuotes([]byte(quotesJSON))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load quotes"})
			return
		}
		user := handlers.GetUserFromSession(c, dbConn.DB)
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
		authGroup.POST("/forgot-password", handlers.ForgotPasswordHandler(dbConn.DB, emailService))
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

		// Writing prompt generation
		apiGroup.POST("/generate-prompt", handlers.GeneratePromptHandler(dbConn.DB, geminiService))
	}

	// Device API routes - token-based auth for Pi client and other devices
	deviceGroup := r.Group("/api/device")
	deviceGroup.Use(middleware.APITokenRequired())
	{
		deviceGroup.POST("/quote", handlers.CreateQuoteHandler(dbConn.DB))
	}

	return r
}

func main() {
	dbConn, err := database.StartDatabase()
	if err != nil {
		panic(fmt.Sprintf("Failed to start database: %v", err))
	}
	defer dbConn.DB.Close()

	// Initialize services
	emailService := services.NewEmailService()
	geminiService := services.NewGeminiService()

	r := setupRouter(dbConn, emailService, geminiService)
	r.Run(":8080")
}
