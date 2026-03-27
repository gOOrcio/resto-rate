package services

import (
	authv1 "api/src/generated/auth/v1"
	"api/src/generated/auth/v1/v1connect"
	"api/src/internal/models"
	"context"
	"errors"
	"fmt"
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
	if err := s.issueSession(ctx, user.ID, token); err != nil {
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
		// Get user ID before deleting session so we can update the sessions set
		result := s.Valkey.Do(ctx, s.Valkey.B().Get().Key("session:"+token).Build())
		if userID, err := result.ToString(); err == nil {
			s.Valkey.Do(ctx, s.Valkey.B().Srem().Key("user_sessions:"+userID).Member(token).Build())
		}
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

func (s *AuthService) UpdateMyProfile(
	ctx context.Context,
	req *connect.Request[authv1.UpdateMyProfileRequest],
) (*connect.Response[authv1.UpdateMyProfileResponse], error) {
	userID, err := getUserIDFromSession(ctx, req.Header(), s.Valkey)
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := s.DB.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("user not found"))
	}

	updates := map[string]interface{}{}

	if req.Msg.Username != "" {
		// Validate username format: 3–30 chars, lowercase letters/digits/underscores
		if !isValidUsername(req.Msg.Username) {
			return nil, connect.NewError(connect.CodeInvalidArgument,
				errors.New("username must be 3–30 characters and contain only lowercase letters, digits, and underscores"))
		}
		// Check uniqueness
		var existing models.User
		if err := s.DB.WithContext(ctx).Where("username = ? AND id != ?", req.Msg.Username, userID).First(&existing).Error; err == nil {
			return nil, connect.NewError(connect.CodeAlreadyExists,
				fmt.Errorf("username '%s' is already taken", req.Msg.Username))
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		updates["Username"] = models.StringPtr(req.Msg.Username)
	}

	if req.Msg.DefaultRegion != "" {
		updates["DefaultRegion"] = req.Msg.DefaultRegion
	}

	if req.Msg.SetIsDarkModeEnabled {
		updates["IsDarkModeEnabled"] = req.Msg.IsDarkModeEnabled
	}

	if len(updates) == 0 {
		return connect.NewResponse(&authv1.UpdateMyProfileResponse{User: user.ToProto()}), nil
	}

	if err := s.DB.WithContext(ctx).Model(&user).Updates(updates).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Re-fetch to get updated values
	if err := s.DB.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&authv1.UpdateMyProfileResponse{User: user.ToProto()}), nil
}

func (s *AuthService) GetMyStats(
	ctx context.Context,
	req *connect.Request[authv1.GetMyStatsRequest],
) (*connect.Response[authv1.GetMyStatsResponse], error) {
	userID, err := getUserIDFromSession(ctx, req.Header(), s.Valkey)
	if err != nil {
		return nil, err
	}

	var reviewCount, wishlistCount, friendCount int64

	if err := s.DB.WithContext(ctx).Table("reviews").Where("user_id = ?", userID).Count(&reviewCount).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if err := s.DB.WithContext(ctx).Table("wishlist_items").Where("user_id = ?", userID).Count(&wishlistCount).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if err := s.DB.WithContext(ctx).Table("friend_requests").
		Where("(sender_id = ? OR receiver_id = ?) AND status = 'accepted'", userID, userID).
		Count(&friendCount).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&authv1.GetMyStatsResponse{
		ReviewCount:   int32(reviewCount),
		WishlistCount: int32(wishlistCount),
		FriendCount:   int32(friendCount),
	}), nil
}

func (s *AuthService) DeleteMyAccount(
	ctx context.Context,
	req *connect.Request[authv1.DeleteMyAccountRequest],
) (*connect.Response[authv1.DeleteMyAccountResponse], error) {
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

	// Wipe all sessions for this user
	if err := s.wipeAllSessions(ctx, userID, token); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Delete the user and dependent rows in a transaction (no DB-level cascade on these FKs).
	if err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&models.Review{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&models.WishlistItem{}).Error; err != nil {
			return err
		}
		if err := tx.Where("sender_id = ? OR receiver_id = ?", userID, userID).Delete(&models.FriendRequest{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.User{}, "id = ?", userID).Error
	}); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res := connect.NewResponse(&authv1.DeleteMyAccountResponse{Success: true})
	res.Header().Set("Set-Cookie", s.sessionCookie("session_token", "", -1))
	return res, nil
}

func (s *AuthService) SignOutAllDevices(
	ctx context.Context,
	req *connect.Request[authv1.SignOutAllDevicesRequest],
) (*connect.Response[authv1.SignOutAllDevicesResponse], error) {
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

	if err := s.wipeAllSessions(ctx, userID, token); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res := connect.NewResponse(&authv1.SignOutAllDevicesResponse{Success: true})
	res.Header().Set("Set-Cookie", s.sessionCookie("session_token", "", -1))
	return res, nil
}

// issueSession creates a session token in Valkey and tracks it in the user's sessions set.
func (s *AuthService) issueSession(ctx context.Context, userID, token string) error {
	setCmd := s.Valkey.B().Set().Key("session:"+token).Value(userID).Ex(24 * time.Hour).Build()
	if err := s.Valkey.Do(ctx, setCmd).Error(); err != nil {
		return err
	}
	// Track in the user's session set (for sign-out-all). Return error so callers know sessions won't be wipeable.
	saddCmd := s.Valkey.B().Sadd().Key("user_sessions:"+userID).Member(token).Build()
	if err := s.Valkey.Do(ctx, saddCmd).Error(); err != nil {
		return err
	}
	// Refresh TTL on the tracking set to prevent unbounded growth from expired individual tokens.
	const sessionTTL = int64(24 * 60 * 60) // 24 h in seconds
	expireCmd := s.Valkey.B().Expire().Key("user_sessions:"+userID).Seconds(sessionTTL).Build()
	if err := s.Valkey.Do(ctx, expireCmd).Error(); err != nil {
		return fmt.Errorf("issueSession: EXPIRE user_sessions:%s: %w", userID, err)
	}
	return nil
}

// wipeAllSessions deletes every tracked session for the given user from Valkey.
// currentToken is always deleted regardless of whether the tracking set exists,
// ensuring the caller's own session is invalidated even for pre-tracking users.
func (s *AuthService) wipeAllSessions(ctx context.Context, userID, currentToken string) error {
	smembersResult := s.Valkey.Do(ctx, s.Valkey.B().Smembers().Key("user_sessions:"+userID).Build())
	tracked, _ := smembersResult.AsStrSlice()

	// Deduplicate: always include the current token.
	toDelete := make(map[string]struct{}, len(tracked)+1)
	toDelete[currentToken] = struct{}{}
	for _, t := range tracked {
		toDelete[t] = struct{}{}
	}

	for t := range toDelete {
		if err := s.Valkey.Do(ctx, s.Valkey.B().Del().Key("session:"+t).Build()).Error(); err != nil {
			return fmt.Errorf("wipeAllSessions: DEL session:%s: %w", t, err)
		}
	}
	if err := s.Valkey.Do(ctx, s.Valkey.B().Del().Key("user_sessions:"+userID).Build()).Error(); err != nil {
		return fmt.Errorf("wipeAllSessions: DEL user_sessions:%s: %w", userID, err)
	}
	return nil
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

// isValidUsername checks the username regex: 3–30 lowercase letters, digits, underscores.
func isValidUsername(s string) bool {
	if len(s) < 3 || len(s) > 30 {
		return false
	}
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}
