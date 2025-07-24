package models

import (
	restaurantpb "api/src/generated/restaurants/v1"
	"gorm.io/gorm"
	"time"
)

type Restaurant struct {
	ULID
	GoogleID  string    `gorm:"uniqueIndex" json:"googleId"`
	Email     string    `gorm:"uniqueIndex" json:"email"`
	Name      string    `gorm:"not null" json:"name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (r *Restaurant) BeforeCreate(tx *gorm.DB) (err error) {
	if err = r.ULID.BeforeCreate(tx); err != nil {
		return err
	}
	return nil
}

func (r *Restaurant) ToProto() *restaurantpb.RestaurantProto {
	return &restaurantpb.RestaurantProto{
		Id:        r.ID,
		GoogleId:  r.GoogleID,
		Email:     r.Email,
		Name:      r.Name,
		CreatedAt: r.CreatedAt.Unix(),
		UpdatedAt: r.UpdatedAt.Unix(),
	}
}
