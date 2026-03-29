package services

import (
	wishlistv1 "api/src/generated/wishlist/v1"
	"api/src/generated/wishlist/v1/v1connect"
	"api/src/internal/models"
	"context"
	"errors"
	"fmt"
	"strings"

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
		return nil, connect.NewError(connect.CodeInternal, errors.New(errDatabaseNotInitialized))
	}

	userID, err := getUserIDFromSession(ctx, req.Header(), s.Valkey)
	if err != nil {
		return nil, err
	}

	if req.Msg.GooglePlacesId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New(errGooglePlacesIDRequired))
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

	tags := req.Msg.TagSlugs
	if tags == nil {
		tags = []string{}
	}

	item := models.WishlistItem{
		UserID:         userID,
		RestaurantID:   restaurant.ID,
		GooglePlacesID: req.Msg.GooglePlacesId,
		Tags:           tags,
	}

	// Use FirstOrCreate so calling Add twice is idempotent; update tags if already exists
	var existing models.WishlistItem
	res := s.DB.WithContext(ctx).
		Where("user_id = ? AND restaurant_id = ?", userID, restaurant.ID).
		FirstOrCreate(&existing, item)
	if res.Error != nil {
		return nil, res.Error
	}

	// If item already existed and tags were explicitly provided, update them.
	// An empty tag list is not treated as "clear tags" — use a dedicated update
	// endpoint for that to avoid silently wiping tags on a plain re-add.
	if res.RowsAffected == 0 && len(req.Msg.TagSlugs) > 0 {
		if err := s.DB.WithContext(ctx).Model(&existing).Update("tags", tags).Error; err != nil {
			return nil, err
		}
		existing.Tags = tags
	}

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
		return nil, connect.NewError(connect.CodeInternal, errors.New(errDatabaseNotInitialized))
	}

	userID, err := getUserIDFromSession(ctx, req.Header(), s.Valkey)
	if err != nil {
		return nil, err
	}

	if req.Msg.GooglePlacesId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New(errGooglePlacesIDRequired))
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
		return nil, connect.NewError(connect.CodeInternal, errors.New(errDatabaseNotInitialized))
	}

	callerID, err := getUserIDFromSession(ctx, req.Header(), s.Valkey)
	if err != nil {
		return nil, err
	}

	// Determine which user's wishlist to return
	targetUserID := callerID
	if req.Msg.TargetUserId != "" && req.Msg.TargetUserId != callerID {
		if err := assertFriendship(ctx, s.DB, callerID, req.Msg.TargetUserId); err != nil {
			return nil, err
		}
		targetUserID = req.Msg.TargetUserId
	}

	needsJoin := req.Msg.City != "" || req.Msg.Country != "" ||
		req.Msg.SortBy == wishlistv1.WishlistSortBy_WISHLIST_SORT_BY_NAME_ASC ||
		req.Msg.SortBy == wishlistv1.WishlistSortBy_WISHLIST_SORT_BY_NAME_DESC

	query := s.DB.WithContext(ctx).
		Preload("Restaurant").
		Where("wishlist_items.user_id = ?", targetUserID)

	if needsJoin {
		query = query.Joins("JOIN restaurants ON restaurants.id = wishlist_items.restaurant_id")
	}

	if req.Msg.GooglePlacesId != "" {
		query = query.Where("wishlist_items.google_places_id = ?", req.Msg.GooglePlacesId)
	}

	if req.Msg.City != "" {
		query = query.Where("restaurants.city ILIKE ?", "%"+req.Msg.City+"%")
	}
	if req.Msg.Country != "" {
		query = query.Where("restaurants.country ILIKE ?", "%"+req.Msg.Country+"%")
	}

	query = applyWishlistTagFilter(query, req.Msg.TagSlugs, req.Msg.TagFilterMode)

	switch req.Msg.SortBy {
	case wishlistv1.WishlistSortBy_WISHLIST_SORT_BY_DATE_ASC:
		query = query.Order("wishlist_items.created_at ASC")
	case wishlistv1.WishlistSortBy_WISHLIST_SORT_BY_NAME_ASC:
		query = query.Order("restaurants.name ASC")
	case wishlistv1.WishlistSortBy_WISHLIST_SORT_BY_NAME_DESC:
		query = query.Order("restaurants.name DESC")
	default: // UNSPECIFIED and DATE_DESC → newest first
		query = query.Order("wishlist_items.created_at DESC")
	}

	var items []models.WishlistItem
	if err := query.Find(&items).Error; err != nil {
		return nil, err
	}

	protos := make([]*wishlistv1.WishlistItemProto, len(items))
	for i, item := range items {
		protos[i] = item.ToProto()
	}

	return connect.NewResponse(&wishlistv1.ListWishlistResponse{Items: protos}), nil
}

// applyWishlistTagFilter adds WHERE clauses for tag filtering based on mode (AND/OR).
func applyWishlistTagFilter(query *gorm.DB, slugs []string, mode wishlistv1.WishlistTagFilterMode) *gorm.DB {
	if len(slugs) == 0 {
		return query
	}
	if mode == wishlistv1.WishlistTagFilterMode_WISHLIST_TAG_FILTER_MODE_AND {
		for _, slug := range slugs {
			query = query.Where("wishlist_items.tags LIKE ?", fmt.Sprintf(`%%"%s"%%`, slug))
		}
		return query
	}
	// OR (default): at least one specified tag must appear
	conditions := make([]string, len(slugs))
	args := make([]interface{}, len(slugs))
	for i, slug := range slugs {
		conditions[i] = "wishlist_items.tags LIKE ?"
		args[i] = fmt.Sprintf(`%%"%s"%%`, slug)
	}
	return query.Where("("+strings.Join(conditions, " OR ")+")", args...)
}
