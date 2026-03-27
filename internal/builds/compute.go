package builds

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"

	"github.com/name/deadlock/internal/domain"
)

const (
	MinFinalItems        = 4
	MinK                 = 1
	MaxK                 = 12
	MinSampleSize        = 30    // minimum games per cluster to include in results
	MinHeroPlayers       = 30    // skip hero if fewer players
	MergeOverlap         = 0.75  // merge clusters whose templates share ≥75% items
	KMeansRestarts       = 3    // random restarts per k
	KMeansMaxIter        = 50   // max iterations per run
	SilhouetteSampleSize = 500 // max points for silhouette computation
	CentroidThreshold    = 0.3  // item in template if ≥30% of cluster has it
	MaxTemplateItems     = 12   // cap template size
)

// PlayerBuild holds a single player's final item set and match outcome.
type PlayerBuild struct {
	HeroID  int
	ItemIDs []int64 // sorted
	Won     bool
}

// CollectPlayerBuild extracts the final build from a player's items.
// Returns nil if the player has fewer than MinFinalItems.
func CollectPlayerBuild(heroID int, team int, winningTeam int, items []domain.ItemPurchase, playerSlot int, shopItemIDs map[int64]bool) *PlayerBuild {
	var finalItems []int64
	for _, it := range items {
		if it.PlayerSlot != playerSlot {
			continue
		}
		if it.SoldTimeS != 0 {
			continue
		}
		if !shopItemIDs[it.ItemID] {
			continue
		}
		finalItems = append(finalItems, it.ItemID)
	}

	if len(finalItems) < MinFinalItems {
		return nil
	}

	sort.Slice(finalItems, func(i, j int) bool { return finalItems[i] < finalItems[j] })

	return &PlayerBuild{
		HeroID:  heroID,
		ItemIDs: finalItems,
		Won:     team == winningTeam,
	}
}

// itemSetKey returns a string key for a sorted item set.
func itemSetKey(items []int64) string {
	parts := make([]string, len(items))
	for i, id := range items {
		parts[i] = fmt.Sprintf("%d", id)
	}
	return strings.Join(parts, ",")
}

// buildItemIndex scans all builds to create item ID → column index mapping.
func buildItemIndex(builds []PlayerBuild) ([]int64, map[int64]int) {
	seen := make(map[int64]bool)
	for _, b := range builds {
		for _, id := range b.ItemIDs {
			seen[id] = true
		}
	}
	items := make([]int64, 0, len(seen))
	for id := range seen {
		items = append(items, id)
	}
	sort.Slice(items, func(i, j int) bool { return items[i] < items[j] })
	idx := make(map[int64]int, len(items))
	for i, id := range items {
		idx[id] = i
	}
	return items, idx
}

// toBinaryMatrix converts builds into binary vectors.
func toBinaryMatrix(builds []PlayerBuild, itemIndex map[int64]int) [][]float64 {
	cols := len(itemIndex)
	matrix := make([][]float64, len(builds))
	for i, b := range builds {
		row := make([]float64, cols)
		for _, id := range b.ItemIDs {
			if col, ok := itemIndex[id]; ok {
				row[col] = 1.0
			}
		}
		matrix[i] = row
	}
	return matrix
}

// euclideanDistSq returns squared Euclidean distance between two vectors.
func euclideanDistSq(a, b []float64) float64 {
	sum := 0.0
	for i := range a {
		d := a[i] - b[i]
		sum += d * d
	}
	return sum
}

// kmeansPP initializes centroids using K-means++.
func kmeansPP(data [][]float64, k int, rng *rand.Rand) [][]float64 {
	n := len(data)
	dims := len(data[0])
	centroids := make([][]float64, 0, k)

	// First centroid: random point
	first := make([]float64, dims)
	copy(first, data[rng.Intn(n)])
	centroids = append(centroids, first)

	// Distance from each point to nearest centroid
	minDist := make([]float64, n)
	for i := range minDist {
		minDist[i] = euclideanDistSq(data[i], centroids[0])
	}

	for len(centroids) < k {
		// Weighted random selection proportional to minDist
		totalDist := 0.0
		for _, d := range minDist {
			totalDist += d
		}
		if totalDist == 0 {
			// All points are at existing centroids; pick random
			c := make([]float64, dims)
			copy(c, data[rng.Intn(n)])
			centroids = append(centroids, c)
			continue
		}

		target := rng.Float64() * totalDist
		cumulative := 0.0
		chosen := 0
		for i, d := range minDist {
			cumulative += d
			if cumulative >= target {
				chosen = i
				break
			}
		}

		c := make([]float64, dims)
		copy(c, data[chosen])
		centroids = append(centroids, c)

		// Update minDist
		for i := range data {
			d := euclideanDistSq(data[i], c)
			if d < minDist[i] {
				minDist[i] = d
			}
		}
	}

	return centroids
}

