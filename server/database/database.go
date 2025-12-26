package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// DBConnection holds the SQL connection for convenience
type DBConnection struct {
	DB *sql.DB
}

// StartDatabase connects to PostgreSQL using environment variables
func StartDatabase() *DBConnection {
	_, filename, _, _ := runtime.Caller(0) // Gets database.go location
	dir := filepath.Dir(filename)          // /path/to/server/database
	parentDir := filepath.Dir(dir)         // /path/to/server (parent)
	envPath := filepath.Join(parentDir, ".env")
	if err := godotenv.Load(envPath); err != nil {
		fmt.Printf("⚠️  .env not found at %s, using system env vars\n", envPath)
	}
	host := os.Getenv("DB_HOSTNAME")
	user := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	portStr := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic("Invalid DB_PORT: " + err.Error())
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}

	fmt.Println("✅ Connected to database successfully.")
	return &DBConnection{DB: db}
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
				strVal := string(v)

				// 🔥 Handle tags like {a,b,c}
				if col == "tags" {
					clean := strings.Trim(strVal, "{}")
					if len(clean) > 0 {
						entry[col] = strings.Split(clean, ",")
					} else {
						entry[col] = []string{}
					}
				} else {
					entry[col] = strVal
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
