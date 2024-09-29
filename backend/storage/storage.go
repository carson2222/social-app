package storage

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

//	type Storage interface {
//		createUser(*Credentials) (int, error)
//		authUser(*Credentials) (int, error)
//		createSession(int) (string, error)
//		verifySession(string) (bool, int, error)
//		killSession(string) error
//		initProfile(int) error
//		updateProfile(int, string, string, string, string) error
//		getProfileByID(int) (Profile, error)
//		areFriends(userId int, friendId int) (bool, error)
//		addFriend(senderId int, receiverId int) error
//		isRequestedFriend(senderId int, receiverId int) (bool, error)
//		acceptFriendRequest(userId int, senderId int) error
//		rejectFriendRequest(userId int, senderId int) error
//		removeFriend(userId int, friendId int) error
//	}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStorage() (*PostgresStore, error) {
	connStr := "user=admin dbname=postgres password=admin sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to pin	g database: %w", err)
	}

	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Init() error {

	tasks := []func() error{
		s.createExtensions,
		s.createUsersTable,
		s.createSessionsTable,
		s.createProfilesTable,
		s.createFriendsTable,
		s.createFriendRequestsTable,
		s.createChatsTable,
		s.createChatUsersTable,
		s.createMessagesTable,
	}

	for _, task := range tasks {
		if err := task(); err != nil {
			return err
		}
	}

	log.Println("Storage initialized")
	return nil
}

func (s *PostgresStore) createExtensions() error {
	_, err := s.db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto;")

	return err
}
