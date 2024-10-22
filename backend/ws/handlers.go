package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/carson2222/social-app/types"
)

// TODO: Save message in database
func (ws *WebSocketServer) handleMessage(client *types.Client, rawMessage []byte) {
	var message types.SendMessage

	err := json.Unmarshal(rawMessage, &message)
	if err != nil {
		log.Printf("Error unmarshaling message: %v\n", err)
		return
	}

	// Validate message content
	if message.Content == "" {
		log.Print("Empty message")
		return
	}

	// Check if user is in chat
	if err := ws.storage.IsUserInChat(client.UserID, message.ChatID); err != nil {
		log.Printf("Error checking if user is in chat: %v\n", err)
		return
	}

	now := time.Now()

	// Insert message into database
	if err := ws.storage.NewMessage(message.ChatID, client.UserID, message.Content, now); err != nil {
		log.Printf("Error inserting message into database: %v\n", err)
		return
	}

	// Format the message that will be served to other users
	outgoingMsg := types.OutgoingBase{
		Type:       message.Type,
		Content:    message.Content,
		VerifyID:   message.ChatID,
		VerifyType: "chatID",
		SentAt:     now,
	}

	// Marshal the outgoing message
	marshaledMsg, err := json.Marshal(outgoingMsg)
	if err != nil {
		log.Printf("Error marshaling message: %v\n", err)
		return
	}

	// Broadcast the message
	ws.broadcast <- marshaledMsg
}

func (ws *WebSocketServer) handleNewChat(client *types.Client, rawMessage []byte) {

}

func (ws *WebSocketServer) handleFriendRequest(client *types.Client, rawMessage []byte) {
	// Handle friend request logic here
}
