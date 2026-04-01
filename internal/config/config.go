// Package config は設定ファイルの読み書きとセッションパスの管理を担う。
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

const appName = "vrchat-tui"

// AppConfig はポーリング間隔などアプリケーション動作の設定を保持する。
type AppConfig struct {
	PollFriendsInterval time.Duration `toml:"poll_friends_interval"`
	PollNotifsInterval  time.Duration `toml:"poll_notifs_interval"`
}

// Config はルート設定を保持する。
type Config struct {
	App AppConfig `toml:"app"`
}

// Default はデフォルト設定を返す。
func Default() *Config {
	return &Config{
		App: AppConfig{
			PollFriendsInterval: 30 * time.Second,
			PollNotifsInterval:  15 * time.Second,
		},
	}
}

// ConfigDir は設定ディレクトリのパスを返す (~/.config/vrchat-tui)。
// ディレクトリが存在しない場合でもパスだけを返し、作成は行わない。
func ConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("config: get user config dir: %w", err)
	}
	return filepath.Join(base, appName), nil
}

// SessionPath はセッション Cookie ファイルのパスを返す。
func SessionPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "session.json"), nil
}

// configFilePath は設定 TOML ファイルのパスを返す。
func configFilePath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

// Load は設定ファイルを読み込む。
// ファイルが存在しない場合は Default() を返し、エラーは返さない。
func Load() (*Config, error) {
	path, err := configFilePath()
	if err != nil {
		return nil, err
	}

	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return nil, fmt.Errorf("config: read file: %w", err)
	}

	if _, err := toml.Decode(string(data), cfg); err != nil {
		return nil, fmt.Errorf("config: decode toml: %w", err)
	}
	return cfg, nil
}

// Save は設定をファイルに書き込む。
// 必要に応じてディレクトリを作成する。
func Save(cfg *Config) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("config: create config dir: %w", err)
	}

	path := filepath.Join(dir, "config.toml")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("config: open file for writing: %w", err)
	}
	defer f.Close()

	if err := toml.NewEncoder(f).Encode(cfg); err != nil {
		return fmt.Errorf("config: encode toml: %w", err)
	}
	return nil
}