// kmeans runs one K-means pass. Returns centroids, assignments, total within-cluster distance.
func kmeans(data [][]float64, k int, maxIter int, rng *rand.Rand) ([][]float64, []int, float64) {
	n := len(data)
	dims := len(data[0])
	centroids := kmeansPP(data, k, rng)
	assignments := make([]int, n)

	for iter := 0; iter < maxIter; iter++ {
		changed := false

		// Assign each point to nearest centroid
		for i, point := range data {
			bestC := 0
			bestD := euclideanDistSq(point, centroids[0])
			for c := 1; c < k; c++ {
				d := euclideanDistSq(point, centroids[c])
				if d < bestD {
					bestD = d
					bestC = c
				}
			}
			if assignments[i] != bestC {
				assignments[i] = bestC
				changed = true
			}
		}

		if !changed {
			break
		}

		// Recompute centroids
		counts := make([]int, k)
		newCentroids := make([][]float64, k)
		for c := 0; c < k; c++ {
			newCentroids[c] = make([]float64, dims)
		}
		for i, point := range data {
			c := assignments[i]
			counts[c]++
			for d := 0; d < dims; d++ {
				newCentroids[c][d] += point[d]
			}
		}
		for c := 0; c < k; c++ {
			if counts[c] == 0 {
				// Empty cluster: re-seed to a random point
				copy(newCentroids[c], data[rng.Intn(n)])
			} else {
				for d := 0; d < dims; d++ {
					newCentroids[c][d] /= float64(counts[c])
				}
			}
		}
		centroids = newCentroids
	}

	// Compute total within-cluster distance
	totalDist := 0.0
	for i, point := range data {
		totalDist += euclideanDistSq(point, centroids[assignments[i]])
	}

	return centroids, assignments, totalDist
}

// kmeansMultiRestart runs K-means multiple times, returns best result.
func kmeansMultiRestart(data [][]float64, k int, restarts int, rng *rand.Rand) ([][]float64, []int) {
	var bestCentroids [][]float64
	var bestAssignments []int
	bestDist := math.MaxFloat64

	for r := 0; r < restarts; r++ {
		c, a, d := kmeans(data, k, KMeansMaxIter, rng)
		if d < bestDist {
			bestDist = d
			bestCentroids = c
			bestAssignments = a
		}
	}
	return bestCentroids, bestAssignments
}

// silhouetteScore computes average silhouette, sampling if n is large.
func silhouetteScore(data [][]float64, assignments []int, k int, rng *rand.Rand) float64 {
	n := len(data)
	if n <= 1 || k <= 1 {
		return 0
	}

	// Sample indices if too many points
	indices := make([]int, n)
	for i := range indices {
		indices[i] = i
	}
	if n > SilhouetteSampleSize {
		rng.Shuffle(n, func(i, j int) { indices[i], indices[j] = indices[j], indices[i] })
		indices = indices[:SilhouetteSampleSize]
	}

	// Group points by cluster for efficient distance computation
	clusterPoints := make([][]int, k)
	for i := 0; i < n; i++ {
		c := assignments[i]
		clusterPoints[c] = append(clusterPoints[c], i)
	}

	totalSil := 0.0
	validCount := 0

	for _, idx := range indices {
		myCluster := assignments[idx]
		myPoints := clusterPoints[myCluster]

		if len(myPoints) <= 1 {
			continue // silhouette undefined for singleton clusters
		}

		// a = mean distance to same-cluster points
		a := 0.0
		for _, j := range myPoints {
			if j != idx {
				a += math.Sqrt(euclideanDistSq(data[idx], data[j]))
			}
		}
		a /= float64(len(myPoints) - 1)

		// b = min mean distance to any other cluster
		b := math.MaxFloat64
		for c := 0; c < k; c++ {
			if c == myCluster || len(clusterPoints[c]) == 0 {
				continue
			}
			meanD := 0.0
			for _, j := range clusterPoints[c] {
				meanD += math.Sqrt(euclideanDistSq(data[idx], data[j]))
			}
			meanD /= float64(len(clusterPoints[c]))
			if meanD < b {
				b = meanD
			}
		}

		if b == math.MaxFloat64 {
			continue
		}

		sil := (b - a) / math.Max(a, b)
		totalSil += sil
		validCount++
	}

	if validCount == 0 {
		return 0
	}
	return totalSil / float64(validCount)
}

