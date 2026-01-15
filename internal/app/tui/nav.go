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

// View renders the navigation bar
func (n NavBar) View() string {
	var items []string

	for _, item := range n.Items {
		keyStyle := n.styles.HelpKey
		var labelStyle lipgloss.Style

		if item.Active {
			labelStyle = n.styles.Active
		} else {
			labelStyle = n.styles.Muted
		}

		items = append(items, keyStyle.Render(item.Key)+":"+labelStyle.Render(item.Label))
	}

	return strings.Join(items, "  ")
}
