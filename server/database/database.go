package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

type DBConnection struct {
	DB *sql.DB
}

// Establishes the connection to the database using the information stored in ../.env
// returns the db connection
func StartDatabase() *sql.DB {
	host := os.Getenv("DB_HOSTNAME")
	user := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	portStr := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic("Invalid DB_PORT: " + err.Error())
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	return db
}

// used for the GET reguest at the "/" end point. Returns all of the quotes in the quote database
// returns the quotes as a json object.
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

	for i := 0; i < count; i++ {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		err := rows.Scan(valuePtrs...)
		if err != nil {
			panic(err)
		}

		entry := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]

			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			entry[col] = v
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

// EXAMPLE USAGE
//	dbConn := StartDatabase()
//	defer dbConn.Close()
//
//	jsonResult := FetchQuotesAsJson(dbConn)
