package main

import "math/rand"

type Account struct {
	ID         int    `json:"id"`
	FIRST_NAME string `json:"firstName"`
	LAST_NAME  string `json:"lastName"`
	HOBBY      string `json:"hobby"`
	AGE        int    `json:"age"`
	ACCOUNT    int64  `json:"account"`
	BALANCE    int64  `json:"balance"`
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
