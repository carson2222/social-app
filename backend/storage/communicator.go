package storage

import (
	"time"
)

func (s *PostgresStore) createChatsTable() error {

	query := `CREATE TABLE IF NOT EXISTS chats (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_group BOOLEAN DEFAULT FALSE,
		name TEXT
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

func (s *PostgresStore) NewMessage(chatID int, senderID int, content string, sentAt time.Time) (int, error) {
	query := `INSERT INTO messages (chat_id, sender_id, content, sent_at) VALUES ($1, $2, $3, $4) RETURNING id;`

	messageID := -1
	err := s.db.QueryRow(query, chatID, senderID, content, sentAt).Scan(&messageID)

	return messageID, err
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

func (s *PostgresStore) InitNewChat(chatName string, members []int) (int, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return -1, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	isGroup := false
	if len(members) > 2 {
		isGroup = true
	}

	var chatNameValue interface{}
	if chatName == "" {
		chatNameValue = nil
	} else {
		chatNameValue = chatName
	}
	// Create new chat
	query1 := `INSERT INTO chats (is_group, name) VALUES ($1, $2) RETURNING id;`

	var chatId int
	err = tx.QueryRow(query1, isGroup, chatNameValue).Scan(&chatId)
	if err != nil {
		return -1, err
	}

	// Add members to chat
	query2 := `INSERT INTO chat_users (chat_id, user_id) VALUES ($1, $2);`

	for _, member := range members {
		_, err = tx.Exec(query2, chatId, member)
		if err != nil {
			return -1, err
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return -1, err
	}

	return chatId, nil
}

func (s *PostgresStore) IsPrivateChatExisting(user1 int, user2 int) (bool, error) {
	var exists bool

	query := `SELECT COUNT(id) > 0 AS chat_exists
FROM chats
JOIN chat_users cu1 ON chats.id = cu1.chat_id
JOIN chat_users cu2 ON chats.id = cu2.chat_id
WHERE chats.is_group = false
AND cu1.user_id = $1
AND cu2.user_id = $2
`

	err := s.db.QueryRow(query, user1, user2).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// func (s *PostgresStore) GetChatsInfo(userId int) error {

// 	query := `SELECT chats.id as chat_id, created_at, is_group, chats.name, messages.sender_id, messages.content, messages.sent_at,
// profiles.name, profiles.surname, profiles.pfp
// FROM chats JOIN chat_users ON chat_users.chat_id = chats.id LEFT JOIN LATERAL (SELECT * FROM messages WHERE chat_id = chats.id
// ORDER BY sent_at desc LIMIT 1) messages on true LEFT JOIN profiles on messages.sender_id = profiles.user_id WHERE chat_users.user_id = 26`
// }
