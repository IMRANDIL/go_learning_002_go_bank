package main

import (
	"math/rand"
	"strconv"
)

type createAccountRequest struct {
	FIRST_NAME string `json:"firstName"`
	LAST_NAME  string `json:"lastName"`
	HOBBY      string `json:"hobby"`
	AGE        int    `json:"age"`
}

type Account struct {
	ID         int    `json:"id"`
	FIRST_NAME string `json:"firstName"`
	LAST_NAME  string `json:"lastName"`
	HOBBY      string `json:"hobby"`
	AGE        int    `json:"age"`
	ACCOUNT    int64  `json:"account"`
	BALANCE    string `json:"balance"`
}

func newAccount(firstName, lastName, hobby string, age int) *Account {
	return &Account{
		//ID:         rand.Intn(10000),
		FIRST_NAME: firstName,
		LAST_NAME:  lastName,
		HOBBY:      hobby,
		ACCOUNT:    int64(rand.Intn(100000000)),
		BALANCE:    strconv.Itoa(rand.Intn(1000000000)), // Convert int to string
		AGE:        age,
	}
}
