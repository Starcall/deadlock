package api

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleGetHeroes(w http.ResponseWriter, r *http.Request) {
	heroes, err := s.db.GetAllHeroes(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch heroes")
		return
	}
	writeJSON(w, http.StatusOK, heroes)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
