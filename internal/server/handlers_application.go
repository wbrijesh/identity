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

func (s *Server) CreateApplicationHandler(w http.ResponseWriter, r *http.Request) {
	adminID, ok := r.Context().Value("adminID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusInternalServerError)
		return
	}

	var app models.Application
	if err := json.NewDecoder(r.Body).Decode(&app); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	app.AdminID = adminID

	if err := utils.CheckNeceassaryFieldsExist(app, []string{"Name", "Description"}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdApp, err := s.db.CreateApplication(r.Context(), &app)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(createdApp)
}

func (s *Server) ListApplicationsHandler(w http.ResponseWriter, r *http.Request) {
	adminID, ok := r.Context().Value("adminID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusInternalServerError)
		return
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	apps, total, err := s.db.ListApplications(r.Context(), offset, limit, adminID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"applications": apps,
		"total":        total,
	})
}

func (s *Server) GenerateRefreshTokenForApplicationHandler(w http.ResponseWriter, r *http.Request) {
	applicationID := chi.URLParam(r, "applicationID")
	if applicationID == "" {
		http.Error(w, "Application ID is required", http.StatusBadRequest)
		return
	}

	// Check if request is coming from the application owner
	adminID, ok := r.Context().Value("adminID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusInternalServerError)
		return
	}
	application, err := s.db.GetApplicationByID(r.Context(), applicationID)
	if err != nil {
		http.Error(w, "Application not found", http.StatusNotFound)
		return
	}
	if application.AdminID != adminID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Generate refresh token
	refreshToken, err := s.db.GenerateRefreshToken(r.Context(), applicationID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// DELETE TemporaryAccessToken before pushing first stable version
	accessToken, err := auth.GenerateAccessTokenFromRefreshToken(refreshToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := struct {
		RefreshToken         string `json:"refresh_token"`
		TemporaryAccessToken string `json:"temporary_access_token"`
	}{
		RefreshToken:         refreshToken,
		TemporaryAccessToken: accessToken,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) UpdateRefreshTokenForApplicationHandler(w http.ResponseWriter, r *http.Request) {
	applicationID := chi.URLParam(r, "applicationID")
	if applicationID == "" {
		http.Error(w, "Application ID is required", http.StatusBadRequest)
		return
	}

	// Check if request is coming from the application owner
	adminID, ok := r.Context().Value("adminID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusInternalServerError)
		return
	}
	application, err := s.db.GetApplicationByID(r.Context(), applicationID)
	if err != nil {
		http.Error(w, "Application not found", http.StatusNotFound)
		return
	}
	if application.AdminID != adminID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Delete existing refresh token
	err = s.db.DeleteRefreshToken(r.Context(), applicationID)
	if err != nil {
		http.Error(w, "Failed to delete existing refresh token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate new refresh token
	newRefreshToken, err := s.db.GenerateRefreshToken(r.Context(), applicationID)
	if err != nil {
		http.Error(w, "Failed to generate new refresh token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// DELETE TemporaryAccessToken before pushing first stable version
	accessToken, err := auth.GenerateAccessTokenFromRefreshToken(newRefreshToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := struct {
		RefreshToken         string `json:"refresh_token"`
		TemporaryAccessToken string `json:"temporary_access_token"`
	}{
		RefreshToken:         newRefreshToken,
		TemporaryAccessToken: accessToken,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
