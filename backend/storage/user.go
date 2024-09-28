package storage

import "github.com/carson2222/social-app/types"

func (s *PostgresStore) AuthUser(c *types.Credentials) (int, error) {
	query := `SELECT id FROM users WHERE email = $1 AND password = crypt($2, password) LIMIT 1;`

	ID := -1
	err := s.db.QueryRow(query, c.Email, c.Password).Scan(&ID)

	if err != nil {
		return -1, err
	}

	return ID, nil
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

func (s *PostgresStore) CreateUser(c *types.Credentials) (int, error) {
	query := `INSERT INTO users (email, password) VALUES ($1, crypt($2, gen_salt('bf'))) RETURNING id;`
	ID := -1
	err := s.db.QueryRow(query, c.Email, c.Password).Scan(&ID)

	if err != nil {
		return -1, err
	}

	return ID, nil
}
