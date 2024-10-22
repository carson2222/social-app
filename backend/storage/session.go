package storage

import (
	"errors"
	"fmt"
	"time"
)

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

func (s *PostgresStore) CreateSession(user_id int) (string, error) {

	query := `INSERT INTO sessions (user_id, session_token)
VALUES ($1, encode($2::text::bytea, 'hex') || encode(gen_random_bytes(32), 'hex')) RETURNING session_token;`

	var sessionToken string
	err := s.db.QueryRow(query, user_id, user_id).Scan(&sessionToken)

	if err != nil {
		return "", err
	}

	return sessionToken, nil
}

func (s *PostgresStore) VerifySession(sessionToken string) (bool, int, error) {

	query := `SELECT is_valid, user_id, expires_at FROM sessions WHERE session_token = $1;`

	isValid := false
	expiresAt := time.Now()
	userId := -1

	err := s.db.QueryRow(query, sessionToken).Scan(&isValid, &userId, &expiresAt)
	if err != nil {
		return false, -1, err
	}

	if time.Now().After(expiresAt) {
		isValid = false
		err = errors.New("session expired")
	}

	if !isValid {
		err = errors.New("session invalid")
	}

	return isValid, userId, err
}

func (s *PostgresStore) KillSession(sessionToken string) error {

	query := `UPDATE sessions SET is_valid = false WHERE session_token = $1;`

	_, err := s.db.Exec(query, sessionToken)
	if err != nil {
		return err
	}

	return nil
}
