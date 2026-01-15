package tui

import (
	"claude-watcher/internal/analytics"
	analyticstui "claude-watcher/internal/analytics/inbound/tui"
	"claude-watcher/internal/pkg/tui/theme"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// App is the main dashboard TUI application
type App struct {
	overview *analyticstui.Overview
	styles   *theme.Styles
	width    int
	height   int
}

// NewApp creates a new dashboard application
func NewApp(analyticsService *analytics.Service) *App {
	return &App{
		overview: analyticstui.NewOverview(analyticsService),
		styles:   theme.Default(),
	}
}

// Init implements tea.Model
func (a App) Init() tea.Cmd {
	return a.overview.Init()
}

// Update implements tea.Model
func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	}

	// Forward to overview screen
	var cmd tea.Cmd
	a.overview, cmd = a.overview.Update(msg)
	return a, cmd
}

// View implements tea.Model
func (a App) View() string {
	header := a.renderHeader()
	content := a.overview.View()

	return lipgloss.JoinVertical(lipgloss.Left, header, "", content)
}

func (a App) renderHeader() string {
	title := a.styles.Title.Copy().
		Foreground(theme.BrightPurple).
		Bold(true).
		Render("Claude Watcher")

	return title
}
