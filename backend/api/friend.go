package api

import (
	"net/http"
	"strconv"

	"github.com/carson2222/social-app/utils"
	"github.com/gorilla/mux"
)

func (s *APIServer) handleAddFriend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userId, _, err := s.authSession(r)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, "Unauthorized:"+err.Error())
		return
	}

	vars := mux.Vars(r)
	friendId, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid id")
		return
	}
	if friendId == userId {
		utils.WriteJSON(w, http.StatusBadRequest, "Cannot add self")
		return
	}

	action := vars["action"]

	switch action {
	case "accept":
		// Check if users are already friends
		areFriends, err := s.storage.AreFriends(userId, friendId)
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, "Failed to check if users are friend:"+err.Error())
			return
		}

		if areFriends {
			utils.WriteJSON(w, http.StatusBadRequest, "Users are already friends")
			return
		}

		// Check if user is already requested
		isRequested, err := s.storage.IsRequestedFriend(friendId, userId)
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, "Failed to check if user is requested friend with the friend:"+err.Error())
			return
		}

		if !isRequested {
			utils.WriteJSON(w, http.StatusBadRequest, "No request to be accepted")
			return
		}

		err = s.storage.AcceptFriendRequest(userId, friendId)
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, "Failed to accept friend request:"+err.Error())
			return
		}
		utils.WriteJSON(w, http.StatusOK, "OK")
		return
	case "reject":
		// Check if users are already friends
		areFriends, err := s.storage.AreFriends(userId, friendId)
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, "Failed to check if users are friend:"+err.Error())
			return
		}

		if areFriends {
			utils.WriteJSON(w, http.StatusBadRequest, "Users are already friends")
			return
		}

		// Check if friend request is already sent
		isRequested, err := s.storage.IsRequestedFriend(friendId, userId)
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, "Failed to check if user is requested friend with the friend:"+err.Error())
			return
		}

		if !isRequested {
			utils.WriteJSON(w, http.StatusBadRequest, "No request to be rejected")
			return
		}

		err = s.storage.RejectFriendRequest(userId, friendId)
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, "Failed to reject friend request:"+err.Error())
			return
		}
		utils.WriteJSON(w, http.StatusOK, "OK")
		return
	case "add":
		// Check if users are already friends
		areFriends, err := s.storage.AreFriends(userId, friendId)
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, "Failed to check if users are friend:"+err.Error())
			return
		}

		if areFriends {
			utils.WriteJSON(w, http.StatusBadRequest, "Users are already friends")
			return
		}

		// Check if friend request is already sent
		isRequested1, err1 := s.storage.IsRequestedFriend(userId, friendId)
		isRequested2, err2 := s.storage.IsRequestedFriend(friendId, userId)

		if err1 != nil || err2 != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, "Failed to check if user is requested friend with the friend:"+err.Error())
			return
		}

		if isRequested1 || isRequested2 {
			utils.WriteJSON(w, http.StatusBadRequest, "User already requested to be friend with the friend")
			return
		}

		err = s.storage.AddFriend(userId, friendId)
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, "Failed to add friend:"+err.Error())
			return
		}
		utils.WriteJSON(w, http.StatusOK, "OK")
		return
	case "remove":
		// Check if users are already friends
		areFriends, err := s.storage.AreFriends(userId, friendId)
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, "Failed to check if users are friend:"+err.Error())
			return
		}

		if !areFriends {
			utils.WriteJSON(w, http.StatusBadRequest, "Users are not friends")
			return
		}

		err = s.storage.RemoveFriend(userId, friendId)
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, "Failed to remove friend:"+err.Error())
			return
		}
		utils.WriteJSON(w, http.StatusOK, "OK")
		return
	default:
		utils.WriteJSON(w, http.StatusBadRequest, "Invalid action")
		return
	}

}
