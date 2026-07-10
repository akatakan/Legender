package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Settings struct {
	SchemaVersion int              `json:"schemaVersion"`
	AutoAccept    bool             `json:"autoAccept"`
	AutoPick      AutoPickSettings `json:"autoPick"`
}

const currentSchemaVersion = 1

type AutoPickSettings struct {
	Enabled        bool  `json:"enabled"`
	PrimaryChamp   int   `json:"primaryChamp"`
	SecondaryChamp int   `json:"secondaryChamp"`
	ChampPool      []int `json:"champPool"`
}

// Store ayarların bellek ve disk üzerindeki tek sahibidir.
type Store struct {
	mu       sync.RWMutex
	path     string
	legacy   []string
	settings Settings
}

func NewStore(path string) *Store {
	return &Store{path: path, settings: defaultSettings()}
}

func NewStoreWithLegacy(path string, legacyPaths ...string) *Store {
	return &Store{path: path, legacy: append([]string(nil), legacyPaths...), settings: defaultSettings()}
}

func DefaultPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "settings.json"
	}
	return filepath.Join(configDir, "Legender", "settings.json")
}

func (s *Store) Load() error {
	file, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		for _, legacyPath := range s.legacy {
			legacyFile, legacyErr := os.ReadFile(legacyPath)
			if legacyErr != nil {
				continue
			}
			if err := s.loadJSON(legacyFile); err != nil {
				return fmt.Errorf("eski ayarlar taşınamadı: %w", err)
			}
			return s.Save()
		}
		return s.Save()
	}
	if err != nil {
		return fmt.Errorf("ayarlar dosyası okunamadı: %w", err)
	}

	return s.loadJSON(file)
}

func (s *Store) Snapshot() Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneSettings(s.settings)
}

func (s *Store) SetAutoAccept(enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settings.AutoAccept = enabled
	return s.saveLocked()
}

func (s *Store) SetAutoPick(cfg AutoPickSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settings.AutoPick = AutoPickSettings{
		Enabled:        cfg.Enabled,
		PrimaryChamp:   cfg.PrimaryChamp,
		SecondaryChamp: cfg.SecondaryChamp,
		ChampPool:      append([]int(nil), cfg.ChampPool...),
	}
	s.settings = migrateSettings(s.settings)
	return s.saveLocked()
}

func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked()
}

func (s *Store) saveLocked() error {
	data, err := json.MarshalIndent(s.settings, "", "  ")
	if err != nil {
		return fmt.Errorf("ayarlar JSON verisine dönüştürülemedi: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("ayarlar dizini oluşturulamadı: %w", err)
	}
	if err := os.WriteFile(s.path, data, 0o600); err != nil {
		return fmt.Errorf("ayarlar dosyası yazılamadı: %w", err)
	}
	return nil
}

func (s *Store) loadJSON(data []byte) error {
	var loaded Settings
	if err := json.Unmarshal(data, &loaded); err != nil {
		return fmt.Errorf("ayarlar JSON verisi çözülemedi: %w", err)
	}
	if loaded.SchemaVersion > currentSchemaVersion {
		return fmt.Errorf("ayar şeması desteklenmiyor: %d (desteklenen: %d)", loaded.SchemaVersion, currentSchemaVersion)
	}
	loaded = migrateSettings(loaded)
	s.mu.Lock()
	s.settings = cloneSettings(loaded)
	s.mu.Unlock()
	return nil
}

func defaultSettings() Settings {
	return Settings{
		SchemaVersion: currentSchemaVersion,
		AutoAccept:    false,
		AutoPick: AutoPickSettings{
			ChampPool: []int{},
		},
	}
}

func migrateSettings(settings Settings) Settings {
	if settings.SchemaVersion == 0 {
		settings.SchemaVersion = currentSchemaVersion
	}
	if settings.AutoPick.PrimaryChamp < 0 {
		settings.AutoPick.PrimaryChamp = 0
	}
	if settings.AutoPick.SecondaryChamp < 0 {
		settings.AutoPick.SecondaryChamp = 0
	}
	seen := make(map[int]struct{}, len(settings.AutoPick.ChampPool))
	pool := make([]int, 0, len(settings.AutoPick.ChampPool))
	for _, championID := range settings.AutoPick.ChampPool {
		if championID <= 0 {
			continue
		}
		if _, exists := seen[championID]; exists {
			continue
		}
		seen[championID] = struct{}{}
		pool = append(pool, championID)
	}
	settings.AutoPick.ChampPool = pool
	return settings
}

func cloneSettings(source Settings) Settings {
	result := source
	result.AutoPick.ChampPool = append([]int(nil), source.AutoPick.ChampPool...)
	return result
}
