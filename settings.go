package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func (s *Store) Settings() SettingsDTO {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return settingsDTO(s.settings)
}

func (s *Store) EmbeddingConfig() EmbeddingConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cfg := EmbeddingConfig{
		URL:       s.settings.EmbeddingURL,
		Model:     s.settings.EmbeddingModel,
		BatchSize: normalizedBatchSize(s.settings.EmbeddingBatchSize),
		MaxTokens: normalizedMaxTokens(s.settings.EmbeddingMaxTokens),
	}
	if s.settings.EmbeddingAPIKeyNonce != "" && s.settings.EmbeddingAPIKeyCipher != "" {
		key, err := s.decrypt(s.settings.EmbeddingAPIKeyNonce, s.settings.EmbeddingAPIKeyCipher)
		if err == nil {
			cfg.APIKey = key
		}
	}
	return cfg
}

func (s *Store) UpdateSettings(input SettingsDTO) (SettingsDTO, error) {
	theme := clean(input.Theme)
	if theme == "" {
		theme = "system"
	}
	if theme != "dark" && theme != "light" && theme != "system" {
		return SettingsDTO{}, errors.New("theme must be dark, light, or system")
	}
	if err := setAutostart(input.Autostart); err != nil {
		return SettingsDTO{}, err
	}

	s.mu.Lock()
	next := s.settings
	next.EmbeddingURL = clean(input.EmbeddingURL)
	next.EmbeddingModel = clean(input.EmbeddingModel)
	next.EmbeddingBatchSize = normalizedBatchSize(input.EmbeddingBatchSize)
	next.EmbeddingMaxTokens = normalizedMaxTokens(input.EmbeddingMaxTokens)
	next.Theme = theme
	next.Autostart = input.Autostart

	if input.ClearEmbeddingKey {
		next.EmbeddingAPIKeyNonce = ""
		next.EmbeddingAPIKeyCipher = ""
	} else if input.EmbeddingAPIKey != "" {
		nonce, cipherText, err := s.encrypt(input.EmbeddingAPIKey)
		if err != nil {
			s.mu.Unlock()
			return SettingsDTO{}, err
		}
		next.EmbeddingAPIKeyNonce = nonce
		next.EmbeddingAPIKeyCipher = cipherText
	}

	s.settings = next
	if err := s.saveSettingsLocked(); err != nil {
		s.mu.Unlock()
		return SettingsDTO{}, err
	}
	output := settingsDTO(s.settings)
	s.mu.Unlock()

	s.precomputeAllIfConfigured(true)
	return output, nil
}

func settingsDTO(settings StoredSettings) SettingsDTO {
	settings = normalizeSettings(settings)
	return SettingsDTO{
		EmbeddingURL:       settings.EmbeddingURL,
		EmbeddingModel:     settings.EmbeddingModel,
		EmbeddingBatchSize: settings.EmbeddingBatchSize,
		EmbeddingMaxTokens: settings.EmbeddingMaxTokens,
		HasEmbeddingAPIKey: settings.EmbeddingAPIKeyCipher != "",
		Theme:              settings.Theme,
		Autostart:          settings.Autostart,
	}
}

func normalizeSettings(settings StoredSettings) StoredSettings {
	settings.EmbeddingBatchSize = normalizedBatchSize(settings.EmbeddingBatchSize)
	settings.EmbeddingMaxTokens = normalizedMaxTokens(settings.EmbeddingMaxTokens)
	return settings
}

func normalizedBatchSize(value int) int {
	if value <= 0 {
		return defaultEmbeddingBatch
	}
	if value > 1000 {
		return 1000
	}
	return value
}

func normalizedMaxTokens(value int) int {
	if value <= 0 {
		return defaultEmbeddingTokens
	}
	if value > 1000000 {
		return 1000000
	}
	return value
}

func setAutostart(enabled bool) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	switch runtime.GOOS {
	case "linux":
		return setLinuxAutostart(exe, enabled)
	case "darwin":
		return setMacAutostart(exe, enabled)
	case "windows":
		return setWindowsAutostart(exe, enabled)
	default:
		return nil
	}
}

func setLinuxAutostart(exe string, enabled bool) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	autoDir := filepath.Join(configDir, "autostart")
	if err := os.MkdirAll(autoDir, 0700); err != nil {
		return err
	}
	path := filepath.Join(autoDir, "itsmyto.desktop")
	if !enabled {
		return removeIfExists(path)
	}
	content := "[Desktop Entry]\nType=Application\nName=ItsMyTo\nExec=" + exe + "\nTerminal=false\nX-GNOME-Autostart-enabled=true\n"
	return os.WriteFile(path, []byte(content), 0600)
}

func setMacAutostart(exe string, enabled bool) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	agentDir := filepath.Join(configDir, "LaunchAgents")
	if err := os.MkdirAll(agentDir, 0700); err != nil {
		return err
	}
	path := filepath.Join(agentDir, "com.itsmyto.app.plist")
	if !enabled {
		return removeIfExists(path)
	}
	content := `<?xml version="1.0" encoding="UTF-8"?><plist version="1.0"><dict><key>Label</key><string>com.itsmyto.app</string><key>ProgramArguments</key><array><string>` + exe + `</string></array><key>RunAtLoad</key><true/></dict></plist>`
	return os.WriteFile(path, []byte(content), 0600)
}

func setWindowsAutostart(exe string, enabled bool) error {
	key := `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`
	if !enabled {
		return exec.Command("reg", "delete", key, "/v", appName, "/f").Run()
	}
	return exec.Command("reg", "add", key, "/v", appName, "/t", "REG_SZ", "/d", exe, "/f").Run()
}

func removeIfExists(path string) error {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
