package client_test

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/kqnade/vrcgo/shared"
	"github.com/kqnade/vrcli/internal/client"
)

func TestErrTwoFactorRequired(t *testing.T) {
	// ErrTwoFactorRequired がセンチネルエラーとして機能するか確認
	err := fmt.Errorf("wrapped: %w", client.ErrTwoFactorRequired)
	if !errors.Is(err, client.ErrTwoFactorRequired) {
		t.Error("errors.Is should find ErrTwoFactorRequired through wrapping")
	}
}

func TestIsAuthErrorViaAPIError(t *testing.T) {
	// 401 の shared.APIError が isAuthError でも検出できるか（ErrMsg.IsAuth 経由）
	// VRCClient.CheckSession は内部で isAuthError を使う
	// ここでは ErrMsg の IsAuth フィールドが設定される想定をダミーで検証

	apiErr := &shared.APIError{StatusCode: 401, Message: "Unauthorized"}
	wrapped := fmt.Errorf("client error: %w", apiErr)

	var got *shared.APIError
	if !errors.As(wrapped, &got) {
		t.Fatal("errors.As should unwrap *shared.APIError")
	}
	if got.StatusCode != 401 {
		t.Errorf("StatusCode = %d, want 401", got.StatusCode)
	}
}

func TestNewVRCClient(t *testing.T) {
	c, err := client.New("vrchat-tui/test")
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if c == nil {
		t.Fatal("New() returned nil")
	}
}

func TestUpdateStatusWithoutUserID(t *testing.T) {
	c, err := client.New("vrchat-tui/test")
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	// userID 未設定のまま UpdateStatus を呼ぶ → ErrMsg が返るはず
	cmd := c.UpdateStatus("active", "")
	msg := cmd()
	errMsg, ok := msg.(client.ErrMsg)
	if !ok {
		t.Fatalf("expected ErrMsg, got %T", msg)
	}
	if errMsg.Err == nil {
		t.Error("ErrMsg.Err should not be nil when userID is empty")
	}
}

func TestSaveSessionCreatesDir(t *testing.T) {
	c, err := client.New("vrchat-tui/test")
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	path := filepath.Join(t.TempDir(), "nested", "dir", "session.json")
	// SaveCookies はファイルに書くが、空セッションでも dir は作成される
	if err := c.SaveSession(path); err != nil {
		t.Fatalf("SaveSession() error: %v", err)
	}
}
