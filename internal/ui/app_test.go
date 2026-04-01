package ui_test

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kqnade/vrcli/internal/client"
	"github.com/kqnade/vrcli/internal/config"
	"github.com/kqnade/vrcli/internal/ui"
)

// newSizedApp は WindowSizeMsg を適用済みの AppModel を返す。
func newSizedApp(t *testing.T) tea.Model {
	t.Helper()
	m := ui.New(nil, config.Default())
	sized, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	return sized
}

func TestAppInit(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("New() panicked: %v", r)
		}
	}()
	_ = ui.New(nil, config.Default())
}

func TestAppViewLoading(t *testing.T) {
	m := ui.New(nil, config.Default())
	view := m.View()
	if view != "Loading..." {
		t.Errorf("View() before WindowSizeMsg = %q, want \"Loading...\"", view)
	}
}

func TestAppWindowResize(t *testing.T) {
	m := newSizedApp(t)
	view := m.View()
	if view == "Loading..." {
		t.Error("View() should render content after WindowSizeMsg")
	}
}

func TestAppQuit(t *testing.T) {
	m := newSizedApp(t)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Fatal("Quit key should return a cmd")
	}
	msg := cmd()
	if msg != tea.Quit() {
		t.Errorf("Quit cmd returned %T, want tea.QuitMsg", msg)
	}
}

func TestAppHelpToggle(t *testing.T) {
	m := newSizedApp(t)

	// '?' でヘルプを表示
	model1, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	view1 := model1.View()

	// もう一度 '?' で非表示
	model2, _ := model1.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	view2 := model2.View()

	if view1 == view2 {
		t.Error("help toggle should change the view")
	}
}

func TestAppAuthOKMsg(t *testing.T) {
	m := newSizedApp(t)

	// vrc=nil なので FetchFriends/FetchNotifs が panic する可能性があるが、
	// Update 自体は panic しないことを確認する
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Update(AuthOKMsg) panicked: %v", r)
		}
	}()

	model, _ := m.Update(client.AuthOKMsg{
		UserID:            "u_test",
		DisplayName:       "TestUser",
		Status:            client.StatusActive,
		StatusDescription: "",
	})
	_ = model.View()
}

func TestAppErrMsg(t *testing.T) {
	m := newSizedApp(t)
	model, _ := m.Update(client.ErrMsg{Err: errors.New("test error")})
	view := model.View()
	_ = view // panic しないこと
}

func TestAppStatusModalOpenClose(t *testing.T) {
	m := newSizedApp(t)

	// 's' でモーダルを開く
	model1, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	view1 := model1.View()
	if view1 == "Loading..." {
		t.Error("modal view should not be loading")
	}

	// Esc でモーダルを閉じる
	model2, _ := model1.Update(tea.KeyMsg{Type: tea.KeyEsc})
	_ = model2.View() // panic しないこと
}

func TestAppFriendRequestIntentMessages(t *testing.T) {
	m := newSizedApp(t)

	// AcceptRequestMsg/RejectRequestMsg は vrc=nil でも panic すべきでない
	// (AppModel.Update が cmd を返すだけ)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Update panicked: %v", r)
		}
	}()

	// cmd が nil でないことだけ確認（実行はしない）
	_, cmd := m.Update(ui.AcceptRequestMsg{NotifID: "n1"})
	if cmd == nil {
		t.Error("AcceptRequestMsg should return a cmd")
	}

	_, cmd2 := m.Update(ui.RejectRequestMsg{NotifID: "n1"})
	if cmd2 == nil {
		t.Error("RejectRequestMsg should return a cmd")
	}
}

func TestAppConfirmStatusMsg(t *testing.T) {
	m := newSizedApp(t)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Update panicked: %v", r)
		}
	}()

	_, cmd := m.Update(ui.ConfirmStatusMsg{Status: client.StatusBusy, Desc: "busy"})
	if cmd == nil {
		t.Error("ConfirmStatusMsg should return a cmd")
	}
}
