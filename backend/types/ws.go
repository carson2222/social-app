package types

import (
	"encoding/json"
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
	Type       string          `json:"type"`
	Data       json.RawMessage `json:"Data"`
	VerifyType string          `json:"verify_type"`
	VerifyIDs  []int           `json:"verify_id"`
}

type Final struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type NewMessage struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	ChatID  int    `json:"chat_id"`
}

type IncomingFR struct {
	Type     string `json:"type"`
	SenderID int    `json:"sender_id"`
}

type IncomingFRData struct {
	SenderID   int `json:"sender_id"`
	ReceiverID int `json:"accepter_id"`
}

type SendFR struct {
	Type       string `json:"type"`
	ReceiverID int    `json:"receiver"`
}

type SendFRData struct {
	ReceiverID int `json:"receiver_id"`
	SenderID   int `json:"sender_id"`
}
type NewMessageData struct {
	Content   string    `json:"content"`
	ChatID    int       `json:"chat_id"`
	SenderID  int       `json:"sender_id"`
	SentAt    time.Time `json:"sent_at"`
	MessageID int       `json:"message_id"`
}

type NewChat struct {
	Type     string `json:"type"`
	Members  []int  `json:"members"`
	ChatName string `json:"chat_name"`
}

type NewChatData struct {
	ChatID   int       `json:"chat_id"`
	Members  []int     `json:"members"`
	ChatName string    `json:"chat_name"`
	SentAt   time.Time `json:"sent_at"`
}
