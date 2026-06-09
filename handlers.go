package main

import "net/http"

func (s *Server) handleListCards(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.List(r.URL.Query().Get("archive") == "1")
	writeResult(w, items, err)
}

func (s *Server) handleCreateCard(w http.ResponseWriter, r *http.Request) {
	var input CardInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	item, err := s.store.Create(input)
	writeResult(w, item, err)
}

func (s *Server) handleUpdateCard(w http.ResponseWriter, r *http.Request) {
	var input CardInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	item, err := s.store.Update(r.PathValue("id"), input)
	writeResult(w, item, err)
}

func (s *Server) handleDeleteCard(w http.ResponseWriter, r *http.Request) {
	writeResult(w, map[string]bool{"ok": true}, s.store.Delete(r.PathValue("id")))
}

func (s *Server) handleArchiveCard(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Archived bool `json:"archived"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	item, err := s.store.Archive(r.PathValue("id"), input.Archived)
	writeResult(w, item, err)
}

func (s *Server) handleRevealCard(w http.ResponseWriter, r *http.Request) {
	key, err := s.store.Reveal(r.PathValue("id"))
	writeResult(w, revealResponse{APIKey: key}, err)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	var input SearchRequest
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	items, err := s.store.Search(input.Query, input.IncludeArchive)
	writeResult(w, items, err)
}

func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	writeResult(w, s.store.Settings(), nil)
}

func (s *Server) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	var input SettingsDTO
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	settings, err := s.store.UpdateSettings(input)
	writeResult(w, settings, err)
}

func (s *Server) handleOpenURL(w http.ResponseWriter, r *http.Request) {
	var input OpenURLRequest
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	err := openExternalURL(input.URL)
	writeResult(w, map[string]bool{"ok": true}, err)
}
