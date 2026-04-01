package ui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kqnade/vrcli/internal/client"
	"github.com/kqnade/vrcli/internal/ui"
)

func newNotifsModel() ui.NotifsModel {
	return ui.NewNotifsModel(ui.NewStyles(), ui.DefaultKeyMap)
}

func TestNewNotifsModel(t *testing.T) {
	m := newNotifsModel()
	if cmd := m.Init(); cmd != nil {
		t.Error("Init() should return nil cmd")
	}
}

func TestNotifsModelViewNoPanic(t *testing.T) {
	m := newNotifsModel()
	_ = m.View()
}

func TestNotifsModelSetNotifsMsg(t *testing.T) {
	m := newNotifsModel()
	notifs := []client.Notif{
		{ID: "n1", SenderUsername: "Alice", IsFriendRequest: true},
		{ID: "n2", SenderUsername: "Bob", IsFriendRequest: true},
	}
	m, _ = m.Update(client.NotifsMsg{Notifs: notifs})
	_ = m.View() // panic しないこと
}

func TestNotifsModelAcceptKey(t *testing.T) {
	m := newNotifsModel()
	m.SetFocus(true)
	notifs := []client.Notif{
		{ID: "n1", SenderUsername: "Alice", IsFriendRequest: true},
	}
	m, _ = m.Update(client.NotifsMsg{Notifs: notifs})

	// 'a' キーで AcceptRequestMsg が返る
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	_ = m2
	if cmd == nil {
		t.Fatal("accept key should return a cmd for friend request")
	}
	msg := cmd()
	ar, ok := msg.(ui.AcceptRequestMsg)
	if !ok {
		t.Fatalf("cmd() returned %T, want ui.AcceptRequestMsg", msg)
	}
	if ar.NotifID != "n1" {
		t.Errorf("AcceptRequestMsg.NotifID = %q, want \"n1\"", ar.NotifID)
	}
}

func TestNotifsModelRejectKey(t *testing.T) {
	m := newNotifsModel()
	m.SetFocus(true)
	notifs := []client.Notif{
		{ID: "n1", SenderUsername: "Alice", IsFriendRequest: true},
	}
	m, _ = m.Update(client.NotifsMsg{Notifs: notifs})

	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	_ = m2
	if cmd == nil {
		t.Fatal("reject key should return a cmd for friend request")
	}
	msg := cmd()
	rr, ok := msg.(ui.RejectRequestMsg)
	if !ok {
		t.Fatalf("cmd() returned %T, want ui.RejectRequestMsg", msg)
	}
	if rr.NotifID != "n1" {
		t.Errorf("RejectRequestMsg.NotifID = %q, want \"n1\"", rr.NotifID)
	}
}

func TestNotifsModelAcceptKeyNonFriendRequest(t *testing.T) {
	m := newNotifsModel()
	m.SetFocus(true)
	// IsFriendRequest = false の通知
	notifs := []client.Notif{
		{ID: "n1", SenderUsername: "Alice", IsFriendRequest: false},
	}
	m, _ = m.Update(client.NotifsMsg{Notifs: notifs})

	// 'a' キーは cmd を返さないはず
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	if cmd != nil {
		t.Error("accept key should not emit msg for non-friend-request")
	}
}

func TestNotifsModelFocusGatesKeys(t *testing.T) {
	m := newNotifsModel()
	m.SetFocus(false)
	notifs := []client.Notif{
		{ID: "n1", SenderUsername: "Alice", IsFriendRequest: true},
	}
	m, _ = m.Update(client.NotifsMsg{Notifs: notifs})

	// フォーカスなし時は 'a' を無視
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	if cmd != nil {
		t.Error("key should be ignored when not focused")
	}
}

func TestNotifsModelFriendRequestAccepted(t *testing.T) {
	m := newNotifsModel()
	notifs := []client.Notif{
		{ID: "n1", SenderUsername: "Alice", IsFriendRequest: true},
		{ID: "n2", SenderUsername: "Bob", IsFriendRequest: true},
	}
	m, _ = m.Update(client.NotifsMsg{Notifs: notifs})

	// n1 を承認 → リストから除去される
	m, _ = m.Update(client.FriendRequestAcceptedMsg{NotifID: "n1"})

	// 除去後、先頭アイテムは n2 のはず。
	// フォーカスを当てて 'a' を押し、NotifID が n2 であることで n1 が消えたことを確認する。
	m.SetFocus(true)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	if cmd == nil {
		t.Fatal("expected AcceptRequestMsg for n2 after n1 was removed")
	}
	msg := cmd()
	ar, ok := msg.(ui.AcceptRequestMsg)
	if !ok {
		t.Fatalf("cmd() returned %T, want ui.AcceptRequestMsg", msg)
	}
	if ar.NotifID != "n2" {
		t.Errorf("after removing n1, first item NotifID = %q, want \"n2\"", ar.NotifID)
	}
}

func TestNotifsModelFriendRequestRejected(t *testing.T) {
	m := newNotifsModel()
	notifs := []client.Notif{
		{ID: "n1", SenderUsername: "Alice", IsFriendRequest: true},
	}
	m, _ = m.Update(client.NotifsMsg{Notifs: notifs})
	m, _ = m.Update(client.FriendRequestRejectedMsg{NotifID: "n1"})
	_ = m.View() // panic しないこと
}

func TestNotifsModelSetSize(t *testing.T) {
	m := newNotifsModel()
	m.SetSize(80, 40)
	_ = m.View()
}
