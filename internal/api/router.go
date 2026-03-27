package api

import (
	"net/http"

	"github.com/name/deadlock/internal/store"
)

// Server holds the API server dependencies.
type Server struct {
	db  *store.DB
	mux *http.ServeMux
}

// NewServer creates a new API server.
func NewServer(db *store.DB, frontendURL string) *Server {
	s := &Server{
		db:  db,
		mux: http.NewServeMux(),
	}

	// Wrap all routes with CORS middleware
	cors := corsMiddleware(frontendURL)

	s.mux.Handle("GET /api/heroes", cors(http.HandlerFunc(s.handleGetHeroes)))
	s.mux.Handle("GET /api/items", cors(http.HandlerFunc(s.handleGetItems)))
	s.mux.Handle("GET /api/wpa/hero/{heroId}", cors(http.HandlerFunc(s.handleGetHeroWPA)))
	s.mux.Handle("GET /api/wpa/hero/{heroId}/item/{itemId}", cors(http.HandlerFunc(s.handleGetHeroItemWPA)))
	s.mux.Handle("GET /api/wpa/contexts", cors(http.HandlerFunc(s.handleGetContexts)))
	s.mux.Handle("GET /api/builds/hero/{heroId}", cors(http.HandlerFunc(s.handleGetHeroBuilds)))
	s.mux.Handle("GET /api/builds/coverage", cors(http.HandlerFunc(s.handleGetBuildCoverage)))
	s.mux.Handle("GET /api/model/stats", cors(http.HandlerFunc(s.handleGetModelStats)))
	s.mux.Handle("GET /api/model/reliability", cors(http.HandlerFunc(s.handleGetReliability)))
	s.mux.Handle("GET /api/status", cors(http.HandlerFunc(s.handleGetStatus)))

	// Handle CORS preflight
	s.mux.Handle("OPTIONS /api/", cors(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})))

	return s
}

// Handler returns the HTTP handler for the server.
func (s *Server) Handler() http.Handler {
	return s.mux
}

func corsMiddleware(frontendURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := frontendURL
			if origin == "" {
				origin = "*"
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
