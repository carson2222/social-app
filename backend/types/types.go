package types

import "time"

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json	:"created_at"`
}

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SuccessAuthResponse struct {
	SessionId string `json:"session_id"`
	Status    string `json:"status"`
	Action    string `json:"action"`
}

type ProfileRequest struct {
	Name    string `json:"name"`
	Surname string `json:"surname"`
	Bio     string `json:"bio"`
	Pfp     bool   `json:"pfp"`
}

type GetProfileRequest struct {
	ID int `json:"id"`
}

type Profile struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Surname string `json:"surname"`
	Bio     string `json:"bio"`
	Pfp     string `json:"pfp"`
}

type BaseMessageWs struct {
	Type string `json:"type"`
}
type SentMessageNewMessageWS struct {
	Type    string `json:"type"`
	ChatId  int    `json:"chat_id"`
	Content string `json:"content"`
}

// {
// 	"type": "new_message",
// 	"chat_id": 1,
// 	"content": "Test"
// }
type SentMessageNewChatWS struct {
	Type    string `json:"type"`
	Members []int  `json:"members"`
	Name    string `json:"name"`
}

// {
// 	"type": "new_chat",
// 	"members": [27],
// 	"name": "x"
// }
type ReceivedMessageNewMessageWS struct {
	Type     string    `json:"type"`
	ChatId   int       `json:"chat_id"`
	SenderId int       `json:"sender_id"`
	Content  string    `json:"content"`
	SentAt   time.Time `json:"sent_at"`
}

type ReceivedMessageNewChatWS struct {
	Type    string    `json:"type"`
	ChatId  int       `json:"chat_id"`
	Members []int     `json:"members"`
	Name    string    `json:"name"`
	SentAt  time.Time `json:"sent_at"`
}

type ChatShortInfo struct {
	ChatId            int       `json:"chat_id"`
	CreatedAt         time.Time `json:"created_at"`
	IsGroup           bool      `json:"is_group"`
	Name              string    `json:"name"`
	LastSenderId      int       `json:"last_sender_id"`
	Message           string    `json:"message"`
	LastActivity      string    `json:"last_activity"`
	LastSenderName    string    `json:"last_sender_name"`
	LastSenderSurname string    `json:"last_sender_surname"`
	LastSenderPfp     string    `json:"last_sender_pfp"`
}
