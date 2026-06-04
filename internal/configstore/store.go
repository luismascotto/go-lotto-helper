package configstore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/luismascotto/lotofacil-checker/internal/model"
)

type Store struct {
	path string
}

func New(path string) *Store {
	return &Store{path: path}
}

func (s *Store) Path() string {
	return s.path
}

func (s *Store) Load() (model.Config, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return model.Config{Bets: []model.Bet{}}, nil
		}
		return model.Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg model.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return model.Config{}, fmt.Errorf("parse config: %w", err)
	}
	if cfg.Bets == nil {
		cfg.Bets = []model.Bet{}
	}
	return cfg, nil
}

func (s *Store) LoadOrCreate() (model.Config, error) {
	_, err := os.Stat(s.path)
	if err == nil {
		return s.Load()
	}
	if !os.IsNotExist(err) {
		return model.Config{}, fmt.Errorf("stat config: %w", err)
	}

	cfg := model.Config{Bets: []model.Bet{}}
	if err := s.Save(cfg); err != nil {
		return model.Config{}, err
	}
	return cfg, nil
}

func (s *Store) Save(cfg model.Config) error {
	if cfg.Bets == nil {
		cfg.Bets = []model.Bet{}
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(s.path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create config directory: %w", err)
		}
	}

	if err := os.WriteFile(s.path, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}
