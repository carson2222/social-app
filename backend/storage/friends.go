package storage

import "fmt"

func (s *PostgresStore) createFriendsTable() error {
	query := `CREATE TABLE IF NOT EXISTS friends (
    user_id INTEGER REFERENCES users (id) ON DELETE CASCADE,
    friend_id INTEGER REFERENCES users (id) ON DELETE CASCADE,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, friend_id)
);`

	_, err := s.db.Exec(query)

	return err
}
func (s *PostgresStore) createFriendRequestsTable() error {
	query := `CREATE TABLE IF NOT EXISTS friend_requests (
		sender_id INTEGER REFERENCES users (id) ON DELETE CASCADE NOT NULL,
		receiver_id INTEGER REFERENCES users (id) ON DELETE CASCADE NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (sender_id, receiver_id)
	);`

	_, err := s.db.Exec(query)

	return err
}

func (s *PostgresStore) AreFriends(userId, friendId int) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM friends WHERE user_id = $1 AND friend_id = $2 OR user_id = $2 AND friend_id = $1);`

	var exists bool
	err := s.db.QueryRow(query, userId, friendId).Scan(&exists)

	return exists, err
}

func (s *PostgresStore) IsRequestedFriend(senderId, receiverId int) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM friend_requests WHERE sender_id = $1 AND receiver_id = $2);`

	var exists bool
	err := s.db.QueryRow(query, senderId, receiverId).Scan(&exists)

	return exists, err
}

func (s *PostgresStore) GetFriends(userId int) (map[int]bool, error) {
	query := `SELECT friend_id FROM friends WHERE user_id = $1;`

	rows, err := s.db.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var friendIDs = make(map[int]bool)
	for rows.Next() {
		var friendId int
		if err := rows.Scan(&friendId); err != nil {
			return nil, err
		}
		friendIDs[friendId] = true
	}

	return friendIDs, nil
}

func (s *PostgresStore) SendFR(senderId, receiverId int) error {
	query := `INSERT INTO friend_requests (sender_id, receiver_id) VALUES ($1, $2);`

	_, err := s.db.Exec(query, senderId, receiverId)
	return err
}
func (s *PostgresStore) RemoveFriend(userId, friendId int) error {
	query := `DELETE FROM friends WHERE user_id = $1 AND friend_id = $2 OR user_id = $2 AND friend_id = $1;`
	_, err := s.db.Exec(query, userId, friendId)
	return err
}

func (s *PostgresStore) AcceptFriendRequest(userId, senderId int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete friend request
	query := `DELETE FROM friend_requests WHERE sender_id = $1 AND receiver_id = $2;`
	_, err = tx.Exec(query, senderId, userId)
	if err != nil {
		return fmt.Errorf("failed to delete friend request: %w", err)
	}

	// Add friend relation
	query = `INSERT INTO friends (user_id, friend_id) VALUES ($1, $2), ($2, $1);`
	_, err = tx.Exec(query, userId, senderId)
	if err != nil {
		return fmt.Errorf("failed to add friend: %w", err)
	}

	return tx.Commit()
}

func (s *PostgresStore) RejectFriendRequest(userId, senderId int) error {
	query := `DELETE FROM friend_requests WHERE sender_id = $1 AND receiver_id = $2;`
	_, err := s.db.Exec(query, senderId, userId)
	return err
}
