package database

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/lib/pq"
)

// GetQuoteByID retrieves a single quote by its ID
func GetQuoteByID(db *sql.DB, quoteID int) (map[string]interface{}, error) {
	query := `
		SELECT quote_id, user_id, quote, author, book, tags
		FROM quotes
		WHERE quote_id = $1
	`

	var (
		qID    int
		userID int
		quote  string
		author string
		book   string
		tags   []byte
	)

	err := db.QueryRow(query, quoteID).Scan(&qID, &userID, &quote, &author, &book, &tags)
	if err == sql.ErrNoRows {
		return nil, errors.New("quote not found")
	}
	if err != nil {
		return nil, err
	}

	// Parse tags array
	tagsStr := string(tags)
	clean := strings.Trim(tagsStr, "{}")
	var tagsList []string
	if len(clean) > 0 {
		tagsList = strings.Split(clean, ",")
	} else {
		tagsList = []string{}
	}

	result := map[string]interface{}{
		"quote_id": qID,
		"user_id":  userID,
		"quote":    quote,
		"author":   author,
		"book":     book,
		"tags":     tagsList,
	}

	return result, nil
}

// UpdateQuote updates a quote's content
func UpdateQuote(db *sql.DB, quoteID int, quote, author, book string, tags []string) error {
	tagsStr := `{` + strings.Join(tags, `,`) + `}`

	query := `
		UPDATE quotes
		SET quote = $1, author = $2, book = $3, tags = $4, updated_at = CURRENT_TIMESTAMP
		WHERE quote_id = $5
	`

	result, err := db.Exec(query, quote, author, book, tagsStr, quoteID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("quote not found")
	}

	return nil
}

// DeleteQuote removes a quote from the database
func DeleteQuote(db *sql.DB, quoteID int) error {
	query := `DELETE FROM quotes WHERE quote_id = $1`

	result, err := db.Exec(query, quoteID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("quote not found")
	}

	return nil
}

// GetQuotesByUserID retrieves all quotes for a specific user
func GetQuotesByUserID(db *sql.DB, userID int) ([]map[string]interface{}, error) {
	query := `
		SELECT quote_id, user_id, quote, author, book, tags
		FROM quotes
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	quotes := make([]map[string]interface{}, 0)

	for rows.Next() {
		var (
			qID    int
			uID    int
			quote  string
			author string
			book   string
			tags   []byte
		)

		if err := rows.Scan(&qID, &uID, &quote, &author, &book, &tags); err != nil {
			return nil, err
		}

		// Parse tags array
		tagsStr := string(tags)
		clean := strings.Trim(tagsStr, "{}")
		var tagsList []string
		if len(clean) > 0 {
			tagsList = strings.Split(clean, ",")
		} else {
			tagsList = []string{}
		}

		result := map[string]interface{}{
			"quote_id": qID,
			"user_id":  uID,
			"quote":    quote,
			"author":   author,
			"book":     book,
			"tags":     tagsList,
		}

		quotes = append(quotes, result)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return quotes, nil
}

// VerifyQuoteOwnership checks if a user owns a specific quote
func VerifyQuoteOwnership(db *sql.DB, quoteID, userID int) (bool, error) {
	query := `SELECT user_id FROM quotes WHERE quote_id = $1`

	var ownerID int
	err := db.QueryRow(query, quoteID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		return false, errors.New("quote not found")
	}
	if err != nil {
		return false, err
	}

	return ownerID == userID, nil
}

// CreateQuote inserts a new quote into the database
func CreateQuote(db *sql.DB, userID int, quote, author, book string, tags []string) (int, error) {
	tagsArray := pq.Array(tags)

	query := `
		INSERT INTO quotes (user_id, quote, author, book, tags)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING quote_id
	`

	var quoteID int
	err := db.QueryRow(query, userID, quote, author, book, tagsArray).Scan(&quoteID)
	if err != nil {
		return 0, err
	}

	return quoteID, nil
}
