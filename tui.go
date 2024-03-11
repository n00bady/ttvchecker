package main

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("55"))

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
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	Select  key.Binding
	Refresh key.Binding
	Quit    key.Binding
}

type model struct {
	table table.Model
	keys  keyMap
	help  help.Model
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

// Maybe I should initialize the table here ?
// But the example doesn't do that...
func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "f5": // refresh the table
			m.table.SetRows(updateStreamers())
			return m, tea.Printf("Refreshed!")
		case "q", "ctrl+c": // quit
			return m, tea.Quit
		case "enter": // select and open to default browser TODO: make it OS agnostic ?
			err := exec.Command("xdg-open", m.table.SelectedRow()[3]).Start()
			if err != nil {
				return m, tea.Println(err)
			}
			return m, tea.Printf("You selected %s", m.table.SelectedRow()[1])
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	tbl := baseStyle.Render(m.table.View())
	hlp := m.help.View(m.keys)
	// +4 for the outer top/bottom and header
	height := m.table.Height() + 4 - strings.Count(tbl, "\n") - strings.Count(hlp, "\n")

	return tbl + strings.Repeat("\n", height) + hlp + "\n"
}

// This is the "main"
func startTUI() error {
	// create the collumns and header titles
	columns := []table.Column{
		{Title: "#", Width: 3},
		{Title: "Streamer", Width: 22},
		{Title: "Live?", Width: 8},
		{Title: "URL", Width: 22 + 22},
	}

	// create the rows
	rows := updateStreamers()
	var tHeight int // max height of the table
	if len(rows) >= 15 {
		tHeight = 15
	} else {
		tHeight = len(rows)
	}

	// make a table style
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		Foreground(lipgloss.Color("#772ce7")).
		BorderForeground(lipgloss.Color("#441194")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#5b16c5")).
		Bold(true)

	// Create the initial table
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tHeight),
		table.WithStyles(s),
	)

	// create the help
	h := help.New()

	// run the model
	m := model{
		table: t,
		help:  h,
	}
	m.keys = keys
	if _, err := tea.NewProgram(m).Run(); err != nil {
		return err
	}

	return nil
}

func updateStreamers() (rows []table.Row) {
	f := openStreamerlist()

	fScanner := bufio.NewScanner(f)
	fScanner.Split(bufio.ScanLines)
	i := 0
	for fScanner.Scan() {
		streamer := fScanner.Text()
		// probably I should use bubbletea to print those
		// but it works fine like that so I will leave it
		// until it bites my ass
		fmt.Println("Checking ", yellow+streamer+reset, "...")

		resp, err := getResponse(url + streamer)
		if err != nil {
			log.Println(err)
		}

		if resp != nil {
			isLive, err := parse(resp)
			if err != nil {
				log.Println(err)
			}
			defer resp.Body.Close()

			if isLive {
				rows = append(rows, table.Row{strconv.Itoa(i), streamer, "LIVE", url + streamer})
			} else {
				rows = append(rows, table.Row{strconv.Itoa(i), streamer, "OFFLINE", url + streamer})
			}
		}
		// add a delay between each request so we won't get banned :S
		i++
		time.Sleep(500 * time.Millisecond)
	}
	defer f.Close()
	clearTerm()

	return rows
}
