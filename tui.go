package main

import (
	"math/rand"
	"os/exec"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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

func (m model) Init() tea.Cmd {
	return tea.Batch(tea.ClearScreen, refreshStreamer(m.streamerlist[0], 0), m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
        case "j","k","u","d":
            m.table, cmd = m.table.Update(msg)
            cmds = append(cmds, cmd)
		case "ctrl+f":
			err := exec.Command("xdg-open", "https://www.twitch.tv/directory/following/live").Start()
			if err != nil {
                cmds = append(cmds, tea.Println(err))
			}
		case "f5": // refresh the table
			if !m.spin {
                m.spin = true
                m.updated = "Checking " + m.streamerlist[0] + "... "
                cmds = append(cmds, refreshStreamer(m.streamerlist[0], 0), m.spinner.Tick)
			}
		case "q", "ctrl+c", "ctrl+d": // quit
			return m, tea.Quit
		case "enter": // select and open to default browser
			m.updated = "You selected " + m.table.SelectedRow()[1]
			err := exec.Command("xdg-open", url+m.table.SelectedRow()[1]).Start()
			if err != nil {
                cmds = append(cmds, tea.Println(err))
			}
		}
	case updatedMsg:
		if m.index >= len(m.streamerlist)-1 {
			m.spin = false
			m.table.SetRows(Rows)
			m.index = 0
			m.updated = "Last updated: " + time.Now().Format("15:04:05")
			Rows = nil
		} else {
            m.index++
            m.updated = "Checking " + m.streamerlist[m.index] + "... "
            cmds = append(cmds, refreshStreamer(m.streamerlist[m.index], m.index))
        }
	case spinner.TickMsg:
		if m.spin {
            m.spinner, cmd = m.spinner.Update(msg)
            cmds = append(cmds, cmd)
		}
	}

    return m, tea.Batch(cmds...)
}

func (m model) View() string {
	tbl := baseStyle.Render(m.table.View())
	hlp := helpStyle.Render(m.help.View(m.keys))
	status := statusStyle.Render(m.updated)
	if m.spin {
		status = statusStyle.Render(m.spinner.View() + m.updated)
	}
	// +4 for the outer top/bottom and header
	// height := m.table.Height() + 4 - strings.Count(tbl, "\n") - strings.Count(hlp, "\n")

	return tbl + "\n" + status + "\n" + hlp + "\n"
}

func InitialModel() model {
	columns := []table.Column{
		{Title: "#", Width: 3},
		{Title: "Streamer", Width: 22},
		{Title: "Live?", Width: 8},
		{Title: "Title", Width: 37}, // so it kinda fits in 80 columns
	}

	// It helps to add the streamerlist at the start to be able to show progress
	// updates but also if you add a streamer while the TUI is open you have to
	// restart.
	streamerlist := initStreamerList()
	// Temp change so it will always show 15 rows until I decide what to do with it
	tHeight := 15

	// Init the table style
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
		// table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tHeight),
		table.WithStyles(s),
	)

	// Init the help
	h := help.New()

	u := "Last update: " + time.Now().Format("15:04:05")

	spin := spinner.New()
	spin.Spinner = spinner.Dot
	spin.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	// Init the actual model
	m := model{
		table:        t,
		keys:         keys,
		help:         h,
		updated:      u,
		spinner:      spin,
		spin:         true,
		index:        0,
		streamerlist: streamerlist,
	}

	return m
}

// This is where it starts
func startTUI() error {
	rand.New(rand.NewSource(time.Now().Unix()))
	m := InitialModel()
	if _, err := tea.NewProgram(m).Run(); err != nil {
		return err
	}

	return nil
}
