package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateUser(*Credentials) error
}
type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStorage() (*PostgresStore, error) {
	connStr := "user=admin dbname=postgres password=admin sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
		return nil, err
	}

	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Init() error {

	if err := s.CreateExtensions(); err != nil {
		return err
	}

	if err := s.CreateUsersTable(); err != nil {
		return err
	}

	log.Println("Storage initialized")

	return nil
}

func (s *PostgresStore) CreateExtensions() error {
	_, err := s.db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto;")

	return err
}

func (s *PostgresStore) CreateUsersTable() error {
	query := `CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`

	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) CreateUser(c *Credentials) error {
	query := `INSERT INTO users (username, password) VALUES ($1, crypt($2, gen_salt('bf')));`

	_, err := s.db.Exec(query, c.Username, c.Password)

	if err != nil {
		return err
	}

	return nil
}
