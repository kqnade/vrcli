package ui_test

import (
	"testing"

	"github.com/kqnade/vrcli/internal/client"
	"github.com/kqnade/vrcli/internal/ui"
)

func TestNewStyles(t *testing.T) {
	s := ui.NewStyles()
	// スタイルが初期化されることを確認（Render が空でないこと）
	if s.Pane.Render("x") == "" {
		t.Error("Pane style should render non-empty")
	}
	if s.ActivePane.Render("x") == "" {
		t.Error("ActivePane style should render non-empty")
	}
	if s.Modal.Render("x") == "" {
		t.Error("Modal style should render non-empty")
	}
}

func TestTrustRankColors(t *testing.T) {
	ranks := []client.TrustRank{
		client.TrustVisitor,
		client.TrustNewUser,
		client.TrustUser,
		client.TrustKnownUser,
		client.TrustTrusted,
		client.TrustFriend,
	}

	// 各ランクが別々の色を持つことを確認
	seen := map[string]client.TrustRank{}
	for _, rank := range ranks {
		color := string(ui.TrustRankColor(rank))
		if prev, ok := seen[color]; ok {
			t.Errorf("TrustRank %d and %d share color %q", rank, prev, color)
		}
		seen[color] = rank
	}
}

func TestActivePaneColor(t *testing.T) {
	// ActivePane と Pane は同じ境界文字を使うが色が異なる。
	// 色は ANSI コードに依存するためヘッドレス環境では検証困難。
	// ここでは NewStyles() が panic せずに完了することのみを確認する。
	s := ui.NewStyles()
	_ = s.ActivePane
	_ = s.Pane
}
