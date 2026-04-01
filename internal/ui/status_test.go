package ui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kqnade/vrcli/internal/client"
	"github.com/kqnade/vrcli/internal/ui"
)

func newStatusModal() ui.StatusModal {
	return ui.NewStatusModal(ui.NewStyles(), ui.DefaultKeyMap)
}

func TestStatusModalInit(t *testing.T) {
	m := newStatusModal()
	if cmd := m.Init(); cmd != nil {
		t.Error("Init() should return nil cmd")
	}
}

func TestStatusModalNotVisible(t *testing.T) {
	m := newStatusModal()
	if m.IsVisible() {
		t.Error("should not be visible initially")
	}
	if v := m.View(); v != "" {
		t.Errorf("View() should return empty when not visible, got %q", v)
	}
}

func TestStatusModalOpen(t *testing.T) {
	m := newStatusModal()
	m.Open(client.StatusActive, "hello")
	if !m.IsVisible() {
		t.Error("should be visible after Open()")
	}
	v := m.View()
	if v == "" {
		t.Error("View() should return non-empty when visible")
	}
}

func TestStatusModalClose(t *testing.T) {
	m := newStatusModal()
	m.Open(client.StatusActive, "")
	m.Close()
	if m.IsVisible() {
		t.Error("should not be visible after Close()")
	}
}

func TestStatusModalEscKey(t *testing.T) {
	m := newStatusModal()
	m.Open(client.StatusActive, "")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.IsVisible() {
		t.Error("Esc should close the modal")
	}
}

func TestStatusModalConfirm(t *testing.T) {
	m := newStatusModal()
	m.Open(client.StatusBusy, "do not disturb")

	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = m2
	if cmd == nil {
		t.Fatal("Enter should return a cmd")
	}
	msg := cmd()
	cs, ok := msg.(ui.ConfirmStatusMsg)
	if !ok {
		t.Fatalf("cmd() returned %T, want ui.ConfirmStatusMsg", msg)
	}
	if cs.Status != client.StatusBusy {
		t.Errorf("ConfirmStatusMsg.Status = %q, want %q", cs.Status, client.StatusBusy)
	}
	if cs.Desc != "do not disturb" {
		t.Errorf("ConfirmStatusMsg.Desc = %q, want \"do not disturb\"", cs.Desc)
	}
	if m2.IsVisible() {
		t.Error("modal should be closed after confirm")
	}
}

func TestStatusModalNavigate(t *testing.T) {
	m := newStatusModal()
	m.Open(client.StatusActive, "")

	// j キーで次のステータスへ
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	// Enter で確認して StatusChoices[1] が返るはず
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter should return a cmd")
	}
	msg := cmd()
	cs, ok := msg.(ui.ConfirmStatusMsg)
	if !ok {
		t.Fatalf("cmd() returned %T, want ui.ConfirmStatusMsg", msg)
	}
	if cs.Status != ui.StatusChoices[1] {
		t.Errorf("ConfirmStatusMsg.Status = %q, want %q", cs.Status, ui.StatusChoices[1])
	}
}

func TestStatusModalIgnoreKeyWhenHidden(t *testing.T) {
	m := newStatusModal()
	// visible == false の状態でキーを送っても何もしない
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("should return nil cmd when modal is hidden")
	}
	if m2.IsVisible() {
		t.Error("modal should remain hidden")
	}
}

func TestStatusChoicesLength(t *testing.T) {
	if len(ui.StatusChoices) != 4 {
		t.Errorf("StatusChoices length = %d, want 4", len(ui.StatusChoices))
	}
}
