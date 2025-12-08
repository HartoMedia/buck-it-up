package http

import (
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	nethttp "net/http"
	"os"
	"strings"

	"buck_It_Up/internal/models"

	"github.com/go-chi/chi/v5"
)

type AuthLevel int

const (
	AuthLevelNone AuthLevel = iota
	AuthLevelReadOnly
	AuthLevelReadWrite
	AuthLevelAll
)

func (r *Router) AuthMiddleware(level AuthLevel) func(nethttp.Handler) nethttp.Handler {
	return func(next nethttp.Handler) nethttp.Handler {
		return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, req *nethttp.Request) {
			if level == AuthLevelNone {
				next.ServeHTTP(w, req)
				return
			}

			authHeader := req.Header.Get("Authorization")
			if authHeader == "" {
				w.Header().Set("X-Auth-Error", "missing authorization header")
				nethttp.Error(w, "authorization required", nethttp.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				w.Header().Set("X-Auth-Error", "invalid format - expected 'Bearer <key_id>:<secret>'")
				nethttp.Error(w, "invalid authorization format", nethttp.StatusUnauthorized)
				return
			}

			credentials := strings.SplitN(parts[1], ":", 2)
			if len(credentials) != 2 {
				w.Header().Set("X-Auth-Error", "invalid credentials format - expected 'key_id:secret'")
				nethttp.Error(w, "invalid credentials format", nethttp.StatusUnauthorized)
				return
			}

			keyID := credentials[0]
			secret := credentials[1]

			if keyID == "admin" {
				adminPassword := os.Getenv("ADMIN_PASSWORD")
				if adminPassword == "" {
					w.Header().Set("X-Auth-Error", "admin authentication not configured")
					nethttp.Error(w, "invalid credentials", nethttp.StatusUnauthorized)
					return
				}

				if subtle.ConstantTimeCompare([]byte(secret), []byte(adminPassword)) != 1 {
					w.Header().Set("X-Auth-Error", "invalid admin password")
					nethttp.Error(w, "invalid credentials", nethttp.StatusUnauthorized)
					return
				}

				authCtx := &AuthContext{
					KeyID:    "admin",
					BucketID: 0,
					Role:     models.RoleAll,
				}
				ctx := SetAuthContext(req.Context(), authCtx)

				next.ServeHTTP(w, req.WithContext(ctx))
				return
			}

			akStore := models.NewAccessKeyStore(r.db)
			ctx := req.Context()
			accessKey, err := akStore.GetByKeyID(ctx, keyID)
			if err != nil {
				if err == sql.ErrNoRows {
					w.Header().Set("X-Auth-Error", "key_id not found")
					nethttp.Error(w, "invalid credentials", nethttp.StatusUnauthorized)
					return
				}
				w.Header().Set("X-Auth-Error", "database error")
				nethttp.Error(w, "internal error", nethttp.StatusInternalServerError)
				return
			}

			if !verifySecret(secret, accessKey.SecretHash) {
				w.Header().Set("X-Auth-Error", "secret mismatch")
				nethttp.Error(w, "invalid credentials", nethttp.StatusUnauthorized)
				return
			}

			if !hasPermission(accessKey.Role, level) {
				nethttp.Error(w, "insufficient permissions", nethttp.StatusForbidden)
				return
			}

			bucketName := chi.URLParam(req, "bucketName")
			if bucketName == "" {
				bucketName = chi.URLParam(req, "name")
			}

			if bucketName != "" {
				bStore := models.NewBucketStore(r.db)
				bucket, err := bStore.GetBucketByName(ctx, bucketName)
				if err != nil {
					if err == sql.ErrNoRows {
						nethttp.Error(w, "bucket not found", nethttp.StatusNotFound)
						return
					}
					nethttp.Error(w, "internal error", nethttp.StatusInternalServerError)
					return
				}

				if bucket.ID != accessKey.BucketID {
					nethttp.Error(w, "access denied to this bucket", nethttp.StatusForbidden)
					return
				}
			}

			authCtx := &AuthContext{
				KeyID:    accessKey.KeyID,
				BucketID: accessKey.BucketID,
				Role:     accessKey.Role,
			}
			ctx = SetAuthContext(ctx, authCtx)

			next.ServeHTTP(w, req.WithContext(ctx))
		})
	}
}

func verifySecret(secret, storedHash string) bool {
	hash := sha256.Sum256([]byte(secret))
	providedHash := base64.StdEncoding.EncodeToString(hash[:])

	return subtle.ConstantTimeCompare([]byte(providedHash), []byte(storedHash)) == 1
}

func hasPermission(role models.AccessKeyRole, required AuthLevel) bool {
	roleLevel := getRoleLevel(role)
	return roleLevel >= int(required)
}

func getRoleLevel(role models.AccessKeyRole) int {
	switch role {
	case models.RoleReadOnly:
		return int(AuthLevelReadOnly)
	case models.RoleReadWrite:
		return int(AuthLevelReadWrite)
	case models.RoleAll:
		return int(AuthLevelAll)
	default:
		return 0
	}
}
