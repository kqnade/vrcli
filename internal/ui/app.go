package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kqnade/vrcli/internal/client"
	"github.com/kqnade/vrcli/internal/config"
)

// ペイン識別子
type pane int

const (
	paneFriends pane = iota
	paneNotifs
)

// ポーリング用内部メッセージ
type pollFriendsMsg struct{}
type pollNotifsMsg struct{}

// clearErrMsg はエラー表示を消去するための内部メッセージ。
// errSeq により古いタイマーが新しいエラーを消さないようにする。
type clearErrMsg struct{ seq int }

// errState はエラー表示の状態を保持する。
type errState struct {
	err error
	seq int
}

// AppModel は Bubbletea のルートモデル。
type AppModel struct {
	friends     FriendsModel
	notifs      NotifsModel
	statusModal StatusModal
	help        help.Model

	vrc *client.VRCClient
	cfg *config.Config

	activePane pane
	styles     Styles
	keys       KeyMap

	width  int
	height int

	currentUser       string
	currentUserID     string
	currentStatus     client.UserStatus
	currentStatusDesc string

	errState   *errState
	errSeqNext int
	showHelp   bool

	// ポーリング中フラグ（前のリクエストが完了してから次をスケジュールするために使う）
	friendsPolling bool
	notifsPolling  bool
}

const errDisplayDuration = 5 * time.Second

// New は AppModel を初期化して返す。
func New(vrc *client.VRCClient, cfg *config.Config) AppModel {
	styles := NewStyles()
	keys := DefaultKeyMap
	friends := NewFriendsModel(styles, keys)
	friends.SetFocus(true) // 初期フォーカスをここで設定（Init は値レシーバのため）
	return AppModel{
		friends:     friends,
		notifs:      NewNotifsModel(styles, keys),
		statusModal: NewStatusModal(styles, keys),
		help:        help.New(),
		vrc:         vrc,
		cfg:         cfg,
		activePane:  paneFriends,
		styles:      styles,
		keys:        keys,
	}
}

// Init は初期コマンドを返す。セッションチェックを開始する。
func (m AppModel) Init() tea.Cmd {
	return m.vrc.CheckSession()
}

// Update はメッセージを処理して状態を更新する。
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// モーダルが表示中はすべてのキーをモーダルに委譲
	if m.statusModal.IsVisible() {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			var cmd tea.Cmd
			m.statusModal, cmd = m.statusModal.Update(keyMsg)
			return m, cmd
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.distributeSize()
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
			m.distributeSize()
			return m, nil

		case key.Matches(msg, m.keys.Tab):
			m.switchPane(paneNotifs)
			return m, nil

		case key.Matches(msg, m.keys.ShiftTab):
			m.switchPane(paneFriends)
			return m, nil

		case key.Matches(msg, m.keys.Status):
			m.statusModal.Open(m.currentStatus, m.currentStatusDesc)
			return m, nil

		case key.Matches(msg, m.keys.Refresh):
			return m, tea.Batch(m.vrc.FetchFriends(), m.vrc.FetchNotifs())
		}

		// アクティブペインにキーを転送
		return m.updateActivePane(msg)

	case client.AuthOKMsg:
		m.currentUserID = msg.UserID
		m.currentUser = msg.DisplayName
		m.currentStatus = msg.Status
		m.currentStatusDesc = msg.StatusDescription
		// 初回フェッチ開始。ポーリングはレスポンス後にスケジュールする
		m.friendsPolling = true
		m.notifsPolling = true
		return m, tea.Batch(m.vrc.FetchFriends(), m.vrc.FetchNotifs())

	case client.FriendsMsg:
		var cmd tea.Cmd
		m.friends, cmd = m.friends.Update(msg)
		// 前回フェッチが完了したので次のポーリングをスケジュール
		if m.friendsPolling {
			cmd = tea.Batch(cmd, m.pollFriends())
		}
		return m, cmd

	case client.NotifsMsg:
		var cmd tea.Cmd
		m.notifs, cmd = m.notifs.Update(msg)
		if m.notifsPolling {
			cmd = tea.Batch(cmd, m.pollNotifs())
		}
		return m, cmd

	case client.ErrMsg:
		// 認証エラー(401)の場合はポーリングを継続しない
		if msg.IsAuth {
			m.friendsPolling = false
			m.notifsPolling = false
			m.showError(msg.Err)
			m.distributeSize()
			return m, tea.Batch(m.scheduleClearErr(), tea.Quit)
		}
		m.showError(msg.Err)
		m.distributeSize()
		return m, m.scheduleClearErr()

	case client.StatusUpdatedMsg:
		m.statusModal.Close()
		m.currentStatus = msg.Status
		m.currentStatusDesc = msg.StatusDesc
		return m, nil

	case client.FriendRequestAcceptedMsg:
		var cmd tea.Cmd
		m.notifs, cmd = m.notifs.Update(msg)
		return m, cmd

	case client.FriendRequestRejectedMsg:
		var cmd tea.Cmd
		m.notifs, cmd = m.notifs.Update(msg)
		return m, cmd

	case AcceptRequestMsg:
		return m, m.vrc.AcceptFriendRequest(msg.NotifID)

	case RejectRequestMsg:
		return m, m.vrc.RejectFriendRequest(msg.NotifID)

	case ConfirmStatusMsg:
		return m, m.vrc.UpdateStatus(string(msg.Status), msg.Desc)

	case pollFriendsMsg:
		// 次のポーリングは FriendsMsg 受信後にスケジュールする（オーバーラップ防止）
		return m, m.vrc.FetchFriends()

	case pollNotifsMsg:
		return m, m.vrc.FetchNotifs()

	case clearErrMsg:
		// 古いタイマーが新しいエラーを消さないようシーケンス番号で確認する
		if m.errState != nil && m.errState.seq == msg.seq {
			m.errState = nil
			m.distributeSize()
		}
		return m, nil
	}

	return m, nil
}

