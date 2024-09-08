package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/wbrijesh/identity/buid"
	"github.com/wbrijesh/identity/internal/auth"
	"github.com/wbrijesh/identity/internal/models"
	"gorm.io/gorm"
)

// Assuming this is defined somewhere in your package
var secretKey = []byte("your-secret-key-here")

func (s *service) CreateApplication(ctx context.Context, app *models.Application) (*models.Application, error) {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	app.ID = buid.GenerateBUID()

	// Set CreatedAt and UpdatedAt
	now := time.Now()
	app.CreatedAt = now
	app.UpdatedAt = now

	if err := tx.Create(app).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return app, nil
}

func (s *service) GetApplicationByID(ctx context.Context, id string) (*models.Application, error) {
	var app models.Application
	if err := s.db.WithContext(ctx).First(&app, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("application with ID %s not found", id)
		}
		return nil, fmt.Errorf("error fetching application: %w", err)
	}
	return &app, nil
}

func (s *service) GetApplicationByAPIKey(ctx context.Context, apiKey string) (*models.Application, error) {
	var app models.Application
	if err := s.db.WithContext(ctx).Where("api_key = ?", apiKey).First(&app).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("application with API key %s not found", apiKey)
		}
		return nil, fmt.Errorf("error fetching application: %w", err)
	}
	return &app, nil
}

func (s *service) UpdateApplication(ctx context.Context, app *models.Application) (*models.Application, error) {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var existingApp models.Application
	if err := tx.First(&existingApp, "id = ?", app.ID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("application with ID %s not found", app.ID)
		}
		return nil, fmt.Errorf("error fetching application: %w", err)
	}

	app.UpdatedAt = time.Now()
	if err := tx.Model(&existingApp).Updates(app).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update application: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return app, nil
}

func (s *service) DeleteApplication(ctx context.Context, id string) error {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Delete(&models.Application{}, "id = ?", id).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete application: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *service) ListApplications(ctx context.Context, offset, limit int, adminID string) ([]*models.Application, int64, error) {
	var apps []*models.Application
	var total int64

	query := s.db.WithContext(ctx).Model(&models.Application{}).Where("admin_id = ?", adminID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("error counting applications: %w", err)
	}

	if limit == 0 {
		offset = 0
		limit = 20
	}

	if err := query.Offset(offset).Limit(limit).Find(&apps).Error; err != nil {
		return nil, 0, fmt.Errorf("error fetching applications: %w", err)
	}

	return apps, total, nil
}

func (s *service) GenerateRefreshToken(ctx context.Context, id string) (string, error) {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Fetch the application
	var app models.Application
	if err := tx.First(&app, "id = ?", id).Error; err != nil {
		tx.Rollback()
		return "", fmt.Errorf("failed to find application: %w", err)
	}

	if app.RefreshToken != "" {
		tx.Rollback()
		return "", fmt.Errorf("application already has a refresh token")
	}

	// Generate a new refresh token
	refreshToken, err := auth.GenerateRefreshToken(app.ID)
	if err != nil {
		tx.Rollback()
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Update the application with the new refresh token
	app.RefreshToken = refreshToken
	app.UpdatedAt = time.Now()

	if err := tx.Save(&app).Error; err != nil {
		tx.Rollback()
		return "", fmt.Errorf("failed to update application with refresh token: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return refreshToken, nil
}

func (s *service) DeleteRefreshToken(ctx context.Context, id string) error {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Fetch the application
	var app models.Application
	if err := tx.First(&app, "id = ?", id).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to find application: %w", err)
	}

	// Clear the refresh token
	app.RefreshToken = ""
	app.UpdatedAt = time.Now()

	if err := tx.Save(&app).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update application and remove refresh token: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
