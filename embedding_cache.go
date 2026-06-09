package main

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
)

type embeddingWork struct {
	ID   string
	Hash string
	Text string
}

type embeddingResult struct {
	embeddingWork
	Vector []float64
}

func (s *Store) precomputeAllIfConfigured(includeArchive bool) {
	cfg := s.EmbeddingConfig()
	if !cfg.ready() {
		return
	}
	if err := s.PrecomputeEmbeddings(cfg, includeArchive); err != nil {
		log.Printf("embedding precompute skipped: %v", err)
	}
}

func (s *Store) PrecomputeEmbeddings(cfg EmbeddingConfig, includeArchive bool) error {
	if !cfg.ready() {
		return nil
	}
	work := s.embeddingWork(cfg.Model, includeArchive)
	if len(work) == 0 {
		return nil
	}

	results := fetchWorkEmbeddings(cfg, work)
	if len(results) == 0 {
		return nil
	}
	return s.saveEmbeddingVectors(cfg.Model, results)
}

func fetchWorkEmbeddings(cfg EmbeddingConfig, work []embeddingWork) []embeddingResult {
	results := make([]embeddingResult, 0, len(work))
	for _, batch := range batchEmbeddingWork(cfg, work) {
		results = append(results, fetchBatchWithFallback(cfg, batch)...)
	}
	return results
}

func (s *Store) embeddingWork(model string, includeArchive bool) []embeddingWork {
	s.mu.RLock()
	defer s.mu.RUnlock()

	work := make([]embeddingWork, 0, len(s.cards))
	for _, card := range s.cards {
		if card.Archived && !includeArchive {
			continue
		}
		if !cardNeedsEmbedding(card, model) {
			continue
		}
		work = append(work, embeddingWork{
			ID:   card.ID,
			Hash: cardEmbeddingHash(card),
			Text: cardEmbeddingText(card),
		})
	}
	return work
}

func (s *Store) saveEmbeddingVectors(model string, results []embeddingResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	changed := false
	for _, item := range results {
		for cardIndex, card := range s.cards {
			if card.ID != item.ID || cardEmbeddingHash(card) != item.Hash {
				continue
			}
			s.cards[cardIndex].EmbeddingModel = model
			s.cards[cardIndex].EmbeddingHash = item.Hash
			s.cards[cardIndex].EmbeddingVector = append([]float64(nil), item.Vector...)
			changed = true
			break
		}
	}
	if !changed {
		return nil
	}
	return s.saveCardsLocked()
}

func cardNeedsEmbedding(card StoredCard, model string) bool {
	return card.EmbeddingModel != model ||
		card.EmbeddingHash != cardEmbeddingHash(card) ||
		len(card.EmbeddingVector) == 0
}

func cardHasEmbedding(card StoredCard, model string) bool {
	return !cardNeedsEmbedding(card, model)
}

func cardEmbeddingText(card StoredCard) string {
	return clean(card.Name) + "\n" + clean(card.Description)
}

func cardEmbeddingHash(card StoredCard) string {
	sum := sha256.Sum256([]byte(cardEmbeddingText(card)))
	return hex.EncodeToString(sum[:])
}

func (cfg EmbeddingConfig) ready() bool {
	return cfg.URL != "" && cfg.Model != ""
}
