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

	md, err := toml.Decode(string(data), cfg)
	if err != nil {
		return nil, fmt.Errorf("config: decode toml: %w", err)
	}
	if undecoded := md.Undecoded(); len(undecoded) > 0 {
		return nil, fmt.Errorf("config: unknown keys in config file: %v", undecoded)
	}
	return cfg, nil
}

// Save は設定をファイルに書き込む。
// 必要に応じてディレクトリを作成する。アトミック書き込みで破損を防ぐ。
func Save(cfg *Config) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("config: create config dir: %w", err)
	}
	// 既存ディレクトリのパーミッションも矯正する
	if err := os.Chmod(dir, 0o700); err != nil {
		return fmt.Errorf("config: chmod config dir: %w", err)
	}

	finalPath := filepath.Join(dir, "config.toml")
	tmpPath := finalPath + ".tmp"

	f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("config: open temp file for writing: %w", err)
	}

	encErr := toml.NewEncoder(f).Encode(cfg)
	syncErr := f.Sync()
	closeErr := f.Close()

	if encErr != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("config: encode toml: %w", encErr)
	}
	if syncErr != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("config: sync temp file: %w", syncErr)
	}
	if closeErr != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("config: close temp file: %w", closeErr)
	}

	if err := os.Chmod(tmpPath, 0o600); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("config: chmod temp file: %w", err)
	}

	if err := os.Rename(tmpPath, finalPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("config: rename temp file: %w", err)
	}
	return nil
}
