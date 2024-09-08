package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/wbrijesh/identity/buid"
	"github.com/wbrijesh/identity/internal/models"
	"gorm.io/gorm"
)

func (s *service) CreateUser(ctx context.Context, user *models.User) (*models.ResponseUser, error) {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if a user with the same email already exists in the same application
	var existingUser models.User
	if err := tx.Where("application_id = ? AND email = ?", user.ApplicationID, user.Email).First(&existingUser).Error; err == nil {
		tx.Rollback()
		return nil, fmt.Errorf("user with email %s already exists in this application", user.Email)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return nil, fmt.Errorf("error checking for existing user: %w", err)
	}

	passwordHash, err := HashPassword(user.PasswordHash)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	user.PasswordHash = passwordHash

	// Set CreatedAt, UpdatedAt and ID
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.ID = buid.GenerateBUID()

	// Create the user
	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return user.ToResponseUser(), nil
}

func (s *service) GetUserByID(ctx context.Context, id string) (*models.ResponseUser, error) {
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found with id %s", id)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user.ToResponseUser(), nil
}

func (s *service) GetUserByEmail(ctx context.Context, applicationID, email string) (*models.ResponseUser, error) {
	var user models.User
	if err := s.db.WithContext(ctx).Where("application_id = ? AND email = ?", applicationID, email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found with email %s in application %s", email, applicationID)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user.ToResponseUser(), nil
}

func (s *service) UpdateUser(ctx context.Context, user *models.User) (*models.ResponseUser, error) {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if the user exists
	var existingUser models.User
	if err := tx.First(&existingUser, "id = ?", user.ID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found with id %s", user.ID)
		}
		return nil, fmt.Errorf("error fetching user: %w", err)
	}

	// Update the user
	user.UpdatedAt = time.Now()
	if err := tx.Model(&existingUser).Updates(user).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return user.ToResponseUser(), nil
}

func (s *service) DeleteUser(ctx context.Context, id string) error {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	result := tx.Delete(&models.User{}, "id = ?", id)
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("user not found with id %s", id)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *service) ListUsers(ctx context.Context, applicationID string, offset, limit int) ([]*models.ResponseUser, int64, error) {
	var users []*models.User
	var total int64

	query := s.db.WithContext(ctx).Model(&models.User{}).Where("application_id = ?", applicationID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	if err := query.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	responseUsers := make([]*models.ResponseUser, len(users))
	for i, user := range users {
		responseUsers[i] = user.ToResponseUser()
	}

	return responseUsers, total, nil
}

func (s *service) AuthenticateUser(ctx context.Context, applicationID, email, password string) (*models.ResponseUser, error) {
	var user models.User
	if err := s.db.WithContext(ctx).Where("application_id = ? AND email = ?", applicationID, email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found with email %s in application %s", email, applicationID)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if err := VerifyPassword(user.PasswordHash, password); err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	return user.ToResponseUser(), nil
}
