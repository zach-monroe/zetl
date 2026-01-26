package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zach-monroe/zetl/server/database"
)

type QuoteRequest struct {
	Quote  string   `json:"quote" binding:"required"`
	Author string   `json:"author" binding:"required"`
	Book   string   `json:"book" binding:"required"`
	Tags   []string `json:"tags"`
	Notes  string   `json:"notes"`
}

// CreateQuoteHandler handles creating a new quote
func CreateQuoteHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user_id from context (set by AuthRequired middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}

		var req QuoteRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Ensure tags is not nil
		if req.Tags == nil {
			req.Tags = []string{}
		}

		// Create quote
		quoteID, err := database.CreateQuote(c.Request.Context(), db, userID.(int), req.Quote, req.Author, req.Book, req.Tags, req.Notes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create quote", "details": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":  "Quote created successfully",
			"quote_id": quoteID,
		})
	}
}

// UpdateQuoteHandler handles updating an existing quote
func UpdateQuoteHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get quote_id from context (set by QuoteOwnershipRequired middleware)
		quoteID, exists := c.Get("quote_id")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quote ID"})
			return
		}

		var req QuoteRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Ensure tags is not nil
		if req.Tags == nil {
			req.Tags = []string{}
		}

		// Update quote
		err := database.UpdateQuote(c.Request.Context(), db, quoteID.(int), req.Quote, req.Author, req.Book, req.Tags, req.Notes)
		if err != nil {
			if errors.Is(err, database.ErrQuoteNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Quote not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update quote"})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Quote updated successfully"})
	}
}

// DeleteQuoteHandler handles deleting a quote
func DeleteQuoteHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get quote_id from context (set by QuoteOwnershipRequired middleware)
		quoteID, exists := c.Get("quote_id")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quote ID"})
			return
		}

		// Delete quote
		err := database.DeleteQuote(c.Request.Context(), db, quoteID.(int))
		if err != nil {
			if errors.Is(err, database.ErrQuoteNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Quote not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete quote"})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Quote deleted successfully"})
	}
}

// GetAllQuotesHandler returns all quotes (public)
func GetAllQuotesHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		quotesJSON := database.FetchQuotesAsJson(db)
		c.Header("Content-Type", "application/json")
		c.String(http.StatusOK, quotesJSON)
	}
}

// GetUserQuotesHandler returns all quotes for a specific user (public)
func GetUserQuotesHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Param("id")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		quotes, err := database.GetQuotesByUserID(c.Request.Context(), db, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch quotes"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"quotes": quotes})
	}
}
