package database

import (
	"go-app/src/services"
	"gorm.io/gorm"
)

func autoMigrateAndSeedRestaurants(db *gorm.DB) error {
	if err := db.AutoMigrate(&services.Restaurant{}); err != nil {
		return err
	}

	var count int64
	if err := db.Model(&services.Restaurant{}).Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		seedRestaurants := []services.Restaurant{
			{GoogleID: "g1", Email: "a@b.com", Name: "Testaurant"},
			{GoogleID: "g2", Email: "c@d.com", Name: "Food Place"},
		}
		if err := db.Create(&seedRestaurants).Error; err != nil {
			return err
		}
	}
	return nil
}

func autoMigrateAndSeedUsers(db *gorm.DB) error {
	if err := db.AutoMigrate(&services.User{}); err != nil {
		return err
	}

	var count int64
	if err := db.Model(&services.User{}).Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		seedUsers := []services.User{
			{GoogleId: "1", Email: "user1@example.com", Name: "User One", Username: "username-a", IsAdmin: true},
			{GoogleId: "2", Email: "user2@example.com", Name: "User Two", Username: "username-b", IsAdmin: false},
		}
		if err := db.Create(&seedUsers).Error; err != nil {
			return err
		}
	}
	return nil
}

func AutoMigrateAndSeed(db *gorm.DB) error {
	if err := autoMigrateAndSeedRestaurants(db); err != nil {
		return err
	}
	if err := autoMigrateAndSeedUsers(db); err != nil {
		return err
	}
	return nil
}
