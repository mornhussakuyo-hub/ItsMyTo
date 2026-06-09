package main

import "unicode/utf8"

func batchEmbeddingWork(cfg EmbeddingConfig, work []embeddingWork) [][]embeddingWork {
	batchSize := normalizedBatchSize(cfg.BatchSize)
	maxTokens := normalizedMaxTokens(cfg.MaxTokens)

	var batches [][]embeddingWork
	var current []embeddingWork
	currentTokens := 0
	for _, item := range work {
		tokens := estimateTokens(item.Text)
		shouldFlush := len(current) >= batchSize
		shouldFlush = shouldFlush || (len(current) > 0 && currentTokens+tokens > maxTokens)
		if shouldFlush {
			batches = append(batches, current)
			current = nil
			currentTokens = 0
		}
		current = append(current, item)
		currentTokens += tokens
	}
	if len(current) > 0 {
		batches = append(batches, current)
	}
	return batches
}

func fetchBatchWithFallback(cfg EmbeddingConfig, work []embeddingWork) []embeddingResult {
	inputs := make([]string, 0, len(work))
	for _, item := range work {
		inputs = append(inputs, item.Text)
	}
	vectors, err := fetchEmbeddings(cfg, inputs)
	if err == nil && len(vectors) == len(work) {
		return pairEmbeddingResults(work, vectors)
	}
	if len(work) <= 1 {
		return nil
	}

	mid := len(work) / 2
	results := fetchBatchWithFallback(cfg, work[:mid])
	results = append(results, fetchBatchWithFallback(cfg, work[mid:])...)
	return results
}

func pairEmbeddingResults(work []embeddingWork, vectors [][]float64) []embeddingResult {
	results := make([]embeddingResult, 0, len(work))
	for index, item := range work {
		results = append(results, embeddingResult{
			embeddingWork: item,
			Vector:        vectors[index],
		})
	}
	return results
}

func estimateTokens(text string) int {
	ascii := 0
	nonASCII := 0
	for len(text) > 0 {
		r, size := utf8.DecodeRuneInString(text)
		text = text[size:]
		if r <= 127 {
			ascii++
		} else {
			nonASCII++
		}
	}
	tokens := (ascii + 3) / 4
	tokens += nonASCII
	if tokens < 1 {
		return 1
	}
	return tokens
}
