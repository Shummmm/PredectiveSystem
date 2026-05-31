package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type MLPredictRequest struct {
	Features [][]float64 `json:"features"`
}

type MLPredictResponse struct {
	Predictions []float64 `json:"predictions"`
}

type MLFeatureImportanceResponse struct {
	Importance map[string]float64 `json:"importance"`
}

func callMLService(features [][]float64) ([]float64, error) {
	reqBody := MLPredictRequest{Features: features}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post("http://localhost:5001/predict", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result MLPredictResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Predictions, nil
}

func getFeatureImportance() (map[string]float64, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("http://localhost:5001/feature_importance")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result MLFeatureImportanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Importance, nil
}
