package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kqnade/vrcli/internal/client"
)

// friendItem は bubbles/list.Item インターフェースを実装するラッパー。
type friendItem struct {
	friend client.Friend
}

func (f friendItem) FilterValue() string { return f.friend.DisplayName }
func (f friendItem) Title() string       { return f.friend.DisplayName }
func (f friendItem) Description() string {
	status := f.friend.Status.Emoji() + " " + string(f.friend.Status)
	if f.friend.StatusDescription != "" {
		status += " — " + f.friend.StatusDescription
	}
	return status
}

// friendDelegate は friendItem のカスタムレンダラー。
type friendDelegate struct {
	styles Styles
}

func (d friendDelegate) Height() int                             { return 2 }
func (d friendDelegate) Spacing() int                           { return 0 }
func (d friendDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d friendDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	fi, ok := item.(friendItem)
	if !ok {
		return
	}

	nameColor := TrustRankColor(fi.friend.TrustRank)
	name := lipgloss.NewStyle().Foreground(nameColor).Render(fi.Title())
	desc := fi.Description()

	if index == m.Index() {
		// 選択時は SelectedItem のスタイルを維持しつつ TrustRank カラーを前景色に適用する
		name = d.styles.SelectedItem.Foreground(nameColor).Render(fi.Title())
		desc = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCCCC")).
			Render(desc)
	} else {
		desc = d.styles.NormalItem.Render(desc)
	}

	fmt.Fprintf(w, "%s\n%s", name, desc)
}

// FriendsModel はフレンドリストペインの Model。
type FriendsModel struct {
	list   list.Model
	keys   KeyMap
	styles Styles

	width  int
	height int

	focused bool
}

// NewFriendsModel は FriendsModel を初期化して返す。
func NewFriendsModel(styles Styles, keys KeyMap) FriendsModel {
	delegate := friendDelegate{styles: styles}
	l := list.New(nil, delegate, 0, 0)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Title = "Friends"
	l.Styles.Title = styles.Header

	return FriendsModel{
		list:   l,
		keys:   keys,
		styles: styles,
	}
}

// Init は初期コマンドを返す（FriendsModel 単体は何もしない）。
func (m FriendsModel) Init() tea.Cmd { return nil }

// Update はメッセージを処理して状態を更新する。
func (m FriendsModel) Update(msg tea.Msg) (FriendsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case client.FriendsMsg:
		items := make([]list.Item, len(msg.Friends))
		for i, f := range msg.Friends {
			items[i] = friendItem{friend: f}
		}
		cmd := m.list.SetItems(items)
		return m, cmd

	case tea.KeyMsg:
		if !m.focused {
			return m, nil
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
func (m FriendsModel) View() string {
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
func (m *FriendsModel) SetFocus(focused bool) {
	m.focused = focused
}

// SetSize はペインのサイズを設定する。
func (m *FriendsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	// border(2) + padding(2) を除いたサイズをリストに渡す
	m.list.SetSize(width-4, height-4)
}

// SelectedFriend は現在選択中のフレンドを返す。選択なしの場合は nil。
func (m FriendsModel) SelectedFriend() *client.Friend {
	item := m.list.SelectedItem()
	if item == nil {
		return nil
	}
	fi, ok := item.(friendItem)
	if !ok {
		return nil
	}
	f := fi.friend
	return &f
}
