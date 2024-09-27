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
