package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kqnade/vrcli/internal/config"
)

func TestDefault(t *testing.T) {
	cfg := config.Default()
	if cfg.App.PollFriendsInterval != 30*time.Second {
		t.Errorf("PollFriendsInterval = %v, want 30s", cfg.App.PollFriendsInterval)
	}
	if cfg.App.PollNotifsInterval != 15*time.Second {
		t.Errorf("PollNotifsInterval = %v, want 15s", cfg.App.PollNotifsInterval)
	}
}

func TestConfigDir(t *testing.T) {
	dir, err := config.ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() error: %v", err)
	}
	if dir == "" {
		t.Fatal("ConfigDir() returned empty string")
	}
}

func TestSessionPath(t *testing.T) {
	path, err := config.SessionPath()
	if err != nil {
		t.Fatalf("SessionPath() error: %v", err)
	}
	if filepath.Base(path) != "session.json" {
		t.Errorf("SessionPath() base = %q, want session.json", filepath.Base(path))
	}
}

func TestLoadNonExistent(t *testing.T) {
	// 存在しないディレクトリを指定 → Default を返す（エラーなし）
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() with missing file should not error: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	want := config.Default()
	want.App.PollFriendsInterval = 10 * time.Second
	want.App.PollNotifsInterval = 5 * time.Second

	if err := config.Save(want); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	got, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if got.App.PollFriendsInterval != want.App.PollFriendsInterval {
		t.Errorf("PollFriendsInterval = %v, want %v", got.App.PollFriendsInterval, want.App.PollFriendsInterval)
	}
	if got.App.PollNotifsInterval != want.App.PollNotifsInterval {
		t.Errorf("PollNotifsInterval = %v, want %v", got.App.PollNotifsInterval, want.App.PollNotifsInterval)
	}
}

func TestSaveCreatesDir(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "nested", "config")
	t.Setenv("XDG_CONFIG_HOME", dir)

	if err := config.Save(config.Default()); err != nil {
		t.Fatalf("Save() should create missing directories: %v", err)
	}

	cfgPath, _ := config.ConfigDir()
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		t.Error("Save() did not create config directory")
	}
}
