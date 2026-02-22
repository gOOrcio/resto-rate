package services

import (
	v1 "api/src/generated/restaurants/v1"
	"api/src/generated/restaurants/v1/v1connect"
	"api/src/internal/models"
	"context"
	"errors"
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"connectrpc.com/connect"
)

type RestaurantsService struct {
	v1connect.UnimplementedRestaurantsServiceHandler
	DB *gorm.DB
}

func NewRestaurantsService(db *gorm.DB) *RestaurantsService {
	return &RestaurantsService{DB: db}
}

func (s *RestaurantsService) CreateRestaurant(
	ctx context.Context,
	req *connect.Request[v1.CreateRestaurantRequest],
) (*connect.Response[v1.CreateRestaurantResponse], error) {
	if req.Msg.Name == "" {
		return nil, fmt.Errorf("restaurant name is required")
	}

	if req.Msg.GooglePlacesId == "" {
		slog.Debug("Google ID is not provided for restaurant", slog.String("name", req.Msg.Name))
	}

	if req.Msg.Address == "" {
		slog.Debug("Address is not provided for restaurant", slog.String("name", req.Msg.Name))
	}
	restaurant := &models.Restaurant{
		GoogleID: req.Msg.GooglePlacesId,
		Address:  req.Msg.Address,
		Name:     req.Msg.Name,
	}

	if err := s.DB.WithContext(ctx).Create(restaurant).Error; err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	res := connect.NewResponse(&v1.CreateRestaurantResponse{
		Restaurant: restaurant.ToProto(),
	})
	return res, nil
}

func (s *RestaurantsService) GetRestaurant(
	ctx context.Context,
	req *connect.Request[v1.GetRestaurantRequest],
) (*connect.Response[v1.GetRestaurantResponse], error) {
	if req.Msg.Id == "" {
		return nil, fmt.Errorf("restaurant ID is required")
	}

	var restaurant models.Restaurant
	if err := s.DB.WithContext(ctx).First(&restaurant, "id = ?", req.Msg.Id).Error; err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("restaurant with ID %s not found", req.Msg.Id)
		}
		return nil, err
	}

	res := connect.NewResponse(&v1.GetRestaurantResponse{
		Restaurant: restaurant.ToProto(),
	})
	return res, nil
}

func (s *RestaurantsService) UpdateRestaurant(
	ctx context.Context,
	req *connect.Request[v1.UpdateRestaurantRequest],
) (*connect.Response[v1.UpdateRestaurantResponse], error) {

	restaurant, err := s.findRestaurantByIDWithContext(ctx, req.Msg.Id)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	updates := map[string]interface{}{
		"Name":    req.Msg.Name,
		"Address": req.Msg.Address,
	}

	if err := s.DB.WithContext(ctx).Model(restaurant).Updates(updates).Error; err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	updated, err := s.findRestaurantByIDWithContext(ctx, req.Msg.Id)
	if err != nil {
		return nil, err
	}

	res := connect.NewResponse(&v1.UpdateRestaurantResponse{
		Restaurant: updated.ToProto(),
	})
	return res, nil
}

func (s *RestaurantsService) DeleteRestaurant(
	ctx context.Context,
	req *connect.Request[v1.DeleteRestaurantRequest],
) (*connect.Response[v1.DeleteRestaurantResponse], error) {
	if req.Msg.Id == "" {
		return nil, fmt.Errorf("restaurant ID is required")
	}

	var restaurant models.Restaurant
	if err := s.DB.WithContext(ctx).First(&restaurant, "id = ?", req.Msg.Id).Error; err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("restaurant with ID %s not found", req.Msg.Id)
		}
		return nil, err
	}

	if err := s.DB.WithContext(ctx).Delete(&restaurant).Error; err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	res := connect.NewResponse(&v1.DeleteRestaurantResponse{
		Success: true,
	})
	return res, nil
}

func (s *RestaurantsService) ListRestaurants(
	ctx context.Context,
	req *connect.Request[v1.ListRestaurantsRequest],
) (*connect.Response[v1.ListRestaurantsResponse], error) {
	page := int(req.Msg.Page)
	pageSize := int(req.Msg.PageSize)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	var restaurants []models.Restaurant
	var total int64

	if err := s.DB.WithContext(ctx).Model(&models.Restaurant{}).Count(&total).Error; err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	if err := s.DB.WithContext(ctx).Offset(offset).Limit(pageSize).Find(&restaurants).Error; err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	restaurantProtos := make([]*v1.RestaurantProto, len(restaurants))
	for i, restaurant := range restaurants {
		restaurantProtos[i] = restaurant.ToProto()
	}

	res := connect.NewResponse(&v1.ListRestaurantsResponse{
		Restaurants: restaurantProtos,
		Total:       int32(total),
		Page:        int32(page),
		PageSize:    int32(pageSize),
	})
	return res, nil
}

func (s *RestaurantsService) findRestaurantByIDWithContext(ctx context.Context, id string) (*models.Restaurant, error) {
	var restaurant models.Restaurant
	if err := s.DB.WithContext(ctx).First(&restaurant, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("restaurant with ID %s not found", id)
		}
		return nil, err
	}
	return &restaurant, nil
}
