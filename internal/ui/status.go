package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kqnade/vrcli/internal/client"
)

// ConfirmStatusMsg はステータス変更確定の意図を伝えるメッセージ。
// AppModel が受け取って VRCClient.UpdateStatus Cmd を発行する。
type ConfirmStatusMsg struct {
	Status client.UserStatus
	Desc   string
}

// StatusChoices は選択可能なステータス一覧（offline は除く）。
var StatusChoices = []client.UserStatus{
	client.StatusActive,
	client.StatusJoinMe,
	client.StatusAskMe,
	client.StatusBusy,
}

// StatusModal はステータス変更オーバーレイ。
type StatusModal struct {
	visible        bool
	selectedStatus int
	descInput      textinput.Model
	inputFocused   bool

	styles Styles
	keys   KeyMap
	width  int
	height int
}

// NewStatusModal は StatusModal を初期化して返す。
func NewStatusModal(styles Styles, keys KeyMap) StatusModal {
	ti := textinput.New()
	ti.Placeholder = "Status message..."
	ti.CharLimit = 255
	return StatusModal{
		styles:    styles,
		keys:      keys,
		descInput: ti,
	}
}

// Init は初期コマンドを返す。
func (m StatusModal) Init() tea.Cmd { return nil }

// Update はメッセージを処理して状態を更新する。
func (m StatusModal) Update(msg tea.Msg) (StatusModal, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Escape):
			m.Close()
			return m, nil

		case key.Matches(msg, m.keys.Tab):
			m.inputFocused = !m.inputFocused
			if m.inputFocused {
				m.descInput.Focus()
			} else {
				m.descInput.Blur()
			}
			return m, nil

		case key.Matches(msg, m.keys.Enter):
			status := StatusChoices[m.selectedStatus]
			desc := m.descInput.Value()
			m.Close()
			return m, func() tea.Msg {
				return ConfirmStatusMsg{Status: status, Desc: desc}
			}

		case key.Matches(msg, m.keys.Down):
			if !m.inputFocused {
				m.selectedStatus = (m.selectedStatus + 1) % len(StatusChoices)
			}

		case key.Matches(msg, m.keys.Up):
			if !m.inputFocused {
				m.selectedStatus = (m.selectedStatus - 1 + len(StatusChoices)) % len(StatusChoices)
			}
		}

		if m.inputFocused {
			var cmd tea.Cmd
			m.descInput, cmd = m.descInput.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// View はモーダル本体の文字列のみ返す。
func (m StatusModal) View() string {
	if !m.visible {
		return ""
	}

	var sb strings.Builder

	title := m.styles.ModalTitle.Render("Set Status")
	sb.WriteString(title)
	sb.WriteString("\n")

	for i, s := range StatusChoices {
		cursor := "  "
		if i == m.selectedStatus {
			cursor = "▶ "
		}
		line := fmt.Sprintf("%s%s %s", cursor, s.Emoji(), string(s))
		if i == m.selectedStatus {
			line = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true).
				Render(line)
		} else {
			line = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#AAAAAA")).
				Render(line)
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	sb.WriteString("\n")

	inputLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Render("Msg: ")
	sb.WriteString(inputLabel)
	sb.WriteString(m.descInput.View())
	sb.WriteString("\n\n")

	hint := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).
		Render("[Tab] switch  [Enter] OK  [Esc] ×")
	sb.WriteString(hint)

	return m.styles.Modal.Render(sb.String())
}

// Open はモーダルを現在のステータスと説明で開く。
func (m *StatusModal) Open(status client.UserStatus, desc string) {
	m.visible = true
	m.inputFocused = false
	m.descInput.Blur()
	m.descInput.SetValue(desc)

	for i, s := range StatusChoices {
		if s == status {
			m.selectedStatus = i
			break
		}
	}
}

// Close はモーダルを閉じる。
func (m *StatusModal) Close() {
	m.visible = false
	m.descInput.Blur()
}

// IsVisible はモーダルが表示中かどうかを返す。
func (m StatusModal) IsVisible() bool {
	return m.visible
}
