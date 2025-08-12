package utils

import (
	"api/src/internal/models"
	"log/slog"
	"os"
	"strings"

	"gorm.io/gorm"
)

func CreateSchema(db *gorm.DB) error {
	slog.Info("Creating database schema...")

	if err := db.AutoMigrate(&models.Restaurant{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		return err
	}

	slog.Info("Database schema created successfully")
	return nil
}

func seedRestaurants(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.Restaurant{}).Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		slog.Info("Seeding restaurants...")
		seedRestaurants := []models.Restaurant{
			{GoogleID: "places/1", Address: "Szkolna warszawa", Name: "Banjaluka"},
			{GoogleID: "places/2", Address: "Plac kazimierza tarnów", Name: "Bistro Przepis"},
		}
		if err := db.Create(&seedRestaurants).Error; err != nil {
			return err
		}
		slog.Info("Restaurants seeded successfully")
	} else {
		slog.Info("Restaurants already present, skipping seed")
	}
	return nil
}

func seedUsers(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.User{}).Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		slog.Info("Seeding users...")
		seedUsers := []models.User{
			{GoogleId: "1", Email: "user1@example.com", Name: "User One", Username: "username-a", IsAdmin: true},
			{GoogleId: "2", Email: "user2@example.com", Name: "User Two", Username: "username-b", IsAdmin: false},
		}
		if err := db.Create(&seedUsers).Error; err != nil {
			return err
		}
		slog.Info("Users seeded successfully")
	} else {
		slog.Info("Users already present, skipping seed")
	}
	return nil
}

func SeedDatabase(db *gorm.DB) error {
	if os.Getenv("ENV") == "dev" && strings.EqualFold(os.Getenv("SEED"), "true") {
		slog.Info("Development environment detected with SEED=true, seeding database...")
		if err := seedRestaurants(db); err != nil {
			return err
		}
		if err := seedUsers(db); err != nil {
			return err
		}
		slog.Info("Database seeding completed")
	}
	return nil
}
