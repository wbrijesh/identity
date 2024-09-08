package server

import (
	"encoding/json"
	"net/http"

	"github.com/wbrijesh/identity/internal/auth"
	"github.com/wbrijesh/identity/internal/models"
	"github.com/wbrijesh/identity/utils"
)

func (s *Server) CreateAdminHandler(w http.ResponseWriter, r *http.Request) {
	var admin models.Admin
	if err := json.NewDecoder(r.Body).Decode(&admin); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := utils.CheckNeceassaryFieldsExist(admin, []string{"Email", "PasswordHash", "FirstName", "LastName"}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdAdmin, err := s.db.CreateAdmin(r.Context(), &admin)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := auth.GenerateAdminJWT(createdAdmin)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"admin": createdAdmin,
		"token": token,
	}

	json.NewEncoder(w).Encode(response)
}

func (s *Server) LoginAdminHandler(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := utils.CheckNeceassaryFieldsExist(creds, []string{"Email", "Password"}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	admin, err := s.db.AuthenticateAdmin(r.Context(), creds.Email, creds.Password)
	if err != nil {
		http.Error(w, "Invalid credentials"+err.Error(), http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateAdminJWT(admin)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"admin": admin,
		"token": token,
	}

	json.NewEncoder(w).Encode(response)
}
