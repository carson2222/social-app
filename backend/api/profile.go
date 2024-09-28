package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/carson2222/social-app/types"
	"github.com/carson2222/social-app/utils"
	"github.com/gorilla/mux"
)

func (s *APIServer) handleProfile(w http.ResponseWriter, r *http.Request) {

	// Update the profile
	if r.Method == http.MethodPost {
		// Verify session
		userId, _, err := s.authSession(r)
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, "Unauthorized:"+err.Error())
			return
		}

		err = s.updateProfile(r, userId)
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, "Failed to update profile:"+err.Error())
			return
		}

		utils.WriteJSON(w, http.StatusOK, "OK")
		return

	}

	// Get the profile
	if r.Method == http.MethodGet {
		// Get seek profile id
		id, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			utils.WriteJSON(w, http.StatusBadRequest, "Invalid id")
			return
		}

		// Verify session
		_, _, err = s.authSession(r)
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, "Unauthorized:"+err.Error())
			return
		}

		// Get profile
		profile, err := s.storage.GetProfileByID(id)
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, "Failed to get profile:"+err.Error())
			return
		}

		utils.WriteJSON(w, http.StatusOK, profile)
		return
	}

	utils.WriteJSON(w, http.StatusMethodNotAllowed, "Method not allowed")
}

func (s *APIServer) updateProfile(r *http.Request, userId int) error {
	// Load JSON data
	data := &types.ProfileRequest{}

	var err error
	if err = json.Unmarshal([]byte(r.FormValue("data")), &data); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Load pfp if added
	pfpSrc := ""
	if data.Pfp == true {
		pfpSrc, err = utils.UploadProfilePicture(r)
		if err != nil {
			log.Println(err)
			return fmt.Errorf("failed to upload profile picture: %w", err)
		}
	}

	// Update profile
	err = s.storage.UpdateProfile(userId, data.Name, data.Surname, data.Bio, pfpSrc)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("failed to update profile: %w", err)
	}

	return nil
}
