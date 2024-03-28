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

// Doesn't look like it wants to align to the right... brobably I am using it wrong(?)
var timeStyle = lipgloss.NewStyle().
	Align(lipgloss.Position(lipgloss.Right)).
	Foreground(lipgloss.Color("240"))

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
	FollowURL: key.NewBinding(
		key.WithKeys("ctrl+f"),
		key.WithHelp("ctrl+f", "open twitch in browser"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	Select    key.Binding
	Refresh   key.Binding
	FollowURL key.Binding
	Quit      key.Binding
}

type model struct {
	table   table.Model
	keys    keyMap
	help    help.Model
	updated string
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
// Made it to clearscreen on init so I don't have to do it twice
// when updateing the tables throught updateStreamers()
func (m model) Init() tea.Cmd { return tea.ClearScreen }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+f":
			err := exec.Command("xdg-open", "https://www.twitch.tv/directory/following/live").Start()
			if err != nil {
				return m, tea.Println(err)
			}
		case "f5": // refresh the table
			t := "Last update: " + time.Now().Format("15:04:05")
			m.table.SetRows(updateStreamers())
			// Can't get rid of tea.Printf() because it breaks the table on refresh, why?!
			// seems like forcing a ClearScreen works as a workaround.
			// updating the model doesn't help the table breaks still
			return model{table: m.table, keys: m.keys, help: m.help, updated: t}, tea.ClearScreen
		case "q", "ctrl+c", "ctrl+d": // quit
			return m, tea.Quit
		case "enter": // select and open to default browser TODO: make it OS agnostic ?
			selection := "You selected " + m.table.SelectedRow()[1]
			err := exec.Command("xdg-open", m.table.SelectedRow()[3]).Start()
			if err != nil {
				return m, tea.Println(err)
			}
			return model{table: m.table, keys: m.keys, help: m.help, updated: selection}, nil
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	tbl := baseStyle.Render(m.table.View())
	hlp := m.help.View(m.keys)
	time_str := timeStyle.Render(m.updated)
	// +4 for the outer top/bottom and header
	height := m.table.Height() + 4 - strings.Count(tbl, "\n") - strings.Count(hlp, "\n")

	return tbl + strings.Repeat("\n", height) + time_str + "\n" + hlp + "\n"
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
		time.Sleep(300 * time.Millisecond)
	}
	defer f.Close()
	// clearTerm()

	return rows
}
