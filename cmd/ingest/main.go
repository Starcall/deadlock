package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/name/deadlock/internal/deadlockapi"
	"github.com/name/deadlock/internal/store"
)

func main() {
	dbPath := flag.String("db", envOr("DB_PATH", "./deadlock.db"), "SQLite database path")
	count := flag.Int("count", 1000, "Number of matches to ingest")
	rateLimit := flag.Int("rate", 8, "API requests per second (API limit: 100/10s)")
	minDuration := flag.Int("min-duration", 900, "Minimum match duration in seconds")
	workers := flag.Int("workers", 8, "Number of concurrent fetch workers")
	patchDays := flag.Int("patch-days", 14, "Only ingest matches from the last N days (current patch window)")
	flag.Parse()

	ctx := context.Background()

	db, err := store.New(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	client := deadlockapi.NewClient(*rateLimit)

	// Step 1: Seed heroes and items
	log.Println("Seeding heroes and items...")
	if err := seedAssets(ctx, client, db); err != nil {
		log.Fatalf("Failed to seed assets: %v", err)
	}

	// Step 2: Discover match IDs
	patchCutoff := time.Now().AddDate(0, 0, -*patchDays).Unix()
	log.Printf("Discovering match IDs (target: %d, patch window: last %d days)...\n", *count, *patchDays)
	matchIDs, err := discoverMatches(ctx, client, db, *count, *minDuration, patchCutoff)
	if err != nil {
		log.Fatalf("Failed to discover matches: %v", err)
	}
	log.Printf("Found %d new match IDs to fetch\n", len(matchIDs))

	if len(matchIDs) == 0 {
		log.Println("No new matches to ingest")
		printStats(ctx, db)
		return
	}

	// Step 3: Fetch match details concurrently
	log.Printf("Fetching match details with %d workers...\n", *workers)
	fetchMatches(ctx, client, db, matchIDs, *workers)

	printStats(ctx, db)
}

func seedAssets(ctx context.Context, client *deadlockapi.Client, db *store.DB) error {
	heroes, err := client.FetchHeroes(ctx)
	if err != nil {
		return fmt.Errorf("fetching heroes: %w", err)
	}
	for _, h := range heroes {
		if err := db.InsertHero(ctx, h); err != nil {
			return fmt.Errorf("inserting hero %s: %w", h.Name, err)
		}
	}
	log.Printf("  Seeded %d heroes\n", len(heroes))

	items, err := client.FetchItems(ctx)
	if err != nil {
		return fmt.Errorf("fetching items: %w", err)
	}
	for _, item := range items {
		if err := db.InsertItem(ctx, item); err != nil {
			return fmt.Errorf("inserting item %s: %w", item.Name, err)
		}
	}
	log.Printf("  Seeded %d shop items\n", len(items))

	return nil
}

func discoverMatches(ctx context.Context, client *deadlockapi.Client, db *store.DB, target, minDuration int, patchCutoff int64) ([]int64, error) {
	var allIDs []int64
	seen := make(map[int64]bool)

	maxID, err := db.GetMaxMatchID(ctx)
	if err != nil {
		return nil, err
	}

	// Start from a recent window and expand backwards if needed.
	// The API returns ~1000 matches per page sorted by match_id ascending,
	// so starting from patchCutoff (14 days ago) would take hundreds of pages
	// to reach today. Instead, start from a few hours ago.
	windowStart := time.Now().Add(-6 * time.Hour).Unix()
	if windowStart < patchCutoff {
		windowStart = patchCutoff
	}

	for len(allIDs) < target {
		q := deadlockapi.MatchQuery{
			MinDurationS: minDuration,
			MinUnixTime:  windowStart,
		}
		if maxID > 0 {
			q.MinMatchID = maxID + 1
		}

		metas, err := client.FetchMatchMetadata(ctx, q)
		if err != nil {
			return nil, fmt.Errorf("fetching metadata: %w", err)
		}

		if len(metas) == 0 {
			// No more matches in this window — expand backwards
			newStart := windowStart - 6*3600 // 6 more hours back
			if newStart < patchCutoff {
				newStart = patchCutoff
			}
			if newStart == windowStart {
				break // Already at patch cutoff, no more matches
			}
			windowStart = newStart
			maxID = 0 // Reset pagination for new window
			log.Printf("  Expanding window to %s...", time.Unix(windowStart, 0).Format("2006-01-02 15:04"))
			continue
		}

		ids := deadlockapi.ParseBulkMatchIDs(metas, minDuration)
		for _, id := range ids {
			if seen[id] {
				continue
			}
			exists, err := db.MatchExists(ctx, id)
			if err != nil {
				return nil, err
			}
			if exists {
				seen[id] = true
				continue
			}
			seen[id] = true
			allIDs = append(allIDs, id)
		}

		// Move window forward
		var maxInBatch int64
		for _, m := range metas {
			if m.MatchID > maxInBatch {
				maxInBatch = m.MatchID
			}
		}
		if maxInBatch <= maxID {
			// No progress in current window — expand backwards
			newStart := windowStart - 6*3600
			if newStart < patchCutoff {
				newStart = patchCutoff
			}
			if newStart == windowStart {
				break
			}
			windowStart = newStart
			maxID = 0
			log.Printf("  Expanding window to %s...", time.Unix(windowStart, 0).Format("2006-01-02 15:04"))
			continue
		}
		maxID = maxInBatch

		log.Printf("  Discovered %d/%d match IDs...\n", len(allIDs), target)

		if len(allIDs) >= target {
			allIDs = allIDs[:target]
			break
		}
	}

	return allIDs, nil
}

func fetchMatches(ctx context.Context, client *deadlockapi.Client, db *store.DB, matchIDs []int64, numWorkers int) {
	var (
		fetched atomic.Int64
		failed  atomic.Int64
		total   = len(matchIDs)
	)

	jobs := make(chan int64, numWorkers*2)
	var wg sync.WaitGroup

	start := time.Now()

	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for matchID := range jobs {
				info, err := client.FetchMatchDetail(ctx, matchID)
				if err != nil {
					log.Printf("  Failed to fetch match %d: %v", matchID, err)
					failed.Add(1)
					continue
				}

				match, players, items, snapshots := deadlockapi.ConvertMatch(info)
				if err := db.InsertMatch(ctx, match, players, items, snapshots); err != nil {
					log.Printf("  Failed to store match %d: %v", matchID, err)
					failed.Add(1)
					continue
				}

				n := fetched.Add(1)
				if n%100 == 0 {
					elapsed := time.Since(start)
					rate := float64(n) / elapsed.Seconds()
					log.Printf("  Progress: %d/%d (%.1f/s, %d failed)", n, total, rate, failed.Load())
				}
			}
		}()
	}

	for _, id := range matchIDs {
		jobs <- id
	}
	close(jobs)
	wg.Wait()

	elapsed := time.Since(start)
	log.Printf("Fetched %d matches in %s (%d failed)", fetched.Load(), elapsed.Round(time.Second), failed.Load())
}

func printStats(ctx context.Context, db *store.DB) {
	count, err := db.CountMatches(ctx)
	if err != nil {
		log.Printf("Failed to count matches: %v", err)
		return
	}
	log.Printf("Database now contains %d matches", count)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
