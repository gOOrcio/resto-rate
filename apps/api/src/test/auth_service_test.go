package test

import (
	authv1 "api/src/generated/auth/v1"
	"api/src/services"
	"context"
	"testing"

	"connectrpc.com/connect"
)

// TestUpdateMyProfile_NoSession verifies that UpdateMyProfile rejects requests with no session cookie.
func TestUpdateMyProfile_NoSession(t *testing.T) {
	svc := &services.AuthService{} // nil Valkey — fails before hitting it
	req := connect.NewRequest(&authv1.UpdateMyProfileRequest{Username: "alice"})
	_, err := svc.UpdateMyProfile(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for missing session, got nil")
	}
	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeUnauthenticated {
		t.Fatalf("expected CodeUnauthenticated, got %v", connectErr.Code())
	}
}

// TestGetMyStats_NoSession verifies that GetMyStats rejects requests with no session cookie.
func TestGetMyStats_NoSession(t *testing.T) {
	svc := &services.AuthService{}
	req := connect.NewRequest(&authv1.GetMyStatsRequest{})
	_, err := svc.GetMyStats(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for missing session, got nil")
	}
	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeUnauthenticated {
		t.Fatalf("expected CodeUnauthenticated, got %v", connectErr.Code())
	}
}

// TestDeleteMyAccount_NoSession verifies that DeleteMyAccount rejects requests with no session cookie.
func TestDeleteMyAccount_NoSession(t *testing.T) {
	svc := &services.AuthService{}
	req := connect.NewRequest(&authv1.DeleteMyAccountRequest{})
	_, err := svc.DeleteMyAccount(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for missing session, got nil")
	}
	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeUnauthenticated {
		t.Fatalf("expected CodeUnauthenticated, got %v", connectErr.Code())
	}
}

// TestSignOutAllDevices_NoSession verifies that SignOutAllDevices rejects requests with no session cookie.
func TestSignOutAllDevices_NoSession(t *testing.T) {
	svc := &services.AuthService{}
	req := connect.NewRequest(&authv1.SignOutAllDevicesRequest{})
	_, err := svc.SignOutAllDevices(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for missing session, got nil")
	}
	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}
	if connectErr.Code() != connect.CodeUnauthenticated {
		t.Fatalf("expected CodeUnauthenticated, got %v", connectErr.Code())
	}
}

// TestLogin_UnspecifiedProvider verifies that Login rejects AUTH_PROVIDER_UNSPECIFIED.
// This test is pure in-process — no DB or network needed.
func TestLogin_UnspecifiedProvider(t *testing.T) {
	svc := &services.AuthService{} // nil DB/Valkey — should fail before hitting them

	req := connect.NewRequest(&authv1.LoginRequest{
		Provider: authv1.AuthProvider_AUTH_PROVIDER_UNSPECIFIED,
		IdToken:  "some-token",
	})

	_, err := svc.Login(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for UNSPECIFIED provider, got nil")
	}

	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T: %v", err, err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Fatalf("expected CodeInvalidArgument, got %v", connectErr.Code())
	}
}

// TestLogin_EmptyToken verifies that Login rejects an empty id_token.
func TestLogin_EmptyToken(t *testing.T) {
	svc := &services.AuthService{}

	req := connect.NewRequest(&authv1.LoginRequest{
		Provider: authv1.AuthProvider_AUTH_PROVIDER_GOOGLE,
		IdToken:  "",
	})

	_, err := svc.Login(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for empty token, got nil")
	}

	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T: %v", err, err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Fatalf("expected CodeInvalidArgument, got %v", connectErr.Code())
	}
}
