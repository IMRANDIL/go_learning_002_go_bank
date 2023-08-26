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

func (s *PostgresStore) Init() error {
	err := s.createAccountTable()
	if err != nil {
		return err
	}

	// You can add more initialization steps here

	return nil
}

func (s *PostgresStore) createAccountTable() error {
	query := `
        CREATE TABLE IF NOT EXISTS accounts (
            id SERIAL PRIMARY KEY,
            first_name VARCHAR(255) NOT NULL,
            last_name VARCHAR(255) NOT NULL,
            hobby VARCHAR(255),
            age INT,
            account_number BIGINT NOT NULL,
            balance DECIMAL(15, 2) NOT NULL
        )
    `

	_, err := s.db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStore) createAccount(*Account) error {
	return nil
}

func (s *PostgresStore) deleteAccount(id int) error {
	return nil
}

func (s *PostgresStore) updateAccount(*Account) error {
	return nil
}

func (s *PostgresStore) getAccountById(id int) (*Account, error) {
	return nil, nil
}
