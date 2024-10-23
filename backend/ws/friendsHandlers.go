package ws

import (
	"encoding/json"
	"log"

	"github.com/carson2222/social-app/types"
)

func (ws *WebSocketServer) handleAcceptFR(client *types.Client, rawMessage []byte) {
	message := types.AcceptFR{}

	if err := json.Unmarshal(rawMessage, &message); err != nil {
		log.Printf("Error unmarshaling message: %v\n", err)
		return
	}

	userId := client.UserID
	senderID := message.SenderID
	if userId == senderID {
		log.Printf("User cannot accept their own friend request")
		return
	}

	// Check if users are already friends
	areFriends, err := ws.storage.AreFriends(userId, senderID)
	if err != nil || areFriends {
		log.Printf("Users are already friends " + err.Error())
		return
	}

	// Check if user is already requested
	isRequested, err := ws.storage.IsRequestedFriend(senderID, userId)
	if err != nil {
		log.Printf("Failed to check if user is requested friend with the friend:" + err.Error())
		return
	}

	if !isRequested {
		log.Printf("There is no pending fr from this user")
		return
	}

	err = ws.storage.AcceptFriendRequest(userId, senderID)
	if err != nil {
		log.Printf("Failed to accept friend request:" + err.Error())
		return
	}

	// Create data
	data := types.AcceptFRData{
		SenderID:   senderID,
		AccepterID: userId,
	}
	dataRaw, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling message: %v\n", err)
		return
	}

	// Create outgoing message
	outgoingMsg := types.OutgoingBase{
		Type:       "acceptFR",
		Data:       dataRaw,
		VerifyType: "userID",
		VerifyIDs:  []int{senderID, userId},
	}
	marshaledMsg, err := json.Marshal(outgoingMsg)
	if err != nil {
		log.Printf("Error marshaling message: %v\n", err)
		return
	}

	// Broadcast message
	ws.broadcast <- marshaledMsg
}

func (ws *WebSocketServer) handleRejectFR(client *types.Client, rawMessage []byte) {
	// Handle friend request logic here
}

func (ws *WebSocketServer) handleSendFR(client *types.Client, rawMessage []byte) {
	// Handle friend request logic here
}

func (ws *WebSocketServer) handleRemoveFriend(client *types.Client, rawMessage []byte) {
	// Handle friend request logic here
}
