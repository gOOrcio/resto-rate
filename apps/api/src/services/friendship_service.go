package services

import (
	v1 "api/src/generated/friendship/v1"
	"api/src/generated/friendship/v1/v1connect"
	"api/src/internal/models"
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"
	"github.com/valkey-io/valkey-go"
	"gorm.io/gorm"
)

type FriendshipService struct {
	v1connect.UnimplementedFriendshipServiceHandler
	DB     *gorm.DB
	Valkey valkey.Client
}

func NewFriendshipService(db *gorm.DB, kv valkey.Client) *FriendshipService {
	return &FriendshipService{DB: db, Valkey: kv}
}

func (s *FriendshipService) SendFriendRequest(
	ctx context.Context,
	req *connect.Request[v1.SendFriendRequestRequest],
) (*connect.Response[v1.SendFriendRequestResponse], error) {
	senderID, err := getUserIDFromSession(ctx, req.Header(), s.Valkey)
	if err != nil {
		return nil, err
	}

	var receiver models.User
	switch {
	case req.Msg.GetReceiverEmail() != "":
		if err := s.DB.WithContext(ctx).Where("email = ?", req.Msg.GetReceiverEmail()).First(&receiver).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, connect.NewError(connect.CodeNotFound, errors.New("user not found"))
			}
			return nil, err
		}
	case req.Msg.GetReceiverUsername() != "":
		handle := strings.ToLower(strings.TrimPrefix(req.Msg.GetReceiverUsername(), "@"))
		if !isValidUsername(handle) {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid username"))
		}
		if err := s.DB.WithContext(ctx).Where("username = ?", handle).First(&receiver).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, connect.NewError(connect.CodeNotFound, errors.New("user not found"))
			}
			return nil, err
		}
	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("receiver_email or receiver_username is required"))
	}

	if receiver.ID == senderID {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("cannot send friend request to yourself"))
	}

	var sender models.User
	if err := s.DB.WithContext(ctx).First(&sender, "id = ?", senderID).Error; err != nil {
		return nil, err
	}

	pairKey := canonicalPairKey(senderID, receiver.ID)

	// Check for an existing request between this pair.
	var existing models.FriendRequest
	err = s.DB.WithContext(ctx).Where("pair_key = ?", pairKey).First(&existing).Error
	if err == nil {
		// If the previous request was declined, allow re-sending by resetting to pending.
		if existing.Status == models.FriendRequestStatusDeclined {
			existing.SenderID = senderID
			existing.ReceiverID = receiver.ID
			existing.Status = models.FriendRequestStatusPending
			existing.Sender = sender
			existing.Receiver = receiver
			if err := s.DB.WithContext(ctx).Save(&existing).Error; err != nil {
				return nil, err
			}
			return connect.NewResponse(&v1.SendFriendRequestResponse{Request: existing.ToProto()}), nil
		}
		return nil, connect.NewError(connect.CodeAlreadyExists, errors.New("friend request already exists"))
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	fr := models.FriendRequest{
		SenderID:   senderID,
		ReceiverID: receiver.ID,
		Status:     models.FriendRequestStatusPending,
		Sender:     sender,
		Receiver:   receiver,
	}
	if err := s.DB.WithContext(ctx).Create(&fr).Error; err != nil {
		return nil, err
	}

	return connect.NewResponse(&v1.SendFriendRequestResponse{Request: fr.ToProto()}), nil
}

func (s *FriendshipService) AcceptFriendRequest(
	ctx context.Context,
	req *connect.Request[v1.AcceptFriendRequestRequest],
) (*connect.Response[v1.AcceptFriendRequestResponse], error) {
	userID, err := getUserIDFromSession(ctx, req.Header(), s.Valkey)
	if err != nil {
		return nil, err
	}

	if req.Msg.RequestId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("request_id is required"))
	}

	var fr models.FriendRequest
	if err := s.DB.WithContext(ctx).Preload("Sender").Preload("Receiver").
		First(&fr, "id = ? AND receiver_id = ? AND status = ?", req.Msg.RequestId, userID, models.FriendRequestStatusPending).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("pending friend request not found"))
		}
		return nil, err
	}

	fr.Status = models.FriendRequestStatusAccepted
	if err := s.DB.WithContext(ctx).Save(&fr).Error; err != nil {
		return nil, err
	}

	return connect.NewResponse(&v1.AcceptFriendRequestResponse{Request: fr.ToProto()}), nil
}

func (s *FriendshipService) DeclineFriendRequest(
	ctx context.Context,
	req *connect.Request[v1.DeclineFriendRequestRequest],
) (*connect.Response[v1.DeclineFriendRequestResponse], error) {
	userID, err := getUserIDFromSession(ctx, req.Header(), s.Valkey)
	if err != nil {
		return nil, err
	}

	if req.Msg.RequestId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("request_id is required"))
	}

	var fr models.FriendRequest
	if err := s.DB.WithContext(ctx).
		First(&fr, "id = ? AND receiver_id = ? AND status = ?", req.Msg.RequestId, userID, models.FriendRequestStatusPending).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("pending friend request not found"))
		}
		return nil, err
	}

	fr.Status = models.FriendRequestStatusDeclined
	if err := s.DB.WithContext(ctx).Save(&fr).Error; err != nil {
		return nil, err
	}

	return connect.NewResponse(&v1.DeclineFriendRequestResponse{Success: true}), nil
}

