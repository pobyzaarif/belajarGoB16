package db

import (
	"database/sql"
	"fmt"
	"os"
)

func InitDB() (db *sql.DB, err error) {
	db, err = sql.Open("mysql", os.Getenv("MYSQL_DSN"))
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return db, err
	}

	err = db.Ping()
	if err != nil {
		fmt.Println("Error pinging database:", err)
		return db, err
	}

	fmt.Println("Successfully connected to the database!")
	fmt.Println()
	return db, err
}
