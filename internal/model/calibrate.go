package model

import (
	"math"
)

// CalibrateModel applies Platt scaling to calibrate predicted probabilities.
// Uses the test set predictions and labels to fit A, B in:
//
//	P(y=1|f) = 1 / (1 + exp(A*f + B))
//
// where f is the raw model output.
func CalibrateModel(m *LogisticModel, X [][]float64, y []float64) {
	n := len(X)
	if n == 0 {
		return
	}

	// Get raw predictions
	preds := make([]float64, n)
	for i := range n {
		preds[i] = m.Predict(X[i])
	}

	// Fit Platt scaling parameters using Newton's method
	// Following Platt (1999) with the improved target values
	nPos := 0.0
	for _, yi := range y {
		if yi > 0.5 {
			nPos++
		}
	}
	nNeg := float64(n) - nPos

	// Target values (regularized)
	tPos := (nPos + 1) / (nPos + 2)
	tNeg := 1.0 / (nNeg + 2)

	targets := make([]float64, n)
	for i, yi := range y {
		if yi > 0.5 {
			targets[i] = tPos
		} else {
			targets[i] = tNeg
		}
	}

	// Initialize parameters
	A := 0.0
	B := math.Log((nNeg + 1) / (nPos + 1))

	// Newton's method
	for iter := range 100 {
		_ = iter
		gradA := 0.0
		gradB := 0.0
		hessAA := 0.0
		hessAB := 0.0
		hessBB := 0.0

		for i := range n {
			fVal := preds[i]
			p := 1.0 / (1.0 + math.Exp(A*fVal+B))
			d1 := targets[i] - p
			d2 := p * (1 - p)

			gradA += fVal * d1
			gradB += d1
			hessAA += fVal * fVal * d2
			hessAB += fVal * d2
			hessBB += d2
		}

		// Avoid singular Hessian
		det := hessAA*hessBB - hessAB*hessAB
		if math.Abs(det) < 1e-10 {
			break
		}

		dA := -(hessBB*gradA - hessAB*gradB) / det
		dB := -(hessAA*gradB - hessAB*gradA) / det

		A += dA
		B += dB

		if math.Abs(dA) < 1e-8 && math.Abs(dB) < 1e-8 {
			break
		}
	}

	m.PlattA = A
	m.PlattB = B
	m.Calibrated = true
}
