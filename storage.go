package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccountById(int) (*Account, error)
	GetAccounts() ([]*Account, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	connStr := "user=postgres dbname=postgres password=root sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Print("🗃️ connected to database")
	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Init() error {
	return s.CreateAccountTable()
}

func (s *PostgresStore) CreateAccountTable() error {
	query := `create table  if not exists account (
	id serial primary key,
	first_name varchar(50),
	last_name varchar(50),
	number serial,
	balance serial,
	created_at timestamp
	)`
	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) CreateAccount(a *Account) error {
	sqlQuery := `insert into 
	account (first_name, last_name, number, balance, created_at) 
	values($1, $2, $3, $4, $5)`

	_, err := s.db.Query(sqlQuery,
		a.FirstName,
		a.LastName,
		a.Number,
		a.Balance,
		a.CreatedAt)

	return err
}

func (s *PostgresStore) DeleteAccount(id int) error {
	query := "delete from account where id = $1"
	_, err := s.db.Query(query, id)
	return err
}

func (s *PostgresStore) UpdateAccount(a *Account) error {
	query := `update account set balance = $1 where id = $2`
	_, err := s.db.Query(query, a.Balance, a.Id)
	return err
}

func (s *PostgresStore) GetAccountById(id int) (*Account, error) {
	sqlQuery := "select * from account where id = $1"
	rows, err := s.db.Query(sqlQuery, id)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoAccount(rows)
	}
	return nil, fmt.Errorf("account %d not found", id)
}

func (s *PostgresStore) GetAccounts() ([]*Account, error) {
	sqlQuery := "select * from account"
	rows, err := s.db.Query(sqlQuery)
	if err != nil {
		return nil, err
	}

	accounts := []*Account{}
	for rows.Next() {
		account, err := scanIntoAccount(rows)

		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

func scanIntoAccount(rows *sql.Rows) (*Account, error) {
	account := &Account{}
	err := rows.Scan(
		&account.Id,
		&account.FirstName,
		&account.LastName,
		&account.Number,
		&account.Balance,
		&account.CreatedAt,
	)

	return account, err

}
