package http

import (
	"buck_It_Up/internal/models"
	"context"
)

type contextKey string

const authContextKey contextKey = "auth"

// AuthContext holds the authenticated access key information
type AuthContext struct {
	AccessKey *models.AccessKey
	BucketID  int64
	Role      models.AccessKeyRole
}

// SetAuthContext stores the auth context in the request context
func SetAuthContext(ctx context.Context, authCtx *AuthContext) context.Context {
	return context.WithValue(ctx, authContextKey, authCtx)
}

// GetAuthContext retrieves the auth context from the request context
func GetAuthContext(ctx context.Context) (*AuthContext, bool) {
	authCtx, ok := ctx.Value(authContextKey).(*AuthContext)
	return authCtx, ok
}
