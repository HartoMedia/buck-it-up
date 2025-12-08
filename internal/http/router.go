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

const MethodList = "LIST"

func init() {
	chi.RegisterMethod(MethodList)
}

func New(db *sql.DB) *Router {
	r := &Router{mux: chi.NewRouter(), db: db}

	r.mux.Use(middleware.RequestID)
	r.mux.Use(middleware.RealIP)
	r.mux.Use(middleware.Logger)
	r.mux.Use(middleware.Recoverer)
	r.mux.Use(middleware.Compress(5))

	//Misc routes - no auth required
	r.mux.Get("/health", r.health)
	r.mux.Get("/echo", r.echo)

	// UI - auth handled in browser
	r.mux.Get("/ui/login", r.uiLogin)
	r.mux.Get("/ui/dashboard", r.uiDashboard)
	r.mux.Get("/ui/bucket/*", r.uiBucketView)
	r.mux.Get("/ui/empty.svg", r.uiRandomEmptySVG)
	r.mux.Get("/ui", func(w nethttp.ResponseWriter, req *nethttp.Request) {
		nethttp.Redirect(w, req, "/ui/login", nethttp.StatusFound)
	})
	r.mux.Get("/ui/", func(w nethttp.ResponseWriter, req *nethttp.Request) {
		nethttp.Redirect(w, req, "/ui/login", nethttp.StatusFound)
	})

	r.mux.Group(func(admin chi.Router) {
		admin.Use(r.AuthMiddleware(AuthLevelAll))
		admin.MethodFunc(MethodList, "/", r.listBuckets)
		admin.Post("/", r.createBucket)
		admin.Get("/{name}/access-keys", r.listAccessKeys)
		admin.Post("/{name}/access-keys/recreate", r.recreateAccessKey)
	})

	r.mux.Group(func(readOnly chi.Router) {
		readOnly.Use(r.AuthMiddleware(AuthLevelReadOnly))
		readOnly.Get("/{name}", r.getBucketByName)
	})

	r.mux.Group(func(readOnly chi.Router) {
		readOnly.Use(r.AuthMiddleware(AuthLevelReadOnly))
		readOnly.MethodFunc(MethodList, "/{bucketName}", r.listBucketContent)
		readOnly.Get("/{bucketName}/all/*", r.getObjectByKey)
		readOnly.Get("/{bucketName}/metadata/*", r.getObjectByKeyOnlyMetadata)
		readOnly.Get("/{bucketName}/content/*", r.getObjectByKeyOnlyContent)
	})

	r.mux.Group(func(readWrite chi.Router) {
		readWrite.Use(r.AuthMiddleware(AuthLevelReadWrite))
		readWrite.Post("/{bucketName}/upload", r.uploadObjectToBucket)
		readWrite.Delete("/{bucketName}/*", r.deleteObjectByKey)
	})

	r.mux.Group(func(all chi.Router) {
		all.Use(r.AuthMiddleware(AuthLevelAll))
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
	if name == "" || strings.Contains(name, "/") {
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
		errMsg := err.Error()
		if strings.Contains(strings.ToLower(errMsg), "unique") {
			nethttp.Error(w, "bucket exists", nethttp.StatusConflict)
			return
		}
		nethttp.Error(w, "internal error", nethttp.StatusInternalServerError)
		return
	}
	bucket.ID = bucketID

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
			_ = store.DeleteBucketByName(ctx, name)
			nethttp.Error(w, "failed to generate access keys", nethttp.StatusInternalServerError)
			return
		}

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
			_ = store.DeleteBucketByName(ctx, name)
			nethttp.Error(w, "failed to create access keys", nethttp.StatusInternalServerError)
			return
		}
		ak.ID = akID

		accessKeys = append(accessKeys, ak.ToResponseWithSecret(secret))
	}

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
	if _, err := store.GetBucketByName(ctx, name); err != nil {
		if err == sql.ErrNoRows {
			nethttp.NotFound(w, req)
			return
		}
		nethttp.Error(w, "internal error", nethttp.StatusInternalServerError)
		return
	}
	if err := store.DeleteBucketByName(ctx, name); err != nil {
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
	objectKey := strings.TrimPrefix(req.URL.Path, "/"+bucketName+"/all/")
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

	contentPath := obj.FilePath
	fmt.Println("Trying to read file at path:", contentPath)
	expectedPrefix := filepath.Join("data", "buckets", strconv.FormatInt(bucket.ID, 10), "objects")
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

	contentPath := obj.FilePath
	expectedPrefix := filepath.Join("data", "buckets", strconv.FormatInt(bucket.ID, 10), "objects")
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

	var body struct {
		ObjectKey     string `json:"object_key"`
		Content       string `json:"content"`
		ContentType   string `json:"content_type"`
		Base64Encoded bool   `json:"base64_encoded,omitempty"`
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

	var contentBytes []byte
	if body.Base64Encoded {
		contentBytes, err = base64.StdEncoding.DecodeString(body.Content)
		if err != nil {
			nethttp.Error(w, "invalid base64 content", nethttp.StatusBadRequest)
			return
		}
	} else {
		contentBytes = []byte(body.Content)
	}
	contentType := strings.TrimSpace(body.ContentType)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	checksum := fmt.Sprintf("%d", len(contentBytes))

	obj := &models.Object{
		BucketID:    bucket.ID,
		ObjectKey:   objectKey,
		FilePath:    "",
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

	obj.ID = objectID

	bucketDir, err := r.ensureBucketObjectsDir(bucket.ID)
	if err != nil {
		_ = oStore.DeleteObject(ctx, bucket.ID, objectKey)
		nethttp.Error(w, "failed to create storage directory", nethttp.StatusInternalServerError)
		return
	}

	filePath := filepath.Join(bucketDir, strconv.FormatInt(objectID, 10))

	if err := os.WriteFile(filePath, contentBytes, 0o644); err != nil {
		_ = oStore.DeleteObject(ctx, bucket.ID, objectKey)
		nethttp.Error(w, "failed to write object file", nethttp.StatusInternalServerError)
		return
	}

	obj.FilePath = filePath
	if err := r.updateObjectFilePath(ctx, objectID, filePath); err != nil {
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
	obj, err := oStore.GetObject(ctx, bucket.ID, objectKey)
	if err != nil {
		if err == sql.ErrNoRows {
			nethttp.NotFound(w, req)
			return
		}
		nethttp.Error(w, "internal error", nethttp.StatusInternalServerError)
		return
	}

	if obj.FilePath != "" {
		expectedPrefix := filepath.Join("data", "buckets", strconv.FormatInt(bucket.ID, 10), "objects")
		normalizedContentPath := filepath.Clean(obj.FilePath)
		normalizedExpectedPrefix := filepath.Clean(expectedPrefix)
		if !strings.HasPrefix(normalizedContentPath, normalizedExpectedPrefix) {
			nethttp.Error(w, "invalid stored path", nethttp.StatusInternalServerError)
			return
		}

		if err := os.Remove(obj.FilePath); err != nil && !os.IsNotExist(err) {
			nethttp.Error(w, "failed to delete object file", nethttp.StatusInternalServerError)
			return
		}
	}

	if err := oStore.DeleteObject(ctx, bucket.ID, objectKey); err != nil {
		nethttp.Error(w, "failed to delete object", nethttp.StatusInternalServerError)
		return
	}

	w.WriteHeader(nethttp.StatusNoContent)
}

func (r *Router) listAccessKeys(w nethttp.ResponseWriter, req *nethttp.Request) {
	name := chi.URLParam(req, "name")
	if name == "" || strings.Contains(name, "/") {
		nethttp.Error(w, "invalid bucket name", nethttp.StatusBadRequest)
		return
	}

	ctx := req.Context()
	bStore := models.NewBucketStore(r.db)
	bucket, err := bStore.GetBucketByName(ctx, name)
	if err != nil {
		if err == sql.ErrNoRows {
			nethttp.NotFound(w, req)
			return
		}
		nethttp.Error(w, "internal error", nethttp.StatusInternalServerError)
		return
	}

	akStore := models.NewAccessKeyStore(r.db)
	accessKeys, err := akStore.ListAccessKeysByBucketID(ctx, bucket.ID)
	if err != nil {
		nethttp.Error(w, "internal error", nethttp.StatusInternalServerError)
		return
	}

	response := make([]*models.AccessKeyResponse, 0, len(accessKeys))
	for _, ak := range accessKeys {
		response = append(response, ak.ToResponse())
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(nethttp.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func (r *Router) recreateAccessKey(w nethttp.ResponseWriter, req *nethttp.Request) {
	name := chi.URLParam(req, "name")
	if name == "" || strings.Contains(name, "/") {
		nethttp.Error(w, "invalid bucket name", nethttp.StatusBadRequest)
		return
	}

	var body struct {
		Role models.AccessKeyRole `json:"role"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		nethttp.Error(w, "invalid json", nethttp.StatusBadRequest)
		return
	}

	if body.Role != models.RoleReadOnly && body.Role != models.RoleReadWrite && body.Role != models.RoleAll {
		nethttp.Error(w, "invalid role: must be 'readOnly', 'readWrite', or 'all'", nethttp.StatusBadRequest)
		return
	}

	ctx := req.Context()
	bStore := models.NewBucketStore(r.db)
	bucket, err := bStore.GetBucketByName(ctx, name)
	if err != nil {
		if err == sql.ErrNoRows {
			nethttp.NotFound(w, req)
			return
		}
		nethttp.Error(w, "internal error", nethttp.StatusInternalServerError)
		return
	}

	akStore := models.NewAccessKeyStore(r.db)

	if err := akStore.DeleteAccessKeyByBucketAndRole(ctx, bucket.ID, body.Role); err != nil {
		nethttp.Error(w, "failed to delete old access key", nethttp.StatusInternalServerError)
		return
	}

	keyID, secret, err := r.generateAccessKey()
	if err != nil {
		nethttp.Error(w, "failed to generate access key", nethttp.StatusInternalServerError)
		return
	}

	secretHash := hashSecret(secret)

	ak := &models.AccessKey{
		BucketID:   bucket.ID,
		KeyID:      keyID,
		SecretHash: secretHash,
		Role:       body.Role,
		CreatedAt:  time.Now().Unix(),
	}

	akID, err := akStore.CreateAccessKey(ctx, ak)
	if err != nil {
		nethttp.Error(w, "failed to create access key", nethttp.StatusInternalServerError)
		return
	}
	ak.ID = akID

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(nethttp.StatusCreated)
	_ = json.NewEncoder(w).Encode(ak.ToResponseWithSecret(secret))
}

func (r *Router) ensureBucketObjectsDir(bucketID int64) (string, error) {
	dataRoot := "data"
	if dr := os.Getenv("BUCKITUP_DATA_PATH"); dr != "" {
		dataRoot = dr
	}
	p := filepath.Join(dataRoot, "buckets", strconv.FormatInt(bucketID, 10), "objects")
	return p, os.MkdirAll(p, 0o755)
}

func (r *Router) updateObjectFilePath(ctx context.Context, objectID int64, filePath string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE objects
		SET file_path = ?
		WHERE id = ?
	`, filePath, objectID)
	return err
}

func (r *Router) generateAccessKey() (keyID string, secret string, err error) {
	keyIDBytes := make([]byte, 20)
	if _, err := rand.Read(keyIDBytes); err != nil {
		return "", "", err
	}
	keyID = base64.URLEncoding.EncodeToString(keyIDBytes)

	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return "", "", err
	}
	secret = base64.URLEncoding.EncodeToString(secretBytes)

	return keyID, secret, nil
}

func hashSecret(secret string) string {
	hash := sha256.Sum256([]byte(secret))
	return base64.StdEncoding.EncodeToString(hash[:])
}
