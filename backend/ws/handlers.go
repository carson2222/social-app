package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/carson2222/social-app/types"
)

// TODO: Save message in database
func (ws *WebSocketServer) handleMessage(client *types.Client, rawMessage []byte) {
	var message types.NewMessage

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
	var messageID int
	// Insert message into database
	if messageID, err = ws.storage.NewMessage(message.ChatID, client.UserID, message.Content, now); err != nil || messageID == -1 {
		log.Printf("Error inserting message into database: %v\n", err)
		return
	}

	// Format Data
	data := types.NewMessageData{
		Content:   message.Content,
		ChatID:    message.ChatID,
		SenderID:  client.UserID,
		SentAt:    now,
		MessageID: messageID,
	}

	dataRaw, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling message: %v\n", err)
		return
	}

	// Format the message that will be served to other users
	outgoingMsg := types.OutgoingBase{
		Type:       message.Type,
		Data:       json.RawMessage(dataRaw),
		VerifyIDs:  []int{message.ChatID},
		VerifyType: "chatID",
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
	var message types.NewChat

	err := json.Unmarshal(rawMessage, &message)
	if err != nil {
		log.Printf("Error unmarshaling message: %v\n", err)
		return
	}

	// Validate members
	areMembersValid := true
	for _, member := range message.Members {
		var exists bool
		if exists, err = ws.storage.IsUserExisting(member); err != nil {
			log.Printf("Error checking if user exists: %v\n", err)
			areMembersValid = false
		}

		if !exists {
			log.Printf("User %d does not exist\n", member)
			areMembersValid = false
		}
	}

	if !areMembersValid || len(message.Members) == 2 {
		log.Print("Invalid members")
		return
	}

	// TODO: Check if all users are friends (LATER)

	// If it's a private chat, check if it's already existing
	if len(message.Members) == 1 {
		if exists, err := ws.storage.IsPrivateChatExisting(message.Members[0], client.UserID); err != nil || !exists {
			log.Printf("Error checking if private chat exists: %v\n", err)
			return
		}
	}

	// Create new chat and broadcast message
	membersWithCreator := append([]int{client.UserID}, message.Members...)

	chatId, err := ws.storage.InitNewChat(message.ChatName, membersWithCreator)
	if err != nil {
		log.Printf("Error creating new chat: %v\n", err)
		return
	}

	// Create a data json
	data := types.NewChatData{
		ChatID:   chatId,
		ChatName: message.ChatName,
		Members:  membersWithCreator,
		SentAt:   time.Now(),
	}
	dataRaw, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error creating content json: %v\n", err)
		return
	}

	// Format the message that will be served to other users
	outgoingMsg := types.OutgoingBase{
		Type:       message.Type,
		Data:       json.RawMessage(dataRaw),
		VerifyIDs:  membersWithCreator,
		VerifyType: "userID",
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

func (ws *WebSocketServer) handleFriendRequest(client *types.Client, rawMessage []byte) {
	// Handle friend request logic here
}
