package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"sort"
	"strings"
)

func toDTO(card StoredCard, keyMask string) CardDTO {
	return CardDTO{
		ID:          card.ID,
		Name:        card.Name,
		BaseURL:     card.BaseURL,
		Description: card.Description,
		DocURL:      card.DocURL,
		APIKeyMask:  keyMask,
		Archived:    card.Archived,
		CreatedAt:   card.CreatedAt,
		UpdatedAt:   card.UpdatedAt,
	}
}

func sortCards(items []CardDTO) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Archived != items[j].Archived {
			return !items[i].Archived
		}
		return items[i].UpdatedAt.After(items[j].UpdatedAt)
	})
}

func maskKey(key string) string {
	runes := []rune(key)
	if len(runes) <= 8 {
		return strings.Repeat("•", max(4, len(runes)))
	}
	return string(runes[:4]) + strings.Repeat("•", 8) + string(runes[len(runes)-4:])
}

func clean(value string) string {
	return strings.TrimSpace(value)
}

func newID() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return fmt.Sprintf("%d", len(raw))
	}
	return base64.RawURLEncoding.EncodeToString(raw[:])
}

func openURL(rawURL string) {
	_ = openExternalURL(rawURL)
}

func openExternalURL(rawURL string) error {
	target, err := normalizeExternalURL(rawURL)
	if err != nil {
		return err
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", target)
	case "darwin":
		cmd = exec.Command("open", target)
	default:
		cmd = exec.Command("xdg-open", target)
	}
	return cmd.Start()
}

func normalizeExternalURL(rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", errors.New("url is required")
	}
	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return "", errors.New("invalid url")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errors.New("only http and https urls are allowed")
	}
	return parsed.String(), nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
