package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type Storage interface {
	createAccount(*Account) error
	deleteAccount(int) error
	updateAccount(*Account) error
	getAccountById(int) (*Account, error)
	allAccounts() ([]*Account, error)
	transferBalance(int64, int64, float64) error
	getUserByUsername(string) (*User, error)
	createUser(*User) error
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

	err = s.createUserTable()
	if err != nil {
		return err
	}

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
            balance DECIMAL(15, 2) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `

	_, err := s.db.Exec(query)
	if err != nil {
		log.Printf("Error inserting account: %v", err)
		return err
	}
	return nil
}

func (s *PostgresStore) createUserTable() error {
	query := `
        CREATE TABLE IF NOT EXISTS users (
            id SERIAL PRIMARY KEY,
            username VARCHAR(255) NOT NULL,
            password VARCHAR(255) NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `

	_, err := s.db.Exec(query)
	if err != nil {
		log.Printf("Error creating users table: %v", err)
		return err
	}
	return nil
}

func (s *PostgresStore) createAccount(account *Account) error {

	query := `
        INSERT INTO accounts (first_name, last_name, hobby, age, account_number, balance)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, first_name, last_name, hobby, age, account_number, balance, created_at, updated_at
    `

	err := s.db.QueryRow(
		query,
		account.FIRST_NAME,
		account.LAST_NAME,
		account.HOBBY,
		account.AGE,
		account.ACCOUNT,
		account.BALANCE,
	).Scan(
		&account.ID,
		&account.FIRST_NAME,
		&account.LAST_NAME,
		&account.HOBBY,
		&account.AGE,
		&account.ACCOUNT,
		&account.BALANCE,
		&account.CREATED_AT,
		&account.UPDATED_AT,
	)

	if err != nil {
		log.Printf("Error inserting account: %+v", err)
		return err
	}

	return nil
}

func (s *PostgresStore) allAccounts() ([]*Account, error) {
	query := `SELECT * FROM accounts;`

	rows, err := s.db.Query(query)
	if err != nil {
		log.Printf("Error fetching accounts: %v", err)
		return nil, err
	}
	defer rows.Close()

	var accounts []*Account

	for rows.Next() {
		account := &Account{}
		err := rows.Scan(
			&account.ID,
			&account.FIRST_NAME,
			&account.LAST_NAME,
			&account.HOBBY,
			&account.AGE,
			&account.ACCOUNT,
			&account.BALANCE,
			&account.CREATED_AT,
			&account.UPDATED_AT,
		)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}
		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
		return nil, err
	}

	return accounts, nil
}

func (s *PostgresStore) deleteAccount(id int) error {
	// Check if the account exists
	existsQuery := `
		SELECT id
		FROM accounts
		WHERE id = $1;
	`

	var accountID int
	err := s.db.QueryRow(existsQuery, id).Scan(&accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("account not found")
		}
		log.Printf("Error checking account existence: %v", err)
		return err
	}

	// Account exists, so proceed with deletion
	deleteQuery := `
		DELETE FROM accounts
		WHERE id = $1;
	`

	_, err = s.db.Exec(deleteQuery, id)
	if err != nil {
		log.Printf("Error deleting account: %v", err)
		return err
	}

	return nil
}

func (s *PostgresStore) updateAccount(updatedAccount *Account) error {
	query := `
        UPDATE accounts
        SET first_name = $1,
            last_name = $2,
            hobby = $3,
            age = $4,
            balance = $5,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $6;
    `

	_, err := s.db.Exec(query,
		updatedAccount.FIRST_NAME,
		updatedAccount.LAST_NAME,
		updatedAccount.HOBBY,
		updatedAccount.AGE,
		updatedAccount.BALANCE,
		updatedAccount.ID,
	)
	if err != nil {
		log.Printf("Error updating account: %v", err)
		return err
	}

	return nil
}

func (s *PostgresStore) getAccountById(id int) (*Account, error) {
	query := `
		SELECT *
		FROM accounts
		WHERE id = $1;
	`

	account := &Account{}

	err := s.db.QueryRow(query, id).Scan(
		&account.ID,
		&account.FIRST_NAME,
		&account.LAST_NAME,
		&account.HOBBY,
		&account.AGE,
		&account.ACCOUNT,
		&account.BALANCE,
		&account.CREATED_AT,
		&account.UPDATED_AT,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account not found")
		}
		log.Printf("Error fetching account by ID: %v", err)
		return nil, err
	}

	return account, nil
}

func (s *PostgresStore) transferBalance(fromAccountNumber, toAccountNumber int64, amount float64) error {
	// Begin a new database transaction
	tx, err := s.db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return err
	}
	defer tx.Rollback() // Rollback the transaction if it's not committed

	// Check if both accounts exist
	existsQuery := `
		SELECT account_number
		FROM accounts
		WHERE account_number IN ($1, $2);
	`

	var accountNumbers []int64
	rows, err := tx.Query(existsQuery, fromAccountNumber, toAccountNumber)
	if err != nil {
		log.Printf("Error checking account existence: %v", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var accountNumber int64
		if err := rows.Scan(&accountNumber); err != nil {
			log.Printf("Error scanning row: %v", err)
			return err
		}
		accountNumbers = append(accountNumbers, accountNumber)
	}

	if len(accountNumbers) != 2 {
		return fmt.Errorf("one or both accounts not found")
	}

	// Fetch the current balance of the 'from' account
	fromAccountQuery := `
		SELECT balance
		FROM accounts
		WHERE account_number = $1;
	`

	var fromBalance float64
	err = tx.QueryRow(fromAccountQuery, fromAccountNumber).Scan(&fromBalance)
	if err != nil {
		log.Printf("Error fetching balance: %v", err)
		return err
	}

	// Check if the 'from' account has enough balance
	if fromBalance < amount {
		return fmt.Errorf("insufficient balance in the 'from' account")
	}

	// Update the balance of the 'from' account
	updateFromBalanceQuery := `
		UPDATE accounts
		SET balance = balance - $1,
			updated_at = CURRENT_TIMESTAMP
		WHERE account_number = $2;
	`

	_, err = tx.Exec(updateFromBalanceQuery, amount, fromAccountNumber)
	if err != nil {
		log.Printf("Error updating 'from' account balance: %v", err)
		return err
	}

	// Update the balance of the 'to' account
	updateToBalanceQuery := `
		UPDATE accounts
		SET balance = balance + $1,
			updated_at = CURRENT_TIMESTAMP
		WHERE account_number = $2;
	`

	_, err = tx.Exec(updateToBalanceQuery, amount, toAccountNumber)
	if err != nil {
		log.Printf("Error updating 'to' account balance: %v", err)
		return err
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}

	return nil
}

func (s *PostgresStore) getUserByUsername(username string) (*User, error) {
	query := `
		SELECT *
		FROM users
		WHERE username = $1;
	`

	user := &User{}

	err := s.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil if user not found
		}
		log.Printf("Error fetching user by username: %v", err)
		return nil, err
	}

	return user, nil
}

func (s *PostgresStore) createUser(user *User) error {
	query := `
		INSERT INTO users (username, password)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at;
	`

	err := s.db.QueryRow(
		query,
		user.Username,
		user.Password,
	).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		log.Printf("Error inserting user: %v", err)
		return err
	}

	return nil
}
