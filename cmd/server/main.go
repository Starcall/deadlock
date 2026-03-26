package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/name/deadlock/internal/api"
	"github.com/name/deadlock/internal/store"
)

func main() {
	dbPath := flag.String("db", envOr("DB_PATH", "./deadlock.db"), "SQLite database path")
	port := flag.String("port", envOr("SERVER_PORT", "8080"), "Server port")
	frontendURL := flag.String("frontend-url", envOr("FRONTEND_URL", "http://localhost:3000"), "Frontend URL for CORS")
	flag.Parse()

	db, err := store.New(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	srv := api.NewServer(db, *frontendURL)

	addr := fmt.Sprintf(":%s", *port)
	log.Printf("Starting server on %s", addr)
	log.Printf("Frontend URL: %s", *frontendURL)

	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
