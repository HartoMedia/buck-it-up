package http

import (
	"embed"
	"io/fs"
	"math/rand"
	nethttp "net/http"
	"strings"
	"time"
)

//go:embed ui_login.html
var loginHTML string

//go:embed ui_dashboard.html
var dashboardHTML string

//go:embed ui_bucket.html
var bucketHTML string

//go:embed empty_svgs/*.svg
var emptySVGsFS embed.FS

var emptySVGFiles []string

func init() {
	// Seed RNG once
	rand.Seed(time.Now().UnixNano())
	// Build list of embedded SVG file paths
	entries, err := emptySVGsFS.ReadDir("empty_svgs")
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".svg") {
				emptySVGFiles = append(emptySVGFiles, "empty_svgs/"+e.Name())
			}
		}
	}
}

func (r *Router) uiLogin(w nethttp.ResponseWriter, _ *nethttp.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(nethttp.StatusOK)
	_, _ = w.Write([]byte(loginHTML))
}

func (r *Router) uiDashboard(w nethttp.ResponseWriter, _ *nethttp.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(nethttp.StatusOK)
	_, _ = w.Write([]byte(dashboardHTML))
}

func (r *Router) uiBucketView(w nethttp.ResponseWriter, _ *nethttp.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(nethttp.StatusOK)
	_, _ = w.Write([]byte(bucketHTML))
}

func (r *Router) uiRandomEmptySVG(w nethttp.ResponseWriter, _ *nethttp.Request) {
	// Serve a random embedded SVG (no per-request reseed)
	if len(emptySVGFiles) == 0 {
		nethttp.Error(w, "no svgs available", nethttp.StatusInternalServerError)
		return
	}
	path := emptySVGFiles[rand.Intn(len(emptySVGFiles))]
	data, err := fs.ReadFile(emptySVGsFS, path)
	if err != nil {
		nethttp.Error(w, "failed to load svg", nethttp.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/svg+xml")
	// Prevent caching so repeated loads can show different SVGs
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(nethttp.StatusOK)
	_, _ = w.Write(data)
}
