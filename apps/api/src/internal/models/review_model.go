package models

import (
	reviewspb "api/src/generated/reviews/v1"
	"time"

	"gorm.io/gorm"
)

type Review struct {
	UUIDv7
	RestaurantID   string    `gorm:"not null;index;uniqueIndex:idx_review_restaurant_user"`
	UserID         string    `gorm:"not null;index;uniqueIndex:idx_review_restaurant_user"`
	GooglePlacesID string    `gorm:"index"`
	Comment        string
	Rating         float64   `gorm:"not null"`
	Tags           []string  `gorm:"serializer:json"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`
}

func (r *Review) BeforeCreate(tx *gorm.DB) (err error) {
	return r.UUIDv7.BeforeCreate(tx)
}

func (r *Review) ToProto() *reviewspb.ReviewProto {
	tags := r.Tags
	if tags == nil {
		tags = []string{}
	}
	return &reviewspb.ReviewProto{
		Id:             r.ID,
		UserId:         r.UserID,
		RestaurantId:   r.RestaurantID,
		GooglePlacesId: r.GooglePlacesID,
		Comment:        r.Comment,
		Rating:         r.Rating,
		Tags:           tags,
		CreatedAt:      r.CreatedAt.Unix(),
		UpdatedAt:      r.UpdatedAt.Unix(),
	}
}
