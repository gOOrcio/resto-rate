package services

import (
	"go-app/src/services/models"
	"gorm.io/gorm"
	"time"
)

type UserService struct {
	DB *gorm.DB
}

type User struct {
	models.ULID
	GoogleId  string    `gorm:"uniqueIndex"`
	Email     string    `gorm:"uniqueIndex"`
	Username  string    `gorm:"uniqueIndex"`
	Name      string    `gorm:"not null"`
	IsAdmin   bool      `gorm:"default:false"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (s *UserService) GetUserByID(id string) (*User, error) {
	var user User
	if err := s.DB.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) GetAllUsers() ([]User, error) {
	var user []User
	if err := s.DB.Find(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) CreateUser(user *User) error {
	return s.DB.Create(user).Error
}
