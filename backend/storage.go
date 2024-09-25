package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type Storage interface {
	createUser(*Credentials) (int, error)
	authUser(*Credentials) (int, error)
	createSession(int) (string, error)
	verifySession(string) (bool, error)
	killSession(string) error
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

	if err := s.createExtensions(); err != nil {
		return err
	}

	if err := s.createUsersTable(); err != nil {
		return err
	}

	if err := s.createSessionsTable(); err != nil {
		return err
	}

	log.Println("Storage initialized")

	return nil
}

func (s *PostgresStore) createExtensions() error {
	_, err := s.db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto;")

	return err
}

func (s *PostgresStore) createUsersTable() error {
	query := `CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`

	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) createSessionsTable() error {
	SESSION_DURATION := 24 // Hours

	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS sessions (
		id SERIAL PRIMARY KEY,
		user_id INTEGER REFERENCES users (id) ON DELETE CASCADE NOT NULL,
		session_token TEXT UNIQUE NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		expires_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP + INTERVAL '%d hours'),
		last_active TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		is_valid BOOLEAN NOT NULL DEFAULT TRUE
	)`, SESSION_DURATION)

	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) createUser(c *Credentials) (int, error) {
	query := `INSERT INTO users (email, password) VALUES ($1, crypt($2, gen_salt('bf'))) RETURNING id;`
	ID := -1
	err := s.db.QueryRow(query, c.Email, c.Password).Scan(&ID)

	if err != nil {
		return -1, err
	}

	return ID, nil
}

func (s *PostgresStore) authUser(c *Credentials) (int, error) {
	query := `SELECT id FROM users WHERE email = $1 AND password = crypt($2, password) LIMIT 1;`

	ID := -1
	err := s.db.QueryRow(query, c.Email, c.Password).Scan(&ID)

	if err != nil {
		return -1, err
	}

	return ID, nil
}

func (s *PostgresStore) createSession(user_id int) (string, error) {

	query := `INSERT INTO sessions (user_id, session_token)
VALUES ($1, encode($2::text::bytea, 'hex') || encode(gen_random_bytes(32), 'hex')) RETURNING session_token;`

	var sessionToken string
	err := s.db.QueryRow(query, user_id, user_id).Scan(&sessionToken)

	if err != nil {
		return "", err
	}

	return sessionToken, nil
}

func (s *PostgresStore) verifySession(sessionToken string) (bool, error) {

	query := `SELECT is_valid FROM sessions WHERE session_token = $1;`

	isValid := false
	err := s.db.QueryRow(query, sessionToken).Scan(&isValid)
	if err != nil {
		return false, err
	}

	return isValid, nil
}

func (s *PostgresStore) killSession(sessionToken string) error {

	query := `UPDATE sessions SET is_valid = false WHERE session_token = $1;`

	_, err := s.db.Exec(query, sessionToken)
	if err != nil {
		return err
	}

	return nil
}
