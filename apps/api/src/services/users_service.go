package services

import (
	v1 "api/src/generated/users/v1"
	"api/src/generated/users/v1/v1connect"
	"api/src/internal/models"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"connectrpc.com/connect"
)

type UserService struct {
	v1connect.UnimplementedUsersServiceHandler
	DB *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{DB: db}
}

func (u *UserService) GetUser(
	ctx context.Context,
	req *connect.Request[v1.GetUserRequest],
) (*connect.Response[v1.GetUserResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("user ID cannot be empty"))
	}
	user, err := u.findUserByIDWithContext(ctx, req.Msg.Id)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&v1.GetUserResponse{User: user.ToProto()}), nil
}

func (u *UserService) ListUsers(
	ctx context.Context,
	req *connect.Request[v1.ListUsersRequest],
) (*connect.Response[v1.ListUsersResponse], error) {
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

	var users []models.User
	var total int64

	if err := u.DB.WithContext(ctx).Model(&models.User{}).Count(&total).Error; err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	if err := u.DB.WithContext(ctx).Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	userProtos := make([]*v1.UserProto, len(users))
	for i, user := range users {
		userProtos[i] = user.ToProto()
	}

	return connect.NewResponse(&v1.ListUsersResponse{
		Users:    userProtos,
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}), nil
}

func (u *UserService) findUserByIDWithContext(ctx context.Context, id string) (*models.User, error) {
	if id == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	var user models.User
	if err := u.DB.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user with ID %s not found", id))
		}
		return nil, err
	}

	return &user, nil
}
