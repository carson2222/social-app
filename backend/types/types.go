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
