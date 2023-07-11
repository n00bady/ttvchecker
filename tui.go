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

// Our model it's only a stantard table model
// from the bubbles lib
type model struct {
    table table.Model
}

// Maybe I should initialize the table here ?
// But the example doesn't do that...
func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "f5":              // refresh the table
            m.table.SetRows(updateStreamers())
            return m, tea.Printf("Refreshed!")
        case "q", "ctrl+c":     // quit
            return m, tea.Quit 
        case "enter":           // select and open to default browser TODO: make it OS agnostic ?
            err := exec.Command("xdg-open", m.table.SelectedRow()[3]).Start()
            if err != nil { return m, tea.Println(err) }
            return m, tea.Printf("You selected %s", m.table.SelectedRow()[1])
        } 
    } 
    m.table, cmd = m.table.Update(msg)
    return m, cmd
}

func (m model) View() string {
    return baseStyle.Render(m.table.View()) + "\n" + dark_gray + "j-down k-up enter-select q-quit" + reset + "\n"
}

// This is the "main"
func startTUI() error {
    // create the collumns and header titles
    columns := []table.Column{
        {Title: "#", Width: 3},
        {Title: "Streamer", Width: 22},
        {Title: "Live?", Width: 8},
        {Title: "URL", Width: 22+22},
    }

    // create the rows
    rows := updateStreamers()
    var tHeight int // max height of the table
    if len(rows) >= 15 {
        tHeight = 15
    } else {
        tHeight = len(rows)
    }

    // Create the initial table
    t := table.New(
        table.WithColumns(columns),
        table.WithRows(rows),
        table.WithFocused(true),
        table.WithHeight(tHeight),
    )

    // initialize a default style and then modify it
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
    // --- These should probably go to a helper function ? ----
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
    // --------------------------------------------------------

    fScanner := bufio.NewScanner(f)
    fScanner.Split(bufio.ScanLines)
    i := 0
    for fScanner.Scan() {
        streamer := fScanner.Text()
        // probably I should use bubbletea to print those
        // but it works fine like that so I will leave it 
        // until it bites my ass
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

            if isLive {
                rows = append(rows, table.Row{strconv.Itoa(i), streamer, "LIVE", url+streamer})
            } else {
                rows = append(rows, table.Row{strconv.Itoa(i), streamer, "OFFLINE", url+streamer})
            }
        }
        // add a delay between each request so we won't get banned :S
        i++
        time.Sleep(1 * time.Second)
    } 
    clearTerm()

    return rows
}

