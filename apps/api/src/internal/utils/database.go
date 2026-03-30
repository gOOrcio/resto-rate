package utils

import (
	"api/src/internal/models"
	"log/slog"
	"os"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RequiredTags is the canonical list of predefined tags seeded into the DB.
// Exported so tests can verify the list without a DB connection.
var RequiredTags = []models.Tag{
	// Cuisine
	{Slug: "italian", Label: "Italian", Category: "Cuisine"},
	{Slug: "japanese", Label: "Japanese", Category: "Cuisine"},
	{Slug: "mexican", Label: "Mexican", Category: "Cuisine"},
	{Slug: "chinese", Label: "Chinese", Category: "Cuisine"},
	{Slug: "indian", Label: "Indian", Category: "Cuisine"},
	{Slug: "french", Label: "French", Category: "Cuisine"},
	{Slug: "thai", Label: "Thai", Category: "Cuisine"},
	{Slug: "american", Label: "American", Category: "Cuisine"},
	{Slug: "mediterranean", Label: "Mediterranean", Category: "Cuisine"},
	{Slug: "korean", Label: "Korean", Category: "Cuisine"},
	// Vibe
	{Slug: "romantic", Label: "Romantic", Category: "Vibe"},
	{Slug: "casual", Label: "Casual", Category: "Vibe"},
	{Slug: "family-friendly", Label: "Family Friendly", Category: "Vibe"},
	{Slug: "date-night", Label: "Date Night", Category: "Vibe"},
	{Slug: "business-lunch", Label: "Business Lunch", Category: "Vibe"},
	{Slug: "lively", Label: "Lively", Category: "Vibe"},
	{Slug: "quiet", Label: "Quiet", Category: "Vibe"},
	{Slug: "trendy", Label: "Trendy", Category: "Vibe"},
	// Price
	{Slug: "budget", Label: "Budget", Category: "Price"},
	{Slug: "mid-range", Label: "Mid-Range", Category: "Price"},
	{Slug: "expensive", Label: "Expensive", Category: "Price"},
	{Slug: "splurge", Label: "Splurge", Category: "Price"},
	// Dietary
	{Slug: "vegan", Label: "Vegan", Category: "Dietary"},
	{Slug: "vegetarian", Label: "Vegetarian", Category: "Dietary"},
	{Slug: "gluten-free", Label: "Gluten-Free", Category: "Dietary"},
	{Slug: "halal", Label: "Halal", Category: "Dietary"},
	{Slug: "kosher", Label: "Kosher", Category: "Dietary"},
	{Slug: "dairy-free", Label: "Dairy-Free", Category: "Dietary"},
	// Group
	{Slug: "solo", Label: "Solo", Category: "Group"},
	{Slug: "couple", Label: "Couple", Category: "Group"},
	{Slug: "small-group", Label: "Small Group", Category: "Group"},
	{Slug: "large-group", Label: "Large Group", Category: "Group"},
	// Occasion
	{Slug: "birthday", Label: "Birthday", Category: "Occasion"},
	{Slug: "anniversary", Label: "Anniversary", Category: "Occasion"},
	{Slug: "brunch", Label: "Brunch", Category: "Occasion"},
	{Slug: "late-night", Label: "Late Night", Category: "Occasion"},
	{Slug: "celebration", Label: "Celebration", Category: "Occasion"},
	{Slug: "quick-bite", Label: "Quick Bite", Category: "Occasion"},
}

// removedSlugs lists tag slugs that were once seeded but have since been
// removed. SeedRequiredData deletes them on every startup so they don't
// linger in existing databases.
var removedSlugs = []string{
	"fast-service", "outdoor-seating", "delivery", "takeaway", "reservations", "dog-friendly",
}

// SeedRequiredData seeds production-required data unconditionally using upsert.
// Safe to call on every startup — idempotent by slug.
func SeedRequiredData(db *gorm.DB) error {
	slog.Info("Seeding required data (tags)...")

	// Remove tags that are no longer part of the canonical list.
	if len(removedSlugs) > 0 {
		if err := db.Where("slug IN ?", removedSlugs).Delete(&models.Tag{}).Error; err != nil {
			return err
		}
	}

	// Copy the list so GORM's ID-setting side-effect doesn't mutate RequiredTags
	tags := make([]models.Tag, len(RequiredTags))
	copy(tags, RequiredTags)

	result := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "slug"}},
		DoUpdates: clause.AssignmentColumns([]string{"label", "category"}),
	}).Create(&tags)
	if result.Error != nil {
		return result.Error
	}

	slog.Info("Required data seeded successfully", slog.Int64("tags", int64(len(tags))))
	return nil
}

func CreateSchema(db *gorm.DB) error {
	slog.Info("Creating database schema...")

	if err := db.AutoMigrate(&models.Restaurant{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&models.Review{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&models.Tag{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&models.WishlistItem{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&models.FriendRequest{}); err != nil {
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
			{GoogleId: models.StringPtr("1"), Email: models.StringPtr("user1@example.com"), Name: "User One", Username: models.StringPtr("username-a")},
			{GoogleId: models.StringPtr("2"), Email: models.StringPtr("user2@example.com"), Name: "User Two", Username: models.StringPtr("username-b")},
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
