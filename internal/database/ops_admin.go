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

func (s *service) CreateAdmin(ctx context.Context, admin *models.Admin) (*models.ResponseAdmin, error) {
	// Start a new database transaction
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Defer a function to handle transaction commit or rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if an admin with the same email already exists
	var existingAdmin models.Admin
	if err := tx.Where("email = ?", admin.Email).First(&existingAdmin).Error; err == nil {
		tx.Rollback()
		return nil, fmt.Errorf("admin with email %s already exists", admin.Email)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return nil, fmt.Errorf("error checking for existing admin: %w", err)
	}

	passwordHash, err := HashPassword(admin.PasswordHash)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	admin.PasswordHash = passwordHash

	// Set CreatedAt and UpdatedAt and ID
	now := time.Now()
	admin.CreatedAt = now
	admin.UpdatedAt = now
	admin.ID = buid.GenerateBUID()

	// Create the admin
	if err := tx.Create(admin).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create admin: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return admin.ToResponseAdmin(), nil
}

func (s *service) GetAdminByID(ctx context.Context, id string) (*models.ResponseAdmin, error) {
	var admin models.Admin
	if err := s.db.WithContext(ctx).First(&admin, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("admin with ID %s not found", id)
		}
		return nil, fmt.Errorf("error fetching admin: %w", err)
	}
	return admin.ToResponseAdmin(), nil
}

func (s *service) GetAdminByEmail(ctx context.Context, email string) (*models.ResponseAdmin, error) {
	var admin models.Admin
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("admin with email %s not found", email)
		}
		return nil, fmt.Errorf("error fetching admin: %w", err)
	}
	return admin.ToResponseAdmin(), nil
}

func (s *service) UpdateAdmin(ctx context.Context, admin *models.Admin) (*models.ResponseAdmin, error) {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if the admin exists
	var existingAdmin models.Admin
	if err := tx.First(&existingAdmin, "id = ?", admin.ID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("admin with ID %s not found", admin.ID)
		}
		return nil, fmt.Errorf("error fetching admin: %w", err)
	}

	// Update the admin
	admin.UpdatedAt = time.Now()
	if err := tx.Model(&existingAdmin).Updates(admin).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update admin: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return admin.ToResponseAdmin(), nil
}

func (s *service) DeleteAdmin(ctx context.Context, id string) error {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if the admin exists
	var admin models.Admin
	if err := tx.First(&admin, "id = ?", id).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("admin with ID %s not found", id)
		}
		return fmt.Errorf("error fetching admin: %w", err)
	}

	// Delete the admin
	if err := tx.Delete(&admin).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete admin: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *service) ListAdmins(ctx context.Context, offset, limit int) ([]*models.ResponseAdmin, int64, error) {
	var admins []*models.Admin
	var total int64

	// Count total number of admins
	if err := s.db.WithContext(ctx).Model(&models.Admin{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("error counting admins: %w", err)
	}

	// Fetch admins with pagination
	if err := s.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&admins).Error; err != nil {
		return nil, 0, fmt.Errorf("error fetching admins: %w", err)
	}

	responseAdmins := make([]*models.ResponseAdmin, 0, len(admins))
	for _, admin := range admins {
		responseAdmins = append(responseAdmins, admin.ToResponseAdmin())
	}

	return responseAdmins, total, nil
}

// AuthenticateAdmin checks if the provided email and password match an admin in the database
func (s *service) AuthenticateAdmin(ctx context.Context, email, password string) (*models.ResponseAdmin, error) {
	var admin models.Admin
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("admin with email %s not found", email)
		}
		return nil, fmt.Errorf("error fetching admin: %w", err)
	}

	if err := VerifyPassword(admin.PasswordHash, password); err != nil {
		return nil, err
	}

	return admin.ToResponseAdmin(), nil
}
