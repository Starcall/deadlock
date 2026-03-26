package api

import (
	"database/sql"
	"net/http"
	"sort"
	"time"

	"github.com/name/deadlock/internal/domain"
	"github.com/name/deadlock/internal/model"
)

func (s *Server) handleGetModelStats(w http.ResponseWriter, r *http.Request) {
	meta, err := s.db.GetActiveModel(r.Context())
	if err != nil {
		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusOK, map[string]any{
				"trained":    false,
				"message":    "No model trained yet. Run compute first.",
			})
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to fetch model stats")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"trained":     true,
		"trained_at":  time.Unix(meta.TrainedAt, 0).Format(time.RFC3339),
		"accuracy":    meta.Accuracy,
		"ece":         meta.ECE,
		"num_matches": meta.NumMatches,
	})
}

func (s *Server) handleGetReliability(w http.ResponseWriter, r *http.Request) {
	meta, err := s.db.GetActiveModel(r.Context())
	if err != nil {
		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusOK, []model.ReliabilityBin{})
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to fetch model")
		return
	}

	m, err := model.DeserializeModel(meta.Weights)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to deserialize model")
		return
	}

	// Return model info (reliability data would need to be stored separately
	// or recomputed; for now return the model params as verification)
	writeJSON(w, http.StatusOK, map[string]any{
		"calibrated": m.Calibrated,
		"platt_a":    m.PlattA,
		"platt_b":    m.PlattB,
	})
}

func (s *Server) handleGetStatus(w http.ResponseWriter, r *http.Request) {
	matchCount, latestTime, modelAcc, err := s.db.GetStatus(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch status")
		return
	}

	resp := map[string]any{
		"match_count":    matchCount,
		"model_accuracy": modelAcc,
	}

	if latestTime > 0 {
		resp["latest_match"] = time.Unix(latestTime, 0).Format(time.RFC3339)
	}

	writeJSON(w, http.StatusOK, resp)
}

// sortWPAResults sorts WPA results by the given field.
func sortWPAResults(results []domain.WPAResult, sortBy string) {
	switch sortBy {
	case "initial_w":
		sort.Slice(results, func(i, j int) bool {
			return results[i].MeanInitialW > results[j].MeanInitialW
		})
	case "win_rate":
		sort.Slice(results, func(i, j int) bool {
			return results[i].WinRate > results[j].WinRate
		})
	case "sample_size":
		sort.Slice(results, func(i, j int) bool {
			return results[i].SampleSize > results[j].SampleSize
		})
	default: // "delta_w" or empty
		sort.Slice(results, func(i, j int) bool {
			return results[i].MeanDeltaW > results[j].MeanDeltaW
		})
	}
}
