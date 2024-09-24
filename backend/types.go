package main

import "time"

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json	:"created_at"`
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
