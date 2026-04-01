// Package ui は TUI のすべての Model・View・スタイルを提供する。
package ui

import (
	"github.com/charmbracelet/bubbles/key"
)

// KeyMap はアプリ全体のキーバインドを定義する。
type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Enter    key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	Status   key.Binding
	Refresh  key.Binding
	Quit     key.Binding
	Help     key.Binding
	Accept   key.Binding
	Reject   key.Binding
	Escape   key.Binding
}

// DefaultKeyMap はデフォルトキーバインドを返す。
var DefaultKeyMap = KeyMap{
	Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Enter:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	Tab:      key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next pane")),
	ShiftTab: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev pane")),
	Status:   key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "set status")),
	Refresh:  key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
	Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Help:     key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Accept:   key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "accept")),
	Reject:   key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "reject")),
	Escape:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "close")),
}

// ShortHelp は日常的に使うキーのみ返す（help.KeyMap インターフェース実装）。
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Tab, k.Status, k.Refresh, k.Quit, k.Help}
}

// FullHelp は全キーを分類して返す（help.KeyMap インターフェース実装）。
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter},
		{k.Tab, k.ShiftTab},
		{k.Status, k.Refresh, k.Quit, k.Help},
		{k.Accept, k.Reject, k.Escape},
	}
}
