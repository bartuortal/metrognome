package main

import (
	"fmt"
	"io/fs"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/bartuortal/metrognome/soundTable"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gopxl/beep/effects"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type (
	errMsg error
)

type model struct {
	soundTable soundTable.SoundTable
	editing    int
	err        error
	textInput  textinput.Model
}

func getMaxFileLength(files []fs.DirEntry) (maxLength int) {
	for _, file := range files {
		maxLength = int(math.Max(float64(len(file.Name())), float64(maxLength)))
	}
	return maxLength
}

func playSounds(m model, ticker *time.Ticker) {
	for range ticker.C {
		for _, row := range m.soundTable.Rows() {
			random := rand.Intn(60 * 60)
			// It is always a number
			freq, _ := strconv.Atoi(row[2])
			if random < freq {
				f, err := os.Open(fmt.Sprintf("./sounds/%s", row[1]))
				if err != nil {
					log.Fatal(err)
				}
				streamer, format, err := mp3.Decode(f)
				if err != nil {
					log.Fatal(err)
				}
				err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/5))
				if err != nil {
					log.Println("error while initializing speaker:", err)
				}
				volume, _ := strconv.ParseFloat(row[3], 64)
				volume = (volume - 100) / 50
				streamerWithVolume := &effects.Volume{
					Streamer: streamer,
					Base:     2,
					Volume:   volume,
					Silent:   false,
				}
				speaker.Play(streamerWithVolume)
			}
		}
	}
}

func main() {
	ticker := time.NewTicker(time.Second)

	entries, err := os.ReadDir("./sounds")
	if err != nil {
		log.Fatal(err)
	}

	tableNameWidth := getMaxFileLength(entries)
	columns := []table.Column{
		{Title: "Id", Width: 2},
		{Title: "Name", Width: tableNameWidth},
		{Title: "Frequency(/hr)", Width: 15},
		{Title: "Sound", Width: 5},
	}

	rows := make([]table.Row, len(entries))

	for i, entry := range entries {
		rows[i] = table.Row{fmt.Sprint(i + 1), entry.Name(), "0", "100"}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	ti := textinput.New()
	ti.Focus()
	ti.Width = 20

	m := model{
		soundTable: soundTable.NewFromTable(t),
		textInput:  ti,
		editing:    -1,
	}
	go playSounds(m, ticker)
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if m.editing < 0 {

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				if m.soundTable.Focused() {
					m.soundTable.Blur()
				} else {
					m.soundTable.Focus()
				}
			case "q", "ctrl+c":
				return m, tea.Quit
			case "enter":
				selectedRow := m.soundTable.SelectedSoundRow()
				m.editing = selectedRow.Id - 1
				m.textInput.Placeholder = fmt.Sprint(selectedRow.Freq)
			case "h", "left":
				selectedRow := m.soundTable.SelectedSoundRow()
				selectedRow.DecrementVolume()
				m.soundTable.UpdateViewport()
			case "l", "right":
				selectedRow := m.soundTable.SelectedSoundRow()
				selectedRow.IncrementVolume()
				m.soundTable.UpdateViewport()
			}
		}
		newTable, cmd := m.soundTable.Update(msg)
		m.soundTable = soundTable.NewFromTable(newTable)
		return m, cmd
	} else {
	outer:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyEnter:
				rows := m.soundTable.Rows()
				if _, err := strconv.Atoi(m.textInput.Value()); err != nil {
					m.editing = -1
					break outer
				}
				rows[m.editing][2] = m.textInput.Value()
				m.soundTable.SetRows(rows)
				m.editing = -1
				m.textInput.SetValue("")

			case tea.KeyCtrlC, tea.KeyEsc:
				return m, tea.Quit
			}

		// We handle errors just like any other message
		case errMsg:
			m.err = msg
			return m, nil
		}

		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

}

func (m model) View() string {
	if m.editing < 0 {
		return baseStyle.Render(m.soundTable.View()) + "\n"
	} else {
		return fmt.Sprintf(
			"What's the new frequancy?\n\n%s",
			m.textInput.View(),
		) + "\n"
	}
}
