package api

import (
	"net/http"
)

func (s *Server) handleGetItems(w http.ResponseWriter, r *http.Request) {
	items, err := s.db.GetAllItems(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch items")
		return
	}
	writeJSON(w, http.StatusOK, items)
}
