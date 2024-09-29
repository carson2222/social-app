package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/carson2222/social-app/types"
	"github.com/carson2222/social-app/utils"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	CheckOrigin: func(r *http.Request) bool {
		return r.Host == "localhost:3000"
	},
}

type Client struct {
	conn    *websocket.Conn
	userID  int
	chatIDs map[int]bool
	send    chan []byte
}

var clients = make(map[*Client]bool)
var clientsMu sync.Mutex
var broadcast = make(chan []byte)

// define our WebSocket endpoint
func (s *APIServer) serveCommunicatorWs(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Host)

	userId, err := s.authenticateWebSocket(r)
	if err != nil || userId == -1 {
		log.Println("failed to authenticate websocket", err)
		utils.WriteJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	chatIDs, err := s.storage.GetUserChats(userId)
	if err != nil {
		log.Println(err)
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	// upgrade this connection to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	client := &Client{
		conn:    conn,
		userID:  userId,
		chatIDs: chatIDs,
		send:    make(chan []byte),
	}

	clientsMu.Lock()
	clients[client] = true
	clientsMu.Unlock()

	go handleReads(client, s)
	go handleWrites(client)
}

func handleReads(client *Client, s *APIServer) {
	defer func() {
		delete(clients, client)
		client.conn.Close()
	}()

	for {
		_, messageBytes, err := client.conn.ReadMessage()

		if err != nil {
			log.Printf("error reading message: %v", err)
			break
		}

		// Unmarshal to get message type
		var baseMsg types.BaseMessageWs
		if err := json.Unmarshal(messageBytes, &baseMsg); err != nil {
			log.Printf("error unmarshaling base message: %v", err)
			continue
		}

		switch baseMsg.Type {
		case "new_message":
			// Unmarshal to get message json
			var messageData types.SentMessageNewMessageWS
			if err := json.Unmarshal(messageBytes, &messageData); err != nil {
				log.Printf("error unmarshaling message: %v", err)
				continue
			}

			// Check if user belongs to the chat
			if err := s.storage.IsUserInChat(client.userID, messageData.ChatId); err != nil {
				log.Printf("User does not belong to the chat: %v", err)
				break
			}

			upgradedMessage := &types.ReceivedMessageNewMessageWS{
				Type:     "new_message",
				ChatId:   messageData.ChatId,
				SenderId: client.userID,
				Content:  messageData.Content,
				SentAt:   time.Now(),
			}

			// Save message to the database
			if err := s.storage.NewMessage(upgradedMessage); err != nil {
				log.Printf("error saving message to the database: %v", err)
				break
			}

			upgradedMessageBytes, err := json.Marshal(upgradedMessage)
			if err != nil {
				log.Printf("error marshaling message: %v", err)
				break
			}

			broadcast <- upgradedMessageBytes

		case "new_chat":
			// TODO: If channel is not group type, check if it is already created
			// Unmarshal to get new chat json
			var chatData types.SentMessageNewChatWS
			if err := json.Unmarshal(messageBytes, &chatData); err != nil {
				log.Printf("error unmarshaling new chat: %v", err)
				continue
			}

			// Get chat creator friend list
			friendList, err := s.storage.GetFriends(client.userID)
			if err != nil {
				log.Printf("error getting friend list: %v", err)
				continue
			}

			// Check if all chat members are in the friend list
			var membersAreFriends = true
			for _, member := range chatData.Members {
				if !friendList[member] {
					membersAreFriends = false
					log.Printf("user %d is not in the friend list", member)
					continue
				}
			}

			if !membersAreFriends {
				log.Printf("not all members are in the friend list")
				continue
			}

			// Add creator to member list
			chatData.Members = append(chatData.Members, client.userID)

			// Initialize new chat
			chatId, err := s.storage.InitNewChat(chatData.Name, chatData.Members)
			if err != nil || chatId == -1 {
				log.Printf("error creating new chat: %v", err)
				continue
			}

			upgradedMessage := &types.ReceivedMessageNewChatWS{
				Type:    "new_chat",
				ChatId:  chatId,
				Members: chatData.Members,
				Name:    chatData.Name,
				SentAt:  time.Now(),
			}

			upgradedMessageBytes, err := json.Marshal(upgradedMessage)
			if err != nil {
				log.Printf("error marshaling message: %v", err)
				break
			}

			log.Printf("new chat created: %d", chatId)
			broadcast <- upgradedMessageBytes

		default:
			log.Printf("unknown message type: %s", baseMsg.Type)
			continue
		}

	}

}

func handleWrites(client *Client) {
	defer func() {
		client.conn.Close()
		delete(clients, client)
	}()

	for message := range client.send {
		// Unmarshal message to get message type
		var baseMsg types.BaseMessageWs
		if err := json.Unmarshal(message, &baseMsg); err != nil {
			log.Printf("error unmarshaling base message: %v", err)
			continue
		}

		switch baseMsg.Type {
		case "new_message":
			// Unmarshal message to types.ReceivedMessageNewMessageWS
			var receivedMessage types.ReceivedMessageNewMessageWS
			err := json.Unmarshal(message, &receivedMessage)
			if err != nil {
				log.Printf("error unmarshaling message: %v", err)
				continue
			}

			// Check if user belongs to the chat
			if client.chatIDs[receivedMessage.ChatId] {

				// Send message to the client
				err = client.conn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Printf("Error writing message: %v\n", err)
					return
				}
			}
		case "new_chat":
			// Unmarshal message to types.ReceivedMessageNewChatWS
			var receivedNewChat types.ReceivedMessageNewChatWS
			err := json.Unmarshal(message, &receivedNewChat)
			if err != nil {
				log.Printf("error unmarshaling message: %v", err)
				continue
			}

			//	Check if user is a member of the new chat
			if !utils.Contains(receivedNewChat.Members, client.userID) {
				continue
			}

			// Send message to the client
			err = client.conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Printf("Error writing message: %v\n", err)
				return
			}

		default:
			log.Printf("unknown message type: %s", baseMsg.Type)
			continue
		}

	}
}

func (s *APIServer) authenticateWebSocket(r *http.Request) (int, error) {
	sessionToken := r.Header.Get("session_token")

	if sessionToken == "" {
		return -1, fmt.Errorf("session token not found")
	}

	isValid, userId, err := s.storage.VerifySession(sessionToken)
	if err != nil || userId == -1 || !isValid {
		return -1, fmt.Errorf("failed to verify session: %w", err)
	}

	return userId, nil
}

func (s *APIServer) handleMessages() {
	for {
		msg := <-broadcast

		// Iterate through clients and send messages
		for client := range clients {
			select {
			case client.send <- msg:
			default:
				close(client.send)
				delete(clients, client)
			}
		}
	}
}
