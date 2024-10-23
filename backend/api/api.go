package api

import (
	"log"
	"net/http"

	"github.com/carson2222/social-app/storage"
	"github.com/carson2222/social-app/ws"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type APIServer struct {
	listenAddr string
	storage    *storage.PostgresStore
	wsServer   *ws.WebSocketServer
}

func NewAPIServer(listenAddr string, storage *storage.PostgresStore, wsServer *ws.WebSocketServer) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		storage:    storage,
		wsServer:   wsServer,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()
	// router.HandleFunc("/communicator/ws", s.serveCommunicatorWs)
	// go s.handleMessages()
	router.HandleFunc("/ws", s.wsServer.ServerWebSocket)

	router.HandleFunc("/auth/login", s.handleLogin)
	router.HandleFunc("/auth/register", s.handleRegister)
	router.HandleFunc("/auth/logout", s.handleLogout)

	router.HandleFunc("/profile", s.handleProfile).Methods("POST")
	router.HandleFunc("/profile/{id}", s.handleProfile).Methods("GET")

	// router.HandleFunc("/friends/{action}/{id}", s.handleAddFriend).Methods("POST")

	// Serve static files
	router.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads/"))))

	// CORS settings
	allowCredentials := handlers.AllowCredentials()

	log.Println("Listening on port " + s.listenAddr)
	http.ListenAndServe(s.listenAddr, handlers.CORS(allowCredentials)(router))
}
