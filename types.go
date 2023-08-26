package main

type Account struct {
	ID         int
	FIRST_NAME string
	LAST_NAME  string
	HOBBY      string
	AGE        int
	ACCOUNT    int64
	BALANCE    int64
}

func newAccount(firstName, lastName, hobby string, account, balance int) *Account {
	return &Account{
		ID:         2,
		FIRST_NAME: firstName,
		LAST_NAME:  lastName,
		HOBBY:      hobby,
		ACCOUNT:    int64(account),
		BALANCE:    int64(balance),
	}
}
