package api

import (
	"net/http"
	"strconv"

	"github.com/name/deadlock/internal/wpa"
)

func (s *Server) handleGetHeroWPA(w http.ResponseWriter, r *http.Request) {
	heroIDStr := r.PathValue("heroId")
	heroID, err := strconv.Atoi(heroIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid hero ID")
		return
	}

	contextKey := r.URL.Query().Get("context")
	if contextKey == "" {
		contextKey = "all"
	}

	minSampleSize := 30
	if v := r.URL.Query().Get("min_sample_size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			minSampleSize = n
		}
	}

	results, err := s.db.GetWPAForHero(r.Context(), heroID, contextKey, minSampleSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch WPA data")
		return
	}

	// Sort by requested field
	sortBy := r.URL.Query().Get("sort")
	sortWPAResults(results, sortBy)

	writeJSON(w, http.StatusOK, results)
}

func (s *Server) handleGetHeroItemWPA(w http.ResponseWriter, r *http.Request) {
	heroIDStr := r.PathValue("heroId")
	heroID, err := strconv.Atoi(heroIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid hero ID")
		return
	}

	itemIDStr := r.PathValue("itemId")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid item ID")
		return
	}

	results, err := s.db.GetWPAForHeroItem(r.Context(), heroID, itemID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch WPA data")
		return
	}

	writeJSON(w, http.StatusOK, results)
}

func (s *Server) handleGetContexts(w http.ResponseWriter, r *http.Request) {
	contexts := wpa.AllContexts()
	type contextInfo struct {
		Key         string `json:"key"`
		Description string `json:"description"`
	}
	out := make([]contextInfo, len(contexts))
	for i, c := range contexts {
		out[i] = contextInfo{Key: c.Key, Description: c.Description}
	}
	writeJSON(w, http.StatusOK, out)
}