// selectK tries k=2..MaxK and returns the k with best silhouette score.
// Returns 1 if no k≥2 produces a silhouette score above the minimum threshold.
func selectK(data [][]float64, rng *rand.Rand) int {
	const minSilhouette = 0.05 // minimum score to justify splitting

	bestK := 1 // default: single cluster
	bestScore := minSilhouette
	declineCount := 0

	for k := 2; k <= MaxK; k++ {
		if k > len(data) {
			break
		}
		_, assignments := kmeansMultiRestart(data, k, KMeansRestarts, rng)
		score := silhouetteScore(data, assignments, k, rng)

		if score > bestScore {
			bestScore = score
			bestK = k
			declineCount = 0
		} else {
			declineCount++
			if declineCount >= 2 {
				break // early termination
			}
		}
	}

	return bestK
}

// templateFromCentroid extracts the most representative items from a centroid.
func templateFromCentroid(centroid []float64, allItems []int64) []int64 {
	type itemWeight struct {
		id     int64
		weight float64
	}
	var candidates []itemWeight
	for i, w := range centroid {
		if w >= CentroidThreshold {
			candidates = append(candidates, itemWeight{allItems[i], w})
		}
	}
	// Sort by weight descending, cap at MaxTemplateItems
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].weight > candidates[j].weight })
	n := len(candidates)
	if n > MaxTemplateItems {
		n = MaxTemplateItems
	}
	result := make([]int64, n)
	for i := 0; i < n; i++ {
		result[i] = candidates[i].id
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

type rankedBuild struct {
	template domain.BuildTemplate
	items    []int64 // parsed template items for overlap comparison
	count    int
}

// parseItemIDs splits a comma-separated string of item IDs into []int64.
func parseItemIDs(s string) []int64 {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]int64, 0, len(parts))
	for _, p := range parts {
		var id int64
		fmt.Sscanf(p, "%d", &id)
		result = append(result, id)
	}
	return result
}

// mergeOverlappingBuilds merges clusters whose templates have Jaccard overlap ≥ MergeOverlap.
// The larger cluster absorbs the smaller one; the larger's template is kept.
func mergeOverlappingBuilds(builds []rankedBuild) []rankedBuild {
	if len(builds) <= 1 {
		return builds
	}

	merged := make([]bool, len(builds))
	for i := 0; i < len(builds); i++ {
		if merged[i] {
			continue
		}
		for j := i + 1; j < len(builds); j++ {
			if merged[j] {
				continue
			}
			if templateJaccard(builds[i].items, builds[j].items) >= MergeOverlap {
				// Absorb j into i
				builds[i].template.FuzzyCount += builds[j].template.FuzzyCount
				builds[i].template.ExactCount += builds[j].template.ExactCount
				builds[i].template.Wins += builds[j].template.Wins
				builds[i].template.Losses += builds[j].template.Losses
				builds[i].count += builds[j].count
				if builds[i].template.FuzzyCount > 0 {
					builds[i].template.WinRate = float64(builds[i].template.Wins) / float64(builds[i].template.FuzzyCount)
				}
				merged[j] = true
			}
		}
	}

	var result []rankedBuild
	for i, b := range builds {
		if !merged[i] {
			result = append(result, b)
		}
	}
	return result
}

