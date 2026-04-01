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
	// j/k が含まれることを確認
	found := false
	for _, b := range bindings {
		for _, k := range b.Keys() {
			if k == "j" || k == "k" {
				found = true
			}
		}
	}
	if !found {
		t.Error("ShortHelp() should include j/k navigation keys")
	}
}

func TestDefaultKeyMapFullHelp(t *testing.T) {
	km := ui.DefaultKeyMap
	groups := km.FullHelp()
	if len(groups) == 0 {
		t.Error("FullHelp() should return non-empty groups")
	}
	totalBindings := 0
	for _, g := range groups {
		totalBindings += len(g)
	}
	// FullHelp は現在 4 グループ・合計 12 バインディングを定義している
	// (navigation:3, pane:2, actions:4, notif:3)
	const wantBindings = 12
	if totalBindings != wantBindings {
		t.Errorf("FullHelp() total bindings = %d, want %d", totalBindings, wantBindings)
	}
}

// TestKeyMapInterfaceCompliance はファイル先頭の var _ help.KeyMap = ... で代替済み。
