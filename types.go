package main

import "time"

const (
	appName         = "ItsMyTo"
	keyFileName     = "master.key"
	cardStoreName   = "cards.json"
	settingsName    = "settings.json"
	maxBodyBytes    = 1 << 20
	similarityFloor = 0.62
)

type Server struct {
	store *Store
}

type StoredCard struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	BaseURL      string    `json:"baseUrl"`
	Description  string    `json:"description"`
	DocURL       string    `json:"docUrl"`
	APIKeyNonce  string    `json:"apiKeyNonce"`
	APIKeyCipher string    `json:"apiKeyCipher"`
	Archived     bool      `json:"archived"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type StoredSettings struct {
	EmbeddingURL          string `json:"embeddingUrl"`
	EmbeddingModel        string `json:"embeddingModel"`
	EmbeddingAPIKeyNonce  string `json:"embeddingApiKeyNonce,omitempty"`
	EmbeddingAPIKeyCipher string `json:"embeddingApiKeyCipher,omitempty"`
	Theme                 string `json:"theme"`
	Autostart             bool   `json:"autostart"`
}

type CardDTO struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	BaseURL     string    `json:"baseUrl"`
	Description string    `json:"description"`
	DocURL      string    `json:"docUrl"`
	APIKeyMask  string    `json:"apiKeyMask"`
	Archived    bool      `json:"archived"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type CardInput struct {
	Name        string `json:"name"`
	BaseURL     string `json:"baseUrl"`
	APIKey      string `json:"apiKey"`
	Description string `json:"description"`
	DocURL      string `json:"docUrl"`
	Archived    *bool  `json:"archived"`
}

type SettingsDTO struct {
	EmbeddingURL       string `json:"embeddingUrl"`
	EmbeddingModel     string `json:"embeddingModel"`
	EmbeddingAPIKey    string `json:"embeddingApiKey,omitempty"`
	HasEmbeddingAPIKey bool   `json:"hasEmbeddingApiKey"`
	ClearEmbeddingKey  bool   `json:"clearEmbeddingKey,omitempty"`
	Theme              string `json:"theme"`
	Autostart          bool   `json:"autostart"`
}

type SearchRequest struct {
	Query          string `json:"query"`
	IncludeArchive bool   `json:"includeArchive"`
}

type OpenURLRequest struct {
	URL string `json:"url"`
}

type EmbeddingConfig struct {
	URL    string
	Model  string
	APIKey string
}

type candidate struct {
	card StoredCard
	key  string
}

type revealResponse struct {
	APIKey string `json:"apiKey"`
}

type errorResponse struct {
	Error string `json:"error"`
}
