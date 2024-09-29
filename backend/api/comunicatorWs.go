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
		var message types.SentMessageWS
		err := client.conn.ReadJSON(&message)

		if err != nil {
			log.Printf("error reading message: %v", err)
			break
		}

		// Check if user belongs to the chat
		if err := s.storage.IsUserInChat(client.userID, message.ChatId); err != nil {
			log.Printf("User does not belong to the chat: %v", err)
			break
		}

		upgradedMessage := &types.ReceivedMessageWS{
			ChatId:   message.ChatId,
			SenderId: client.userID,
			Content:  message.Content,
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
	}

}

func handleWrites(client *Client) {
	defer func() {
		client.conn.Close()
		delete(clients, client)
	}()

	for message := range client.send {
		// Unmarshal message to types.ReceivedMessageWS
		var receivedMessage types.ReceivedMessageWS
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
