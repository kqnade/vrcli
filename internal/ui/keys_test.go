package ui_test

import (
	"testing"

	"github.com/kqnade/vrcli/internal/ui"
)

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
	if totalBindings < 10 {
		t.Errorf("FullHelp() total bindings = %d, want at least 10", totalBindings)
	}
}

func TestKeyMapInterfaceCompliance(t *testing.T) {
	// KeyMap が help.KeyMap インターフェースを満たすか（ShortHelp/FullHelp が呼べる）
	km := ui.DefaultKeyMap
	short := km.ShortHelp()
	full := km.FullHelp()
	if len(short) == 0 || len(full) == 0 {
		t.Error("KeyMap must implement both ShortHelp and FullHelp")
	}
}
