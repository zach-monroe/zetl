package database

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	"github.com/zach-monroe/zetl/server/models"
)

// GetQuoteByID retrieves a single quote by its ID
func GetQuoteByID(ctx context.Context, db *sql.DB, quoteID int) (map[string]interface{}, error) {
	query := `
		SELECT quote_id, user_id, quote, author, book, tags, COALESCE(notes, '') as notes
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
		notes  string
	)

	err := db.QueryRowContext(ctx, query, quoteID).Scan(&qID, &userID, &quote, &author, &book, &tags, &notes)
	if err == sql.ErrNoRows {
		return nil, ErrQuoteNotFound
	}
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"quote_id": qID,
		"user_id":  userID,
		"quote":    quote,
		"author":   author,
		"book":     book,
		"tags":     ParsePostgresTags(tags),
		"notes":    notes,
	}

	return result, nil
}

// UpdateQuote updates a quote's content
func UpdateQuote(ctx context.Context, db *sql.DB, quoteID int, quote, author, book string, tags []string, notes string) error {
	tagsStr := FormatPostgresTags(tags)

	query := `
		UPDATE quotes
		SET quote = $1, author = $2, book = $3, tags = $4, notes = $5, updated_at = CURRENT_TIMESTAMP
		WHERE quote_id = $6
	`

	result, err := db.ExecContext(ctx, query, quote, author, book, tagsStr, notes, quoteID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrQuoteNotFound
	}

	return nil
}

// DeleteQuote removes a quote from the database
func DeleteQuote(ctx context.Context, db *sql.DB, quoteID int) error {
	query := `DELETE FROM quotes WHERE quote_id = $1`

	result, err := db.ExecContext(ctx, query, quoteID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrQuoteNotFound
	}

	return nil
}

// GetQuotesByUserID retrieves all quotes for a specific user as maps
func GetQuotesByUserID(ctx context.Context, db *sql.DB, userID int) ([]map[string]interface{}, error) {
	quotes, err := FetchQuotesByUserID(ctx, db, userID)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(quotes))
	for i, q := range quotes {
		result[i] = q.ToMap()
	}
	return result, nil
}

// VerifyQuoteOwnership checks if a user owns a specific quote
func VerifyQuoteOwnership(ctx context.Context, db *sql.DB, quoteID, userID int) (bool, error) {
	query := `SELECT user_id FROM quotes WHERE quote_id = $1`

	var ownerID int
	err := db.QueryRowContext(ctx, query, quoteID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		return false, ErrQuoteNotFound
	}
	if err != nil {
		return false, err
	}

	return ownerID == userID, nil
}

// CreateQuote inserts a new quote into the database
func CreateQuote(ctx context.Context, db *sql.DB, userID int, quote, author, book string, tags []string, notes string) (int, error) {
	tagsArray := pq.Array(tags)

	query := `
		INSERT INTO quotes (user_id, quote, author, book, tags, notes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING quote_id
	`

	var quoteID int
	err := db.QueryRowContext(ctx, query, userID, quote, author, book, tagsArray, notes).Scan(&quoteID)
	if err != nil {
		return 0, err
	}

	return quoteID, nil
}

// FetchQuotesByUserID retrieves all quotes for a specific user as models.Quotes
func FetchQuotesByUserID(ctx context.Context, db *sql.DB, userID int) (models.Quotes, error) {
	query := `
		SELECT quote_id, user_id, quote, author, book, tags, COALESCE(notes, '') as notes
		FROM quotes
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	quotes := make(models.Quotes, 0)

	for rows.Next() {
		var (
			qID    int
			uID    int
			quote  string
			author string
			book   string
			tags   []byte
			notes  string
		)

		if err := rows.Scan(&qID, &uID, &quote, &author, &book, &tags, &notes); err != nil {
			return nil, err
		}

		q := models.Quote{
			QuoteID: qID,
			UserID:  uID,
			Quote:   quote,
			Author:  author,
			Book:    book,
			Tags:    ParsePostgresTags(tags),
			Notes:   notes,
		}

		quotes = append(quotes, q)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return quotes, nil
}
