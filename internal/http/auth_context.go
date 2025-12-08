package http

import (
	"buck_It_Up/internal/models"
	"context"
)

type contextKey string

const authContextKey contextKey = "auth"

type AuthContext struct {
	KeyID    string               `json:"key_id"`
	BucketID int64                `json:"bucket_id"`
	Role     models.AccessKeyRole `json:"role"`
}

func SetAuthContext(ctx context.Context, authCtx *AuthContext) context.Context {
	return context.WithValue(ctx, authContextKey, authCtx)
}

func GetAuthContext(ctx context.Context) (*AuthContext, bool) {
	authCtx, ok := ctx.Value(authContextKey).(*AuthContext)
	return authCtx, ok
}
