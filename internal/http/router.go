package http

import (
	"database/sql"
	"encoding/json"
	"io"
	nethttp "net/http"
	"strings"

	"buck_It_Up/internal/models"
)

type Router struct {
	mux *nethttp.ServeMux
	db  *sql.DB
}

func New(db *sql.DB) *Router {
	r := &Router{mux: nethttp.NewServeMux(), db: db}

	// Register handlers using stdlib mux
	r.mux.HandleFunc("/buckets/", r.getBucketByName) // matches /buckets/{name}
	r.mux.HandleFunc("/", r.index)
	r.mux.HandleFunc("/health", r.health)
	r.mux.HandleFunc("/echo", r.echo)

	return r
}

func (r *Router) Handler() nethttp.Handler {
	return r.mux
}

func (r *Router) getBucketByName(w nethttp.ResponseWriter, req *nethttp.Request) {
	// Expect path like /buckets/{name}
	name := strings.TrimPrefix(req.URL.Path, "/buckets/")
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

func (r *Router) index(w nethttp.ResponseWriter, req *nethttp.Request) {
	w.WriteHeader(nethttp.StatusOK)
	_, _ = w.Write([]byte("Hello from stdlib net/http"))
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
