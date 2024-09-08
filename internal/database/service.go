package database

import (
	"context"

	"github.com/wbrijesh/identity/internal/models"
)

type Service interface {
	// Database specific methods
	Health() map[string]string
	Close() error
	RunMigrations() error

	// Admin CRUD operations
	CreateAdmin(ctx context.Context, admin *models.Admin) (*models.ResponseAdmin, error)
	GetAdminByID(ctx context.Context, id string) (*models.ResponseAdmin, error)
	GetAdminByEmail(ctx context.Context, email string) (*models.ResponseAdmin, error)
	UpdateAdmin(ctx context.Context, admin *models.Admin) (*models.ResponseAdmin, error)
	DeleteAdmin(ctx context.Context, id string) error
	ListAdmins(ctx context.Context, offset, limit int) ([]*models.ResponseAdmin, int64, error)

	// Application CRUD operations
	CreateApplication(ctx context.Context, app *models.Application) (*models.Application, error)
	GetApplicationByID(ctx context.Context, id string) (*models.Application, error)
	GetApplicationByAPIKey(ctx context.Context, apiKey string) (*models.Application, error)
	UpdateApplication(ctx context.Context, app *models.Application) (*models.Application, error)
	DeleteApplication(ctx context.Context, id string) error
	ListApplications(ctx context.Context, offset int, limit int, adminID string) ([]*models.Application, int64, error)
	GenerateRefreshToken(ctx context.Context, id string) (string, error)
	DeleteRefreshToken(ctx context.Context, id string) error

	// User CRUD operations
	CreateUser(ctx context.Context, user *models.User) (*models.ResponseUser, error)
	GetUserByID(ctx context.Context, id string) (*models.ResponseUser, error)
	GetUserByEmail(ctx context.Context, applicationID, email string) (*models.ResponseUser, error)
	UpdateUser(ctx context.Context, user *models.User) (*models.ResponseUser, error)
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, applicationID string, offset, limit int) ([]*models.ResponseUser, int64, error)

	// Additional utility methods
	AuthenticateAdmin(ctx context.Context, email, password string) (*models.ResponseAdmin, error)
	AuthenticateUser(ctx context.Context, applicationID, email, password string) (*models.ResponseUser, error)
}
