package main

import "math/rand"

type Account struct {
	ID         int
	FIRST_NAME string
	LAST_NAME  string
	HOBBY      string
	AGE        int
	ACCOUNT    int64
	BALANCE    int64
}

func newAccount(firstName, lastName, hobby string, age, balance int) *Account {
	return &Account{
		ID:         rand.Intn(10000),
		FIRST_NAME: firstName,
		LAST_NAME:  lastName,
		HOBBY:      hobby,
		ACCOUNT:    int64(rand.Intn(100000000)),
		BALANCE:    int64(balance),
		AGE:        age,
	}
}
