package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// DBConnection holds the SQL connection for convenience
type DBConnection struct {
	DB *sql.DB
}

// StartDatabase connects to PostgreSQL using environment variables
func StartDatabase() (*DBConnection, error) {
	_, filename, _, _ := runtime.Caller(0) // Gets database.go location
	dir := filepath.Dir(filename)          // /path/to/server/database
	parentDir := filepath.Dir(dir)         // /path/to/server (parent)
	envPath := filepath.Join(parentDir, ".env")
	if err := godotenv.Load(envPath); err != nil {
		fmt.Printf(".env not found at %s, using system env vars\n", envPath)
	}
	host := os.Getenv("DB_HOSTNAME")
	user := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	portStr := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	fmt.Println("Connected to database successfully.")
	return &DBConnection{DB: db}, nil
}
func FetchQuotesAsJson(db *sql.DB) string {
	query := "SELECT * FROM quotes"
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		panic(err)
	}

	count := len(columns)
	tableData := make([]map[string]interface{}, 0)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)

	for i := range values {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			panic(err)
		}

		entry := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]

			// Handle SQL byte arrays, JSONB, and tags array conversion cleanly
			switch v := val.(type) {
			case []byte:
				if col == "tags" {
					entry[col] = ParsePostgresTags(v)
				} else {
					entry[col] = string(v)
				}
			default:
				entry[col] = v
			}
		}
		tableData = append(tableData, entry)
	}

	if err = rows.Err(); err != nil {
		panic(err)
	}

	jsonData, err := json.MarshalIndent(tableData, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(jsonData)
}

func AddQuoteToDatabase(db *sql.DB, jsonBytes []byte) error {
	var q map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &q); err != nil {
		return err
	}

	userID := int(q["user_id"].(float64))
	quote := q["quote"].(string)
	author := q["author"].(string)
	book := q["book"].(string)
	tagsRaw := q["tags"].([]interface{})
	tags := make([]string, len(tagsRaw))
	for i, t := range tagsRaw {
		tags[i] = t.(string)
	}

	tagsStr := FormatPostgresTags(tags)

	_, err := db.Exec(`
        INSERT INTO quotes (user_id, quote, author, book, tags)
        VALUES ($1, $2, $3, $4, $5)
    `, userID, quote, author, book, tagsStr)

	return err
}
