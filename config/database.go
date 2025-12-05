package config

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var db *sql.DB

// Open Connection
func OpenConnection() error {
	var err error
	db, err = setupConnection()
	log.Printf("Error OpenConnection: %v\n", err)
	log.Printf("DB open: %v\n", db)
	return err
}

// setupConnection adalah
func setupConnection() (*sql.DB, error) {
	var connection = fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s",
		DBUser, DBPass, DBName, DBHost, DBPort, SSLMode)
	fmt.Println("Connection Info:", DBDriver, connection)

	db, err := sql.Open(DBDriver, connection)
	log.Printf("Error setupConnection: %v\n", err)
	log.Printf("db setup: %v\n", db)
	if err != nil {
		return db, errors.New("Connection closed: Failed Connect Database")
	}

	return db, nil
}

// CloseConnectionDB adalah
func CloseConnectionDB() {
	db.Close()
}

// DBConnection adalah
func DBConnection() *sql.DB {
	return db
}
