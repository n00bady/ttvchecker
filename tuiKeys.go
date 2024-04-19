package main

import "github.com/charmbracelet/bubbles/key"

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("f5"),
		key.WithHelp("f5", "refresh"),
	),
	Following: key.NewBinding(
		key.WithKeys("ctrl+f"),
		key.WithHelp("ctrl+f", "open twitch in browser"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c", "ctrl+d"),
		key.WithHelp("q", "quit"),
	),
}

type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	Select    key.Binding
	Refresh   key.Binding
	Following key.Binding
	Quit      key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Select, k.Refresh, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Select},
		{k.Refresh, k.Quit},
	}
}
