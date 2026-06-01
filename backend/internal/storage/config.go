package storage

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"omar-kada/air-compose/internal/models"

	"github.com/fsnotify/fsnotify"
	"github.com/go-viper/mapstructure/v2"
	"go.yaml.in/yaml/v3"
)

// ConfigStore stores and retreives the configuration
type ConfigStore interface {
	Update(cfg models.Config) error
	Get() models.Config
	ToYaml(cfg models.Config) ([]byte, error)
	SetOnChange(fn func(oldConfig, newConfig models.Config))
	WatchFile() error
}

type configStore struct {
	onConfigUpdate func(oldConfig, newConfig models.Config)
	configFilePath string
	cfg            models.Config
	mu             sync.RWMutex

	watcher *fsnotify.Watcher
}

// NewConfigStore creates a new config file storage
func NewConfigStore(filePath string) (ConfigStore, error) {
	cfg, err := readConfig(filePath)
	if err != nil {
		return nil, err
	}
	configStore := &configStore{
		configFilePath: filePath,
		cfg:            cfg,
	}
	return configStore, err
}

func (s *configStore) Update(cfg models.Config) (err error) {
	slog.Debug("updating configuration file")
	s.mu.Lock()
	defer s.mu.Unlock()

	oldCfg, err := readConfig(s.configFilePath)
	if err != nil {
		return err
	}

	if models.IsObfuscated(cfg.Settings.Git.Token) {
		cfg.Settings.Git.Token = oldCfg.Settings.Git.Token // keep old token when obfuscated
	}
	if models.IsObfuscated(cfg.Settings.Notifications.NotificationURL) {
		cfg.Settings.Notifications.NotificationURL = oldCfg.Settings.Notifications.NotificationURL // keep old url when obfuscated
	}
	if models.IsObfuscated(cfg.Settings.Oidc.ClientSecret) {
		cfg.Settings.Oidc.ClientSecret = oldCfg.Settings.Oidc.ClientSecret // keep old client secret when obfuscated
	}

	if s.onConfigUpdate != nil {
		defer func() {
			if err != nil { // check no error occurred when updating the config
				return
			}
			s.onConfigUpdate(oldCfg, cfg)
		}()
	}

	bs, err := s.ToYaml(cfg)
	if err != nil {
		return err
	}
	if s.pauseWatchFile() {
		defer func() {
			watchErr := s.WatchFile()
			if err == nil {
				err = watchErr
			}
		}()
	}
	if err := os.WriteFile(s.configFilePath, bs, 0o644); err != nil {
		return fmt.Errorf("error writing config file %s: %w", s.configFilePath, err)
	}
	s.cfg = cfg

	return nil
}

func (*configStore) ToYaml(cfg models.Config) ([]byte, error) {
	var m map[string]any
	encCfg := &mapstructure.DecoderConfig{
		TagName:          "mapstructure",
		Result:           &m,
		WeaklyTypedInput: true,
	}
	encoder, err := mapstructure.NewDecoder(encCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create encoder: %w", err)
	}
	if err := encoder.Decode(cfg); err != nil {
		return nil, fmt.Errorf("error encoding config: %w", err)
	}

	bs, err := yaml.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("error marshaling config: %w", err)
	}
	return bs, nil
}

func (s *configStore) SetOnChange(fn func(oldConfig, newConfig models.Config)) {
	slog.Debug("setting OnConfigUpdate")
	s.onConfigUpdate = fn
}

// readConfig reads the configuration from the config file
func readConfig(path string) (models.Config, error) {
	bs, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
				return models.Config{}, err
			}
			return models.Config{}, nil
		}
		return models.Config{}, fmt.Errorf("error reading config file %s: %w", path, err)
	}

	var m map[string]any
	if err := yaml.Unmarshal(bs, &m); err != nil {
		return models.Config{}, fmt.Errorf("error unmarshaling yaml %s: %w", path, err)
	}

	return decodeConfig(m)
}

// Get retreives the configution
func (s *configStore) Get() models.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg
}

func (s *configStore) pauseWatchFile() bool {
	if s.watcher != nil {
		s.watcher.Close()
		s.watcher = nil
		return true
	}
	return false
}

func (s *configStore) WatchFile() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	// Start listening in a goroutine
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Has(fsnotify.Write) {
					slog.Info("config file update detected")
					newCfg, err := readConfig(s.configFilePath)
					if err != nil {
						slog.Error("error reading new config file", "err", err)
					} else {
						s.mu.Lock()
						defer s.mu.Unlock()
						oldCfg := s.cfg
						s.cfg = newCfg
						if s.onConfigUpdate != nil {
							defer s.onConfigUpdate(oldCfg, newCfg)
						}
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				slog.Debug("error on config file watcher", "err", err.Error())
			}
		}
	}()

	// Add the file to watch
	err = watcher.Add(s.configFilePath)
	if err != nil {
		s.pauseWatchFile()
		s.watcher = watcher
	}
	return err
}

func decodeConfig(configMap map[string]any) (models.Config, error) {
	var cfg models.Config
	decCfg := &mapstructure.DecoderConfig{
		TagName:          "mapstructure",
		Result:           &cfg,
		WeaklyTypedInput: true,
	}
	decoder, err := mapstructure.NewDecoder(decCfg)
	if err != nil {
		return models.Config{}, fmt.Errorf("failed to create decoder: %w", err)
	}
	if err := decoder.Decode(configMap); err != nil {
		return models.Config{}, fmt.Errorf("error decoding merged config: %w", err)
	}
	return cfg, nil
}
