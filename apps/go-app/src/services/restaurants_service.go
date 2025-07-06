package services

import (
	"go-app/src/services/models"
	"gorm.io/gorm"
	"time"
)

type RestaurantsService struct {
	DB *gorm.DB
}

type Restaurant struct {
	models.ULID
	GoogleID  string    `gorm:"uniqueIndex" json:"googleId"`
	Email     string    `gorm:"uniqueIndex" json:"email"`
	Name      string    `gorm:"not null" json:"name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (s *RestaurantsService) GetRestaurantByID(id string) (*Restaurant, error) {
	var restaurant Restaurant
	if err := s.DB.First(&restaurant, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &restaurant, nil
}

func (s *RestaurantsService) GetAllRestaurants() ([]Restaurant, error) {
	var restaurant []Restaurant
	if err := s.DB.Find(&restaurant).Error; err != nil {
		return nil, err
	}
	return restaurant, nil
}

func (s *RestaurantsService) CreateRestaurant(restaurant *Restaurant) error {
	return s.DB.Create(restaurant).Error
}