// templateJaccard computes |A ∩ B| / |A ∪ B| for two sorted item slices.
func templateJaccard(a, b []int64) float64 {
	setA := make(map[int64]bool, len(a))
	for _, id := range a {
		setA[id] = true
	}
	intersection := 0
	union := len(a)
	for _, id := range b {
		if setA[id] {
			intersection++
		} else {
			union++
		}
	}
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

// ComputeBuildWinRates clusters player builds per hero using K-means and returns templates + coverage.
func ComputeBuildWinRates(allBuilds []PlayerBuild) ([]domain.BuildTemplate, []domain.HeroBuildCoverage) {
	// Group builds by hero
	heroBuilds := make(map[int][]PlayerBuild)
	for _, b := range allBuilds {
		heroBuilds[b.HeroID] = append(heroBuilds[b.HeroID], b)
	}

	var allTemplates []domain.BuildTemplate
	var allCoverage []domain.HeroBuildCoverage

	for heroID, builds := range heroBuilds {
		totalPlayers := len(builds)

		if totalPlayers < MinHeroPlayers {
			allCoverage = append(allCoverage, domain.HeroBuildCoverage{
				HeroID:          heroID,
				TotalPlayers:    totalPlayers,
				ClassifiedCount: 0,
				Coverage:        0,
			})
			continue
		}

		// Build binary item matrix
		allItems, itemIndex := buildItemIndex(builds)
		data := toBinaryMatrix(builds, itemIndex)

		// Deterministic RNG per hero
		rng := rand.New(rand.NewSource(int64(heroID)))

		// Find optimal k
		k := selectK(data, rng)

		// Final clustering with best k
		var centroids [][]float64
		var assignments []int
		if k == 1 {
			// Single cluster: centroid = mean of all points
			dims := len(data[0])
			centroid := make([]float64, dims)
			for _, row := range data {
				for d := 0; d < dims; d++ {
					centroid[d] += row[d]
				}
			}
			n := float64(len(data))
			for d := range centroid {
				centroid[d] /= n
			}
			centroids = [][]float64{centroid}
			assignments = make([]int, len(data))
			// all zeros by default
		} else {
			centroids, assignments = kmeansMultiRestart(data, k, KMeansRestarts, rng)
		}

		// Compute stats per cluster
		type clusterStats struct {
			total int
			wins  int
		}
		stats := make([]clusterStats, k)
		for i, b := range builds {
			c := assignments[i]
			stats[c].total++
			if b.Won {
				stats[c].wins++
			}
		}

		// Extract templates and filter small clusters
		var ranked []rankedBuild

		classifiedCount := 0
		for c := 0; c < k; c++ {
			if stats[c].total < MinSampleSize {
				continue
			}

			templateItems := templateFromCentroid(centroids[c], allItems)
			if len(templateItems) == 0 {
				continue
			}

			// Count exact matches: players whose items are a superset of template
			templateSet := make(map[int64]bool, len(templateItems))
			for _, id := range templateItems {
				templateSet[id] = true
			}
			exactCount := 0
			for i, b := range builds {
				if assignments[i] != c {
					continue
				}
				match := 0
				for _, id := range b.ItemIDs {
					if templateSet[id] {
						match++
					}
				}
				if match == len(templateItems) {
					exactCount++
				}
			}

			winRate := 0.0
			if stats[c].total > 0 {
				winRate = float64(stats[c].wins) / float64(stats[c].total)
			}
			losses := stats[c].total - stats[c].wins

			classifiedCount += stats[c].total

			ranked = append(ranked, rankedBuild{
				template: domain.BuildTemplate{
					HeroID:           heroID,
					ItemIDs:          itemSetKey(templateItems),
					ExactCount:       exactCount,
					FuzzyCount:       stats[c].total,
					Wins:             stats[c].wins,
					Losses:           losses,
					WinRate:          winRate,
					TotalHeroPlayers: totalPlayers,
				},
				items: templateItems,
				count: stats[c].total,
			})
		}

		// Merge clusters with highly overlapping templates
		ranked = mergeOverlappingBuilds(ranked)

		// Sort by popularity, assign ranks
		sort.Slice(ranked, func(i, j int) bool { return ranked[i].count > ranked[j].count })
		for i := range ranked {
			ranked[i].template.BuildRank = i + 1
			allTemplates = append(allTemplates, ranked[i].template)
		}

		coverage := 0.0
		if totalPlayers > 0 {
			coverage = float64(classifiedCount) / float64(totalPlayers)
		}
		allCoverage = append(allCoverage, domain.HeroBuildCoverage{
			HeroID:          heroID,
			TotalPlayers:    totalPlayers,
			ClassifiedCount: classifiedCount,
			Coverage:        coverage,
		})
	}

	return allTemplates, allCoverage
}
