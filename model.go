package main

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
	"sync"
	"time"
)

var (
	topLineFmt = lipgloss.NewStyle()
	topLeftFmt = topLineFmt.Copy().
			Align(lipgloss.Left)
	topRightFmt = topLineFmt.Copy().
			Align(lipgloss.Right)
	infoStyle = lipgloss.NewStyle().
			Align(lipgloss.Right).
			Border(lipgloss.RoundedBorder())
)

type model struct {
	sync.RWMutex
	viewport  viewport.Model     // Viewport bubble for displaying cmd output
	ready     bool               // Has the model been configured?
	w, h      int                // Width & height of the terminal
	cancel    context.CancelFunc // Context's cancel function
	cmdOutput string             // Output from the last run of the command
	lastRun   time.Time          // Time the last run finished
	interval  time.Duration      // Frequency with which the command is run
	cmdText   string             // Text of the command
}

func newModel(cmd string, cancel context.CancelFunc, i time.Duration) *model {
	return &model{
		cancel:    cancel,
		cmdOutput: cmd,
		interval:  i,
	}
}

func (m *model) updateOutput(output string) {
	m.Lock()
	defer m.Unlock()
	m.cmdOutput = output
	m.lastRun = time.Now()
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if k := msg.String(); k == "ctrl+c" || k == "q" || k == "esc" {
			m.cancel()
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		topLeftFmt.Width(msg.Width / 2)
		topRightFmt.Width(msg.Width / 2)

		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.viewport.SetContent(m.cmdOutput)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}
	}

	// Handle keyboard and mouse events in the viewport
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *model) View() string {
	// Check if setup has completed...
	if !m.ready {
		return "Initializing..."
	}

	// Format the top line, body, and footer
	head := m.headerView()
	body := m.bodyView()
	foot := m.footerView()

	// Combine and return
	return fmt.Sprintf("%s\n%s\n%s", head, body, foot)
}

func (m *model) headerView() string {
	m.RLock()
	defer m.RUnlock()
	topLeft := fmt.Sprintf("Every %s: %q", m.interval, m.cmdText)
	topRight := m.lastRun.Format(time.ANSIC)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		topLeftFmt.Render(topLeft),
		topRightFmt.Render(topRight),
	)
}

func (m *model) bodyView() string {
	m.RLock()
	defer m.RUnlock()
	return m.viewport.View()
}

func (m *model) footerView() string {
	m.RLock()
	defer m.RUnlock()
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Top, line, info)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
