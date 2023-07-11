package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().
    BorderStyle(lipgloss.NormalBorder()).
    BorderForeground(lipgloss.Color("240"))

type model struct {
    table table.Model
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "f5":
            m.table.SetRows(updateStreamers())
            return m, tea.Printf("Refreshed!")
        case "q", "ctrl+c":
            return m, tea.Quit 
        case "enter":
            err := exec.Command("xdg-open", m.table.SelectedRow()[3]).Start()
            if err != nil { return m, tea.Println(err) }
            return m, tea.Printf("You selected %s", m.table.SelectedRow()[1])
        } 
    } 
    m.table, cmd = m.table.Update(msg)
    return m, cmd
}

func (m model) View() string {
    return baseStyle.Render(m.table.View()) + "\n"
}

// This is the main
func startTUI() error {
    // construct the collumns and header titles
    columns := []table.Column{
        {Title: "#", Width: 3},
        {Title: "Streamer", Width: 22},
        {Title: "Live?", Width: 8},
        {Title: "URL", Width: 22+22},
    }

    rows := updateStreamers()
    var tHeight int // max height of the table
    if len(rows) >= 15 {
        tHeight = 15
    } else {
        tHeight = len(rows)
    }

    t := table.New(
        table.WithColumns(columns),
        table.WithRows(rows),
        table.WithFocused(true),
        table.WithHeight(tHeight),
    )

    s := table.DefaultStyles()
    s.Header = s.Header.
        BorderStyle(lipgloss.NormalBorder()).
        BorderForeground(lipgloss.Color("240")).
        BorderBottom(true).
        Bold(true)
    s.Selected = s.Selected.
        Foreground(lipgloss.Color("229")).
        Background(lipgloss.Color("57")).
        Bold(true)
    t.SetStyles(s)

    // run the model
    m := model{t}
    if _, err := tea.NewProgram(m).Run(); err != nil {
        return err
    }

    return nil
}

func updateStreamers() (rows []table.Row) { 
    streamerlist := createStreamerlist()
    f, err := os.OpenFile(streamerlist, os.O_RDONLY, 644)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    defer f.Close()

    // check if the file is empty there is no point to continue
    // if there are no streamer in the file
    fi, err := f.Stat()
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    if fi.Size() == 0 {
        fmt.Println("The "+streamerlist+" is empty!")
        os.Exit(1)
    }

    var results []stream
    fScanner := bufio.NewScanner(f)
    fScanner.Split(bufio.ScanLines)
    for fScanner.Scan() {
        streamer := fScanner.Text()
        fmt.Println("Checking ", yellow+streamer+reset, "...")

        resp, err := getResponse(url+streamer)
        if err != nil {
            fmt.Println(err) 
        }

        if resp != nil {
            isLive, err := parse(resp)
            if err != nil {
                fmt.Println(err)
            }
            defer resp.Body.Close()

            results = append(results, stream{name: streamer, live: isLive, link: url+streamer})
        }
        // add a delay between each request so we won't get banned :S
        time.Sleep(1 * time.Second)
    }
    
    for i, st := range results {
        if st.live {
            rows = append(rows, table.Row{strconv.Itoa(i), st.name, "LIVE", st.link })
        } else {
            rows = append(rows, table.Row{strconv.Itoa(i), st.name, "OFFLINE", st.link })
        }
    }

    clearTerm()

    return rows
}

