package types

import (
	"github.com/gorilla/websocket"
)

type Client struct {
	Conn    *websocket.Conn
	UserID  int
	ChatIDs map[int]bool
	Send    chan []byte
}

type IncomingBase struct {
	Type    string `json:"type"`    // The type of message, e.g., "message", "friend"
	Content string `json:"content"` // The actual content of the message
}

type OutgoingBase struct {
	Type       string `json:"type"`        // The type of message, e.g., "message", "friend"
	Content    string `json:"content"`     // The actual content of the message
	VerifyType string `json:"verify_type"` // Verification type, eg by chat_id or user_id
	VerifyID   int    `json:"verify_id"`   // Verification ID, eg chat_id or user_id
}
type SendMessage struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	ChatID  int    `json:"chat_id"`
}
