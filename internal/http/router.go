package http

import (
	"database/sql"
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

	// Misc routes
	r.mux.Get("/health", r.health)
	r.mux.Get("/echo", r.echo) // echo only supports GET body for parity; could be POST

	// Bucket routes at root
	// LIST / -> list all buckets
	r.mux.MethodFunc(MethodList, "/", r.listBuckets)
	// POST / -> add a new bucket
	r.mux.Post("/", r.createBucket)
	// GET /{name} -> get single bucket by name
	r.mux.Get("/{name}", r.getBucketByName)
	// DELETE /{name} -> delete bucket by name
	r.mux.Delete("/{name}", r.deleteBucketByName)

	// Object or Bucket interal routes
	r.mux.MethodFunc(MethodList, "/{bucketName}", r.listBucketContent)
	// GET /{bucketName}/{objectKey} -> get single object metadata + content
	r.mux.Get("/{bucketName}/*", r.getObjectByKey)

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
	if _, err := store.NewBucket(ctx, bucket); err != nil {
		// Check for uniqueness violation (sqlite returns error string containing "UNIQUE")
		errMsg := err.Error()
		if strings.Contains(strings.ToLower(errMsg), "unique") {
			nethttp.Error(w, "bucket exists", nethttp.StatusConflict)
			return
		}
		nethttp.Error(w, "internal error", nethttp.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(nethttp.StatusCreated)
	_ = json.NewEncoder(w).Encode(bucket)
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
	objectKey := strings.TrimPrefix(req.URL.Path, "/"+bucketName+"/")
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
		Content: string(data), // raw content; could base64 encode if binary
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(nethttp.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
