package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gen2brain/beeep"
)

const hourglassASCII = `
██╗  ██╗ ██████╗ ██╗   ██╗██████╗  ██████╗ ██╗      █████╗ ███████╗███████╗
██║  ██║██╔═══██╗██║   ██║██╔══██╗██╔════╝ ██║     ██╔══██╗██╔════╝██╔════╝
███████║██║   ██║██║   ██║██████╔╝██║  ███╗██║     ███████║███████╗███████╗
██╔══██║██║   ██║██║   ██║██╔══██╗██║   ██║██║     ██╔══██║╚════██║╚════██║
██║  ██║╚██████╔╝╚██████╔╝██║  ██║╚██████╔╝███████╗██║  ██║███████║███████║
╚═╝  ╚═╝ ╚═════╝  ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚═╝  ╚═╝╚══════╝╚══════╝
`

type appState int

const (
	landingState appState = iota
	inputState
	timerState
)

type model struct {
	state     appState
	textInput textinput.Model
	timer     timer.Model
	keymap    keymap
	help      help.Model
	quitting  bool
	timeout   time.Duration
	err       string
	notifIcon []byte
	menuIndex int
}

type keymap struct {
	start key.Binding
	stop  key.Binding
	reset key.Binding
	quit  key.Binding
}

func (m model) Init() tea.Cmd {
	switch m.state {
	case landingState:
		return nil
	case inputState:
		return textinput.Blink
	default:
		return m.timer.Init()
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timer.TickMsg:
		if m.state == timerState {
			var cmd tea.Cmd
			m.timer, cmd = m.timer.Update(msg)
			return m, cmd
		}

	case timer.StartStopMsg:
		if m.state == timerState {
			var cmd tea.Cmd
			m.timer, cmd = m.timer.Update(msg)
			m.keymap.stop.SetEnabled(m.timer.Running())
			m.keymap.start.SetEnabled(!m.timer.Running())
			return m, cmd
		}

	case timer.TimeoutMsg:
		_ = beeep.Alert("hourglass", "Time is up!!", m.notifIcon)

		m.quitting = true
		return m, tea.Quit

	case tea.KeyMsg:
		switch m.state {
		case landingState:
			switch msg.String() {
			case "up", "k":
				if m.menuIndex > 0 {
					m.menuIndex--
				}
			case "down", "j":
				if m.menuIndex < 1 {
					m.menuIndex++
				}
			case "enter":
				switch m.menuIndex {
				case 0: // Start Timer
					m.state = inputState
					return m, textinput.Blink
				case 1: // Quit
					m.quitting = true
					return m, tea.Quit
				}
			case "q", "ctrl+c":
				m.quitting = true
				return m, tea.Quit
			}
			return m, nil
		}

		switch {
		case key.Matches(msg, m.keymap.quit):
			m.quitting = true
			return m, tea.Quit
		}

		switch m.state {
		case inputState:
			switch msg.String() {
			case "enter":
				// Parse the duration from text input
				duration, err := time.ParseDuration(m.textInput.Value())
				if err != nil {
					m.err = "Invalid duration format. Use formats like '5m', '30s', '1h30m'"
					return m, nil
				}
				if duration <= 0 {
					m.err = "Duration must be greater than 0"
					return m, nil
				}

				// Set the timeout and switch to timer state
				m.timeout = duration
				m.timer = timer.NewWithInterval(duration, time.Millisecond)
				m.state = timerState
				m.err = ""
				return m, m.timer.Init()
			default:
				var cmd tea.Cmd
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
		case timerState:
			switch {
			case key.Matches(msg, m.keymap.reset):
				m.timer.Timeout = m.timeout
			case key.Matches(msg, m.keymap.start, m.keymap.stop):
				return m, m.timer.Toggle()
			}
		}
	}

	return m, nil
}

func (m model) helpView() string {
	return "\n" + m.help.ShortHelpView([]key.Binding{
		m.keymap.start,
		m.keymap.stop,
		m.keymap.reset,
		m.keymap.quit,
	})
}

func (m model) View() string {
	switch m.state {
	case landingState:
		menuItems := []string{"Start Timer", "Quit"}
		var menu string
		for i, item := range menuItems {
			cursor := "  "
			if i == m.menuIndex {
				cursor = "> "
			}
			menu += cursor + item + "\n"
		}

		return hourglassASCII + "\n\n" + menu + "\nUse ↑/↓ or j/k to navigate, Enter to select, q to quit"

	case inputState:
		s := "Enter timer duration (e.g., 5m, 30s, 1h30m):\n\n"
		s += m.textInput.View() + "\n"
		if m.err != "" {
			s += "\nError: " + m.err + "\n"
		}
		s += "\nPress Enter to start timer, q to quit"
		return s

	case timerState:
		s := m.timer.View()

		if m.timer.Timedout() {
			s = "All done!!"
		}
		s += "\n"

		if !m.quitting {
			s = "Timer: " + s
			s += m.helpView()
		}
		return s
	}

	return ""
}

func main() {
	beeep.AppName = "hourglass"

	// Initialize text input for duration entry
	ti := textinput.New()
	ti.Placeholder = "5m"
	ti.Focus()
	ti.CharLimit = 20
	ti.Width = 20

	m := model{
		state:     landingState,
		textInput: ti,
		keymap: keymap{
			start: key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "start"),
			),
			stop: key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "stop"),
			),
			reset: key.NewBinding(
				key.WithKeys("r"),
				key.WithHelp("r", "reset"),
			),
			quit: key.NewBinding(
				key.WithKeys("q", "ctrl+c"),
				key.WithHelp("q", "quit"),
			),
		},
		help: help.New(),
	}
	m.keymap.start.SetEnabled(false)

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("uh oh, we encountered an issue:", err)
		os.Exit(1)
	}
}
