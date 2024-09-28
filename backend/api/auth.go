package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/mail"

	"github.com/carson2222/social-app/types"
	"github.com/carson2222/social-app/utils"
)

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	credentials, err := s.createCredentials(r)

	if err != nil {
		log.Println(err)
		utils.WriteJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	userId, err := s.storage.AuthUser(credentials)
	if err != nil || userId == -1 {
		log.Println(err)
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	sessionId, err := s.storage.CreateSession(userId)
	if err != nil || sessionId == "" {
		log.Println(err)
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, &types.SuccessAuthResponse{SessionId: sessionId, Status: "OK", Action: "login"})
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
		utils.WriteJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	userId, err := s.storage.CreateUser(credentials)
	if err != nil || userId == -1 {
		log.Println(err)
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := s.storage.InitProfile(userId); err != nil {
		log.Println(err)
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	sessionId, err := s.storage.CreateSession(userId)
	if err != nil || sessionId == "" {
		log.Println(err)
		utils.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, &types.SuccessAuthResponse{SessionId: sessionId, Status: "OK", Action: "register"})
	log.Println("Account created")

}

func (s *APIServer) handleLogout(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		utils.WriteJSON(w, http.StatusMethodNotAllowed, nil)
		return
	}

	sessionId, err := s.getSession(r)
	if err != nil || sessionId == "" {
		utils.WriteJSON(w, http.StatusUnauthorized, "Unauthorized"+err.Error())
		return
	}

	isValid, _, err := s.storage.VerifySession(sessionId)
	if err != nil || !isValid {
		utils.WriteJSON(w, http.StatusUnauthorized, "Unauthorized"+err.Error())
		return
	}

	err = s.storage.KillSession(sessionId)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, "Error deleting session:"+err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, "OK")
}

func (s *APIServer) createCredentials(r *http.Request) (*types.Credentials, error) {
	// TODO: Improve validation

	credentials := &types.Credentials{}

	if err := json.Unmarshal([]byte(r.FormValue("data")), &credentials); err != nil {
		return nil, err
	}

	if _, err := mail.ParseAddress(credentials.Email); err != nil {
		return nil, err
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
func (s *APIServer) authSession(r *http.Request) (int, string, error) {

	sessionId, err := s.getSession(r)
	if err != nil || sessionId == "" {
		return -1, "", fmt.Errorf("failed to get session: %w", err)
	}

	isValid, userId, err := s.storage.VerifySession(sessionId)
	if err != nil || userId == -1 || !isValid {
		return -1, "", fmt.Errorf("failed to verify session: %w", err)
	}

	return userId, sessionId, err
}

func (s *APIServer) getSession(r *http.Request) (string, error) {

	cookie, err := r.Cookie("session_token")

	if err != nil {
		return "", err
	}

	return cookie.Value, nil
}
