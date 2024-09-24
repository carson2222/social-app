package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type APIServer struct {
	listenAddr string
	storage    Storage
}

func NewAPIServer(listenAddr string, storage Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		storage:    storage,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()
	router.HandleFunc("/login", s.handleLogin)
	router.HandleFunc("/register", s.handleRegister)

	log.Println("Listening on port " + s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	log.Println("Login request")
}
func (s *APIServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	credentials, err := s.createCredentials(r)

	if err != nil {
		log.Println(err)
		WriteJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := s.storage.CreateUser(credentials); err != nil {
		log.Println(err)
		WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Println("Register request")
}

func (s *APIServer) createCredentials(r *http.Request) (*Credentials, error) {
	// TODO: Improve validation

	credentials := &Credentials{}

	if err := json.NewDecoder(r.Body).Decode(credentials); err != nil {
		return nil, err
	}

	if credentials.Username == "" {
		return nil, errors.New("username is empty")
	}

	if len(credentials.Username) < 3 {
		return nil, errors.New("username must be at least 3 characters long")
	}

	if len(credentials.Username) > 20 {
		return nil, errors.New("username must be at most 20 characters long")
	}

	if credentials.Password == "" {
		return nil, errors.New("password is empty")
	}

	if len(credentials.Password) < 8 {
		return nil, errors.New("password must be at least 8 characters long")
	}

	if len(credentials.Password) > 50 {
		return nil, errors.New("password must be at most 50 characters long")
	}

	return credentials, nil
}
