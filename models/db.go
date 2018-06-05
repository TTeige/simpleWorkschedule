package models

import (
	_ "github.com/lib/pq"
	"database/sql"
	"fmt"
)

func OpenDatabase(DB_USER string, DB_NAME string, DB_PASSWORD string) (*sql.DB, error) {
	dbStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", DB_USER, DB_PASSWORD, DB_NAME)
	db, err := sql.Open("postgres", dbStr)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
