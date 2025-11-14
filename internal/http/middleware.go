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

// AuthLevel represents the required authentication level for a route
type AuthLevel int

const (
	AuthLevelNone AuthLevel = iota
	AuthLevelReadOnly
	AuthLevelReadWrite
	AuthLevelAll
)

// AuthMiddleware creates a middleware that enforces authentication and authorization
func (r *Router) AuthMiddleware(level AuthLevel) func(nethttp.Handler) nethttp.Handler {
	return func(next nethttp.Handler) nethttp.Handler {
		return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, req *nethttp.Request) {
			// No auth required for this route
			if level == AuthLevelNone {
				next.ServeHTTP(w, req)
				return
			}

			// Extract credentials from Authorization header
			// Expected format: "Bearer <key_id>:<secret>" or "Bearer admin:<admin_password>"
			authHeader := req.Header.Get("Authorization")
			if authHeader == "" {
				w.Header().Set("X-Auth-Error", "missing authorization header")
				nethttp.Error(w, "authorization required", nethttp.StatusUnauthorized)
				return
			}

			// Parse the header
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				w.Header().Set("X-Auth-Error", "invalid format - expected 'Bearer <key_id>:<secret>'")
				nethttp.Error(w, "invalid authorization format", nethttp.StatusUnauthorized)
				return
			}

			// Extract key_id and secret
			credentials := strings.SplitN(parts[1], ":", 2)
			if len(credentials) != 2 {
				w.Header().Set("X-Auth-Error", "invalid credentials format - expected 'key_id:secret'")
				nethttp.Error(w, "invalid credentials format", nethttp.StatusUnauthorized)
				return
			}

			keyID := credentials[0]
			secret := credentials[1]

			// Check if this is an admin login
			if keyID == "admin" {
				adminPassword := os.Getenv("ADMIN_PASSWORD")
				if adminPassword == "" {
					// Admin password not configured
					w.Header().Set("X-Auth-Error", "admin authentication not configured")
					nethttp.Error(w, "invalid credentials", nethttp.StatusUnauthorized)
					return
				}

				// Verify admin password using constant-time comparison
				if subtle.ConstantTimeCompare([]byte(secret), []byte(adminPassword)) != 1 {
					w.Header().Set("X-Auth-Error", "invalid admin password")
					nethttp.Error(w, "invalid credentials", nethttp.StatusUnauthorized)
					return
				}

				// Admin authenticated - create admin auth context with full access
				authCtx := &AuthContext{
					KeyID:    "admin",
					BucketID: 0, // Admin has access to all buckets
					Role:     models.RoleAll,
				}
				ctx := SetAuthContext(req.Context(), authCtx)

				// Continue to the next handler with admin privileges
				next.ServeHTTP(w, req.WithContext(ctx))
				return
			}

			// Look up the access key in the database
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

			// Verify the secret hash
			if !verifySecret(secret, accessKey.SecretHash) {
				w.Header().Set("X-Auth-Error", "secret mismatch")
				nethttp.Error(w, "invalid credentials", nethttp.StatusUnauthorized)
				return
			}

			// Check if the access key has sufficient permissions
			if !hasPermission(accessKey.Role, level) {
				nethttp.Error(w, "insufficient permissions", nethttp.StatusForbidden)
				return
			}

			// Extract bucket name from URL if present
			bucketName := chi.URLParam(req, "bucketName")
			if bucketName == "" {
				bucketName = chi.URLParam(req, "name")
			}

			// If a bucket is specified in the route, verify access
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

				// Verify the access key has access to this specific bucket
				if bucket.ID != accessKey.BucketID {
					nethttp.Error(w, "access denied to this bucket", nethttp.StatusForbidden)
					return
				}
			}

			// Store auth context in request context
			authCtx := &AuthContext{
				KeyID:    accessKey.KeyID,
				BucketID: accessKey.BucketID,
				Role:     accessKey.Role,
			}
			ctx = SetAuthContext(ctx, authCtx)

			// Continue to the next handler
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	}
}

// verifySecret compares the provided secret with the stored hash
func verifySecret(secret, storedHash string) bool {
	// Hash the provided secret
	hash := sha256.Sum256([]byte(secret))
	providedHash := base64.StdEncoding.EncodeToString(hash[:])

	// Use constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare([]byte(providedHash), []byte(storedHash)) == 1
}

// hasPermission checks if the access key role has sufficient permissions
func hasPermission(role models.AccessKeyRole, required AuthLevel) bool {
	roleLevel := getRoleLevel(role)
	return roleLevel >= int(required)
}

// getRoleLevel returns the numeric level for a role
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
