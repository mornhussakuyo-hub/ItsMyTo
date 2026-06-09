package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func loadOrCreateKey(dir string) ([]byte, error) {
	if encoded := strings.TrimSpace(os.Getenv("ITSMYTO_MASTER_KEY")); encoded != "" {
		key, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil || len(key) != 32 {
			return nil, errors.New("ITSMYTO_MASTER_KEY must be base64 encoded 32 bytes")
		}
		return key, nil
	}

	path := filepath.Join(dir, keyFileName)
	raw, err := os.ReadFile(path)
	if err == nil {
		key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(raw)))
		if err != nil || len(key) != 32 {
			return nil, errors.New("invalid local encryption key")
		}
		_ = os.Chmod(path, 0600)
		return key, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	encoded := base64.StdEncoding.EncodeToString(key)
	return key, os.WriteFile(path, []byte(encoded), 0600)
}

func (s *Store) encrypt(value string) (string, string, error) {
	nonce := make([]byte, s.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", "", err
	}
	ciphertext := s.aead.Seal(nil, nonce, []byte(value), nil)
	return base64.StdEncoding.EncodeToString(nonce), base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *Store) decrypt(nonceValue, cipherValue string) (string, error) {
	nonce, err := base64.StdEncoding.DecodeString(nonceValue)
	if err != nil {
		return "", err
	}
	ciphertext, err := base64.StdEncoding.DecodeString(cipherValue)
	if err != nil {
		return "", err
	}
	plain, err := s.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func (s *Store) decryptCard(card StoredCard) (string, error) {
	return s.decrypt(card.APIKeyNonce, card.APIKeyCipher)
}
