package tui

import (
	"strings"

	"claude-watcher/internal/pkg/tui/theme"

	"github.com/charmbracelet/lipgloss"
)

// NavItem represents a navigation item
type NavItem struct {
	Key    string
	Label  string
	Active bool
}

// NavBar renders a navigation bar
type NavBar struct {
	Items  []NavItem
	styles *theme.Styles
}

// NewNavBar creates a new navigation bar
func NewNavBar(items []NavItem) *NavBar {
	return &NavBar{
		Items:  items,
		styles: theme.Default(),
	}
}

// View renders the navigation bar as NASA-style tabs
func (n NavBar) View() string {
	var items []string

	for _, item := range n.Items {
		var rendered string
		if item.Active {
			// Active: white on NASA red
			rendered = n.styles.Active.Render(item.Label)
		} else {
			// Inactive: clean key hint
			key := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#757575")).
				Render("[" + item.Key + "]")
			label := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9E9E9E")).
				Render(item.Label)
			rendered = key + " " + label
		}
		items = append(items, rendered)
	}

	// Clean separator
	sep := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#424242")).
		Render("  │  ")

	return strings.Join(items, sep)
}
