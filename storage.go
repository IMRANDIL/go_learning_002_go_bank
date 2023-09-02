package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type Storage interface {
	createAccount(*Account) error
	deleteAccount(int) error
	updateAccount(*Account) error
	getAccountById(int) (*Account, error)
	allAccounts(int) ([]*Account, error)
	transferBalance(int64, int64, float64) error
	getUserByUsername(string) (*User, error)
	createUser(*User) error
	authenticateUser(string, string) (*User, error)
	getUserDetails(int) (getUserDetailsRequest, error)
}

type PostgresStore struct {
	db *sql.DB
}

func newPostgesStore() (*PostgresStore, error) {

	password := os.Getenv("PASSWORD")
	connStr := fmt.Sprintf("user=postgres dbname=goLearning_db password=%s sslmode=disable", password)

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
            user_id INT NOT NULL,
            first_name VARCHAR(255) NOT NULL,
            last_name VARCHAR(255) NOT NULL,
            hobby VARCHAR(255),
            age INT,
            account_number BIGINT NOT NULL,
            balance DECIMAL(15, 2) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
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
	// Check if an account with the same user ID already exists
	query := `
        SELECT COUNT(*) FROM accounts WHERE user_id = $1
    `
	var count int
	err := s.db.QueryRow(query, account.UserID).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("user ID already exists")
	}

	// Proceed to insert the new account
	insertQuery := `
        INSERT INTO accounts (user_id, first_name, last_name, hobby, age, account_number, balance)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, user_id, first_name, last_name, hobby, age, account_number, balance, created_at, updated_at
    `

	err = s.db.QueryRow(
		insertQuery,
		account.UserID,
		account.FIRST_NAME,
		account.LAST_NAME,
		account.HOBBY,
		account.AGE,
		account.ACCOUNT,
		account.BALANCE,
	).Scan(
		&account.ID,
		&account.UserID,
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

func (s *PostgresStore) allAccounts(batchSize int) ([]*Account, error) {
	var accounts []*Account
	offset := 0

	for {
		query := `SELECT * FROM accounts ORDER BY id ASC LIMIT $1 OFFSET $2;`

		rows, err := s.db.Query(query, batchSize, offset)
		if err != nil {
			log.Printf("Error fetching accounts: %v", err)
			return nil, err
		}

		var batchAccounts []*Account

		for rows.Next() {
			account := &Account{}
			err := rows.Scan(
				&account.ID,
				&account.UserID,
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
			batchAccounts = append(batchAccounts, account)
		}

		if err := rows.Err(); err != nil {
			log.Printf("Error iterating rows: %v", err)
			return nil, err
		}

		// If there are no more rows, break the loop
		if len(batchAccounts) == 0 {
			break
		}

		accounts = append(accounts, batchAccounts...)
		offset += batchSize
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
			return nil, nil // Return nil and no error if user not found
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

func (s *PostgresStore) getUserDetails(id int) (getUserDetailsRequest, error) {
	// Initialize an empty getUserDetailsRequest struct to store the result
	var user getUserDetailsRequest

	query := `
		SELECT
			u.id,
			u.username,
			a.id AS account_id,
			a.first_name AS account_first_name,
			a.last_name AS account_last_name,
			a.hobby AS account_hobby,
			a.age AS account_age,
			a.account_number AS bank_account,
			a.balance AS account_balance
			
		FROM
			users u
		LEFT JOIN
			accounts a
		ON
			u.id = a.user_id
		WHERE
			u.id = $1
	`

	rows, err := s.db.Query(query, id)
	//fmt.Println(rows)
	if err != nil {
		log.Printf("Error retrieving user and accounts: %v", err)
		return user, err
	}
	defer rows.Close()

	// Map to store user and associated accounts
	userAccountsMap := make(map[int]getUserDetailsRequest)

	// Iterate through the rows and populate the user and accounts map
	for rows.Next() {
		var userID int
		var username string
		var accountID int
		var accountFirstName, accountLastName, accountHobby string
		var accountAge int
		var accountAccount int64
		var accountBalance float64

		err := rows.Scan(
			&userID,
			&username,
			&accountID,
			&accountFirstName,
			&accountLastName,
			&accountHobby,
			&accountAge,
			&accountAccount,
			&accountBalance,
		)

		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return user, err
		}

		// Check if the user is already in the map, if not, create a new user
		if _, ok := userAccountsMap[userID]; !ok {
			user = getUserDetailsRequest{
				ID:       userID,
				Username: username,
			}
			userAccountsMap[userID] = user
		}

		// Create a temporary user variable to update the Accounts field
		tempUser := userAccountsMap[userID]

		// If there is an associated account, add it to the user's accounts slice
		if accountID != 0 {
			account := AccountsRequest{
				ID:         accountID,
				FIRST_NAME: accountFirstName,
				LAST_NAME:  accountLastName,
				HOBBY:      accountHobby,
				AGE:        accountAge,
				ACCOUNT:    accountAccount,
				BALANCE:    accountBalance,
			}
			tempUser.Accounts = append(tempUser.Accounts, account)
		}

		// Update the user in the map with the temporary user
		userAccountsMap[userID] = tempUser
	}

	// Check if the user exists, and return it along with associated accounts
	if user, ok := userAccountsMap[id]; ok {
		return user, nil
	}

	return user, fmt.Errorf("User not found")
}

func (s *PostgresStore) authenticateUser(username, password string) (*User, error) {
	user, err := s.getUserByUsername(username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil // User not found
	}

	// Compare the hashed password with the provided password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		// Return a custom error message here
		return nil, fmt.Errorf("invalid password") //password don't match
	}

	return user, nil
}
