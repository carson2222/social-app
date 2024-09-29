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

type SentMessageWS struct {
	ChatId  int    `json:"chat_id"`
	Content string `json:"content"`
}
type ReceivedMessageWS struct {
	ChatId   int       `json:"chat_id"`
	SenderId int       `json:"sender_id"`
	Content  string    `json:"content"`
	SentAt   time.Time `json:"sent_at"`
}
