package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/mail"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/handlers"
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
	// router.HandleFunc("/")
	router.HandleFunc("/auth/login", s.handleLogin)
	router.HandleFunc("/auth/register", s.handleRegister)
	router.HandleFunc("/auth/logout", s.handleLogout)
	router.HandleFunc("/profile/{id}", s.handleProfile)
	router.HandleFunc("/friends/{action}/{id}", s.handleAddFriend)

	// Serve static files
	router.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads/"))))

	// CORS settings
	allowCredentials := handlers.AllowCredentials()

	log.Println("Listening on port " + s.listenAddr)
	http.ListenAndServe(s.listenAddr, handlers.CORS(allowCredentials)(router))
}

func (s *APIServer) handleAddFriend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userId, _, err := s.authSession(r)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, "Unauthorized:"+err.Error())
		return
	}

	vars := mux.Vars(r)
	friendId, err := strconv.Atoi(vars["id"])
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid id")
		return
	}
	if friendId == userId {
		WriteJSON(w, http.StatusBadRequest, "Cannot add self")
		return
	}

	action := vars["action"]

	switch action {
	case "accept":
		// Check if users are already friends
		areFriends, err := s.storage.areFriends(userId, friendId)
		if err != nil {
			WriteJSON(w, http.StatusInternalServerError, "Failed to check if users are friend:"+err.Error())
			return
		}

		if areFriends {
			WriteJSON(w, http.StatusBadRequest, "Users are already friends")
			return
		}

		// Check if user is already requested
		isRequested, err := s.storage.isRequestedFriend(friendId, userId)
		if err != nil {
			WriteJSON(w, http.StatusInternalServerError, "Failed to check if user is requested friend with the friend:"+err.Error())
			return
		}

		if !isRequested {
			WriteJSON(w, http.StatusBadRequest, "No request to be accepted")
			return
		}

		err = s.storage.acceptFriendRequest(userId, friendId)
		if err != nil {
			WriteJSON(w, http.StatusInternalServerError, "Failed to accept friend request:"+err.Error())
			return
		}
		WriteJSON(w, http.StatusOK, "OK")
		return
	case "reject":
		// Check if users are already friends
		areFriends, err := s.storage.areFriends(userId, friendId)
		if err != nil {
			WriteJSON(w, http.StatusInternalServerError, "Failed to check if users are friend:"+err.Error())
			return
		}

		if areFriends {
			WriteJSON(w, http.StatusBadRequest, "Users are already friends")
			return
		}

		// Check if friend request is already sent
		isRequested, err := s.storage.isRequestedFriend(friendId, userId)
		if err != nil {
			WriteJSON(w, http.StatusInternalServerError, "Failed to check if user is requested friend with the friend:"+err.Error())
			return
		}

		if !isRequested {
			WriteJSON(w, http.StatusBadRequest, "No request to be rejected")
			return
		}

		err = s.storage.rejectFriendRequest(userId, friendId)
		if err != nil {
			WriteJSON(w, http.StatusInternalServerError, "Failed to reject friend request:"+err.Error())
			return
		}
		WriteJSON(w, http.StatusOK, "OK")
		return
	case "add":
		// Check if users are already friends
		areFriends, err := s.storage.areFriends(userId, friendId)
		if err != nil {
			WriteJSON(w, http.StatusInternalServerError, "Failed to check if users are friend:"+err.Error())
			return
		}

		if areFriends {
			WriteJSON(w, http.StatusBadRequest, "Users are already friends")
			return
		}

		// Check if friend request is already sent
		isRequested1, err1 := s.storage.isRequestedFriend(userId, friendId)
		isRequested2, err2 := s.storage.isRequestedFriend(friendId, userId)

		if err1 != nil || err2 != nil {
			WriteJSON(w, http.StatusInternalServerError, "Failed to check if user is requested friend with the friend:"+err.Error())
			return
		}

		if isRequested1 || isRequested2 {
			WriteJSON(w, http.StatusBadRequest, "User already requested to be friend with the friend")
			return
		}

		err = s.storage.addFriend(userId, friendId)
		if err != nil {
			WriteJSON(w, http.StatusInternalServerError, "Failed to add friend:"+err.Error())
			return
		}
		WriteJSON(w, http.StatusOK, "OK")
		return
	case "remove":
		// Check if users are already friends
		areFriends, err := s.storage.areFriends(userId, friendId)
		if err != nil {
			WriteJSON(w, http.StatusInternalServerError, "Failed to check if users are friend:"+err.Error())
			return
		}

		if !areFriends {
			WriteJSON(w, http.StatusBadRequest, "Users are not friends")
			return
		}

		err = s.storage.removeFriend(userId, friendId)
		if err != nil {
			WriteJSON(w, http.StatusInternalServerError, "Failed to remove friend:"+err.Error())
			return
		}
		WriteJSON(w, http.StatusOK, "OK")
		return
	default:
		WriteJSON(w, http.StatusBadRequest, "Invalid action")
		return
	}

}
func (s *APIServer) handleProfile(w http.ResponseWriter, r *http.Request) {

	// Update the profile
	if r.Method == http.MethodPost {
		// Verify session
		userId, _, err := s.authSession(r)
		if err != nil {
			WriteJSON(w, http.StatusInternalServerError, "Unauthorized:"+err.Error())
			return
		}

		err = s.updateProfile(r, userId)
		if err != nil {
			WriteJSON(w, http.StatusInternalServerError, "Failed to update profile:"+err.Error())
			return
		}

		WriteJSON(w, http.StatusOK, "OK")
		return

	}

	// Get the profile
	if r.Method == http.MethodGet {
		// Get seek profile id
		id, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			WriteJSON(w, http.StatusBadRequest, "Invalid id")
			return
		}

		// Verify session
		_, _, err = s.authSession(r)
		if err != nil {
			WriteJSON(w, http.StatusInternalServerError, "Unauthorized:"+err.Error())
			return
		}

		// Get profile
		profile, err := s.storage.getProfileByID(id)
		if err != nil {
			WriteJSON(w, http.StatusInternalServerError, "Failed to get profile:"+err.Error())
			return
		}

		WriteJSON(w, http.StatusOK, profile)
		return
	}

	WriteJSON(w, http.StatusMethodNotAllowed, "Method not allowed")
}
func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) {
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

	userId, err := s.storage.authUser(credentials)
	if err != nil || userId == -1 {
		log.Println(err)
		WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	sessionId, err := s.storage.createSession(userId)
	if err != nil || sessionId == "" {
		log.Println(err)
		WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, &SuccessAuthResponse{SessionId: sessionId, Status: "OK", Action: "login"})
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

	userId, err := s.storage.createUser(credentials)
	if err != nil || userId == -1 {
		log.Println(err)
		WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := s.storage.initProfile(userId); err != nil {
		log.Println(err)
		WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	sessionId, err := s.storage.createSession(userId)
	if err != nil || sessionId == "" {
		log.Println(err)
		WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, &SuccessAuthResponse{SessionId: sessionId, Status: "OK", Action: "register"})
	log.Println("Account created")

}

func (s *APIServer) handleLogout(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		WriteJSON(w, http.StatusMethodNotAllowed, nil)
		return
	}

	sessionId, err := s.getSession(r)
	if err != nil || sessionId == "" {
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized"+err.Error())
		return
	}

	isValid, _, err := s.storage.verifySession(sessionId)
	if err != nil || !isValid {
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized"+err.Error())
		return
	}

	err = s.storage.killSession(sessionId)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, "Error deleting session:"+err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, "OK")
}

func (s *APIServer) createCredentials(r *http.Request) (*Credentials, error) {
	// TODO: Improve validation

	credentials := &Credentials{}

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

func (s *APIServer) getSession(r *http.Request) (string, error) {

	cookie, err := r.Cookie("session_token")

	if err != nil {
		return "", err
	}

	return cookie.Value, nil
}

func (s *APIServer) uploadProfilePicture(r *http.Request) (string, error) {
	// Parse file
	r.ParseMultipartForm(10 << 20) // Ograniczenie do 10MB
	file, _, err := r.FormFile("profile_picture")
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Validate if it's a PNG or JPG file
	imgType, err := validateFileType(file)
	if err != nil {
		return "", err
	}

	if imgType != "png" && imgType != "jpg" {
		return "", errors.New("Only PNG and JPG files are allowed")
	}

	// Save image to uploads folder
	fileName, err := s.generateFileName()
	if err != nil {
		return "", err
	}

	filePath := fmt.Sprintf("./uploads/pfp/%s", fileName+"."+imgType)
	f, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(f, file)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func (s *APIServer) generateFileName() (string, error) {
	timestamp := time.Now().UnixNano()

	// Generate random bytes
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	input := fmt.Sprintf("%d%s", timestamp, hex.EncodeToString(randomBytes))

	// Encrypt name with SHA256
	hash := sha256.New()
	hash.Write([]byte(input))
	hashedName := hex.EncodeToString(hash.Sum(nil))

	return hashedName, nil
}

func validateFileType(file multipart.File) (string, error) {
	// Read a small portion of the file to detect its MIME type
	buffer := make([]byte, 512) // 512 bytes are enough to sniff the content type
	_, err := file.Read(buffer)
	if err != nil {
		return "", err
	}

	// Reset the file read pointer after reading
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	// Detect the content type (MIME type)
	mimeType := http.DetectContentType(buffer)

	// Check if it's a valid PNG or JPG/JPEG
	if mimeType == "image/png" {
		return "png", nil
	}
	if mimeType == "image/jpeg" {
		return "jpg", nil
	}

	return "", nil
}

func (s *APIServer) authSession(r *http.Request) (int, string, error) {

	sessionId, err := s.getSession(r)
	if err != nil || sessionId == "" {
		return -1, "", fmt.Errorf("failed to get session: %w", err)
	}

	isValid, userId, err := s.storage.verifySession(sessionId)
	if err != nil || userId == -1 || !isValid {
		return -1, "", fmt.Errorf("failed to verify session: %w", err)
	}

	return userId, sessionId, err
}

func (s *APIServer) updateProfile(r *http.Request, userId int) error {
	// Load JSON data
	data := &ProfileRequest{}

	var err error
	if err = json.Unmarshal([]byte(r.FormValue("data")), &data); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Load pfp if added
	pfpSrc := ""
	if data.Pfp == true {
		pfpSrc, err = s.uploadProfilePicture(r)
		if err != nil {
			log.Println(err)
			return fmt.Errorf("failed to upload profile picture: %w", err)
		}
	}

	// Update profile
	err = s.storage.updateProfile(userId, data.Name, data.Surname, data.Bio, pfpSrc)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("failed to update profile: %w", err)
	}

	return nil
}
