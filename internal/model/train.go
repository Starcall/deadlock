package model

import (
	"context"
	"log"
	"math/rand/v2"

	"github.com/name/deadlock/internal/domain"
	"github.com/name/deadlock/internal/store"
)

// TrainConfig controls model training.
type TrainConfig struct {
	SnapshotInterval int     // Sample one snapshot per this many seconds (e.g. 300 = 5 min)
	TestSplit        float64 // Fraction of data for test set (e.g. 0.2)
	Lambda           float64 // L2 regularization strength
	LearningRate     float64 // Gradient descent learning rate
	Epochs           int     // Number of training epochs
}

// DefaultTrainConfig returns sensible defaults.
func DefaultTrainConfig() TrainConfig {
	return TrainConfig{
		SnapshotInterval: 300,
		TestSplit:        0.2,
		Lambda:           1.0,
		LearningRate:     0.01,
		Epochs:           1000,
	}
}

// TrainResult holds the output of model training.
type TrainResult struct {
	Model        *LogisticModel
	TrainSamples int
	TestSamples  int
	Accuracy     float64
	ECE          float64
	Reliability  []ReliabilityBin
}

// TrainFromDB trains a win probability model from stored match data.
func TrainFromDB(ctx context.Context, db *store.DB, cfg TrainConfig) (*TrainResult, error) {
	matchIDs, err := db.GetAllMatchIDs(ctx)
	if err != nil {
		return nil, err
	}
	log.Printf("Training on %d matches", len(matchIDs))

	// Build training dataset
	var X [][]float64
	var y []float64

	for i, matchID := range matchIDs {
		match, err := db.GetMatch(ctx, matchID)
		if err != nil {
			continue
		}
		players, err := db.GetMatchPlayers(ctx, matchID)
		if err != nil {
			continue
		}
		snaps, err := db.GetMatchSnapshots(ctx, matchID)
		if err != nil {
			continue
		}

		if len(players) != 12 || len(snaps) == 0 {
			continue
		}

		// Group snapshots by player slot
		snapsBySlot := groupSnapsBySlot(snaps)

		// Compute average badge
		avgBadge := computeAvgBadge(match)

		// Sample one snapshot per interval
		label := 0.0
		if match.WinningTeam == 0 {
			label = 1.0
		}

		for t := cfg.SnapshotInterval; t < match.DurationS; t += cfg.SnapshotInterval {
			state := BuildGameState(snapsBySlot, players, t, match.DurationS, avgBadge)
			if state == nil {
				continue
			}
			features := ExtractFeatures(state)
			X = append(X, features)
			y = append(y, label)
		}

		if (i+1)%1000 == 0 {
			log.Printf("  Processed %d/%d matches (%d samples)", i+1, len(matchIDs), len(X))
		}
	}

	log.Printf("Total samples: %d", len(X))

	if len(X) == 0 {
		return nil, nil
	}

	// Shuffle and split
	perm := rand.Perm(len(X))
	splitIdx := int(float64(len(X)) * (1 - cfg.TestSplit))

	trainX := make([][]float64, splitIdx)
	trainY := make([]float64, splitIdx)
	testX := make([][]float64, len(X)-splitIdx)
	testY := make([]float64, len(X)-splitIdx)

	for i, pi := range perm {
		if i < splitIdx {
			trainX[i] = X[pi]
			trainY[i] = y[pi]
		} else {
			testX[i-splitIdx] = X[pi]
			testY[i-splitIdx] = y[pi]
		}
	}

	log.Printf("Train: %d, Test: %d", len(trainX), len(testX))

	// Train model
	model := NewLogisticModel(FeatureCount)
	model.Train(trainX, trainY, cfg.Lambda, cfg.LearningRate, cfg.Epochs)

	// Calibrate on test set
	CalibrateModel(model, testX, testY)

	// Evaluate
	acc := ComputeAccuracy(model, testX, testY)
	ece, reliability := ComputeECE(model, testX, testY, 10)

	log.Printf("Test Accuracy: %.4f, ECE: %.4f", acc, ece)

	return &TrainResult{
		Model:        model,
		TrainSamples: len(trainX),
		TestSamples:  len(testX),
		Accuracy:     acc,
		ECE:          ece,
		Reliability:  reliability,
	}, nil
}

func groupSnapsBySlot(snaps []domain.StatSnapshot) map[int][]domain.StatSnapshot {
	m := make(map[int][]domain.StatSnapshot)
	for _, s := range snaps {
		m[s.PlayerSlot] = append(m[s.PlayerSlot], s)
	}
	return m
}

func computeAvgBadge(match *domain.Match) float64 {
	var total, count float64
	if match.AvgBadgeTeam0 != nil {
		total += float64(*match.AvgBadgeTeam0)
		count++
	}
	if match.AvgBadgeTeam1 != nil {
		total += float64(*match.AvgBadgeTeam1)
		count++
	}
	if count == 0 {
		return 50.0 // default
	}
	return total / count
}
