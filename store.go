package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type Store struct {
	mu       sync.RWMutex
	dir      string
	cardPath string
	setPath  string
	aead     cipher.AEAD
	cards    []StoredCard
	settings StoredSettings
}

func NewStore() (*Store, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(configDir, appName)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	key, err := loadOrCreateKey(dir)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	store := &Store{
		dir:      dir,
		cardPath: filepath.Join(dir, cardStoreName),
		setPath:  filepath.Join(dir, settingsName),
		aead:     aead,
		settings: StoredSettings{Theme: "system"},
	}
	if err := store.load(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *Store) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := readJSONFile(s.cardPath, &s.cards); err != nil {
		return err
	}
	if err := readJSONFile(s.setPath, &s.settings); err != nil {
		return err
	}
	if s.settings.Theme == "" {
		s.settings.Theme = "system"
	}
	return nil
}

func readJSONFile(path string, value any) error {
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, value)
}

func (s *Store) saveCardsLocked() error {
	raw, err := json.MarshalIndent(s.cards, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(s.cardPath, raw, 0600)
}

func (s *Store) saveSettingsLocked() error {
	raw, err := json.MarshalIndent(s.settings, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(s.setPath, raw, 0600)
}

func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
