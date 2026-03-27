package test

import (
	friendshipv1 "api/src/generated/friendship/v1"
	"api/src/services"
	"context"
	"testing"

	"connectrpc.com/connect"
)

func TestFriendshipService_SendFriendRequest_NilDB(t *testing.T) {
	svc := &services.FriendshipService{}
	req := connect.NewRequest(&friendshipv1.SendFriendRequestRequest{
		Receiver: &friendshipv1.SendFriendRequestRequest_ReceiverEmail{ReceiverEmail: "friend@example.com"},
	})
	_, err := svc.SendFriendRequest(context.Background(), req)
	if err == nil {
		t.Fatal("expected error from nil DB, got nil")
	}
}

func TestFriendshipService_AcceptFriendRequest_NilDB(t *testing.T) {
	svc := &services.FriendshipService{}
	req := connect.NewRequest(&friendshipv1.AcceptFriendRequestRequest{
		RequestId: "some-id",
	})
	_, err := svc.AcceptFriendRequest(context.Background(), req)
	if err == nil {
		t.Fatal("expected error from nil DB, got nil")
	}
}

func TestFriendshipService_DeclineFriendRequest_NilDB(t *testing.T) {
	svc := &services.FriendshipService{}
	req := connect.NewRequest(&friendshipv1.DeclineFriendRequestRequest{
		RequestId: "some-id",
	})
	_, err := svc.DeclineFriendRequest(context.Background(), req)
	if err == nil {
		t.Fatal("expected error from nil DB, got nil")
	}
}

func TestFriendshipService_RemoveFriend_NilDB(t *testing.T) {
	svc := &services.FriendshipService{}
	req := connect.NewRequest(&friendshipv1.RemoveFriendRequest{
		FriendUserId: "some-user-id",
	})
	_, err := svc.RemoveFriend(context.Background(), req)
	if err == nil {
		t.Fatal("expected error from nil DB, got nil")
	}
}

func TestFriendshipService_ListFriends_NilDB(t *testing.T) {
	svc := &services.FriendshipService{}
	req := connect.NewRequest(&friendshipv1.ListFriendsRequest{})
	_, err := svc.ListFriends(context.Background(), req)
	if err == nil {
		t.Fatal("expected error from nil DB, got nil")
	}
}

func TestFriendshipService_ListPendingRequests_NilDB(t *testing.T) {
	svc := &services.FriendshipService{}
	req := connect.NewRequest(&friendshipv1.ListPendingRequestsRequest{})
	_, err := svc.ListPendingRequests(context.Background(), req)
	if err == nil {
		t.Fatal("expected error from nil DB, got nil")
	}
}

func TestFriendshipService_FindUserByHandle_NilDB(t *testing.T) {
	svc := &services.FriendshipService{}
	req := connect.NewRequest(&friendshipv1.FindUserByHandleRequest{Username: "alice"})
	_, err := svc.FindUserByHandle(context.Background(), req)
	if err == nil {
		t.Fatal("expected error from nil DB, got nil")
	}
}

func TestFriendshipService_FindUserByHandle_EmptyUsername(t *testing.T) {
	svc := &services.FriendshipService{}
	req := connect.NewRequest(&friendshipv1.FindUserByHandleRequest{Username: ""})
	_, err := svc.FindUserByHandle(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for empty username, got nil")
	}
	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Fatalf("expected CodeInvalidArgument, got %v", connectErr.Code())
	}
}

func TestFriendshipService_FindUserByHandle_InvalidUsername(t *testing.T) {
	svc := &services.FriendshipService{}
	req := connect.NewRequest(&friendshipv1.FindUserByHandleRequest{Username: "!!bad!!"})
	_, err := svc.FindUserByHandle(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for invalid username, got nil")
	}
	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Fatalf("expected CodeInvalidArgument, got %v", connectErr.Code())
	}
}

func TestFriendshipService_SendFriendRequest_UsernameNilDB(t *testing.T) {
	// Validates that the username branch is wired — auth check fires first (nil Valkey), so any error is expected.
	svc := &services.FriendshipService{}
	req := connect.NewRequest(&friendshipv1.SendFriendRequestRequest{
		Receiver: &friendshipv1.SendFriendRequestRequest_ReceiverUsername{ReceiverUsername: "alice"},
	})
	_, err := svc.SendFriendRequest(context.Background(), req)
	if err == nil {
		t.Fatal("expected error from nil Valkey/DB, got nil")
	}
}
