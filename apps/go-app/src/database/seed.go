package database

import (
	"go-app/src/services/models"
	"gorm.io/gorm"
	"log"
)

func autoMigrateAndSeedRestaurants(db *gorm.DB) error {
	if err := db.AutoMigrate(&models.Restaurant{}); err != nil {
		return err
	}

	var count int64
	if err := db.Model(&models.Restaurant{}).Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		seedRestaurants := []models.Restaurant{
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
	if err := db.AutoMigrate(&models.User{}); err != nil {
		return err
	}

	var count int64
	if err := db.Model(&models.User{}).Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		seedUsers := []models.User{
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
	log.Println("Seeding finished")
	return nil
}
