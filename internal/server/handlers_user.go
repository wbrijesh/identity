package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/wbrijesh/identity/internal/auth"
	"github.com/wbrijesh/identity/internal/models"
	"github.com/wbrijesh/identity/utils"
)

func (s *Server) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if request is coming from the application owner
	adminID, ok := r.Context().Value("adminID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusInternalServerError)
		return
	}
	adminApps, _, err := s.db.ListApplications(r.Context(), 0, 200, adminID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	adminOwnedApp := false
	for _, app := range adminApps {
		if app.ID == user.ApplicationID {
			adminOwnedApp = true
			break
		}
	}
	if !adminOwnedApp {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := utils.CheckNeceassaryFieldsExist(user, []string{"Email", "PasswordHash", "FirstName", "LastName", "ApplicationID"}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdUser, err := s.db.CreateUser(r.Context(), &user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := auth.GenerateUserJWT(createdUser)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"user":  createdUser,
		"token": token,
	}

	json.NewEncoder(w).Encode(response)
}

func (s *Server) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		ApplicationID string `json:"application_id"`
		Email         string `json:"email"`
		Password      string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if request is coming from the application owner
	adminID, ok := r.Context().Value("adminID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusInternalServerError)
		return
	}
	adminApps, _, err := s.db.ListApplications(r.Context(), 0, 200, adminID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	adminOwnedApp := false
	for _, app := range adminApps {
		if app.ID == creds.ApplicationID {
			adminOwnedApp = true
			break
		}
	}
	if !adminOwnedApp {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := utils.CheckNeceassaryFieldsExist(creds, []string{"ApplicationID", "Email", "Password"}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := s.db.AuthenticateUser(r.Context(), creds.ApplicationID, creds.Email, creds.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateUserJWT(user)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"user":  user,
		"token": token,
	}

	json.NewEncoder(w).Encode(response)
}

func (s *Server) ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	applicationID := chi.URLParam(r, "applicationID")
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	// Check if request is coming from the application owner
	adminID, ok := r.Context().Value("adminID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusInternalServerError)
		return
	}
	adminApps, _, err := s.db.ListApplications(r.Context(), 0, 200, adminID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	adminOwnedApp := false
	for _, app := range adminApps {
		if app.ID == applicationID {
			adminOwnedApp = true
			break
		}
	}
	if !adminOwnedApp {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	users, total, err := s.db.ListUsers(r.Context(), applicationID, offset, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"users": users,
		"total": total,
	})
}
