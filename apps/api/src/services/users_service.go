package services

import (
	v1 "api/src/generated/users/v1"
	"api/src/generated/users/v1/v1connect"
	"api/src/services/models"
	"api/src/services/utils"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserService struct {
	v1connect.UnimplementedUsersServiceHandler
	DB *gorm.DB
}

func (u *UserService) CreateUser(
	ctx context.Context,
	req *connect.Request[v1.CreateUserRequest],
) (*connect.Response[v1.UserProto], error) {
	return u.createUser(ctx, req, false)
}

func (u *UserService) CreateAdminUser(
	ctx context.Context,
	req *connect.Request[v1.CreateUserRequest],
) (*connect.Response[v1.UserProto], error) {
	return u.createUser(ctx, req, true)
}

func (u *UserService) GetUser(
	ctx context.Context,
	req *connect.Request[v1.GetUserRequest],
) (*connect.Response[v1.UserProto], error) {
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

	res := connect.NewResponse(user.ToProto())
	return res, nil
}

func (u *UserService) UpdateUser(
	ctx context.Context,
	req *connect.Request[v1.UpdateUserRequest],
) (*connect.Response[v1.UserProto], error) {
	user, err := u.findUserByIDWithContext(ctx, req.Msg.User.Id)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	updates := map[string]interface{}{
		"Name":  req.Msg.User.Name,
		"Email": req.Msg.User.Email,
	}

	if err := u.DB.WithContext(ctx).Model(user).Updates(updates).Error; err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	res := connect.NewResponse(user.ToProto())
	return res, nil
}

func (u *UserService) DeleteUser(
	ctx context.Context,
	req *connect.Request[v1.DeleteUserRequest],
) (*connect.Response[emptypb.Empty], error) {
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

	res := connect.NewResponse(&emptypb.Empty{})
	return res, nil
}

func (u *UserService) ListUsers(
	ctx context.Context,
	req *connect.Request[v1.ListUsersRequest],
) (*connect.Response[v1.ListUsersResponse], error) {
	var users []*models.User
	totalCount, err := utils.Paginate(u.DB.WithContext(ctx), &users, req.Msg.Page, req.Msg.PageSize)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	var userProtos []*v1.UserProto
	for _, user := range users {
		userProtos = append(userProtos, user.ToProto())
	}

	nextPageToken := utils.CalculateNextPageToken(req.Msg.Page, req.Msg.PageSize, len(users), totalCount)

	response := &v1.ListUsersResponse{
		Users:         userProtos,
		TotalCount:    int32(totalCount),
		Page:          req.Msg.Page,
		PageSize:      req.Msg.PageSize,
		NextPageToken: nextPageToken,
	}

	res := connect.NewResponse(response)
	return res, nil
}

func (u *UserService) findUserByIDWithContext(ctx context.Context, id string) (*models.User, error) {
	if id == "" {
		return nil, fmt.Errorf("invalid user id")
	}

	var user models.User
	if err := u.DB.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
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
) (*connect.Response[v1.UserProto], error) {
	user := &models.User{
		GoogleId: req.Msg.User.GoogleId,
		Email:    req.Msg.User.Email,
		Name:     req.Msg.User.Name,
		IsAdmin:  isAdmin,
	}

	if err := u.DB.WithContext(ctx).Create(user).Error; err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	res := connect.NewResponse(user.ToProto())
	return res, nil
}
