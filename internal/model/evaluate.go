package model

import (
	"math"
)

// ReliabilityBin is one bin in a reliability diagram.
type ReliabilityBin struct {
	BinStart      float64 `json:"bin_start"`
	BinEnd        float64 `json:"bin_end"`
	MeanPredicted float64 `json:"mean_predicted"`
	MeanActual    float64 `json:"mean_actual"`
	Count         int     `json:"count"`
}

// ComputeAccuracy returns classification accuracy on test data.
func ComputeAccuracy(m *LogisticModel, X [][]float64, y []float64) float64 {
	if len(X) == 0 {
		return 0
	}
	correct := 0
	for i := range X {
		pred := m.PredictCalibrated(X[i])
		predicted := 0.0
		if pred >= 0.5 {
			predicted = 1.0
		}
		if predicted == y[i] {
			correct++
		}
	}
	return float64(correct) / float64(len(X))
}

// ComputeECE computes Expected Calibration Error and returns reliability bins.
func ComputeECE(m *LogisticModel, X [][]float64, y []float64, numBins int) (float64, []ReliabilityBin) {
	if len(X) == 0 {
		return 0, nil
	}

	bins := make([]ReliabilityBin, numBins)
	binWidth := 1.0 / float64(numBins)

	// Initialize bins
	for i := range numBins {
		bins[i].BinStart = float64(i) * binWidth
		bins[i].BinEnd = float64(i+1) * binWidth
	}

	// Accumulate predictions and actuals per bin
	binPredSum := make([]float64, numBins)
	binActSum := make([]float64, numBins)

	for i := range X {
		pred := m.PredictCalibrated(X[i])
		binIdx := int(pred / binWidth)
		if binIdx >= numBins {
			binIdx = numBins - 1
		}
		if binIdx < 0 {
			binIdx = 0
		}

		binPredSum[binIdx] += pred
		binActSum[binIdx] += y[i]
		bins[binIdx].Count++
	}

	// Compute ECE
	ece := 0.0
	n := float64(len(X))
	for i := range numBins {
		if bins[i].Count == 0 {
			continue
		}
		cnt := float64(bins[i].Count)
		bins[i].MeanPredicted = binPredSum[i] / cnt
		bins[i].MeanActual = binActSum[i] / cnt
		ece += (cnt / n) * math.Abs(bins[i].MeanPredicted-bins[i].MeanActual)
	}

	return ece, bins
}
