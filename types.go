package main

import (
	"math/rand"
	"time"
)

type LoginRequest struct {
	Number   int64  `json:"number"`
	Password string `json:"password"`
}

type TransferAmountRequest struct {
	ToAccount int `json:"to_account"`
	Amount    int `json:"amount"`
}

type CreateAccountRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password"`
}

type Account struct {
	Id        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Balance   int64     `json:"balance"`
	Number    int64     `json:"number"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

func NewAccount(firstName, lastName, hashedPassword string) *Account {
	return &Account{
		FirstName: firstName,
		LastName:  lastName,
		Number:    int64(rand.Intn(1000000)),
		Password:  hashedPassword,
		CreatedAt: time.Now().UTC(),
	}
}
