package main

import (
	"context"
	"log"
	"regexp"
	"strings"
	"sync"
)

func (s *Store) StreamSearch(ctx context.Context, query string, includeArchive bool) (<-chan CardDTO, <-chan error) {
	out := make(chan CardDTO)
	errs := make(chan error, 1)
	go func() {
		defer close(out)
		defer close(errs)
		query = strings.TrimSpace(query)
		if query == "" {
			items, err := s.List(includeArchive)
			if err != nil {
				errs <- err
				return
			}
			for _, item := range items {
				sendCard(ctx, out, item)
			}
			return
		}

		candidates, err := s.searchCandidates(includeArchive)
		if err != nil {
			errs <- err
			return
		}
		cfg := s.EmbeddingConfig()

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			streamHardMatches(ctx, query, candidates, out)
		}()
		go func() {
			defer wg.Done()
			streamRegexMatches(ctx, query, candidates, out)
		}()
		if cfg.ready() {
			wg.Add(1)
			go func() {
				defer wg.Done()
				s.streamSemanticMatches(ctx, query, includeArchive, cfg, out)
			}()
		}
		wg.Wait()
	}()
	return out, errs
}

func sendCard(ctx context.Context, out chan<- CardDTO, item CardDTO) bool {
	select {
	case <-ctx.Done():
		return false
	case out <- item:
		return true
	}
}

func streamHardMatches(ctx context.Context, query string, items []candidate, out chan<- CardDTO) {
	q := strings.ToLower(query)
	for _, item := range items {
		fields := []string{item.card.Name, item.key, item.card.BaseURL, item.card.Description}
		for _, field := range fields {
			if strings.Contains(strings.ToLower(field), q) {
				if !sendCard(ctx, out, toDTO(item.card, maskKey(item.key))) {
					return
				}
				break
			}
		}
	}
}

func streamRegexMatches(ctx context.Context, query string, items []candidate, out chan<- CardDTO) {
	pattern, err := regexp.Compile("(?i)" + query)
	if err != nil {
		return
	}
	for _, item := range items {
		fields := []string{item.card.Name, item.key, item.card.Description}
		for _, field := range fields {
			if pattern.MatchString(field) {
				if !sendCard(ctx, out, toDTO(item.card, maskKey(item.key))) {
					return
				}
				break
			}
		}
	}
}

func (s *Store) streamSemanticMatches(ctx context.Context, query string, includeArchive bool, cfg EmbeddingConfig, out chan<- CardDTO) {
	if err := s.PrecomputeEmbeddings(cfg, includeArchive); err != nil {
		log.Printf("embedding precompute skipped: %v", err)
	}
	items, err := s.searchCandidates(includeArchive)
	if err != nil {
		return
	}
	vector, err := queryEmbedding(cfg, query)
	if err != nil {
		log.Printf("embedding query skipped: %v", err)
		return
	}
	for _, item := range items {
		if !cardHasEmbedding(item.card, cfg.Model) {
			continue
		}
		if cosine(vector, item.card.EmbeddingVector) >= similarityFloor {
			if !sendCard(ctx, out, toDTO(item.card, maskKey(item.key))) {
				return
			}
		}
	}
}
