package main

import (
	"errors"
	"os"
	"strings"
	"time"
)

func (s *Store) List(includeArchive bool) ([]CardDTO, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.listLocked(includeArchive)
}

func (s *Store) listLocked(includeArchive bool) ([]CardDTO, error) {
	items := make([]CardDTO, 0, len(s.cards))
	for _, card := range s.cards {
		if card.Archived && !includeArchive {
			continue
		}
		key, err := s.decryptCard(card)
		if err != nil {
			return nil, err
		}
		items = append(items, toDTO(card, maskKey(key)))
	}
	sortCards(items)
	return items, nil
}

func (s *Store) Create(input CardInput) (CardDTO, error) {
	if err := validateCard(input, true); err != nil {
		return CardDTO{}, err
	}
	nonce, ciphertext, err := s.encrypt(input.APIKey)
	if err != nil {
		return CardDTO{}, err
	}
	now := time.Now()
	card := StoredCard{
		ID:           newID(),
		Name:         clean(input.Name),
		BaseURL:      clean(input.BaseURL),
		Description:  clean(input.Description),
		DocURL:       clean(input.DocURL),
		APIKeyNonce:  nonce,
		APIKeyCipher: ciphertext,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.cards = append(s.cards, card)
	return toDTO(card, maskKey(input.APIKey)), s.saveCardsLocked()
}

func (s *Store) Update(id string, input CardInput) (CardDTO, error) {
	if err := validateCard(input, false); err != nil {
		return CardDTO{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, card := range s.cards {
		if card.ID != id {
			continue
		}
		updated, err := s.applyCardUpdate(card, input)
		if err != nil {
			return CardDTO{}, err
		}
		s.cards[i] = updated
		if err := s.saveCardsLocked(); err != nil {
			return CardDTO{}, err
		}
		key, err := s.decryptCard(updated)
		return toDTO(updated, maskKey(key)), err
	}
	return CardDTO{}, os.ErrNotExist
}

func (s *Store) applyCardUpdate(card StoredCard, input CardInput) (StoredCard, error) {
	card.Name = clean(input.Name)
	card.BaseURL = clean(input.BaseURL)
	card.Description = clean(input.Description)
	card.DocURL = clean(input.DocURL)
	if input.Archived != nil {
		card.Archived = *input.Archived
	}
	if input.APIKey != "" {
		nonce, ciphertext, err := s.encrypt(input.APIKey)
		if err != nil {
			return StoredCard{}, err
		}
		card.APIKeyNonce = nonce
		card.APIKeyCipher = ciphertext
	}
	card.UpdatedAt = time.Now()
	return card, nil
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, card := range s.cards {
		if card.ID == id {
			s.cards = append(s.cards[:i], s.cards[i+1:]...)
			return s.saveCardsLocked()
		}
	}
	return os.ErrNotExist
}

func (s *Store) Archive(id string, archived bool) (CardDTO, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, card := range s.cards {
		if card.ID != id {
			continue
		}
		card.Archived = archived
		card.UpdatedAt = time.Now()
		s.cards[i] = card
		if err := s.saveCardsLocked(); err != nil {
			return CardDTO{}, err
		}
		key, err := s.decryptCard(card)
		return toDTO(card, maskKey(key)), err
	}
	return CardDTO{}, os.ErrNotExist
}

func (s *Store) Reveal(id string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, card := range s.cards {
		if card.ID == id {
			return s.decryptCard(card)
		}
	}
	return "", os.ErrNotExist
}

func validateCard(input CardInput, requireKey bool) error {
	if clean(input.Name) == "" {
		return errors.New("name is required")
	}
	if clean(input.BaseURL) == "" {
		return errors.New("base url is required")
	}
	if requireKey && strings.TrimSpace(input.APIKey) == "" {
		return errors.New("api key is required")
	}
	return nil
}
