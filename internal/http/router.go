package http

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	nethttp "net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"buck_It_Up/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Router struct {
	mux chi.Router
	db  *sql.DB
}

// Custom HTTP method for listing resources
const MethodList = "LIST"

func init() {
	// Register custom HTTP method with chi
	chi.RegisterMethod(MethodList)
}

func New(db *sql.DB) *Router {
	// Create chi router
	r := &Router{mux: chi.NewRouter(), db: db}

	// Basic middleware stack (recoverer + requestID + logger minimal)
	r.mux.Use(middleware.RequestID)
	r.mux.Use(middleware.RealIP)
	r.mux.Use(middleware.Logger)
	r.mux.Use(middleware.Recoverer)
	r.mux.Use(middleware.Compress(5)) // Enable gzip compression with compression level 5

	//Misc routes - no auth required
	r.mux.Get("/health", r.health)
	r.mux.Get("/echo", r.echo)

	// Bucket routes at root (admin only - will be protected later)
	// LIST / -> list all buckets
	r.mux.MethodFunc(MethodList, "/", r.listBuckets)
	// POST / -> add a new bucket
	r.mux.Post("/", r.createBucket)

	// Protected bucket routes - read only access
	r.mux.Group(func(readOnly chi.Router) {
		readOnly.Use(r.AuthMiddleware(AuthLevelReadOnly))
		// GET /{name} -> get single bucket by name (only for their bucket)
		readOnly.Get("/{name}", r.getBucketByName)
	})

	// Protected bucket content routes - read only access
	r.mux.Group(func(readOnly chi.Router) {
		readOnly.Use(r.AuthMiddleware(AuthLevelReadOnly))
		// LIST /{bucketName} -> list bucket content
		readOnly.MethodFunc(MethodList, "/{bucketName}", r.listBucketContent)
		// GET /{bucketName}/all/* -> get object metadata + content
		readOnly.Get("/{bucketName}/all/*", r.getObjectByKey)
		// GET /{bucketName}/metadata/* -> get object metadata only
		readOnly.Get("/{bucketName}/metadata/*", r.getObjectByKeyOnlyMetadata)
		// GET /{bucketName}/content/* -> get object content only
		readOnly.Get("/{bucketName}/content/*", r.getObjectByKeyOnlyContent)
	})

	// Protected object routes - read/write access
	r.mux.Group(func(readWrite chi.Router) {
		readWrite.Use(r.AuthMiddleware(AuthLevelReadWrite))
		// POST /{bucketName}/upload -> upload object to bucket
		readWrite.Post("/{bucketName}/upload", r.uploadObjectToBucket)
		// DELETE /{bucketName}/* -> delete object by key
		readWrite.Delete("/{bucketName}/*", r.deleteObjectByKey)
	})

	// Protected bucket routes - all access (bucket deletion)
	r.mux.Group(func(all chi.Router) {
		all.Use(r.AuthMiddleware(AuthLevelAll))
		// DELETE /{name} -> delete bucket by name
		all.Delete("/{name}", r.deleteBucketByName)
	})

	return r
}

func (r *Router) Handler() nethttp.Handler {
	return r.mux
}

