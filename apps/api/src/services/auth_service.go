package services

import (
	authv1 "api/src/generated/auth/v1"
	"api/src/generated/auth/v1/v1connect"
	"api/src/internal/models"
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/valkey-io/valkey-go"
	"google.golang.org/api/idtoken"
	"gorm.io/gorm"

	"connectrpc.com/connect"
)

// ProviderClaims holds the normalized identity from any OAuth provider.
// ProviderID is the stable unique identifier (Google: "sub").
// Name may be empty for Apple Sign-In on repeat logins.
type ProviderClaims struct {
	ProviderID string
	Email      string
	Name       string
}

type AuthService struct {
	v1connect.UnimplementedAuthServiceHandler
	DB             *gorm.DB
	Valkey         valkey.Client
	GoogleClientID string
	SecureCookie   bool
}

func NewAuthService(db *gorm.DB, kv valkey.Client, googleClientID string, secureCookie bool) *AuthService {
	return &AuthService{DB: db, Valkey: kv, GoogleClientID: googleClientID, SecureCookie: secureCookie}
}

func (s *AuthService) sessionCookie(name, value string, maxAge int) string {
	cookie := name + "=" + value + "; HttpOnly; Path=/; Max-Age=" + strconv.Itoa(maxAge) + "; SameSite=Lax"
	if s.SecureCookie {
		cookie += "; Secure"
	}
	return cookie
}

func (s *AuthService) Login(
	ctx context.Context,
	req *connect.Request[authv1.LoginRequest],
) (*connect.Response[authv1.LoginResponse], error) {
	if req.Msg.Provider == authv1.AuthProvider_AUTH_PROVIDER_UNSPECIFIED {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("provider is required"))
	}
	if req.Msg.IdToken == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("id_token is required"))
	}

	claims, err := verifyIDToken(ctx, req.Msg.Provider, req.Msg.IdToken, s.GoogleClientID)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	user, err := s.upsertUser(ctx, req.Msg.Provider, claims)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	token := uuid.New().String()
	setCmd := s.Valkey.B().Set().Key("session:"+token).Value(user.ID).Ex(24 * time.Hour).Build()
	if err := s.Valkey.Do(ctx, setCmd).Error(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res := connect.NewResponse(&authv1.LoginResponse{
		User: user.ToProto(),
	})
	res.Header().Set("Set-Cookie", s.sessionCookie("session_token", token, 86400))
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
	res.Header().Set("Set-Cookie", s.sessionCookie("session_token", "", -1))
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

	return connect.NewResponse(&authv1.GetCurrentUserResponse{User: user.ToProto()}), nil
}

// verifyIDToken verifies a provider-issued JWT and returns normalized claims.
// To add Apple: implement AUTH_PROVIDER_APPLE case below.
func verifyIDToken(ctx context.Context, provider authv1.AuthProvider, token, clientID string) (ProviderClaims, error) {
	switch provider {
	case authv1.AuthProvider_AUTH_PROVIDER_GOOGLE:
		payload, err := idtoken.Validate(ctx, token, clientID)
		if err != nil {
			return ProviderClaims{}, err
		}
		name, _ := payload.Claims["name"].(string)
		email, _ := payload.Claims["email"].(string)
		return ProviderClaims{
			ProviderID: payload.Subject,
			Email:      email,
			Name:       name,
		}, nil
	// case authv1.AuthProvider_AUTH_PROVIDER_APPLE:
	//     TODO: implement Apple Sign-In verification
	default:
		return ProviderClaims{}, errors.New("unsupported auth provider")
	}
}

// upsertUser finds user by provider ID or creates a new one, keeping email/name up to date.
func (s *AuthService) upsertUser(ctx context.Context, provider authv1.AuthProvider, claims ProviderClaims) (*models.User, error) {
	var user models.User

	switch provider {
	case authv1.AuthProvider_AUTH_PROVIDER_GOOGLE:
		err := s.DB.WithContext(ctx).Where("google_id = ?", claims.ProviderID).First(&user).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// New user — create
			user = models.User{
				GoogleId: models.StringPtr(claims.ProviderID),
				Email:    models.StringPtr(claims.Email),
				Name:     claims.Name,
				Username: nil, // optional; set later via profile
			}
			if createErr := s.DB.WithContext(ctx).Create(&user).Error; createErr != nil {
				return nil, createErr
			}
		} else {
			// Existing user — update email and name in case they changed in Google
			if err := s.DB.WithContext(ctx).Model(&user).Updates(map[string]interface{}{
				"Email": models.StringPtr(claims.Email),
				"Name":  claims.Name,
			}).Error; err != nil {
				return nil, err
			}
			// Refresh in-memory struct so LoginResponse reflects the updated values
			user.Email = models.StringPtr(claims.Email)
			user.Name = claims.Name
		}
	default:
		return nil, errors.New("unsupported provider for upsert")
	}

	return &user, nil
}

func sessionToken(h http.Header) string {
	r := &http.Request{Header: h}
	if c, err := r.Cookie("session_token"); err == nil {
		return c.Value
	}
	return ""
}
