package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

type Storage interface {
	createAccount(*Account) error
	deleteAccount(int) error
	updateAccount(*Account) error
	getAccountById(int) (*Account, error)
}

type PostgresStore struct {
	db *sql.DB
}

func newPostgesStore() (*PostgresStore, error) {
	connStr := "user=postgres dbname=goLearning_db password=Dil@2580123 sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) createAccount(*Account) error {
	return nil
}
