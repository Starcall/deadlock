package model

import (
	"encoding/json"
	"fmt"
	"math"
)

// LogisticModel implements L2-regularized logistic regression.
type LogisticModel struct {
	Weights []float64 `json:"weights"` // Feature weights
	Bias    float64   `json:"bias"`    // Intercept

	// Platt scaling parameters (post-calibration)
	PlattA float64 `json:"platt_a"`
	PlattB float64 `json:"platt_b"`

	FeatureCount int  `json:"feature_count"`
	Calibrated   bool `json:"calibrated"`
}

// NewLogisticModel creates a new logistic regression model.
func NewLogisticModel(featureCount int) *LogisticModel {
	return &LogisticModel{
		Weights:      make([]float64, featureCount),
		Bias:         0,
		FeatureCount: featureCount,
	}
}

// Predict returns the raw sigmoid probability for a feature vector.
func (m *LogisticModel) Predict(features []float64) float64 {
	z := m.Bias
	for i, w := range m.Weights {
		if i < len(features) {
			z += w * features[i]
		}
	}
	return sigmoid(z)
}

// PredictCalibrated returns the calibrated probability using Platt scaling.
func (m *LogisticModel) PredictCalibrated(features []float64) float64 {
	if !m.Calibrated {
		return m.Predict(features)
	}
	raw := m.Predict(features)
	// Platt scaling: P(y=1|f) = 1 / (1 + exp(A*f + B))
	return 1.0 / (1.0 + math.Exp(m.PlattA*raw+m.PlattB))
}

// Train fits the model using gradient descent with L2 regularization.
func (m *LogisticModel) Train(X [][]float64, y []float64, lambda float64, lr float64, epochs int) {
	n := len(X)
	if n == 0 {
		return
	}

	for epoch := range epochs {
		// Compute gradients
		gradW := make([]float64, m.FeatureCount)
		gradB := 0.0

		totalLoss := 0.0

		for i := range n {
			pred := m.Predict(X[i])
			err := pred - y[i]

			gradB += err
			for j := range m.FeatureCount {
				if j < len(X[i]) {
					gradW[j] += err * X[i][j]
				}
			}

			// Log loss for monitoring
			if pred > 1e-15 && pred < 1-1e-15 {
				totalLoss += -y[i]*math.Log(pred) - (1-y[i])*math.Log(1-pred)
			}
		}

		// Update weights with L2 regularization
		for j := range m.FeatureCount {
			gradW[j] = gradW[j]/float64(n) + lambda*m.Weights[j]
			m.Weights[j] -= lr * gradW[j]
		}
		m.Bias -= lr * gradB / float64(n)

		if (epoch+1)%100 == 0 {
			avgLoss := totalLoss / float64(n)
			_ = avgLoss // Available for logging if needed
		}
	}
}

// Serialize encodes the model to JSON bytes.
func (m *LogisticModel) Serialize() ([]byte, error) {
	return json.Marshal(m)
}

// DeserializeModel decodes a model from JSON bytes.
func DeserializeModel(data []byte) (*LogisticModel, error) {
	var m LogisticModel
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("deserializing model: %w", err)
	}
	return &m, nil
}

func sigmoid(z float64) float64 {
	if z > 500 {
		return 1.0
	}
	if z < -500 {
		return 0.0
	}
	return 1.0 / (1.0 + math.Exp(-z))
}
