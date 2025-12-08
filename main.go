package main

import (
	"log"
	"os"

	nethttp "net/http"

	"buck_It_Up/internal/db"
	httpinternal "buck_It_Up/internal/http"
)

func main() {
	dbPath := os.Getenv("BUCKITUP_DB_PATH")
	if dbPath == "" {
		dbPath = "data.db"
	}

	d := db.Open(dbPath)
	defer d.Close()

	r := httpinternal.New(d)

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}

	log.Printf("starting server on %s (db=%s)", addr, dbPath)
	if err := nethttp.ListenAndServe(addr, r.Handler()); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
