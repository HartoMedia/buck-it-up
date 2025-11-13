package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	dbpkg "buck_It_Up/internal/db"
	buckhttp "buck_It_Up/internal/http"
)

func main() {
	dbPath := os.Getenv("BUCKET_DB_PATH")
	if dbPath == "" {
		dbPath = "buckitup.db"
	}

	db := dbpkg.Open(dbPath)
	defer db.Close()

	router := buckhttp.New(db)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router.Handler(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	fmt.Println("Starting server on http://localhost:8080")
	log.Fatal(srv.ListenAndServe())
}
