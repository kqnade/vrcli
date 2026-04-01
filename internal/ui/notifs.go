package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kqnade/vrcli/internal/client"
)

// 意図メッセージ型（AppModel が API Cmd を発行するために使う）

// AcceptRequestMsg はフレンドリクエスト承認の意図を伝えるメッセージ。
type AcceptRequestMsg struct{ NotifID string }

// RejectRequestMsg はフレンドリクエスト拒否の意図を伝えるメッセージ。
type RejectRequestMsg struct{ NotifID string }

// notifItem は bubbles/list.Item インターフェースを実装するラッパー。
type notifItem struct {
	notif client.Notif
}

func (n notifItem) FilterValue() string { return n.notif.SenderUsername }
func (n notifItem) Title() string {
	prefix := ""
	if n.notif.IsFriendRequest {
		prefix = "[FR] "
	}
	return prefix + n.notif.SenderUsername
}
func (n notifItem) Description() string {
	msg := n.notif.Message
	if msg == "" {
		msg = n.notif.Type
	}
	at := n.notif.CreatedAtText
	if at == "" {
		at = "--"
	}
	return fmt.Sprintf("%s  %s", msg, at)
}

// notifDelegate は notifItem のカスタムレンダラー。
type notifDelegate struct {
	styles Styles
}

func (d notifDelegate) Height() int                             { return 2 }
func (d notifDelegate) Spacing() int                           { return 0 }
func (d notifDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d notifDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	ni, ok := item.(notifItem)
	if !ok {
		return
	}

	title := ni.Title()
	desc := ni.Description()

	if !ni.notif.Seen {
		title = lipgloss.NewStyle().Bold(true).Render(title)
	}

	if index == m.Index() {
		title = d.styles.SelectedItem.Render(ni.Title())
		desc = lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Render(desc)
	} else {
		desc = d.styles.NormalItem.Render(desc)
	}

	fmt.Fprintf(w, "%s\n%s", title, desc)
}

// NotifsModel は通知ペインの Model。
type NotifsModel struct {
	list   list.Model
	keys   KeyMap
	styles Styles

	width  int
	height int

	focused bool
}

// NewNotifsModel は NotifsModel を初期化して返す。
func NewNotifsModel(styles Styles, keys KeyMap) NotifsModel {
	delegate := notifDelegate{styles: styles}
	l := list.New(nil, delegate, 0, 0)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Title = "Notifications"
	l.Styles.Title = styles.Header

	return NotifsModel{
		list:   l,
		keys:   keys,
		styles: styles,
	}
}

// Init は初期コマンドを返す。
func (m NotifsModel) Init() tea.Cmd { return nil }

// Update はメッセージを処理して状態を更新する。
func (m NotifsModel) Update(msg tea.Msg) (NotifsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case client.NotifsMsg:
		items := make([]list.Item, len(msg.Notifs))
		for i, n := range msg.Notifs {
			items[i] = notifItem{notif: n}
		}
		cmd := m.list.SetItems(items)
		return m, cmd

	case client.FriendRequestAcceptedMsg:
		m.removeNotif(msg.NotifID)
		return m, nil

	case client.FriendRequestRejectedMsg:
		m.removeNotif(msg.NotifID)
		return m, nil

	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}
		// a: フレンドリクエスト承認
		if key.Matches(msg, m.keys.Accept) {
			if n := m.selectedNotif(); n != nil && n.IsFriendRequest {
				return m, func() tea.Msg { return AcceptRequestMsg{NotifID: n.ID} }
			}
		}
		// d: フレンドリクエスト拒否
		if key.Matches(msg, m.keys.Reject) {
			if n := m.selectedNotif(); n != nil && n.IsFriendRequest {
				return m, func() tea.Msg { return RejectRequestMsg{NotifID: n.ID} }
			}
		}
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View はペインのレンダリング文字列を返す。
func (m NotifsModel) View() string {
	style := m.styles.Pane
	if m.focused {
		style = m.styles.ActivePane
	}
	return style.
		Width(m.width).
		Height(m.height).
		Render(m.list.View())
}

// SetFocus はフォーカス状態を設定する。
func (m *NotifsModel) SetFocus(focused bool) {
	m.focused = focused
}

// SetSize はペインのサイズを設定する。
func (m *NotifsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width-4, height-4)
}

// selectedNotif は現在選択中の通知を返す。選択なしの場合は nil。
func (m NotifsModel) selectedNotif() *client.Notif {
	item := m.list.SelectedItem()
	if item == nil {
		return nil
	}
	ni, ok := item.(notifItem)
	if !ok {
		return nil
	}
	n := ni.notif
	return &n
}

// removeNotif は指定 ID の通知をリストから除去する。
func (m *NotifsModel) removeNotif(notifID string) {
	items := m.list.Items()
	newItems := make([]list.Item, 0, len(items))
	for _, item := range items {
		ni, ok := item.(notifItem)
		if !ok || ni.notif.ID != notifID {
			newItems = append(newItems, item)
		}
	}
	m.list.SetItems(newItems)
}
