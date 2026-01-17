package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

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

	// HTML template
	r.LoadHTMLFiles("../client/index.html")
	r.Static("/css", "../client/css")

	// Public routes
	r.GET("/", func(c *gin.Context) {
		// Fetch quotes from database as JSON and unmarshal
		quotesJSON := database.FetchQuotesAsJson(dbConn.DB)
		quotes, err := UnmarshalQuotes([]byte(quotesJSON))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load quotes"})
			return
		}
		c.HTML(http.StatusOK, "index.html", gin.H{"items": quotes})
	})

	// Public API routes
	r.GET("/user/:id/quotes", handlers.GetUserQuotesHandler(dbConn.DB))

	// Authentication routes
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/signup", handlers.SignupHandler(dbConn.DB))
		authGroup.POST("/login", handlers.LoginHandler(dbConn.DB))
		authGroup.POST("/logout", handlers.LogoutHandler())
	}

	// Protected routes - require authentication
	apiGroup := r.Group("/api")
	apiGroup.Use(middleware.AuthRequired())
	{
		// Get current user
		apiGroup.GET("/user", handlers.GetCurrentUserHandler(dbConn.DB))

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
