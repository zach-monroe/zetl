package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zach-monroe/zetl/server/database"
	"github.com/zach-monroe/zetl/server/services"
)

// GeneratePromptRequest represents the request body for generating a prompt
type GeneratePromptRequest struct {
	QuoteIDs []int `json:"quote_ids" binding:"required"`
}

// GeneratePromptHandler handles the POST /api/generate-prompt endpoint
func GeneratePromptHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Initialize Gemini service
		gemini := services.NewGeminiService()
		if !gemini.IsConfigured() {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Writing prompt generation is not configured"})
			return
		}

		var req GeneratePromptRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		if len(req.QuoteIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No quotes selected"})
			return
		}

		if len(req.QuoteIDs) > 10 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Maximum 10 quotes allowed"})
			return
		}

		// Fetch quotes from database
		var quotes []services.QuoteInput
		for _, quoteID := range req.QuoteIDs {
			quote, err := database.GetQuoteByID(db, quoteID)
			if err != nil {
				// Skip quotes that don't exist
				continue
			}

			quotes = append(quotes, services.QuoteInput{
				Quote:  quote["quote"].(string),
				Author: quote["author"].(string),
				Book:   quote["book"].(string),
			})
		}

		if len(quotes) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No valid quotes found"})
			return
		}

		// Generate writing prompt
		prompt, err := gemini.GenerateWritingPrompt(quotes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate prompt: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"prompt": prompt})
	}
}