func (r *Router) listBucketContent(w nethttp.ResponseWriter, req *nethttp.Request) {
	bucketName := chi.URLParam(req, "bucketName")
	if bucketName == "" || strings.Contains(bucketName, "/") {
		nethttp.Error(w, "invalid bucket name", nethttp.StatusBadRequest)
		return
	}
	store := models.NewObjectStore(r.db)
	ctx := req.Context()
	c, err := store.ListObjectsByBucketName(ctx, bucketName)
	if err != nil {
		if err == sql.ErrNoRows {
			nethttp.NotFound(w, req)
			return
		}
		nethttp.Error(w, "internal error", nethttp.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(nethttp.StatusOK)
	_ = json.NewEncoder(w).Encode(c)
}

func (r *Router) getBucketByName(w nethttp.ResponseWriter, req *nethttp.Request) {
	name := chi.URLParam(req, "name")
	if name == "" || strings.Contains(name, "/") { // chi already won't match embedded slash, but keep validation
		nethttp.Error(w, "invalid bucket name", nethttp.StatusBadRequest)
		return
	}

	store := models.NewBucketStore(r.db)
	ctx := req.Context()
	b, err := store.GetBucketByName(ctx, name)
	if err != nil {
		if err == sql.ErrNoRows {
			nethttp.NotFound(w, req)
			return
		}
		nethttp.Error(w, "internal error", nethttp.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(nethttp.StatusOK)
	_ = json.NewEncoder(w).Encode(b)
}

// LIST /
func (r *Router) listBuckets(w nethttp.ResponseWriter, req *nethttp.Request) {
	store := models.NewBucketStore(r.db)
	ctx := req.Context()
	buckets, err := store.ListBuckets(ctx)
	if err != nil {
		nethttp.Error(w, "internal error", nethttp.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(nethttp.StatusOK)
	_ = json.NewEncoder(w).Encode(buckets)
}

func (r *Router) index(w nethttp.ResponseWriter, req *nethttp.Request) {
	w.WriteHeader(nethttp.StatusOK)
	_, _ = w.Write([]byte("Hello from chi router"))
}

func (r *Router) health(w nethttp.ResponseWriter, req *nethttp.Request) {
	w.WriteHeader(nethttp.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (r *Router) echo(w nethttp.ResponseWriter, req *nethttp.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		nethttp.Error(w, "invalid body", nethttp.StatusBadRequest)
		return
	}
	w.WriteHeader(nethttp.StatusOK)
	_, _ = w.Write(body)
}

func (r *Router) createBucket(w nethttp.ResponseWriter, req *nethttp.Request) {
	// Expect JSON body {"name":"bucketName"}
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		nethttp.Error(w, "invalid json", nethttp.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		nethttp.Error(w, "name required", nethttp.StatusBadRequest)
		return
	}
	if strings.Contains(name, "/") {
		nethttp.Error(w, "invalid name", nethttp.StatusBadRequest)
		return
	}
	store := models.NewBucketStore(r.db)
	ctx := req.Context()
	bucket := &models.Bucket{Name: name, CreatedAt: time.Now().Unix()}
	bucketID, err := store.NewBucket(ctx, bucket)
	if err != nil {
		// Check for uniqueness violation (sqlite returns error string containing "UNIQUE")
		errMsg := err.Error()
		if strings.Contains(strings.ToLower(errMsg), "unique") {
			nethttp.Error(w, "bucket exists", nethttp.StatusConflict)
			return
		}
		nethttp.Error(w, "internal error", nethttp.StatusInternalServerError)
		return
	}
	bucket.ID = bucketID

	// Generate 3 access keys - one for each role
	akStore := models.NewAccessKeyStore(r.db)
	roles := []models.AccessKeyRole{
		models.RoleReadOnly,
		models.RoleReadWrite,
		models.RoleAll,
	}

	accessKeys := make([]*models.AccessKeyWithSecretResponse, 0, 3)

	for _, role := range roles {
		keyID, secret, err := r.generateAccessKey()
		if err != nil {
			// Rollback: delete the bucket (this will cascade delete any created keys)
			_ = store.DeleteBucketByName(ctx, name)
			nethttp.Error(w, "failed to generate access keys", nethttp.StatusInternalServerError)
			return
		}

		// Hash the secret for storage
		secretHash := hashSecret(secret)

		ak := &models.AccessKey{
			BucketID:   bucketID,
			KeyID:      keyID,
			SecretHash: secretHash,
			Role:       role,
			CreatedAt:  time.Now().Unix(),
		}

		akID, err := akStore.CreateAccessKey(ctx, ak)
		if err != nil {
			// Rollback: delete the bucket
			_ = store.DeleteBucketByName(ctx, name)
			nethttp.Error(w, "failed to create access keys", nethttp.StatusInternalServerError)
			return
		}
		ak.ID = akID

		accessKeys = append(accessKeys, ak.ToResponseWithSecret(secret))
	}

	// Return bucket with access keys
	response := struct {
		*models.BucketResponse
		AccessKeys []*models.AccessKeyWithSecretResponse `json:"access_keys"`
	}{
		BucketResponse: bucket.ToResponse(),
		AccessKeys:     accessKeys,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(nethttp.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}

func (r *Router) deleteBucketByName(w nethttp.ResponseWriter, req *nethttp.Request) {
	name := chi.URLParam(req, "name")
	if name == "" || strings.Contains(name, "/") {
		nethttp.Error(w, "invalid bucket name", nethttp.StatusBadRequest)
		return
	}
	store := models.NewBucketStore(r.db)
	ctx := req.Context()
	// Ensure bucket exists first
	if _, err := store.GetBucketByName(ctx, name); err != nil {
		if err == sql.ErrNoRows {
			nethttp.NotFound(w, req)
			return
		}
		nethttp.Error(w, "internal error", nethttp.StatusInternalServerError)
		return
	}
	if err := store.DeleteBucketByName(ctx, name); err != nil {
		// Likely FK constraint violation if objects/access keys exist
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "constraint") {
			nethttp.Error(w, "bucket not empty", nethttp.StatusConflict)
			return
		}
		nethttp.Error(w, "internal error", nethttp.StatusInternalServerError)
		return
	}
	w.WriteHeader(nethttp.StatusNoContent)
}

func (r *Router) getObjectByKey(w nethttp.ResponseWriter, req *nethttp.Request) {
	bucketName := chi.URLParam(req, "bucketName")
	if bucketName == "" || strings.Contains(bucketName, "/") {
		nethttp.Error(w, "invalid bucket name", nethttp.StatusBadRequest)
		return
	}
	// Remaining path after bucketName is the object key (chi wildcard *) but the objectKey can contain slashes
	objectKey := strings.TrimPrefix(req.URL.Path, "/"+bucketName+"/all/")
	objectKey = strings.TrimSpace(objectKey)
	if objectKey == "" || strings.Contains(objectKey, "\x00") { // basic validation
		nethttp.Error(w, "invalid object key", nethttp.StatusBadRequest)
		return
	}

	ctx := req.Context()
	bStore := models.NewBucketStore(r.db)
	bucket, err := bStore.GetBucketByName(ctx, bucketName)
	if err != nil {
		if err == sql.ErrNoRows {
			nethttp.NotFound(w, req)
			return
		}
		nethttp.Error(w, "internal error", nethttp.StatusInternalServerError)
		return
	}
	fmt.Println("Found Bucket:", bucket)

	oStore := models.NewObjectStore(r.db)
	obj, err := oStore.GetObject(ctx, bucket.ID, objectKey)
	fmt.Println("Object:", obj)
	if err != nil {
		if err == sql.ErrNoRows {
			nethttp.NotFound(w, req)
			return
		}
		nethttp.Error(w, "internal error", nethttp.StatusInternalServerError)
		return
	}

	// Read object content from its file_path
	contentPath := obj.FilePath
	// Security: ensure file path is inside expected bucket directory
	fmt.Println("Trying to read file at path:", contentPath)
	expectedPrefix := filepath.Join("data", "buckets", strconv.FormatInt(bucket.ID, 10), "objects")
	// Normalize both paths for comparison (handle forward/backward slashes)
	normalizedContentPath := filepath.Clean(contentPath)
	normalizedExpectedPrefix := filepath.Clean(expectedPrefix)
	if !strings.HasPrefix(normalizedContentPath, normalizedExpectedPrefix) {
		nethttp.Error(w, "invalid stored path", nethttp.StatusInternalServerError)
		return
	}
	data, err := os.ReadFile(contentPath)
	if err != nil {
		if os.IsNotExist(err) {
			nethttp.Error(w, "object file missing", nethttp.StatusInternalServerError)
			return
		}
		nethttp.Error(w, "failed to read object", nethttp.StatusInternalServerError)
		return
	}

	resp := struct {
		*models.Object `json:"object"`
		Content        string `json:"content"`
	}{
		Object:  obj,
		Content: base64.StdEncoding.EncodeToString(data),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(nethttp.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (r *Router) getObjectByKeyOnlyMetadata(w nethttp.ResponseWriter, req *nethttp.Request) {
	bucketName := chi.URLParam(req, "bucketName")
	if bucketName == "" || strings.Contains(bucketName, "/") {
		nethttp.Error(w, "invalid bucket name", nethttp.StatusBadRequest)
		return
	}
	// Remaining path after bucketName/metadata/ is the object key
	objectKey := strings.TrimPrefix(req.URL.Path, "/"+bucketName+"/metadata/")
	objectKey = strings.TrimSpace(objectKey)
	if objectKey == "" || strings.Contains(objectKey, "\x00") {
		nethttp.Error(w, "invalid object key", nethttp.StatusBadRequest)
		return
	}

	ctx := req.Context()
	bStore := models.NewBucketStore(r.db)
	bucket, err := bStore.GetBucketByName(ctx, bucketName)
	if err != nil {
		if err == sql.ErrNoRows {
			nethttp.NotFound(w, req)
			return
		}
		nethttp.Error(w, "internal error", nethttp.StatusInternalServerError)
		return
	}

	oStore := models.NewObjectStore(r.db)
	obj, err := oStore.GetObject(ctx, bucket.ID, objectKey)
	if err != nil {
		if err == sql.ErrNoRows {
			nethttp.NotFound(w, req)
			return
		}
		nethttp.Error(w, "internal error", nethttp.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(nethttp.StatusOK)
	_ = json.NewEncoder(w).Encode(obj)
}

func (r *Router) getObjectByKeyOnlyContent(w nethttp.ResponseWriter, req *nethttp.Request) {
	bucketName := chi.URLParam(req, "bucketName")
	if bucketName == "" || strings.Contains(bucketName, "/") {
		nethttp.Error(w, "invalid bucket name", nethttp.StatusBadRequest)
		return
	}
	// Remaining path after bucketName/content/ is the object key
	objectKey := strings.TrimPrefix(req.URL.Path, "/"+bucketName+"/content/")
	objectKey = strings.TrimSpace(objectKey)
	if objectKey == "" || strings.Contains(objectKey, "\x00") {
		nethttp.Error(w, "invalid object key", nethttp.StatusBadRequest)
		return
	}

	ctx := req.Context()
	bStore := models.NewBucketStore(r.db)
	bucket, err := bStore.GetBucketByName(ctx, bucketName)
	if err != nil {
		if err == sql.ErrNoRows {
			nethttp.NotFound(w, req)
			return
		}
		nethttp.Error(w, "internal error", nethttp.StatusInternalServerError)
		return
	}

	oStore := models.NewObjectStore(r.db)
	obj, err := oStore.GetObject(ctx, bucket.ID, objectKey)
	if err != nil {
		if err == sql.ErrNoRows {
			nethttp.NotFound(w, req)
			return
		}
		nethttp.Error(w, "internal error", nethttp.StatusInternalServerError)
		return
	}

	// Read object content from its file_path
	contentPath := obj.FilePath
	// Security: ensure file path is inside expected bucket directory
	expectedPrefix := filepath.Join("data", "buckets", strconv.FormatInt(bucket.ID, 10), "objects")
	// Normalize both paths for comparison (handle forward/backward slashes)
	normalizedContentPath := filepath.Clean(contentPath)
	normalizedExpectedPrefix := filepath.Clean(expectedPrefix)
	if !strings.HasPrefix(normalizedContentPath, normalizedExpectedPrefix) {
		nethttp.Error(w, "invalid stored path", nethttp.StatusInternalServerError)
		return
	}
	data, err := os.ReadFile(contentPath)
	if err != nil {
		if os.IsNotExist(err) {
			nethttp.Error(w, "object file missing", nethttp.StatusInternalServerError)
			return
		}
		nethttp.Error(w, "failed to read object", nethttp.StatusInternalServerError)
		return
	}

	// Set content type from metadata if available
	if obj.ContentType != "" {
		w.Header().Set("Content-Type", obj.ContentType)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	w.WriteHeader(nethttp.StatusOK)
	_, _ = w.Write(data)
}

func (r *Router) uploadObjectToBucket(w nethttp.ResponseWriter, req *nethttp.Request) {
	bucketName := chi.URLParam(req, "bucketName")
	if bucketName == "" || strings.Contains(bucketName, "/") {
		nethttp.Error(w, "invalid bucket name", nethttp.StatusBadRequest)
		return
	}

	// Parse JSON body
	var body struct {
		ObjectKey     string `json:"object_key"`
		Content       string `json:"content"`
		ContentType   string `json:"content_type"`
		Base64Encoded bool   `json:"base64_encoded,omitempty"` // Optional flag to indicate content is base64
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		nethttp.Error(w, "invalid json", nethttp.StatusBadRequest)
		return
	}

	objectKey := strings.TrimSpace(body.ObjectKey)
	if objectKey == "" || strings.Contains(objectKey, "\x00") {
		nethttp.Error(w, "invalid object key", nethttp.StatusBadRequest)
		return
	}

	ctx := req.Context()
	bStore := models.NewBucketStore(r.db)
	bucket, err := bStore.GetBucketByName(ctx, bucketName)
	if err != nil {
		if err == sql.ErrNoRows {
			nethttp.NotFound(w, req)
			return
		}
		nethttp.Error(w, "internal error", nethttp.StatusInternalServerError)
		return
	}

	// Create object record first to get the object ID
	var contentBytes []byte
	if body.Base64Encoded {
		// Decode base64 content
		contentBytes, err = base64.StdEncoding.DecodeString(body.Content)
		if err != nil {
			nethttp.Error(w, "invalid base64 content", nethttp.StatusBadRequest)
			return
		}
	} else {
		// Use raw content
		contentBytes = []byte(body.Content)
	}
	contentType := strings.TrimSpace(body.ContentType)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Calculate checksum (simple approach - could use SHA256 or MD5)
	checksum := fmt.Sprintf("%d", len(contentBytes))

	// Create the object in DB (we need the ID to determine the file path)
	obj := &models.Object{
		BucketID:    bucket.ID,
		ObjectKey:   objectKey,
		FilePath:    "", // Will be set after we get the ID
		Size:        int64(len(contentBytes)),
		ContentType: contentType,
		Checksum:    checksum,
		CreatedAt:   time.Now().Unix(),
	}

	oStore := models.NewObjectStore(r.db)
	objectID, err := oStore.PutObject(ctx, obj)
	if err != nil {
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "unique") {
			nethttp.Error(w, "object already exists", nethttp.StatusConflict)
			return
		}
		nethttp.Error(w, "failed to create object", nethttp.StatusInternalServerError)
		return
	}

	// Now we have the object ID, determine the file path
	obj.ID = objectID

	// Use storage helper to ensure directory exists and get file path
	bucketDir, err := r.ensureBucketObjectsDir(bucket.ID)
	if err != nil {
		// Rollback: delete the DB record
		_ = oStore.DeleteObject(ctx, bucket.ID, objectKey)
		nethttp.Error(w, "failed to create storage directory", nethttp.StatusInternalServerError)
		return
	}

	filePath := filepath.Join(bucketDir, strconv.FormatInt(objectID, 10))

	// Write content to file
	if err := os.WriteFile(filePath, contentBytes, 0o644); err != nil {
		// Rollback: delete the DB record
		_ = oStore.DeleteObject(ctx, bucket.ID, objectKey)
		nethttp.Error(w, "failed to write object file", nethttp.StatusInternalServerError)
		return
	}

	// Update the object record with the file path
	obj.FilePath = filePath
	if err := r.updateObjectFilePath(ctx, objectID, filePath); err != nil {
		// Rollback: delete file and DB record
		_ = os.Remove(filePath)
		_ = oStore.DeleteObject(ctx, bucket.ID, objectKey)
		nethttp.Error(w, "failed to update object metadata", nethttp.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(nethttp.StatusCreated)
	_ = json.NewEncoder(w).Encode(obj)
}

func (r *Router) deleteObjectByKey(w nethttp.ResponseWriter, req *nethttp.Request) {
	bucketName := chi.URLParam(req, "bucketName")
	if bucketName == "" || strings.Contains(bucketName, "/") {
		nethttp.Error(w, "invalid bucket name", nethttp.StatusBadRequest)
		return
	}
	// Remaining path after bucketName is the object key (chi wildcard *)
	objectKey := strings.TrimPrefix(req.URL.Path, "/"+bucketName+"/")
	objectKey = strings.TrimSpace(objectKey)
	if objectKey == "" || strings.Contains(objectKey, "\x00") {
		nethttp.Error(w, "invalid object key", nethttp.StatusBadRequest)
		return
	}

	ctx := req.Context()
	bStore := models.NewBucketStore(r.db)
	bucket, err := bStore.GetBucketByName(ctx, bucketName)
	if err != nil {
		if err == sql.ErrNoRows {
			nethttp.NotFound(w, req)
			return
		}
		nethttp.Error(w, "internal error", nethttp.StatusInternalServerError)
		return
	}

	oStore := models.NewObjectStore(r.db)
	// Get the object first to find its file path
	obj, err := oStore.GetObject(ctx, bucket.ID, objectKey)
	if err != nil {
		if err == sql.ErrNoRows {
			nethttp.NotFound(w, req)
			return
		}
		nethttp.Error(w, "internal error", nethttp.StatusInternalServerError)
		return
	}

	// Delete the file from disk first
	if obj.FilePath != "" {
		// Security: ensure file path is inside expected bucket directory
		expectedPrefix := filepath.Join("data", "buckets", strconv.FormatInt(bucket.ID, 10), "objects")
		normalizedContentPath := filepath.Clean(obj.FilePath)
		normalizedExpectedPrefix := filepath.Clean(expectedPrefix)
		if !strings.HasPrefix(normalizedContentPath, normalizedExpectedPrefix) {
			nethttp.Error(w, "invalid stored path", nethttp.StatusInternalServerError)
			return
		}

		// Delete the file (ignore error if file doesn't exist)
		if err := os.Remove(obj.FilePath); err != nil && !os.IsNotExist(err) {
			nethttp.Error(w, "failed to delete object file", nethttp.StatusInternalServerError)
			return
		}
	}

	// Delete the object record from database
	if err := oStore.DeleteObject(ctx, bucket.ID, objectKey); err != nil {
		nethttp.Error(w, "failed to delete object", nethttp.StatusInternalServerError)
		return
	}

	w.WriteHeader(nethttp.StatusNoContent)
}

// Helper to ensure bucket objects directory exists
func (r *Router) ensureBucketObjectsDir(bucketID int64) (string, error) {
	dataRoot := "data"
	if dr := os.Getenv("BUCK_DATA_PATH"); dr != "" {
		dataRoot = dr
	}
	p := filepath.Join(dataRoot, "buckets", strconv.FormatInt(bucketID, 10), "objects")
	return p, os.MkdirAll(p, 0o755)
}

// Helper to update object file path in database
func (r *Router) updateObjectFilePath(ctx context.Context, objectID int64, filePath string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE objects
		SET file_path = ?
		WHERE id = ?
	`, filePath, objectID)
	return err
}

// generateAccessKey generates a secure random key ID and secret
func (r *Router) generateAccessKey() (keyID string, secret string, err error) {
	// Generate keyID (20 bytes -> ~27 chars base64)
	keyIDBytes := make([]byte, 20)
	if _, err := rand.Read(keyIDBytes); err != nil {
		return "", "", err
	}
	keyID = base64.URLEncoding.EncodeToString(keyIDBytes)

	// Generate secret (32 bytes -> ~43 chars base64)
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return "", "", err
	}
	secret = base64.URLEncoding.EncodeToString(secretBytes)

	return keyID, secret, nil
}

// hashSecret creates a SHA256 hash of the secret for storage
func hashSecret(secret string) string {
	hash := sha256.Sum256([]byte(secret))
	return base64.StdEncoding.EncodeToString(hash[:])
}
