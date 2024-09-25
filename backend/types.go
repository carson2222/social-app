package main

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
