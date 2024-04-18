package main

import (
	"bufio"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("55"))

var statusStyle = lipgloss.NewStyle().
	Width(78).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("54")).
	Padding(0, 1, 0, 1).
	Foreground(lipgloss.Color("242"))

var helpStyle = lipgloss.NewStyle().
	Padding(0, 1, 0, 1).
	Width(78).
	Align(lipgloss.Left)

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

type updatedMsg string

var Rows []table.Row

type model struct {
	table        table.Model
	keys         keyMap
	help         help.Model
	updated      string
	spinner      spinner.Model
	spin         bool
	index        int
	streamerlist []string
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
			m.spin = true
			m.updated = "Checking " + m.streamerlist[0] + "... "
			return m, tea.Batch(refreshStreamer(m.streamerlist[0], 0), m.spinner.Tick)
		case "q", "ctrl+c", "ctrl+d": // quit
			return m, tea.Quit
		case "enter": // select and open to default browser
			m.updated = "You selected " + m.table.SelectedRow()[1]
			err := exec.Command("xdg-open", url + m.table.SelectedRow()[1]).Start()
			if err != nil {
				return m, tea.Println(err)
			}
			return m, nil
		}
	case updatedMsg:
		if m.index >= len(m.streamerlist)-1 {
			m.spin = false
			m.table.SetRows(Rows)
			m.index = 0
			m.updated = "Last updated: " + time.Now().Format("15:04:05")
			Rows = nil
			return m, cmd
		}
		m.index++
		m.updated = "Checking " + m.streamerlist[m.index] + "... "
		return m, refreshStreamer(m.streamerlist[m.index], m.index)
	case spinner.TickMsg:
		var cmd tea.Cmd
		if m.spin {
			m.spinner, cmd = m.spinner.Update(msg)
		}
		return m, cmd
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	tbl := baseStyle.Render(m.table.View())
	hlp := helpStyle.Render(m.help.View(m.keys))
	status := statusStyle.Render(m.updated)
	if m.spin {
		status = statusStyle.Render(m.spinner.View() + m.updated)
	}
	// +4 for the outer top/bottom and header
	height := m.table.Height() + 4 - strings.Count(tbl, "\n") - strings.Count(hlp, "\n")

	return tbl + strings.Repeat("\n", height) + status + "\n" + hlp + "\n"
}

// This is the "main"
func startTUI() error {
	// create the columns and header titles
	columns := []table.Column{
		{Title: "#", Width: 3},
		{Title: "Streamer", Width: 22},
		{Title: "Live?", Width: 8},
		{Title: "Title", Width: 37}, // so it fits in 80 columns
	}

	// create the rows
    // TODO: Change the 1st time of checking streamers to construct the table 
    // so it doesn't use the old code
	rows := updateStreamers()
    // It helps to add the streamerlist at the start to be able to show progress
    // updates but also if you add a streamer while the TUI is open you have to
    // restart.
	streamerlist := initStreamerList()
	// var tHeight int // max height of the table
	// if len(rows) >= 15 {
	// 	tHeight = 15
	// } else {
	// 	tHeight = len(rows)
	// }
    tHeight := 15 //Temp change so it will always until I decide what to do with it

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

	u := "Last update: " + time.Now().Format("15:04:05")

	spin := spinner.New()
	spin.Spinner = spinner.Dot
	spin.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	// run the model
	m := model{
		table:        t,
		help:         h,
		updated:      u,
		spinner:      spin,
		spin:         false,
		index:        0,
		streamerlist: streamerlist,
	}
	m.keys = keys
	if _, err := tea.NewProgram(m).Run(); err != nil {
		return err
	}

	return nil
}

// Takes the streamer string and the index, update the Rows and 
// returns a tea.Cmd to show progress
func refreshStreamer(streamer string, index int) tea.Cmd {
	d := 300 * time.Millisecond
	resp, _ := getResponse(url + streamer)

	if resp != nil {
		isLive, title, _ := parse(resp)
		defer resp.Body.Close()

		if isLive {
			Rows = append(Rows, table.Row{strconv.Itoa(index), streamer, "LIVE", title})
		} else {
			Rows = append(Rows, table.Row{strconv.Itoa(index), streamer, "OFFLINE", title})
		}
	}
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return updatedMsg("done")
	})
}

// Constructs a list of all the streamers to go with the model
func initStreamerList() []string {
	var streamerlist []string
	f := openStreamerlist()

	fScanner := bufio.NewScanner(f)
	fScanner.Split(bufio.ScanLines)
	for fScanner.Scan() {
		streamerlist = append(streamerlist, fScanner.Text())
	}

	return streamerlist
}

// Soon hopefully this will not be needed...
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
		// fmt.Println("Checking ", yellow+streamer+reset, "...")

		resp, err := getResponse(url + streamer)
		if err != nil {
			log.Println(err)
		}

		if resp != nil {
			isLive, title, err := parse(resp)
			if err != nil {
				log.Println(err)
			}
			defer resp.Body.Close()

			if isLive {
				rows = append(rows, table.Row{strconv.Itoa(i), streamer, "LIVE", title})
			} else {
				rows = append(rows, table.Row{strconv.Itoa(i), streamer, "OFFLINE", title})
			}
		}
		// add a delay between each request so we won't get banned :S
		i++
		time.Sleep(300 * time.Millisecond)
	}
	defer f.Close()

	return rows
}
