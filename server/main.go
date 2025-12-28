package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/zach-monroe/zetl/server/database"
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

	// HTML template
	r.LoadHTMLFiles("../client/index.html")
	r.Static("/css", "../client/css")

	// Fetch quotes from database as JSON and unmarshal
	quotesJSON := database.FetchQuotesAsJson(dbConn.DB)
	quotes, err := UnmarshalQuotes([]byte(quotesJSON))
	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal quotes: %v", err))
	}

	// Root endpoint renders HTML
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{"items": quotes})
	})

	// Example user route (replace with real DB query later)
	r.GET("/user/:name", func(c *gin.Context) {
		user := c.Param("name")
		c.JSON(http.StatusOK, gin.H{"user": user, "status": "no value"})
	})

	r.POST("/quote", func(c *gin.Context) {
		// Read raw JSON body directly
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
			return
		}

		// Pass raw bytes straight to database (no intermediate struct)
		if err := database.AddQuoteToDatabase(dbConn.DB, bodyBytes); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"status": "success"})
	})

	// Auth group
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"foo":  "bar",
		"manu": "123",
	}))

	authorized.POST("/admin", func(c *gin.Context) {
		user := c.MustGet(gin.AuthUserKey).(string)
		var input struct {
			Value string `json:"value" binding:"required"`
		}

		if c.ShouldBindJSON(&input) == nil {
			// You can add database update code here:
			// database.SaveAdminValue(dbConn.DB, user, input.Value)
			c.JSON(http.StatusOK, gin.H{"user": user, "status": "ok"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		}
	})

	return r
}

func main() {
	dbConn := database.StartDatabase()
	defer dbConn.DB.Close()

	r := setupRouter(dbConn)
	r.Run(":8080")
}
