package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
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
	queryVector, err := queryEmbedding(cfg, query)
	if err != nil {
		return
	}
	for _, item := range items {
		if !cardHasEmbedding(item.card, cfg.Model) {
			continue
		}
		score := cosine(queryVector, item.card.EmbeddingVector)
		if score >= similarityFloor {
			add(item.card.ID, 60+score)
		}
	}
}

func queryEmbedding(cfg EmbeddingConfig, query string) ([]float64, error) {
	vectors, err := fetchEmbeddings(cfg, []string{query})
	if err != nil {
		return nil, err
	}
	if len(vectors) != 1 {
		return nil, fmt.Errorf("embedding query returned %d vectors", len(vectors))
	}
	return vectors[0], nil
}

func fetchEmbeddings(cfg EmbeddingConfig, inputs []string) ([][]float64, error) {
	body, err := json.Marshal(embeddingRequest{Model: cfg.Model, Input: inputs})
	if err != nil {
		return nil, err
	}
	endpoint, err := embeddingEndpoint(cfg.URL)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
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
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("embedding service status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var parsed embeddingResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, embeddingMaxBodyBytes)).Decode(&parsed); err != nil {
		return nil, err
	}
	vectors := make([][]float64, 0, len(parsed.Data))
	for _, item := range parsed.Data {
		vectors = append(vectors, item.Embedding)
	}
	return vectors, nil
}

func embeddingEndpoint(rawURL string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("invalid embedding url")
	}
	path := strings.TrimRight(parsed.Path, "/")
	if !strings.HasSuffix(path, "/embeddings") {
		path += "/embeddings"
	}
	parsed.Path = path
	return parsed.String(), nil
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
