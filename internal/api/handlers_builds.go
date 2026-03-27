package api

import (
	"net/http"
	"strconv"
	"strings"
)

type heroBuildResponse struct {
	HeroID       int             `json:"hero_id"`
	TotalPlayers int             `json:"total_players"`
	Coverage     float64         `json:"coverage"`
	Builds       []buildResponse `json:"builds"`
}

type buildResponse struct {
	BuildRank  int     `json:"build_rank"`
	ItemIDs    []int64 `json:"item_ids"`
	ExactCount int     `json:"exact_count"`
	FuzzyCount int     `json:"fuzzy_count"`
	Wins       int     `json:"wins"`
	Losses     int     `json:"losses"`
	WinRate    float64 `json:"win_rate"`
}

func (s *Server) handleGetHeroBuilds(w http.ResponseWriter, r *http.Request) {
	heroIDStr := r.PathValue("heroId")
	heroID, err := strconv.Atoi(heroIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid hero ID")
		return
	}

	templates, err := s.db.GetBuildTemplatesForHero(r.Context(), heroID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch builds")
		return
	}

	coverage, err := s.db.GetBuildCoverageForHero(r.Context(), heroID)
	if err != nil {
		// No coverage data yet
		writeJSON(w, http.StatusOK, heroBuildResponse{
			HeroID: heroID,
			Builds: make([]buildResponse, 0),
		})
		return
	}

	resp := heroBuildResponse{
		HeroID:       heroID,
		TotalPlayers: coverage.TotalPlayers,
		Coverage:     coverage.Coverage,
		Builds:       make([]buildResponse, 0, len(templates)),
	}

	for _, t := range templates {
		resp.Builds = append(resp.Builds, buildResponse{
			BuildRank:  t.BuildRank,
			ItemIDs:    parseItemIDs(t.ItemIDs),
			ExactCount: t.ExactCount,
			FuzzyCount: t.FuzzyCount,
			Wins:       t.Wins,
			Losses:     t.Losses,
			WinRate:    t.WinRate,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleGetBuildCoverage(w http.ResponseWriter, r *http.Request) {
	coverages, err := s.db.GetBuildCoverage(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch coverage")
		return
	}
	writeJSON(w, http.StatusOK, coverages)
}

func parseItemIDs(csv string) []int64 {
	if csv == "" {
		return make([]int64, 0)
	}
	parts := strings.Split(csv, ",")
	ids := make([]int64, 0, len(parts))
	for _, p := range parts {
		id, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
		if err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}
