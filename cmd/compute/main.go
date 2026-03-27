package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/name/deadlock/internal/builds"
	"github.com/name/deadlock/internal/domain"
	"github.com/name/deadlock/internal/model"
	"github.com/name/deadlock/internal/store"
	"github.com/name/deadlock/internal/wpa"
)

func main() {
	dbPath := flag.String("db", envOr("DB_PATH", "./deadlock.db"), "SQLite database path")
	snapshotInterval := flag.Int("snapshot-interval", 300, "Snapshot sampling interval (seconds)")
	testSplit := flag.Float64("test-split", 0.2, "Test set fraction")
	lambda := flag.Float64("lambda", 1.0, "L2 regularization strength")
	lr := flag.Float64("lr", 0.01, "Learning rate")
	epochs := flag.Int("epochs", 1000, "Training epochs")
	patchDays := flag.Int("patch-days", 14, "Only compute WPA from matches in the last N days")
	flag.Parse()

	ctx := context.Background()

	db, err := store.New(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Step 1: Train win probability model
	log.Println("=== Training Win Probability Model ===")
	cfg := model.TrainConfig{
		SnapshotInterval: *snapshotInterval,
		TestSplit:        *testSplit,
		Lambda:           *lambda,
		LearningRate:     *lr,
		Epochs:           *epochs,
	}

	result, err := model.TrainFromDB(ctx, db, cfg)
	if err != nil {
		log.Fatalf("Training failed: %v", err)
	}
	if result == nil {
		log.Fatal("No training data available. Run ingest first.")
	}

	log.Printf("Model trained: accuracy=%.4f, ECE=%.4f, train=%d, test=%d",
		result.Accuracy, result.ECE, result.TrainSamples, result.TestSamples)

	// Save model
	weights, err := result.Model.Serialize()
	if err != nil {
		log.Fatalf("Failed to serialize model: %v", err)
	}

	matchCount, err := db.CountMatches(ctx)
	if err != nil {
		log.Fatalf("Failed to count matches: %v", err)
	}

	if err := db.InsertModelMetadata(ctx, domain.ModelMetadata{
		TrainedAt:  time.Now().Unix(),
		Accuracy:   result.Accuracy,
		ECE:        result.ECE,
		NumMatches: matchCount,
		Weights:    weights,
		IsActive:   true,
	}); err != nil {
		log.Fatalf("Failed to save model metadata: %v", err)
	}
	log.Println("Model saved to database")

	// Step 2: Compute WPA for all matches
	log.Println("=== Computing WPA ===")
	if err := db.ClearWPAResults(ctx); err != nil {
		log.Fatalf("Failed to clear WPA results: %v", err)
	}

	// Build set of known shop item IDs to filter non-shop items from WPA
	allItems, err := db.GetAllItems(ctx)
	if err != nil {
		log.Fatalf("Failed to get items: %v", err)
	}
	shopItemIDs := make(map[int64]bool, len(allItems))
	for _, item := range allItems {
		shopItemIDs[item.ID] = true
	}
	log.Printf("Filtering WPA to %d known shop items", len(shopItemIDs))

	patchCutoff := time.Now().AddDate(0, 0, -*patchDays).Unix()
	log.Printf("Using matches from last %d days (since %s)", *patchDays, time.Unix(patchCutoff, 0).Format("2006-01-02"))
	matchIDs, err := db.GetMatchIDsSince(ctx, patchCutoff)
	if err != nil {
		log.Fatalf("Failed to get match IDs: %v", err)
	}

	contexts := wpa.AllContexts()
	var allEvents []wpa.PurchaseEvent
	var allBuilds []builds.PlayerBuild

	for i, matchID := range matchIDs {
		match, err := db.GetMatch(ctx, matchID)
		if err != nil {
			continue
		}
		players, err := db.GetMatchPlayers(ctx, matchID)
		if err != nil {
			continue
		}
		allMatchItems, err := db.GetMatchItems(ctx, matchID)
		if err != nil {
			continue
		}
		// Only include known shop items in WPA computation
		var items []domain.ItemPurchase
		for _, it := range allMatchItems {
			if shopItemIDs[it.ItemID] {
				items = append(items, it)
			}
		}
		snaps, err := db.GetMatchSnapshots(ctx, matchID)
		if err != nil {
			continue
		}

		events := wpa.ComputeMatchWPA(result.Model, match, players, items, snaps)
		allEvents = append(allEvents, events...)

		// Collect final builds for build win rate analysis
		for _, p := range players {
			pb := builds.CollectPlayerBuild(p.HeroID, p.Team, match.WinningTeam, allMatchItems, p.PlayerSlot, shopItemIDs)
			if pb != nil {
				allBuilds = append(allBuilds, *pb)
			}
		}

		if (i+1)%500 == 0 {
			log.Printf("  Processed %d/%d matches (%d events)", i+1, len(matchIDs), len(allEvents))
		}
	}

	log.Printf("Total purchase events: %d", len(allEvents))

	// Aggregate
	results := wpa.AggregateWPA(allEvents, contexts)
	log.Printf("Aggregated into %d hero/item/context results", len(results))

	// Store results
	if err := db.BulkUpsertWPAResults(ctx, results); err != nil {
		log.Fatalf("Failed to store WPA results: %v", err)
	}

	log.Println("=== WPA Computation Complete ===")

	// Step 3: Compute Build Win Rates
	log.Println("=== Computing Build Win Rates ===")
	log.Printf("Collected %d player builds across %d heroes", len(allBuilds), countUniqueHeroes(allBuilds))

	if err := db.ClearBuildResults(ctx); err != nil {
		log.Fatalf("Failed to clear build results: %v", err)
	}

	templates, coverages := builds.ComputeBuildWinRates(allBuilds)
	log.Printf("Generated %d build templates for %d heroes", len(templates), len(coverages))

	if err := db.BulkInsertBuildTemplates(ctx, templates); err != nil {
		log.Fatalf("Failed to store build templates: %v", err)
	}
	if err := db.BulkInsertBuildCoverage(ctx, coverages); err != nil {
		log.Fatalf("Failed to store build coverage: %v", err)
	}

	for _, c := range coverages {
		if c.TotalPlayers >= 50 {
			log.Printf("  Hero %d: coverage=%.1f%% (%d/%d classified)", c.HeroID, c.Coverage*100, c.ClassifiedCount, c.TotalPlayers)
		}
	}

	log.Println("=== Build Computation Complete ===")
}

func countUniqueHeroes(builds []builds.PlayerBuild) int {
	seen := make(map[int]bool)
	for _, b := range builds {
		seen[b.HeroID] = true
	}
	return len(seen)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
