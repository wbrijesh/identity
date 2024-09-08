package models

import (
	"time"

	"gorm.io/gorm"
)

type Admin struct {
	gorm.Model

	ID        string    `gorm:"primaryKey;default:gen_random_uuid()" json:"ID"`
	CreatedAt time.Time `json:"CreatedAt"`
	UpdatedAt time.Time `json:"UpdatedAt"`

	Email        string `gorm:"uniqueIndex;not null" json:"Email"`
	PasswordHash string `gorm:"not null" json:"PasswordHash"`
	FirstName    string `json:"FirstName"`
	LastName     string `json:"LastName"`

	Applications []Application `gorm:"foreignKey:AdminID" json:"Applications,omitempty"`
}

type Application struct {
	gorm.Model

	ID        string    `gorm:"primaryKey;default:gen_random_uuid()" json:"ID"`
	CreatedAt time.Time `json:"CreatedAt"`
	UpdatedAt time.Time `json:"UpdatedAt"`

	Name        string `gorm:"not null" json:"Name"`
	Description string `json:"Description"`

	AdminID string `gorm:"not null" json:"AdminID"`
	Admin   *Admin `gorm:"foreignKey:AdminID" json:"Admin,omitempty"`

	RefreshToken string `json:"RefreshToken"`

	Users []User `gorm:"foreignKey:ApplicationID" json:"Users,omitempty"`
}

type User struct {
	gorm.Model

	ID        string    `gorm:"primaryKey;default:gen_random_uuid()" json:"ID"`
	CreatedAt time.Time `json:"CreatedAt"`
	UpdatedAt time.Time `json:"UpdatedAt"`

	Email        string `gorm:"uniqueIndex;not null" json:"Email"`
	PasswordHash string `gorm:"not null" json:"PasswordHash"`
	FirstName    string `json:"FirstName"`
	LastName     string `json:"LastName"`

	ApplicationID string       `gorm:"not null" json:"ApplicationID"`
	Application   *Application `gorm:"foreignKey:ApplicationID" json:"Application,omitempty"`
}

type ResponseUser struct {
	ID        string    `gorm:"primaryKey;default:gen_random_uuid()" json:"ID"`
	CreatedAt time.Time `json:"CreatedAt"`
	UpdatedAt time.Time `json:"UpdatedAt"`

	Email     string `gorm:"uniqueIndex;not null" json:"Email"`
	FirstName string `json:"FirstName"`
	LastName  string `json:"LastName"`

	ApplicationID string       `gorm:"not null" json:"ApplicationID"`
	Application   *Application `gorm:"foreignKey:ApplicationID" json:"Application,omitempty"`
}

type ResponseAdmin struct {
	ID        string    `gorm:"primaryKey;default:gen_random_uuid()" json:"ID"`
	CreatedAt time.Time `json:"CreatedAt"`
	UpdatedAt time.Time `json:"UpdatedAt"`

	Email     string `gorm:"uniqueIndex;not null" json:"Email"`
	FirstName string `json:"FirstName"`
	LastName  string `json:"LastName"`

	Applications []Application `gorm:"foreignKey:AdminID" json:"Applications,omitempty"`
}

func (u *User) ToResponseUser() *ResponseUser {
	return &ResponseUser{
		ID:            u.ID,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
		Email:         u.Email,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		ApplicationID: u.ApplicationID,
		Application:   u.Application,
	}
}

func (a *Admin) ToResponseAdmin() *ResponseAdmin {
	return &ResponseAdmin{
		ID:           a.ID,
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
		Email:        a.Email,
		FirstName:    a.FirstName,
		LastName:     a.LastName,
		Applications: a.Applications,
	}
}
