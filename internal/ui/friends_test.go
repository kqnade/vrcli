package ui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kqnade/vrcli/internal/client"
	"github.com/kqnade/vrcli/internal/ui"
)

func newFriendsModel() ui.FriendsModel {
	return ui.NewFriendsModel(ui.NewStyles(), ui.DefaultKeyMap)
}

func TestNewFriendsModel(t *testing.T) {
	m := newFriendsModel()
	// Init は nil cmd を返す
	if cmd := m.Init(); cmd != nil {
		t.Error("Init() should return nil cmd")
	}
}

func TestFriendsModelViewNoPanic(t *testing.T) {
	m := newFriendsModel()
	_ = m.View() // panic しないこと
}

func TestFriendsModelSetFriendsMsg(t *testing.T) {
	m := newFriendsModel()
	friends := []client.Friend{
		{ID: "u1", DisplayName: "Alice", Status: client.StatusActive, TrustRank: client.TrustUser},
		{ID: "u2", DisplayName: "Bob", Status: client.StatusBusy, TrustRank: client.TrustKnownUser},
	}
	m2, _ := m.Update(client.FriendsMsg{Friends: friends})

	f := m2.SelectedFriend()
	if f == nil {
		t.Fatal("SelectedFriend() should return non-nil after FriendsMsg")
	}
	if f.DisplayName != "Alice" {
		t.Errorf("SelectedFriend().DisplayName = %q, want \"Alice\"", f.DisplayName)
	}
}

func TestFriendsModelSelectedFriendEmpty(t *testing.T) {
	m := newFriendsModel()
	if f := m.SelectedFriend(); f != nil {
		t.Error("SelectedFriend() should return nil when list is empty")
	}
}

func TestFriendsModelSetSize(t *testing.T) {
	m := newFriendsModel()
	m.SetSize(80, 40) // panic しないこと
	_ = m.View()
}

func TestFriendsModelFocusGatesKeys(t *testing.T) {
	m := newFriendsModel()
	friends := []client.Friend{
		{ID: "u1", DisplayName: "Alice"},
		{ID: "u2", DisplayName: "Bob"},
	}
	m, _ = m.Update(client.FriendsMsg{Friends: friends})

	// フォーカスなしでは j キーを無視すること
	m.SetFocus(false)
	before := m.SelectedFriend()
	if before == nil {
		t.Fatal("SelectedFriend() should be non-nil before key press")
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	after := m.SelectedFriend()
	if after == nil {
		t.Fatal("SelectedFriend() should be non-nil after key press")
	}
	if before.ID != after.ID {
		t.Error("key should be ignored when not focused")
	}
}

func TestFriendsModelFocusEnabled(t *testing.T) {
	m := newFriendsModel()
	friends := []client.Friend{
		{ID: "u1", DisplayName: "Alice"},
		{ID: "u2", DisplayName: "Bob"},
	}
	m, _ = m.Update(client.FriendsMsg{Friends: friends})
	m.SetFocus(true)

	// フォーカスありでは j キーが伝播する
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	f := m.SelectedFriend()
	if f == nil {
		t.Fatal("SelectedFriend() should return non-nil")
	}
	if f.ID != "u2" {
		t.Errorf("SelectedFriend().ID = %q, want \"u2\" after j", f.ID)
	}
}
