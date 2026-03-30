package models

import (
	wishlistv1 "api/src/generated/wishlist/v1"
	"time"

	"gorm.io/gorm"
)

type WishlistItem struct {
	UUIDv7
	UserID         string     `gorm:"not null;index;uniqueIndex:idx_wishlist_user_restaurant"`
	RestaurantID   string     `gorm:"not null;index;uniqueIndex:idx_wishlist_user_restaurant"`
	Restaurant     Restaurant `gorm:"foreignKey:RestaurantID"`
	GooglePlacesID string     `gorm:"not null;index"`
	Tags           []string   `gorm:"serializer:json"`
	CreatedAt      time.Time  `gorm:"autoCreateTime"`
}

func (w *WishlistItem) BeforeCreate(tx *gorm.DB) (err error) {
	return w.UUIDv7.BeforeCreate(tx)
}

func (w *WishlistItem) ToProto() *wishlistv1.WishlistItemProto {
	tags := w.Tags
	if tags == nil {
		tags = []string{}
	}
	return &wishlistv1.WishlistItemProto{
		Id:                       w.ID,
		GooglePlacesId:           w.GooglePlacesID,
		RestaurantName:           w.Restaurant.Name,
		RestaurantAddress:        w.Restaurant.Address,
		City:                     w.Restaurant.City,
		Country:                  w.Restaurant.Country,
		RestaurantPhotoReference: w.Restaurant.PhotoReference,
		CreatedAt:                w.CreatedAt.Unix(),
		Tags:                     tags,
	}
}
