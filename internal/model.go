package internal

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fogleman/ease"
)

const (
	progressBarWidth  = 100
	progressFullChar  = "█"
	progressEmptyChar = "░"
	dotChar           = " • "
)

var (
	ticksStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("79"))
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	subtleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	dotStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("236")).Render(dotChar)
	progressEmpty = subtleStyle.Render(progressEmptyChar)
	progressFull  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00")).Render(progressFullChar)
)

type (
	tickMsg  struct{}
	frameMsg struct{}
)

type model struct {
	UserPassword textinput.Model
	Entered      bool
	Ticks        int
	Frames       int
	Progress     float64
	Loaded       bool
	Quitting     bool
}

func InitialModel() model {
	i := textinput.New()
	i.Prompt = ">>"
	i.Placeholder = "passwprd"
	i.Cursor.Style = cursorStyle
	i.Width = 32
	i.EchoMode = textinput.EchoPassword
	i.EchoCharacter = '•'
	i.Focus()

	return model{
		UserPassword: i,
		Entered:      false,
		Ticks:        100,
		Frames:       0,
		Progress:     0,
		Loaded:       false,
		Quitting:     false,
	} //model{0, false, 10, 0, 0, false, false}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tick(), textinput.Blink)
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

func frame() tea.Cmd {
	return tea.Tick(time.Second/60, func(t time.Time) tea.Msg {
		return frameMsg{}
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		if k == "esc" || k == "ctrl+c" {
			m.Quitting = true
			return m, tea.Quit
		}
	}

	if !m.Entered {
		return updateEnter(msg, m)
	}

	return updateEntered(msg, m)
}

func updateEnter(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.UserPassword, cmd = m.UserPassword.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.Entered = true
			return m, frame()
		}

	case tickMsg:
		if m.Ticks == 0 {
			m.Quitting = true
			return m, tea.Quit
		}
		m.Ticks--
		return m, tick()
	}

	return m, cmd
}

func updateEntered(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	if m.UserPassword.Value() != "password" {
		m.Quitting = true
		return m, tea.Quit
	}
	switch msg.(type) {
	case frameMsg:
		if !m.Loaded {
			m.Frames++
			m.Progress = ease.InOutCirc(float64(m.Frames) / float64(100))
			if m.Progress >= 1 {
				m.Progress = 1
				m.Loaded = true
				m.Ticks = 3
				return m, tick()
			}
			return m, frame()
		}

	case tickMsg:
		if m.Loaded {
			if m.Ticks == 0 {
				m.Quitting = true
				return m, tea.Quit
			}
			m.Ticks--
			return m, tick()
		}
	}

	return m, nil
}
func (m model) View() string {

	if m.Quitting {
		return "\n  See you later!\n\n"
	}

	if !m.Entered {
		return enterView(m)
	}
	return enteredView(m)
}
func enterView(m model) string {

	tpl := "Введите пороль для авторизации:\n\n"
	tpl += "%s\n\n"
	tpl += "Program quits in %s seconds\n\n"
	tpl += subtleStyle.Render("enter: entered") + dotStyle +
		subtleStyle.Render("ctrl+c, esc: quit")

	return fmt.Sprintf(tpl, m.UserPassword.View(), ticksStyle.Render(strconv.Itoa(m.Ticks)))
}
func enteredView(m model) string {

	label := "Password verification..."
	if m.Loaded {
		label = fmt.Sprintf("Successful. Exiting in %s seconds...", ticksStyle.Render(strconv.Itoa(m.Ticks)))
	}

	return label + "\n" + progressbar(m.Progress) + "%"
}

func progressbar(percent float64) string {
	w := float64(progressBarWidth)

	fullSize := int(math.Round(w * percent))
	fullCells := strings.Repeat(progressFull, fullSize)

	emptySize := int(w) - fullSize
	emptyCells := strings.Repeat(progressEmpty, emptySize)

	return fmt.Sprintf("%s%s %3.0f", fullCells, emptyCells, math.Round(percent*100))
}
