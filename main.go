package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gen2brain/beeep"
)

var (
	centeredStyle = lipgloss.NewStyle().
			Align(lipgloss.Center)

	asciiStyle = lipgloss.NewStyle().
			Align(lipgloss.Center).
			Bold(true).
			Foreground(lipgloss.Color("#FFD700")) // Gold color

	menuStyle = lipgloss.NewStyle().
			Align(lipgloss.Center).
			Padding(1, 2)

	menuItemStyle = lipgloss.NewStyle().
			Padding(0, 1)

	selectedMenuItemStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Foreground(lipgloss.Color("#00FF00")) // Green for selected

	instructionsStyle = lipgloss.NewStyle().
				Align(lipgloss.Center).
				Faint(true)

	inputStyle = lipgloss.NewStyle().
			Align(lipgloss.Center).
			Padding(2, 0)

	timerStyle = lipgloss.NewStyle().
			Align(lipgloss.Center).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("#FF0000")) // Red for errors
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
	completionState
)

type model struct {
	state           appState
	textInput       textinput.Model
	timer           timer.Model
	keymap          keymap
	help            help.Model
	quitting        bool
	timeout         time.Duration
	err             string
	notifIcon       []byte
	menuIndex       int
	width           int
	height          int
	completionIndex int // For completion screen menu
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

		// Transition to completion state instead of quitting
		m.state = completionState
		m.completionIndex = 0 // Default to "Restart Timer"
		return m, nil

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

		case completionState:
			switch msg.String() {
			case "up", "k":
				if m.completionIndex > 0 {
					m.completionIndex--
				}
			case "down", "j":
				if m.completionIndex < 1 {
					m.completionIndex++
				}
			case "enter":
				switch m.completionIndex {
				case 0: // Restart Timer
					m.timer = timer.NewWithInterval(m.timeout, time.Millisecond)
					m.state = timerState
					return m, m.timer.Init()
				case 1: // Return to Menu
					m.state = landingState
					m.menuIndex = 0
					return m, nil
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
		return m.renderLandingPage()
	case inputState:
		return m.renderInputPage()
	case timerState:
		return m.renderTimerPage()
	case completionState:
		return m.renderCompletionPage()
	}
	return ""
}

func (m model) renderLandingPage() string {
	// ASCII art centered
	ascii := asciiStyle.Render(hourglassASCII)

	// Menu centered
	menu := m.renderMenu()

	// Instructions centered
	instructions := instructionsStyle.Render("Use ↑/↓ or j/k to navigate, Enter to select, q to quit")

	// Combine with vertical centering
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		ascii,
		"\n\n",
		menu,
		"\n",
		instructions,
	)

	// Place content in center of screen
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

func (m model) renderMenu() string {
	menuItems := []string{"Start Timer", "Quit"}
	var menuLines []string

	for i, item := range menuItems {
		if i == m.menuIndex {
			menuLines = append(menuLines, selectedMenuItemStyle.Render("> "+item))
		} else {
			menuLines = append(menuLines, menuItemStyle.Render("  "+item))
		}
	}

	return menuStyle.Render(strings.Join(menuLines, "\n"))
}

func (m model) renderInputPage() string {
	var content strings.Builder

	content.WriteString("Enter timer duration (e.g., 5m, 30s, 1h30m):\n\n")
	content.WriteString(m.textInput.View())
	content.WriteString("\n")

	if m.err != "" {
		content.WriteString("\n")
		content.WriteString(errorStyle.Render("Error: " + m.err))
		content.WriteString("\n")
	}

	content.WriteString("\nPress Enter to start timer, q to quit")

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		inputStyle.Render(content.String()),
	)
}

func (m model) renderTimerPage() string {
	var content strings.Builder

	timerDisplay := m.timer.View()
	if m.timer.Timedout() {
		timerDisplay = "All done!!"
	}

	content.WriteString(timerStyle.Render("Timer: " + timerDisplay))
	content.WriteString("\n")

	if !m.quitting {
		content.WriteString(m.helpView())
	}

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content.String(),
	)
}

func (m model) renderCompletionPage() string {
	// Completion message
	completionMsg := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Bold(true).
		Foreground(lipgloss.Color("#00FF00")).
		Render("🎉 Timer Complete! 🎉")

	// Menu options
	menuItems := []string{"Restart Timer for: " + m.textInput.Value(), "Return to Menu"}
	var menuLines []string

	for i, item := range menuItems {
		if i == m.completionIndex {
			menuLines = append(menuLines, selectedMenuItemStyle.Render("> "+item))
		} else {
			menuLines = append(menuLines, menuItemStyle.Render("  "+item))
		}
	}

	menu := menuStyle.Render(strings.Join(menuLines, "\n"))

	// Instructions
	instructions := instructionsStyle.Render("Use ↑/↓ or j/k to navigate, Enter to select, q to quit")

	// Combine all elements
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		completionMsg,
		"\n\n",
		menu,
		"\n",
		instructions,
	)

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

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("uh oh, we encountered an issue:", err)
		os.Exit(1)
	}
}
