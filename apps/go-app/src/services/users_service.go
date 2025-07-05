package services

import (
	"gorm.io/gorm"
	"time"
)

type UserService struct {
	DB *gorm.DB
}

type User struct {
	id        string    `gorm:"primaryKey"`
	googleId  string    `gorm:"uniqueIndex"`
	email     string    `gorm:"uniqueIndex"`
	name      string    `gorm:"not null"`
	isAdmin   bool      `gorm:"default:false"`
	createdAt time.Time `gorm:"autoCreateTime"`
	updatedAt time.Time `gorm:"autoUpdateTime"`
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