// View はアプリ全体のレンダリング文字列を返す。
func (m AppModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// モーダルオーバーレイ
	if m.statusModal.IsVisible() {
		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			m.statusModal.View())
	}

	// ヘッダー
	header := m.renderHeader()

	// ペイン行
	friendsView := m.friends.View()
	notifsView := m.notifs.View()
	panes := lipgloss.JoinHorizontal(lipgloss.Top, friendsView, notifsView)

	// フッター
	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left, header, panes, footer)
}

// --- プライベートメソッド ---

func (m *AppModel) distributeSize() {
	headerHeight := 1
	footerHeight := 1
	if m.showHelp {
		footerHeight = 4
	}
	if m.errState != nil {
		footerHeight++
	}
	paneHeight := m.height - headerHeight - footerHeight
	paneHeight = max(paneHeight, 1)

	halfWidth := m.width / 2
	m.friends.SetSize(halfWidth, paneHeight)
	m.notifs.SetSize(m.width-halfWidth, paneHeight)
}

func (m *AppModel) switchPane(p pane) {
	m.activePane = p
	m.friends.SetFocus(p == paneFriends)
	m.notifs.SetFocus(p == paneNotifs)
}

func (m AppModel) updateActivePane(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.activePane {
	case paneFriends:
		m.friends, cmd = m.friends.Update(msg)
	case paneNotifs:
		m.notifs, cmd = m.notifs.Update(msg)
	}
	return m, cmd
}

func (m *AppModel) showError(err error) {
	seq := m.errSeqNext
	m.errSeqNext++
	m.errState = &errState{err: err, seq: seq}
}

func (m AppModel) scheduleClearErr() tea.Cmd {
	seq := m.errSeqNext - 1 // 直前に showError で割り当てたシーケンス番号
	return tea.Tick(errDisplayDuration, func(_ time.Time) tea.Msg {
		return clearErrMsg{seq: seq}
	})
}

const minPollInterval = 5 * time.Second

func (m AppModel) pollFriends() tea.Cmd {
	interval := max(m.cfg.App.PollFriendsInterval, minPollInterval)
	return tea.Tick(interval, func(_ time.Time) tea.Msg {
		return pollFriendsMsg{}
	})
}

func (m AppModel) pollNotifs() tea.Cmd {
	interval := max(m.cfg.App.PollNotifsInterval, minPollInterval)
	return tea.Tick(interval, func(_ time.Time) tea.Msg {
		return pollNotifsMsg{}
	})
}

func (m AppModel) renderHeader() string {
	statusEmoji := m.currentStatus.Emoji()
	statusStr := string(m.currentStatus)
	if statusStr == "" {
		statusStr = "..."
	}
	title := fmt.Sprintf("vrchat-tui | %s %s %s", m.currentUser, statusEmoji, statusStr)
	return lipgloss.NewStyle().
		Width(m.width).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#1E1B4B")).
		Padding(0, 1).
		Render(title)
}

func (m AppModel) renderFooter() string {
	var parts []string
	if m.errState != nil {
		parts = append(parts, m.styles.ErrorBar.Render("Error: "+m.errState.err.Error()))
	}
	if m.showHelp {
		parts = append(parts, m.help.FullHelpView(m.keys.FullHelp()))
	} else {
		parts = append(parts, m.styles.StatusBar.Render(m.help.ShortHelpView(m.keys.ShortHelp())))
	}
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
