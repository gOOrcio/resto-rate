package services

import (
	"context"
	"errors"
	"fmt"
	userpb "github.com/goorcio/resto-rate-protos/generated/go/resto-rate/generated/go/users/v1"
	"go-app/src/services/models"
	"go-app/src/services/utils"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type UserService struct {
	userpb.UnimplementedUsersServiceServer
	DB *gorm.DB
}

func (u *UserService) CreateUser(ctx context.Context, request *userpb.CreateUserRequest) (*userpb.UserProto, error) {
	return u.createUser(ctx, request, false)
}

func (u *UserService) CreateAdminUser(ctx context.Context, request *userpb.CreateUserRequest) (*userpb.UserProto, error) {
	return u.createUser(ctx, request, true)
}

func (u *UserService) GetUser(ctx context.Context, request *userpb.GetUserRequest) (*userpb.UserProto, error) {
	user, err := u.findUserByIDWithContext(ctx, request.Id)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	return user.ToProto(), nil
}

func (u *UserService) UpdateUser(ctx context.Context, request *userpb.UpdateUserRequest) (*userpb.UserProto, error) {
	user, err := u.findUserByIDWithContext(ctx, request.User.Id)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	updates := map[string]interface{}{
		"Name":  request.User.Name,
		"Email": request.User.Email,
	}

	if err := u.DB.WithContext(ctx).Model(user).Updates(updates).Error; err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	return user.ToProto(), nil
}

func (u *UserService) DeleteUser(ctx context.Context, request *userpb.DeleteUserRequest) (*emptypb.Empty, error) {
	user, err := u.findUserByIDWithContext(ctx, request.Id)
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

	return &emptypb.Empty{}, nil
}

func (u *UserService) ListUsers(ctx context.Context, request *userpb.ListUsersRequest) (*userpb.ListUsersResponse, error) {
	var users []*models.User
	totalCount, err := utils.Paginate(u.DB.WithContext(ctx), &users, request.Page, request.PageSize)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	var userProtos []*userpb.UserProto
	for _, user := range users {
		userProtos = append(userProtos, user.ToProto())
	}

	var nextPageToken string
	if int64((request.Page-1)*request.PageSize+int32(len(users))) < totalCount {
		nextPageToken = fmt.Sprintf("%d", request.Page+1)
	}

	return &userpb.ListUsersResponse{
		Users:         userProtos,
		TotalCount:    int32(totalCount),
		Page:          request.Page,
		PageSize:      request.PageSize,
		NextPageToken: nextPageToken,
	}, nil
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

func (u *UserService) createUser(ctx context.Context, request *userpb.CreateUserRequest, isAdmin bool) (*userpb.UserProto, error) {
	user := &models.User{
		GoogleId: request.User.GoogleId,
		Email:    request.User.Email,
		Name:     request.User.Name,
		IsAdmin:  isAdmin,
	}

	if err := u.DB.WithContext(ctx).Create(user).Error; err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}

	return user.ToProto(), nil
}
