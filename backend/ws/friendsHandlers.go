package ws

import (
	"encoding/json"
	"log"

	"github.com/carson2222/social-app/types"
)

func (ws *WebSocketServer) handleAcceptFR(client *types.Client, rawMessage []byte) {
	message := types.IncomingFR{}

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
	data := types.IncomingFRData{
		SenderID:   senderID,
		ReceiverID: userId,
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
	message := types.IncomingFR{}

	if err := json.Unmarshal(rawMessage, &message); err != nil {
		log.Printf("Error unmarshaling message: %v\n", err)
		return
	}

	userId := client.UserID
	friendId := message.SenderID
	if userId == friendId {
		log.Printf("User cannot reject their own friend request")
		return
	}

	// Check if users are already friends
	areFriends, err := ws.storage.AreFriends(userId, friendId)
	if err != nil {
		log.Print("Failed to check if users are friend:" + err.Error())
		return
	}

	if areFriends {
		log.Print("Users are already friends")
		return
	}

	// Check if friend request is already sent
	isRequested, err := ws.storage.IsRequestedFriend(friendId, userId)
	if err != nil {
		log.Print("Failed to check if user is requested friend with the friend:" + err.Error())
		return
	}

	if !isRequested {
		log.Print("There is no pending fr from this user")
		return
	}

	err = ws.storage.RejectFriendRequest(userId, friendId)
	if err != nil {
		log.Print("Failed to reject friend request:" + err.Error())
		return
	}

	// Create data
	data := types.IncomingFRData{
		SenderID:   friendId,
		ReceiverID: userId,
	}
	dataRaw, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling message: %v\n", err)
		return
	}

	// Create outgoing message
	outgoingMsg := types.OutgoingBase{
		Type:       "rejectFR",
		Data:       dataRaw,
		VerifyType: "userID",
		VerifyIDs:  []int{userId},
	}
	marshaledMsg, err := json.Marshal(outgoingMsg)
	if err != nil {
		log.Printf("Error marshaling message: %v\n", err)
		return
	}

	// Broadcast message
	ws.broadcast <- marshaledMsg
}

func (ws *WebSocketServer) handleSendFR(client *types.Client, rawMessage []byte) {
	message := types.SendFR{}

	if err := json.Unmarshal(rawMessage, &message); err != nil {
		log.Printf("Error unmarshaling message: %v\n", err)
		return
	}

	userId := client.UserID
	friendId := message.ReceiverID

	if userId == friendId {
		log.Print("User cannot send friend request to themselves")
		return
	}

	// Check if users are already friends
	areFriends, err := ws.storage.AreFriends(userId, friendId)
	if err != nil {
		log.Print("Failed to check if users are friend:" + err.Error())
		return
	}

	if areFriends {
		log.Print("Users are already friends")
		return
	}

	// Check if friend request is already sent
	isRequested1, err1 := ws.storage.IsRequestedFriend(userId, friendId)
	isRequested2, err2 := ws.storage.IsRequestedFriend(friendId, userId)

	if err1 != nil || err2 != nil {
		log.Print("Failed to check if user is requested friend with the friend:" + err1.Error() + err2.Error())
		return
	}

	if isRequested1 || isRequested2 {
		log.Print("User already requested to be friend with the friend")
		return
	}

	err = ws.storage.SendFR(userId, friendId)
	if err != nil {
		log.Print("Failed to add friend:" + err.Error())
		return
	}

	// Create data
	data := types.SendFRData{
		SenderID:   userId,
		ReceiverID: friendId,
	}
	dataRaw, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling message: %v\n", err)
		return
	}

	// Create outgoing message
	outgoingMsg := types.OutgoingBase{
		Type:       "sendFR",
		Data:       dataRaw,
		VerifyType: "userID",
		VerifyIDs:  []int{userId, friendId},
	}
	marshaledMsg, err := json.Marshal(outgoingMsg)
	if err != nil {
		log.Printf("Error marshaling message: %v\n", err)
		return
	}

	// Broadcast message
	ws.broadcast <- marshaledMsg
}

func (ws *WebSocketServer) handleRemoveFriend(client *types.Client, rawMessage []byte) {

	message := types.RemoveFriend{}
	if err := json.Unmarshal(rawMessage, &message); err != nil {
		log.Print("Error unmarshaling message: " + err.Error())
	}

	userId := client.UserID
	friendId := message.FriendID

	if userId == friendId {
		log.Print("User cannot remove friend from themselves")
		return
	}

	// Check if users are already friends
	areFriends, err := ws.storage.AreFriends(userId, friendId)
	if err != nil {
		log.Print("Failed to check if users are friend:" + err.Error())
		return
	}

	if !areFriends {
		log.Print("Users are not friends")
		return
	}

	err = ws.storage.RemoveFriend(userId, friendId)
	if err != nil {
		log.Print("Failed to remove friend:" + err.Error())
		return
	}

	// Create data
	data := types.RemoveFriendData{
		FriendID: friendId,
	}
	dataRaw, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling message: %v\n", err)
		return
	}

	// Create outgoing message
	outgoingMsg := types.OutgoingBase{
		Type:       "removeFriend",
		Data:       dataRaw,
		VerifyType: "userID",
		VerifyIDs:  []int{userId},
	}

	marshaledMsg, err := json.Marshal(outgoingMsg)
	if err != nil {
		log.Printf("Error marshaling message: %v\n", err)
		return
	}

	// Broadcast message
	ws.broadcast <- marshaledMsg
}
