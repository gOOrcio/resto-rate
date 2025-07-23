package services

import (
	"context"
	"errors"
	"fmt"
	restaurantpb "go-app/src/generated/resto-rate/generated/go/restaurants/v1"
	"go-app/src/services/models"
	"go-app/src/services/utils"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type RestaurantsService struct {
	restaurantpb.UnimplementedRestaurantServiceServer
	DB *gorm.DB
}

func (s *RestaurantsService) CreateRestaurant(ctx context.Context, req *restaurantpb.CreateRestaurantRequest) (*restaurantpb.RestaurantProto, error) {
	restaurant := &models.Restaurant{
		GoogleID: req.Restaurant.GoogleId,
		Email:    req.Restaurant.Email,
		Name:     req.Restaurant.Name,
	}

	if err := s.DB.WithContext(ctx).Create(restaurant).Error; err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	return restaurant.ToProto(), nil
}

func (s *RestaurantsService) GetRestaurant(ctx context.Context, req *restaurantpb.GetRestaurantRequest) (*restaurantpb.RestaurantProto, error) {
	if req.Id == "" {
		return nil, fmt.Errorf("restaurant ID is required")
	}

	var restaurant models.Restaurant
	if err := s.DB.WithContext(ctx).First(&restaurant, "id = ?", req.Id).Error; err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("restaurant with ID %s not found", req.Id)
		}
		return nil, err
	}

	return restaurant.ToProto(), nil
}

func (s *RestaurantsService) UpdateRestaurant(ctx context.Context, req *restaurantpb.UpdateRestaurantRequest) (*restaurantpb.RestaurantProto, error) {
	restaurant, err := s.findRestaurantByIDWithContext(ctx, req.Restaurant.Id)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	updates := map[string]interface{}{
		"Name":  req.Restaurant.Name,
		"Email": req.Restaurant.Email,
	}

	if err := s.DB.WithContext(ctx).Model(&restaurant).Updates(updates).Error; err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	return restaurant.ToProto(), nil
}

func (s *RestaurantsService) DeleteRestaurant(ctx context.Context, req *restaurantpb.DeleteRestaurantRequest) (*emptypb.Empty, error) {
	restaurant, err := s.findRestaurantByIDWithContext(ctx, req.Id)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	if err := s.DB.WithContext(ctx).Delete(&restaurant).Error; err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *RestaurantsService) ListRestaurants(ctx context.Context, req *restaurantpb.ListRestaurantsRequest) (*restaurantpb.ListRestaurantsResponse, error) {
	var restaurants []*models.Restaurant
	totalCount, err := utils.Paginate(s.DB.WithContext(ctx), &restaurants, req.Page, req.PageSize)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	var restaurantProtos []*restaurantpb.RestaurantProto
	for _, restaurant := range restaurants {
		restaurantProtos = append(restaurantProtos, restaurant.ToProto())
	}

	var nextPageToken string
	if int64((req.Page-1)*req.PageSize+int32(len(restaurants))) < totalCount {
		nextPageToken = fmt.Sprintf("%d", req.Page+1)
	}

	return &restaurantpb.ListRestaurantsResponse{
		Restaurants:   restaurantProtos,
		TotalCount:    int32(totalCount),
		Page:          req.Page,
		PageSize:      req.PageSize,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *RestaurantsService) findRestaurantByIDWithContext(ctx context.Context, id string) (*models.Restaurant, error) {
	if id == "" {
		return nil, fmt.Errorf("restaurant ID is required")
	}

	var restaurant models.Restaurant
	if err := s.DB.WithContext(ctx).First(&restaurant, "id = ?", id).Error; err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("restaurant with ID %s not found", id)
		}
		return nil, err
	}

	return &restaurant, nil
}
