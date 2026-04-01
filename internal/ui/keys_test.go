package ui_test

import (
	"testing"

	"github.com/charmbracelet/bubbles/help"
	"github.com/kqnade/vrcli/internal/ui"
)

// コンパイル時に KeyMap が help.KeyMap インターフェースを満たすことを保証する。
var _ help.KeyMap = ui.DefaultKeyMap

func TestDefaultKeyMapShortHelp(t *testing.T) {
	km := ui.DefaultKeyMap
	bindings := km.ShortHelp()
	if len(bindings) == 0 {
		t.Error("ShortHelp() should return non-empty bindings")
	}
	// j と k がそれぞれ含まれることを確認（OR ではなく個別にチェック）
	foundJ, foundK := false, false
	for _, b := range bindings {
		for _, k := range b.Keys() {
			if k == "j" {
				foundJ = true
			}
			if k == "k" {
				foundK = true
			}
		}
	}
	if !foundJ {
		t.Error("ShortHelp() should include 'j' navigation key")
	}
	if !foundK {
		t.Error("ShortHelp() should include 'k' navigation key")
	}
}

func TestDefaultKeyMapFullHelp(t *testing.T) {
	km := ui.DefaultKeyMap
	groups := km.FullHelp()

	// FullHelp は 4 グループ: navigation(3), pane(2), actions(4), notif(3)
	wantGroupSizes := []int{3, 2, 4, 3}
	if len(groups) != len(wantGroupSizes) {
		t.Fatalf("FullHelp() group count = %d, want %d", len(groups), len(wantGroupSizes))
	}
	for i, want := range wantGroupSizes {
		if got := len(groups[i]); got != want {
			t.Errorf("FullHelp() group[%d] size = %d, want %d", i, got, want)
		}
	}
}

// TestKeyMapInterfaceCompliance はファイル先頭の var _ help.KeyMap = ... で代替済み。
