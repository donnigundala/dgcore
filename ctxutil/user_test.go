package ctxutil

import (
	"context"
	"testing"
)

// TestWithUser_Storage tests storing user in context
func TestWithUser_Storage(t *testing.T) {
	ctx := context.Background()
	user := map[string]interface{}{
		"id":   123,
		"name": "John Doe",
	}

	ctx = WithUser(ctx, user)

	retrieved := UserFromContext(ctx)
	if retrieved == nil {
		t.Fatal("expected user, got nil")
	}

	userMap, ok := retrieved.(map[string]interface{})
	if !ok {
		t.Fatal("expected user to be map[string]interface{}")
	}

	if userMap["id"] != 123 {
		t.Errorf("expected id=123, got %v", userMap["id"])
	}

	if userMap["name"] != "John Doe" {
		t.Errorf("expected name=John Doe, got %v", userMap["name"])
	}
}

// TestUserFromContext_NotFound tests retrieving non-existent user
func TestUserFromContext_NotFound(t *testing.T) {
	ctx := context.Background()

	retrieved := UserFromContext(ctx)
	if retrieved != nil {
		t.Errorf("expected nil, got %v", retrieved)
	}
}

// TestWithUserID_Storage tests storing user ID in context
func TestWithUserID_Storage(t *testing.T) {
	ctx := context.Background()
	userID := 12345

	ctx = WithUserID(ctx, userID)

	retrieved := UserIDFromContext(ctx)
	if retrieved == nil {
		t.Fatal("expected user ID, got nil")
	}

	if retrieved != userID {
		t.Errorf("expected user ID %d, got %v", userID, retrieved)
	}
}

// TestWithUserID_StringID tests storing string user ID
func TestWithUserID_StringID(t *testing.T) {
	ctx := context.Background()
	userID := "user-abc-123"

	ctx = WithUserID(ctx, userID)

	retrieved := UserIDFromContext(ctx)
	if retrieved == nil {
		t.Fatal("expected user ID, got nil")
	}

	if retrieved != userID {
		t.Errorf("expected user ID %s, got %v", userID, retrieved)
	}
}

// TestUserIDFromContext_NotFound tests retrieving non-existent user ID
func TestUserIDFromContext_NotFound(t *testing.T) {
	ctx := context.Background()

	retrieved := UserIDFromContext(ctx)
	if retrieved != nil {
		t.Errorf("expected nil, got %v", retrieved)
	}
}

// TestUserContext_Combined tests storing both user and user ID
func TestUserContext_Combined(t *testing.T) {
	ctx := context.Background()

	user := map[string]string{
		"name":  "Jane Doe",
		"email": "jane@example.com",
	}
	userID := 456

	ctx = WithUser(ctx, user)
	ctx = WithUserID(ctx, userID)

	retrievedUser := UserFromContext(ctx)
	retrievedID := UserIDFromContext(ctx)

	if retrievedUser == nil {
		t.Error("expected user, got nil")
	}

	if retrievedID != userID {
		t.Errorf("expected user ID %d, got %v", userID, retrievedID)
	}
}

// TestUserContext_StructType tests storing struct as user
func TestUserContext_StructType(t *testing.T) {
	type User struct {
		ID    int
		Name  string
		Email string
	}

	ctx := context.Background()
	user := User{
		ID:    789,
		Name:  "Bob Smith",
		Email: "bob@example.com",
	}

	ctx = WithUser(ctx, user)

	retrieved := UserFromContext(ctx)
	if retrieved == nil {
		t.Fatal("expected user, got nil")
	}

	userStruct, ok := retrieved.(User)
	if !ok {
		t.Fatal("expected user to be User struct")
	}

	if userStruct.ID != 789 {
		t.Errorf("expected ID=789, got %d", userStruct.ID)
	}

	if userStruct.Name != "Bob Smith" {
		t.Errorf("expected name=Bob Smith, got %s", userStruct.Name)
	}
}
