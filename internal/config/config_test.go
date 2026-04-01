package config_test

import (
	"os"
	"path/filepath"
	"runtime"
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

	cfgPath, err := config.ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() error: %v", err)
	}
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		t.Error("Save() did not create config directory")
	}
}

func TestLoadInvalidTOML(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	// 壊れた TOML を書き込む
	cfgDir := filepath.Join(dir, "vrchat-tui")
	if err := os.MkdirAll(cfgDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.toml"), []byte("[[invalid toml"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := config.Load(); err == nil {
		t.Error("Load() with invalid TOML should return error")
	}
}

func TestLoadUnknownKey(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfgDir := filepath.Join(dir, "vrchat-tui")
	if err := os.MkdirAll(cfgDir, 0o700); err != nil {
		t.Fatal(err)
	}
	content := "[app]\nunknown_key = true\n"
	if err := os.WriteFile(filepath.Join(cfgDir, "config.toml"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := config.Load(); err == nil {
		t.Error("Load() with unknown key should return error")
	}
}

func TestSaveFilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits are not enforced on Windows")
	}
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	if err := config.Save(config.Default()); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	cfgDir, err := config.ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() error: %v", err)
	}

	// ディレクトリは 0700
	info, err := os.Stat(cfgDir)
	if err != nil {
		t.Fatalf("Stat(cfgDir) error: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o700 {
		t.Errorf("config dir perm = %o, want 0700", perm)
	}

	// ファイルは 0600
	info, err = os.Stat(filepath.Join(cfgDir, "config.toml"))
	if err != nil {
		t.Fatalf("Stat(config.toml) error: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("config.toml perm = %o, want 0600", perm)
	}
}
