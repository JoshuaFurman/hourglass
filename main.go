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
	"github.com/charmbracelet/lipgloss"
	"github.com/gen2brain/beeep"
)

type appState int

const (
	menuState appState = iota
	inputState
	timerState
	completedState
)

type model struct {
	state      appState
	textInput  textinput.Model
	timer      timer.Model
	keymap     keymap
	help       help.Model
	quitting   bool
	timeout    time.Duration
	err        string
	notifIcon  []byte
	menuCursor int
	menuItems  []string
	width      int
	height     int
}

type keymap struct {
	start key.Binding
	stop  key.Binding
	reset key.Binding
	quit  key.Binding
}

func (m model) Init() tea.Cmd {
	switch m.state {
	case menuState:
		return nil
	case inputState:
		return textinput.Blink
	case timerState:
		return m.timer.Init()
	case completedState:
		return nil
	}
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

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

		// Switch to completed state and wait for user input
		m.state = completedState
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			m.quitting = true
			return m, tea.Quit
		}

		switch m.state {
		case menuState:
			switch msg.String() {
			case "up", "k":
				if m.menuCursor > 0 {
					m.menuCursor--
				}
			case "down", "j":
				if m.menuCursor < len(m.menuItems)-1 {
					m.menuCursor++
				}
			case "enter":
				switch m.menuCursor {
				case 0: // Start Timer
					m.state = inputState
					m.err = ""
					return m, textinput.Blink
				case 1: // Quit
					m.quitting = true
					return m, tea.Quit
				}
			}
		}

		switch m.state {
		case inputState:
			switch msg.String() {
			case "esc":
				m.state = menuState
				m.menuCursor = 0
				m.err = ""
				return m, nil
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
			case msg.String() == "esc":
				m.state = menuState
				m.menuCursor = 0
				return m, nil
			}
		case completedState:
			// Any key press returns to menu
			switch msg.String() {
			case "enter", " ", "esc":
				m.state = menuState
				m.menuCursor = 0
				return m, nil
			}
		}
	}

	return m, nil
}

func (m model) menuView() string {
	var s string

	// ASCII Art Title
	title := `
██   ██  ██████  ██    ██ ██████   ██████  ██       █████  ███████ ███████ 
██   ██ ██    ██ ██    ██ ██   ██ ██       ██      ██   ██ ██      ██      
███████ ██    ██ ██    ██ ██████  ██   ███ ██      ███████ ███████ ███████ 
██   ██ ██    ██ ██    ██ ██   ██ ██    ██ ██      ██   ██      ██      ██ 
██   ██  ██████   ██████  ██   ██  ██████  ███████ ██   ██ ███████ ███████ 
`

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#2E8B8B")).
		Bold(true)

	s += titleStyle.Render(title) + "\n\n"

	// Menu items with boxes
	for i, item := range m.menuItems {
		if m.menuCursor == i {
			// Selected item with highlighted box
			selectedStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#2E8B8B")).
				Bold(true).
				Padding(0, 2).
				Margin(0, 2)
			s += selectedStyle.Render(item) + "\n\n"
		} else {
			// Unselected item
			itemStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#CCCCCC")).
				Padding(0, 2).
				Margin(0, 2)
			s += itemStyle.Render(item) + "\n\n"
		}
	}

	// Subtitle
	subtitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Italic(true).
		Margin(1, 0).
		Render("Track your time and stay focused")

	s += subtitle + "\n\n"

	// Help text
	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Faint(true).
		Render("Use ↑/↓ arrows or j/k to navigate • Enter to select • q to quit")

	s += helpText

	return s
}

func (m model) completedView() string {
	var s string

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#2E8B8B")).
		Render("⏰ Timer Complete")

	s += title + "\n\n"

	// Completion message
	completionMsg := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4ECDC4")).
		Bold(true).
		Render("🎉 All done!! 🎉")

	s += completionMsg + "\n\n"

	// Timer duration display
	durationMsg := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC")).
		Render(fmt.Sprintf("Timer ran for: %v", m.timeout))

	s += durationMsg + "\n\n"

	// Help text
	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Faint(true).
		Render("Press Enter, Space, or Esc to return to menu")

	s += helpText

	return s
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
	var content string

	switch m.state {
	case menuState:
		content = m.menuView()
	case completedState:
		content = m.completedView()
	case inputState:
		var s string

		// Title
		title := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#2E8B8B")).
			Render("⏰ Set Timer Duration")

		s += title + "\n\n"

		// Input instructions
		instructions := "Enter timer duration (e.g., 5m, 30s, 1h30m):"
		s += instructions + "\n\n"

		// Input field
		s += m.textInput.View() + "\n"

		// Error message
		if m.err != "" {
			errorMsg := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF6B6B")).
				Bold(true).
				Render("Error: " + m.err)
			s += "\n" + errorMsg + "\n"
		}

		// Help text
		helpText := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Faint(true).
			Render("Enter to start timer • Esc to go back • q to quit")
		s += "\n" + helpText

		content = s
	case timerState:
		var s string

		// Title
		title := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#2E8B8B")).
			Render("⏰ Timer Running")

		s += title + "\n\n"

		// Timer display
		timerDisplay := m.timer.View()
		if m.timer.Timedout() {
			timerDisplay = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#4ECDC4")).
				Bold(true).
				Render("🎉 All done!! 🎉")
		}
		s += timerDisplay + "\n"

		if !m.quitting {
			s += m.helpView()
			s += "\n\n" + lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666666")).
				Faint(true).
				Render("Esc to return to menu")
		}
		content = s
	}

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
	)
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
		state:      menuState,
		textInput:  ti,
		menuItems:  []string{"Start Timer", "Quit"},
		menuCursor: 0,
		width:      80,
		height:     24,
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

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("uh oh, we encountered an issue:", err)
		os.Exit(1)
	}
}
