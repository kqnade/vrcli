package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/kqnade/vrcli/internal/client"
)

// Trust Rank カラー定数（ダークターミナル前提）。
const (
	ColorVisitor   = lipgloss.Color("#9E9E9E")
	ColorNewUser   = lipgloss.Color("#2196F3")
	ColorUser      = lipgloss.Color("#4CAF50")
	ColorKnownUser = lipgloss.Color("#FF9800")
	ColorTrusted   = lipgloss.Color("#9C27B0")
	ColorFriend    = lipgloss.Color("#FFEB3B")
)

// Styles はアプリ全体の Lipgloss スタイルを保持する。
type Styles struct {
	// レイアウト
	Pane       lipgloss.Style
	ActivePane lipgloss.Style

	// リストアイテム
	SelectedItem lipgloss.Style
	NormalItem   lipgloss.Style

	// ステータスバー
	StatusBar lipgloss.Style
	ErrorBar  lipgloss.Style

	// モーダル
	Modal      lipgloss.Style
	ModalTitle lipgloss.Style

	// ヘッダー
	Header lipgloss.Style
}

// NewStyles はデフォルトスタイルを生成して返す。
func NewStyles() Styles {
	pane := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#555555")).
		Padding(0, 1)

	activePane := pane.
		BorderForeground(lipgloss.Color("#7C3AED"))

	selectedItem := lipgloss.NewStyle().
		Background(lipgloss.Color("#3730A3")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)

	normalItem := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DDDDDD"))

	statusBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9E9E9E"))

	errorBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F44336")).
		Bold(true)

	modal := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#7C3AED")).
		Padding(1, 2).
		Background(lipgloss.Color("#1A1A2E"))

	modalTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		MarginBottom(1)

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#1E1B4B")).
		Padding(0, 1)

	return Styles{
		Pane:         pane,
		ActivePane:   activePane,
		SelectedItem: selectedItem,
		NormalItem:   normalItem,
		StatusBar:    statusBar,
		ErrorBar:     errorBar,
		Modal:        modal,
		ModalTitle:   modalTitle,
		Header:       header,
	}
}

// TrustRankColor は TrustRank に対応する Lipgloss Color を返す。
func TrustRankColor(rank client.TrustRank) lipgloss.Color {
	switch rank {
	case client.TrustNewUser:
		return ColorNewUser
	case client.TrustUser:
		return ColorUser
	case client.TrustKnownUser:
		return ColorKnownUser
	case client.TrustTrusted:
		return ColorTrusted
	case client.TrustFriend:
		return ColorFriend
	default:
		return ColorVisitor
	}
}
