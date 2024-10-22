package ws

import (
	"encoding/json"
	"log"

	"github.com/carson2222/social-app/types"
)

func (ws *WebSocketServer) handleMessage(client *types.Client, rawMessage []byte) {
	var message types.SendMessage

	err := json.Unmarshal(rawMessage, &message)
	if err != nil {
		log.Printf("Error unmarshaling message: %v\n", err)
		return
	}

	outgoingMsg := types.OutgoingBase{
		Type:       message.Type,
		Content:    message.Content,
		VerifyID:   message.ChatID,
		VerifyType: "chatID",
	}
	// Handle chat message logic here
	marshaledMsg, err := json.Marshal(outgoingMsg)
	if err != nil {
		log.Printf("Error marshaling message: %v\n", err)
		return
	}

	// err = client.Conn.WriteMessage(websocket.TextMessage, marshaledMsg)
	// if err != nil {
	// 	log.Printf("Error writing message: %v\n", err)
	// 	return
	// }
	ws.broadcast <- marshaledMsg
}

func (ws *WebSocketServer) handleFriendRequest(client *types.Client, rawMessage []byte) {
	// Handle friend request logic here
}
