package types

import (
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn    *websocket.Conn
	UserID  int
	ChatIDs map[int]bool
	Send    chan []byte
}

type IncomingBase struct {
	Type string `json:"type"`
}

type OutgoingBase struct {
	Type       string    `json:"type"`
	Content    string    `json:"content"`
	VerifyType string    `json:"verify_type"`
	VerifyID   int       `json:"verify_id"`
	SentAt     time.Time `json:"sent_at"`
}
type SendMessage struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	ChatID  int    `json:"chat_id"`
}
