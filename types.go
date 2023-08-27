package main

import (
	"math/rand"
	"time"

	_ "github.com/google/uuid"
)

type createAccountRequest struct {
	FIRST_NAME string  `json:"firstName"`
	LAST_NAME  string  `json:"lastName"`
	HOBBY      string  `json:"hobby"`
	AGE        int     `json:"age"`
	BALANCE    float64 `json:"balance"`
}

type Account struct {
	ID         int     `json:"id"`
	FIRST_NAME string  `json:"firstName"`
	LAST_NAME  string  `json:"lastName"`
	HOBBY      string  `json:"hobby"`
	AGE        int     `json:"age"`
	ACCOUNT    int64   `json:"account"`
	BALANCE    float64 `json:"balance"`
	CREATED_AT string  `json:"created_at"`
	UPDATED_AT string  `json:"updated_at"`
}

func newAccount(firstName, lastName, hobby string, age int, balance float64) *Account {
	return &Account{
		//ID:         rand.Intn(10000),
		FIRST_NAME: firstName,
		LAST_NAME:  lastName,
		HOBBY:      hobby,
		ACCOUNT:    generateUniqueAccountNumber(),
		AGE:        age,
		BALANCE:    balance,
	}
}

func generateUniqueAccountNumber() int64 {
	// Seed the random number generator with the current time
	rand.Seed(time.Now().UnixNano())

	// Generate a random int64 within the range of a bigint column
	return rand.Int63n(9223372036854775807) // Max value for a signed 64-bit integer
}
