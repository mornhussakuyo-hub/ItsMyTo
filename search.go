package main

import (
	"regexp"
	"sort"
	"strings"
	"sync"
)

func (s *Store) Search(query string, includeArchive bool) ([]CardDTO, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return s.List(includeArchive)
	}

	cfg := s.EmbeddingConfig()
	if cfg.ready() {
		_ = s.PrecomputeEmbeddings(cfg, includeArchive)
	}

	candidates, err := s.searchCandidates(includeArchive)
	if err != nil {
		return nil, err
	}

	var mu sync.Mutex
	scoreByID := map[string]float64{}
	add := func(id string, score float64) {
		mu.Lock()
		defer mu.Unlock()
		if score > scoreByID[id] {
			scoreByID[id] = score
		}
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		hardMatch(query, candidates, add)
	}()
	go func() {
		defer wg.Done()
		regexMatch(query, candidates, add)
	}()

	if cfg.ready() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			semanticMatch(query, candidates, cfg, add)
		}()
	}
	wg.Wait()
	return s.searchResults(candidates, scoreByID), nil
}

func (s *Store) searchCandidates(includeArchive bool) ([]candidate, error) {
	s.mu.RLock()
	cards := append([]StoredCard(nil), s.cards...)
	s.mu.RUnlock()

	items := make([]candidate, 0, len(cards))
	for _, card := range cards {
		if card.Archived && !includeArchive {
			continue
		}
		key, err := s.decryptCard(card)
		if err != nil {
			return nil, err
		}
		items = append(items, candidate{card: card, key: key})
	}
	return items, nil
}

func (s *Store) searchResults(items []candidate, scoreByID map[string]float64) []CardDTO {
	results := make([]CardDTO, 0, len(scoreByID))
	for _, item := range items {
		if _, ok := scoreByID[item.card.ID]; !ok {
			continue
		}
		results = append(results, toDTO(item.card, maskKey(item.key)))
	}
	sort.Slice(results, func(i, j int) bool {
		left := scoreByID[results[i].ID]
		right := scoreByID[results[j].ID]
		if left == right {
			return results[i].UpdatedAt.After(results[j].UpdatedAt)
		}
		return left > right
	})
	return results
}

func hardMatch(query string, items []candidate, add func(string, float64)) {
	q := strings.ToLower(query)
	for _, item := range items {
		fields := []string{item.card.Name, item.key, item.card.BaseURL, item.card.Description}
		for index, field := range fields {
			if strings.Contains(strings.ToLower(field), q) {
				score := 100.0 - float64(index)
				if strings.EqualFold(item.card.Name, query) {
					score = 130
				}
				add(item.card.ID, score)
				break
			}
		}
	}
}

func regexMatch(query string, items []candidate, add func(string, float64)) {
	pattern, err := regexp.Compile("(?i)" + query)
	if err != nil {
		return
	}
	for _, item := range items {
		fields := []string{item.card.Name, item.key, item.card.Description}
		for index, field := range fields {
			if pattern.MatchString(field) {
				add(item.card.ID, 80-float64(index))
				break
			}
		}
	}
}
