package services

import (
	v1 "api/src/generated/restaurants/v1"
	"api/src/generated/restaurants/v1/v1connect"
	"api/src/services/models"
	"api/src/services/utils"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RestaurantsService struct {
	v1connect.UnimplementedRestaurantServiceHandler
	DB *gorm.DB
}

func (s *RestaurantsService) CreateRestaurant(
	ctx context.Context,
	req *connect.Request[v1.CreateRestaurantRequest],
) (*connect.Response[v1.RestaurantProto], error) {
	restaurant := &models.Restaurant{
		GoogleID: req.Msg.Restaurant.GoogleId,
		Email:    req.Msg.Restaurant.Email,
		Name:     req.Msg.Restaurant.Name,
	}

	if err := s.DB.WithContext(ctx).Create(restaurant).Error; err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	res := connect.NewResponse(restaurant.ToProto())
	return res, nil
}

func (s *RestaurantsService) GetRestaurant(
	ctx context.Context,
	req *connect.Request[v1.GetRestaurantRequest],
) (*connect.Response[v1.RestaurantProto], error) {
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

	res := connect.NewResponse(restaurant.ToProto())
	return res, nil
}

func (s *RestaurantsService) UpdateRestaurant(
	ctx context.Context,
	req *connect.Request[v1.UpdateRestaurantRequest],
) (*connect.Response[v1.RestaurantProto], error) {
	restaurant, err := s.findRestaurantByIDWithContext(ctx, req.Msg.Restaurant.Id)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	updates := map[string]interface{}{
		"Name":  req.Msg.Restaurant.Name,
		"Email": req.Msg.Restaurant.Email,
	}

	if err := s.DB.WithContext(ctx).Model(&restaurant).Updates(updates).Error; err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	res := connect.NewResponse(restaurant.ToProto())
	return res, nil
}

func (s *RestaurantsService) DeleteRestaurant(
	ctx context.Context,
	req *connect.Request[v1.DeleteRestaurantRequest],
) (*connect.Response[emptypb.Empty], error) {
	restaurant, err := s.findRestaurantByIDWithContext(ctx, req.Msg.Id)
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

	res := connect.NewResponse(&emptypb.Empty{})
	return res, nil
}

func (s *RestaurantsService) ListRestaurants(
	ctx context.Context,
	req *connect.Request[v1.ListRestaurantsRequest],
) (*connect.Response[v1.ListRestaurantsResponse], error) {
	var restaurants []*models.Restaurant
	totalCount, err := utils.Paginate(s.DB.WithContext(ctx), &restaurants, req.Msg.Page, req.Msg.PageSize)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	var restaurantProtos []*v1.RestaurantProto
	for _, restaurant := range restaurants {
		restaurantProtos = append(restaurantProtos, restaurant.ToProto())
	}

	var nextPageToken string
	if int64((req.Msg.Page-1)*req.Msg.PageSize+int32(len(restaurants))) < totalCount {
		nextPageToken = fmt.Sprintf("%d", req.Msg.Page+1)
	}

	response := &v1.ListRestaurantsResponse{
		Restaurants:   restaurantProtos,
		TotalCount:    int32(totalCount),
		Page:          req.Msg.Page,
		PageSize:      req.Msg.PageSize,
		NextPageToken: nextPageToken,
	}

	res := connect.NewResponse(response)
	return res, nil
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
