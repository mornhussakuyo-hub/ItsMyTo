package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

type embeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}

func semanticMatch(query string, items []candidate, cfg EmbeddingConfig, add func(string, float64)) {
	inputs := make([]string, 0, len(items)+1)
	inputs = append(inputs, query)
	for _, item := range items {
		inputs = append(inputs, item.card.Name+"\n"+item.card.Description)
	}
	vectors, err := fetchEmbeddings(cfg, inputs)
	if err != nil || len(vectors) != len(inputs) {
		return
	}
	queryVector := vectors[0]
	for i, item := range items {
		score := cosine(queryVector, vectors[i+1])
		if score >= similarityFloor {
			add(item.card.ID, 60+score)
		}
	}
}

func fetchEmbeddings(cfg EmbeddingConfig, inputs []string) ([][]float64, error) {
	body, err := json.Marshal(embeddingRequest{Model: cfg.Model, Input: inputs})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, cfg.URL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}

	client := http.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("embedding service status %d", resp.StatusCode)
	}

	var parsed embeddingResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxBodyBytes)).Decode(&parsed); err != nil {
		return nil, err
	}
	vectors := make([][]float64, 0, len(parsed.Data))
	for _, item := range parsed.Data {
		vectors = append(vectors, item.Embedding)
	}
	return vectors, nil
}

func cosine(a, b []float64) float64 {
	if len(a) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, na, nb float64
	for i := range a {
		dot += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}
