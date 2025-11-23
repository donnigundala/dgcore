package ctxutil

import "context"

const (
	userKey   contextKey = "user"
	userIDKey contextKey = "user_id"
)

// WithUser stores a user in the context.
func WithUser(ctx context.Context, user interface{}) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// UserFromContext retrieves the user from the context.
func UserFromContext(ctx context.Context) interface{} {
	return ctx.Value(userKey)
}

// WithUserID stores a user ID in the context.
func WithUserID(ctx context.Context, userID interface{}) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFromContext retrieves the user ID from the context.
func UserIDFromContext(ctx context.Context) interface{} {
	return ctx.Value(userIDKey)
}
