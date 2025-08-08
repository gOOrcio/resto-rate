package database

import (
	"api/src/internal/models"
	"log"

	"gorm.io/gorm"
)

func CreateSchema(db *gorm.DB) error {
	log.Println("Creating database schema...")

	if err := db.AutoMigrate(&models.Restaurant{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		return err
	}

	log.Println("Database schema created successfully")
	return nil
}

func seedRestaurants(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.Restaurant{}).Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		log.Println("Seeding restaurants...")
		seedRestaurants := []models.Restaurant{
			{GoogleID: "g1", Email: "a@b.com", Name: "Testaurant"},
			{GoogleID: "g2", Email: "c@d.com", Name: "Food Place"},
		}
		if err := db.Create(&seedRestaurants).Error; err != nil {
			return err
		}
		log.Println("Restaurants seeded successfully")
	} else {
		log.Println("Restaurants already present, skipping seed")
	}
	return nil
}

func seedUsers(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.User{}).Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		log.Println("Seeding users...")
		seedUsers := []models.User{
			{GoogleId: "1", Email: "user1@example.com", Name: "User One", Username: "username-a", IsAdmin: true},
			{GoogleId: "2", Email: "user2@example.com", Name: "User Two", Username: "username-b", IsAdmin: false},
		}
		if err := db.Create(&seedUsers).Error; err != nil {
			return err
		}
		log.Println("Users seeded successfully")
	} else {
		log.Println("Users already present, skipping seed")
	}
	return nil
}

func SeedDatabase(db *gorm.DB) error {
	if err := seedRestaurants(db); err != nil {
		return err
	}
	if err := seedUsers(db); err != nil {
		return err
	}
	log.Println("Database seeding completed")
	return nil
}
