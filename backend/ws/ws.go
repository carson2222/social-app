package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/carson2222/social-app/storage"
	"github.com/carson2222/social-app/types"
	"github.com/carson2222/social-app/utils"
	"github.com/gorilla/websocket"
)

type WebSocketServer struct {
	clients   map[*types.Client]bool                 // Registered clients
	broadcast chan []byte                            // Broadcast channel for all messages
	handlers  map[string]func(*types.Client, []byte) // Event handlers
	storage   *storage.PostgresStore
}

func NewWebSocketServer(storage *storage.PostgresStore) *WebSocketServer {
	wsServer := &WebSocketServer{
		clients:   make(map[*types.Client]bool),
		broadcast: make(chan []byte),
		handlers:  make(map[string]func(*types.Client, []byte)),
		storage:   storage,
	}

	wsServer.registerHandlers()

	go wsServer.BroadcastMessages()

	return wsServer
}

func (ws *WebSocketServer) ServerWebSocket(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Host)

	// Check if the user is authenticated
	userId, err := ws.authenticateWebSocket(r)
	log.Print(userId)
	if err != nil || userId == -1 {
		log.Println("failed to authenticate websocket", err)
		utils.WriteJSON(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get the user's chat IDs
	chatIDs, err := ws.storage.GetUserChats(userId)
	if err != nil {
		log.Println(err)
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	// upgrade this connection to a WebSocket connection
	upgrader := ws.createUpgrader()
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	client := &types.Client{
		Conn:    conn,
		UserID:  userId,
		ChatIDs: chatIDs,
		Send:    make(chan []byte),
	}

	log.Print(client)

	var clientsMu sync.Mutex
	clientsMu.Lock()
	ws.clients[client] = true
	clientsMu.Unlock()

	go ws.handleReads(client)
	go ws.handleWrites(client)
}

func (ws *WebSocketServer) registerHandlers() {
	ws.handlers = make(map[string]func(*types.Client, []byte))
	ws.handlers["sendMessage"] = ws.handleMessage
	ws.handlers["friend"] = ws.handleFriendRequest
	// You can add more handlers as needed.
}

// func (ws *WebSocketServer) registerVerifiers() {
// 	ws.handlers = make(map[string]func(*types.Client, []byte))
// 	ws.handlers["sendMessage"] = ws.handleMessage
// 	ws.handlers["friend"] = ws.handleFriendRequest
// 	// You can add more handlers as needed.
// }

func (ws *WebSocketServer) handleReads(client *types.Client) {
	defer func() {
		delete(ws.clients, client)
		client.Conn.Close()
	}()

	for {
		_, messageBytes, err := client.Conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		var baseIncoming types.IncomingBase
		if err := json.Unmarshal(messageBytes, &baseIncoming); err != nil {
			log.Printf("error unmarshaling base message: %v", err)
			continue
		}

		// Check if we have a handler for the incoming message type
		if handler, ok := ws.handlers[baseIncoming.Type]; ok {
			handler(client, messageBytes)
		} else {
			log.Printf("No handler for message type: %s", baseIncoming.Type)
		}

	}
}

func (ws *WebSocketServer) handleWrites(client *types.Client) {
	defer func() {
		delete(ws.clients, client)
		client.Conn.Close()
	}()

	for message := range client.Send {
		err := client.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("Error writing message to client: %v", err)
			break
		}
	}
}

func (ws *WebSocketServer) authenticateWebSocket(r *http.Request) (int, error) {
	sessionToken := r.Header.Get("session_token")

	log.Print(sessionToken)
	if sessionToken == "" {
		return -1, fmt.Errorf("session token not found")
	}

	isValid, userId, err := ws.storage.VerifySession(sessionToken)
	if err != nil || userId == -1 || !isValid {
		return -1, fmt.Errorf("failed to verify session: %w", err)
	}

	return userId, nil
}

func (ws *WebSocketServer) createUpgrader() websocket.Upgrader {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,

		CheckOrigin: func(r *http.Request) bool {
			return r.Host == "localhost:3000"
		},
	}

	return upgrader
}

func (ws *WebSocketServer) BroadcastMessages() {
	for {
		msg := <-ws.broadcast

		// Unmarshal the message to get metadata like recipient ID
		var outgoingMsg types.OutgoingBase
		if err := json.Unmarshal(msg, &outgoingMsg); err != nil {
			log.Printf("Error unmarshaling broadcast message: %v", err)
			continue
		}

		// Iterate through clients and send messages
		for client := range ws.clients {
			isUserTheReceiver := false
			if outgoingMsg.VerifyType == "chatID" && client.ChatIDs[outgoingMsg.VerifyID] {
				isUserTheReceiver = true
			}

			if outgoingMsg.VerifyType == "userID" && client.UserID == outgoingMsg.VerifyID {
				isUserTheReceiver = true
			}

			if isUserTheReceiver {
				select {
				case client.Send <- []byte(outgoingMsg.Content):
				default:
					close(client.Send)
					delete(ws.clients, client)
				}
			}
		}
	}
}
