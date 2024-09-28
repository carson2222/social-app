package storage

import "github.com/carson2222/social-app/types"

func (s *PostgresStore) createProfilesTable() error {

	query := `CREATE TABLE IF NOT EXISTS profiles (
		user_id INTEGER REFERENCES users (id) ON DELETE CASCADE NOT NULL,
		name TEXT,
		surname TEXT,
		bio TEXT,
		pfp TEXT
	)`

	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) InitProfile(id int) error {
	query := `INSERT INTO profiles (user_id) VALUES ($1);`

	_, err := s.db.Exec(query, id)
	return err
}

func (s *PostgresStore) GetProfileByID(id int) (types.Profile, error) {
	query := `SELECT user_id as id, name, surname, bio, pfp FROM profiles WHERE user_id = $1;`

	var profile types.Profile
	err := s.db.QueryRow(query, id).Scan(&profile.ID, &profile.Name, &profile.Surname, &profile.Bio, &profile.Pfp)

	if err != nil {
		return types.Profile{}, err
	}

	return profile, nil

}

func (s *PostgresStore) UpdateProfile(id int, name, surname, bio, pfp string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if name != "" {
		_, err := tx.Exec("UPDATE profiles SET name = $1 WHERE user_id = $2", name, id)
		if err != nil {
			return err
		}
	}

	if surname != "" {
		_, err := tx.Exec("UPDATE profiles SET surname = $1 WHERE user_id = $2", surname, id)
		if err != nil {
			return err
		}
	}

	if bio != "" {
		_, err := tx.Exec("UPDATE profiles SET bio = $1 WHERE user_id = $2", bio, id)
		if err != nil {
			return err
		}
	}

	if pfp != "" {
		_, err := tx.Exec("UPDATE profiles SET pfp = $1 WHERE user_id = $2", pfp, id)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
