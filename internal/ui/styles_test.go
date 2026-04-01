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

	// 各ランクが空でなく、かつ別々の色を持つことを確認
	seen := map[string]client.TrustRank{}
	for _, rank := range ranks {
		color := string(ui.TrustRankColor(rank))
		if color == "" {
			t.Errorf("TrustRankColor(%d) returned empty string", rank)
			continue
		}
		if prev, ok := seen[color]; ok {
			t.Errorf("TrustRank %d and %d share color %q", rank, prev, color)
		}
		seen[color] = rank
	}
}