func (s *FriendshipService) RemoveFriend(
	ctx context.Context,
	req *connect.Request[v1.RemoveFriendRequest],
) (*connect.Response[v1.RemoveFriendResponse], error) {
	userID, err := getUserIDFromSession(ctx, req.Header(), s.Valkey)
	if err != nil {
		return nil, err
	}

	if req.Msg.FriendUserId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("friend_user_id is required"))
	}

	result := s.DB.WithContext(ctx).Where(
		"((sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)) AND status = ?",
		userID, req.Msg.FriendUserId, req.Msg.FriendUserId, userID, models.FriendRequestStatusAccepted,
	).Delete(&models.FriendRequest{})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("friendship not found"))
	}

	return connect.NewResponse(&v1.RemoveFriendResponse{Success: true}), nil
}

func (s *FriendshipService) ListFriends(
	ctx context.Context,
	req *connect.Request[v1.ListFriendsRequest],
) (*connect.Response[v1.ListFriendsResponse], error) {
	userID, err := getUserIDFromSession(ctx, req.Header(), s.Valkey)
	if err != nil {
		return nil, err
	}

	var friendRequests []models.FriendRequest
	if err := s.DB.WithContext(ctx).Preload("Sender").Preload("Receiver").
		Where("(sender_id = ? OR receiver_id = ?) AND status = ?", userID, userID, models.FriendRequestStatusAccepted).
		Find(&friendRequests).Error; err != nil {
		return nil, err
	}

	friends := make([]*v1.FriendProto, len(friendRequests))
	for i, fr := range friendRequests {
		if fr.SenderID == userID {
			friends[i] = &v1.FriendProto{
				UserId:       fr.ReceiverID,
				Name:         fr.Receiver.Name,
				Email:        derefStr(fr.Receiver.Email),
				Username:     derefStr(fr.Receiver.Username),
				FriendsSince: fr.UpdatedAt.Unix(),
			}
		} else {
			friends[i] = &v1.FriendProto{
				UserId:       fr.SenderID,
				Name:         fr.Sender.Name,
				Email:        derefStr(fr.Sender.Email),
				Username:     derefStr(fr.Sender.Username),
				FriendsSince: fr.UpdatedAt.Unix(),
			}
		}
	}

	return connect.NewResponse(&v1.ListFriendsResponse{Friends: friends}), nil
}

func (s *FriendshipService) ListPendingRequests(
	ctx context.Context,
	req *connect.Request[v1.ListPendingRequestsRequest],
) (*connect.Response[v1.ListPendingRequestsResponse], error) {
	userID, err := getUserIDFromSession(ctx, req.Header(), s.Valkey)
	if err != nil {
		return nil, err
	}

	var friendRequests []models.FriendRequest
	if err := s.DB.WithContext(ctx).Preload("Sender").Preload("Receiver").
		Where("receiver_id = ? AND status = ?", userID, models.FriendRequestStatusPending).
		Find(&friendRequests).Error; err != nil {
		return nil, err
	}

	protos := make([]*v1.FriendRequestProto, len(friendRequests))
	for i, fr := range friendRequests {
		protos[i] = fr.ToProto()
	}

	return connect.NewResponse(&v1.ListPendingRequestsResponse{Requests: protos}), nil
}

func (s *FriendshipService) FindUserByHandle(
	ctx context.Context,
	req *connect.Request[v1.FindUserByHandleRequest],
) (*connect.Response[v1.FindUserByHandleResponse], error) {
	// Validate input before auth so callers get a clear error without needing a valid session.
	handle := strings.ToLower(strings.TrimPrefix(req.Msg.Username, "@"))
	if handle == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("username is required"))
	}
	if !isValidUsername(handle) {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid username"))
	}

	if _, err := getUserIDFromSession(ctx, req.Header(), s.Valkey); err != nil {
		return nil, err
	}

	var user models.User
	if err := s.DB.WithContext(ctx).Where("username = ?", handle).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("user not found"))
		}
		return nil, err
	}

	return connect.NewResponse(&v1.FindUserByHandleResponse{
		Id:       user.ID,
		Username: derefStr(user.Username),
		Name:     user.Name,
	}), nil
}

// canonicalPairKey returns a deterministic, unordered key for a user pair
// so that (A,B) and (B,A) produce the same key.
func canonicalPairKey(a, b string) string {
	if a < b {
		return a + ":" + b
	}
	return b + ":" + a
}

// derefStr safely dereferences a string pointer, returning "" for nil.
func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// getFriendIDs returns the user IDs of all accepted friends for the given user.
func getFriendIDs(ctx context.Context, db *gorm.DB, userID string) ([]string, error) {
	var friendRequests []models.FriendRequest
	if err := db.WithContext(ctx).Select("sender_id, receiver_id").
		Where("(sender_id = ? OR receiver_id = ?) AND status = ?", userID, userID, models.FriendRequestStatusAccepted).
		Find(&friendRequests).Error; err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(friendRequests))
	for _, fr := range friendRequests {
		if fr.SenderID == userID {
			ids = append(ids, fr.ReceiverID)
		} else {
			ids = append(ids, fr.SenderID)
		}
	}
	return ids, nil
}
