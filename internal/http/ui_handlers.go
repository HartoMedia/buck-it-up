package http

import (
	_ "embed"
	nethttp "net/http"
)

//go:embed ui_login.html
var loginHTML string

//go:embed ui_dashboard.html
var dashboardHTML string

//go:embed ui_bucket.html
var bucketHTML string

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
