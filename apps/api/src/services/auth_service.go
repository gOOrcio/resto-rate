package services

import (
	authv1 "api/src/generated/auth/v1"
	"api/src/generated/auth/v1/v1connect"
	"api/src/internal/models"
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/valkey-io/valkey-go"
	"gorm.io/gorm"

	"connectrpc.com/connect"
)

type AuthService struct {
	v1connect.UnimplementedAuthServiceHandler
	DB     *gorm.DB
	Valkey valkey.Client
}

func NewAuthService(db *gorm.DB, kv valkey.Client) *AuthService {
	return &AuthService{DB: db, Valkey: kv}
}

func (s *AuthService) Login(
	ctx context.Context,
	req *connect.Request[authv1.LoginRequest],
) (*connect.Response[authv1.LoginResponse], error) {
	if req.Msg.Username == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("username is required"))
	}

	var user models.User
	err := s.DB.WithContext(ctx).Where("username = ?", req.Msg.Username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user = models.User{
				Username: req.Msg.Username,
				Name:     req.Msg.Username,
			}
			if createErr := s.DB.WithContext(ctx).Create(&user).Error; createErr != nil {
				return nil, connect.NewError(connect.CodeInternal, createErr)
			}
		} else {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	token := uuid.New().String()
	setCmd := s.Valkey.B().Set().Key("session:"+token).Value(user.ID).Ex(24 * time.Hour).Build()
	if err := s.Valkey.Do(ctx, setCmd).Error(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res := connect.NewResponse(&authv1.LoginResponse{
		User: user.ToProto(),
	})
	res.Header().Set("Set-Cookie", "session_token="+token+"; HttpOnly; Path=/; Max-Age=86400; SameSite=Lax")
	return res, nil
}

func (s *AuthService) Logout(
	ctx context.Context,
	req *connect.Request[authv1.LogoutRequest],
) (*connect.Response[authv1.LogoutResponse], error) {
	token := sessionToken(req.Header())
	if token != "" {
		s.Valkey.Do(ctx, s.Valkey.B().Del().Key("session:"+token).Build())
	}

	res := connect.NewResponse(&authv1.LogoutResponse{Success: true})
	res.Header().Set("Set-Cookie", "session_token=; HttpOnly; Path=/; Max-Age=-1; SameSite=Lax")
	return res, nil
}

func (s *AuthService) GetCurrentUser(
	ctx context.Context,
	req *connect.Request[authv1.GetCurrentUserRequest],
) (*connect.Response[authv1.GetCurrentUserResponse], error) {
	token := sessionToken(req.Header())
	if token == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("no session"))
	}

	result := s.Valkey.Do(ctx, s.Valkey.B().Get().Key("session:"+token).Build())
	if result.Error() != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("session expired or invalid"))
	}

	userID, err := result.ToString()
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid session data"))
	}

	var user models.User
	if err := s.DB.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("user not found"))
	}

	res := connect.NewResponse(&authv1.GetCurrentUserResponse{
		User: user.ToProto(),
	})
	return res, nil
}

func sessionToken(h http.Header) string {
	r := &http.Request{Header: h}
	if c, err := r.Cookie("session_token"); err == nil {
		return c.Value
	}
	return ""
}
