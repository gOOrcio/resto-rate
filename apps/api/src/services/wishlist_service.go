package services

import (
	wishlistv1 "api/src/generated/wishlist/v1"
	"api/src/generated/wishlist/v1/v1connect"
	"api/src/internal/models"
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/valkey-io/valkey-go"
	"gorm.io/gorm"
)

type WishlistService struct {
	v1connect.UnimplementedWishlistServiceHandler
	DB     *gorm.DB
	Valkey valkey.Client
}

func NewWishlistService(db *gorm.DB, kv valkey.Client) *WishlistService {
	return &WishlistService{DB: db, Valkey: kv}
}

func (s *WishlistService) AddToWishlist(
	ctx context.Context,
	req *connect.Request[wishlistv1.AddToWishlistRequest],
) (*connect.Response[wishlistv1.AddToWishlistResponse], error) {
	if s.DB == nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("database not initialized"))
	}

	userID, err := getUserIDFromSession(ctx, req.Header(), s.Valkey)
	if err != nil {
		return nil, err
	}

	if req.Msg.GooglePlacesId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("google_places_id is required"))
	}

	var restaurant models.Restaurant
	result := s.DB.WithContext(ctx).
		Where(models.Restaurant{GoogleID: req.Msg.GooglePlacesId}).
		Attrs(models.Restaurant{
			Name:    req.Msg.RestaurantName,
			Address: req.Msg.RestaurantAddress,
			City:    req.Msg.City,
			Country: req.Msg.Country,
		}).
		FirstOrCreate(&restaurant)
	if result.Error != nil {
		return nil, result.Error
	}

	item := models.WishlistItem{
		UserID:         userID,
		RestaurantID:   restaurant.ID,
		GooglePlacesID: req.Msg.GooglePlacesId,
	}

	// Use FirstOrCreate so calling Add twice is idempotent
	var existing models.WishlistItem
	res := s.DB.WithContext(ctx).
		Where("user_id = ? AND restaurant_id = ?", userID, restaurant.ID).
		FirstOrCreate(&existing, item)
	if res.Error != nil {
		return nil, res.Error
	}

	// Preload restaurant for ToProto
	existing.Restaurant = restaurant

	return connect.NewResponse(&wishlistv1.AddToWishlistResponse{
		Item: existing.ToProto(),
	}), nil
}

func (s *WishlistService) RemoveFromWishlist(
	ctx context.Context,
	req *connect.Request[wishlistv1.RemoveFromWishlistRequest],
) (*connect.Response[wishlistv1.RemoveFromWishlistResponse], error) {
	if s.DB == nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("database not initialized"))
	}

	userID, err := getUserIDFromSession(ctx, req.Header(), s.Valkey)
	if err != nil {
		return nil, err
	}

	if req.Msg.GooglePlacesId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("google_places_id is required"))
	}

	result := s.DB.WithContext(ctx).
		Where("user_id = ? AND google_places_id = ?", userID, req.Msg.GooglePlacesId).
		Delete(&models.WishlistItem{})
	if result.Error != nil {
		return nil, result.Error
	}

	return connect.NewResponse(&wishlistv1.RemoveFromWishlistResponse{Success: result.RowsAffected > 0}), nil
}

func (s *WishlistService) ListWishlist(
	ctx context.Context,
	req *connect.Request[wishlistv1.ListWishlistRequest],
) (*connect.Response[wishlistv1.ListWishlistResponse], error) {
	if s.DB == nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("database not initialized"))
	}

	userID, err := getUserIDFromSession(ctx, req.Header(), s.Valkey)
	if err != nil {
		return nil, err
	}

	query := s.DB.WithContext(ctx).
		Preload("Restaurant").
		Where("user_id = ?", userID)

	if req.Msg.GooglePlacesId != "" {
		query = query.Where("google_places_id = ?", req.Msg.GooglePlacesId)
	}

	var items []models.WishlistItem
	if err := query.Order("created_at desc").Find(&items).Error; err != nil {
		return nil, err
	}

	protos := make([]*wishlistv1.WishlistItemProto, len(items))
	for i, item := range items {
		protos[i] = item.ToProto()
	}

	return connect.NewResponse(&wishlistv1.ListWishlistResponse{Items: protos}), nil
}

