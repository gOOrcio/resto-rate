package models

import (
	restaurantpb "api/src/generated/restaurants/v1"
	"time"

	"gorm.io/gorm"
)

type Restaurant struct {
	UUIDv7
	GoogleID  string    `gorm:"uniqueIndex" json:"googleId"`
	Address   string    `gorm:"uniqueIndex" json:"email"`
	Name      string    `gorm:"not null" json:"name"`
	City      string    `gorm:"index"`
	Country   string    `gorm:"index"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (r *Restaurant) BeforeCreate(tx *gorm.DB) (err error) {
	if err = r.UUIDv7.BeforeCreate(tx); err != nil {
		return err
	}
	return nil
}

func (r *Restaurant) ToProto() *restaurantpb.RestaurantProto {
	return &restaurantpb.RestaurantProto{
		Id:             r.ID,
		GooglePlacesId: r.GoogleID,
		Address:        r.Address,
		Name:           r.Name,
		City:           r.City,
		Country:        r.Country,
		CreatedAt:      r.CreatedAt.Unix(),
		UpdatedAt:      r.UpdatedAt.Unix(),
	}
}
