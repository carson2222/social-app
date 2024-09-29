package storage

import (
	"github.com/carson2222/social-app/types"
)

func (s *PostgresStore) createChatsTable() error {

	query := `CREATE TABLE IF NOT EXISTS chats (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_group BOOLEAN DEFAULT FALSE
);`

	_, err := s.db.Exec(query)

	return err
}

func (s *PostgresStore) createMessagesTable() error {
	query := `CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    chat_id INTEGER REFERENCES chats(id) ON DELETE CASCADE,
    sender_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    content TEXT NOT NULL,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

	_, err := s.db.Exec(query)

	return err
}
func (s *PostgresStore) createChatUsersTable() error {
	query := `CREATE TABLE IF NOT EXISTS chat_users (
    chat_id INTEGER REFERENCES chats(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (chat_id, user_id)
);
`

	_, err := s.db.Exec(query)

	return err
}

func (s *PostgresStore) GetUserChats(userId int) (map[int]bool, error) {
	query := `SELECT chat_id FROM chat_users WHERE user_id = $1;`

	rows, err := s.db.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chatIDs = make(map[int]bool)
	for rows.Next() {
		var chatId int
		if err := rows.Scan(&chatId); err != nil {
			return nil, err
		}
		chatIDs[chatId] = true
	}

	return chatIDs, nil
}

func (s *PostgresStore) NewMessage(message *types.ReceivedMessageWS) error {
	query := `INSERT INTO messages (chat_id, sender_id, content, sent_at) VALUES ($1, $2, $3, $4);`

	_, err := s.db.Exec(query, message.ChatId, message.SenderId, message.Content, message.SentAt)

	return err
}

func (s *PostgresStore) IsUserInChat(userId, chatId int) error {

	query := `SELECT EXISTS (SELECT 1 FROM chat_users WHERE user_id = $1 AND chat_id = $2);`

	var exists bool
	err := s.db.QueryRow(query, userId, chatId).Scan(&exists)

	if err != nil || !exists {
		return err
	}

	return nil
}
