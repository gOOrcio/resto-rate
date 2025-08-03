package services

import (
	v1 "api/src/generated/users/v1"
	"api/src/generated/users/v1/v1connect"
	"api/src/services/models"
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

func (u *UserService) CreateUser(
	ctx context.Context,
	req *connect.Request[v1.CreateUserRequest],
) (*connect.Response[v1.CreateUserResponse], error) {
	return u.createUser(ctx, req, false)
}

func (u *UserService) CreateAdminUser(
	ctx context.Context,
	req *connect.Request[v1.CreateUserRequest],
) (*connect.Response[v1.CreateUserResponse], error) {
	return u.createUser(ctx, req, true)
}

func (u *UserService) GetUser(
	ctx context.Context,
	req *connect.Request[v1.GetUserRequest],
) (*connect.Response[v1.GetUserResponse], error) {
	if req.Msg.Id == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	user, err := u.findUserByIDWithContext(ctx, req.Msg.Id)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	res := connect.NewResponse(&v1.GetUserResponse{
		User: user.ToProto(),
	})
	return res, nil
}

func (u *UserService) UpdateUser(
	ctx context.Context,
	req *connect.Request[v1.UpdateUserRequest],
) (*connect.Response[v1.UpdateUserResponse], error) {
	user, err := u.findUserByIDWithContext(ctx, req.Msg.Id)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	updates := map[string]interface{}{
		"Name":     req.Msg.Name,
		"Email":    req.Msg.Email,
		"Username": req.Msg.Username,
	}

	if err := u.DB.WithContext(ctx).Model(user).Updates(updates).Error; err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	res := connect.NewResponse(&v1.UpdateUserResponse{
		User: user.ToProto(),
	})
	return res, nil
}

func (u *UserService) DeleteUser(
	ctx context.Context,
	req *connect.Request[v1.DeleteUserRequest],
) (*connect.Response[v1.DeleteUserResponse], error) {
	user, err := u.findUserByIDWithContext(ctx, req.Msg.Id)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	if err := u.DB.WithContext(ctx).Delete(user).Error; err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	res := connect.NewResponse(&v1.DeleteUserResponse{
		Success: true,
	})
	return res, nil
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

	res := connect.NewResponse(&v1.ListUsersResponse{
		Users:    userProtos,
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	})
	return res, nil
}

func (u *UserService) findUserByIDWithContext(ctx context.Context, id string) (*models.User, error) {
	if id == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	var user models.User
	if err := u.DB.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with ID %s not found", id)
		}
		return nil, err
	}

	return &user, nil
}

func (u *UserService) createUser(
	ctx context.Context,
	req *connect.Request[v1.CreateUserRequest],
	isAdmin bool,
) (*connect.Response[v1.CreateUserResponse], error) {
	if req.Msg.Email == "" {
		return nil, fmt.Errorf("email is required")
	}

	if req.Msg.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	if req.Msg.Username == "" {
		return nil, fmt.Errorf("username is required")
	}

	user := &models.User{
		GoogleId: req.Msg.GoogleId,
		Email:    req.Msg.Email,
		Username: req.Msg.Username,
		Name:     req.Msg.Name,
		IsAdmin:  isAdmin,
	}

	if err := u.DB.WithContext(ctx).Create(user).Error; err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	res := connect.NewResponse(&v1.CreateUserResponse{
		User: user.ToProto(),
	})
	return res, nil
}
