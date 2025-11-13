package http

import (
	"database/sql"
	"encoding/json"
	"io"
	nethttp "net/http"
	"strings"

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

func New(db *sql.DB) *Router {
	// Create chi router
	r := &Router{mux: chi.NewRouter(), db: db}

	// Basic middleware stack (recoverer + requestID + logger minimal)
	r.mux.Use(middleware.RequestID)
	r.mux.Use(middleware.RealIP)
	r.mux.Use(middleware.Logger)
	r.mux.Use(middleware.Recoverer)

	// Routes
	r.mux.Get("/health", r.health)
	r.mux.Get("/echo", r.echo) // echo only supports GET body for parity; could be POST
	r.mux.Get("/", r.index)

	// Bucket routes group
	r.mux.Route("/buckets", func(rg chi.Router) {
		// LIST /buckets -> list all buckets
		rg.MethodFunc(MethodList, "/", r.listBuckets)

		// GET /buckets/{name}
		rg.Get("/{name}", r.getBucketByName)
	})

	return r
}

func (r *Router) Handler() nethttp.Handler {
	return r.mux
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

// LIST /buckets
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
