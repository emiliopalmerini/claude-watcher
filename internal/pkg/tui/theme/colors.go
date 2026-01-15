package theme

import "github.com/charmbracelet/lipgloss"

// Anthropic-inspired retrofuturistic color palette
// Bold, systematic design language
var (
	// Primary - High contrast foundation
	White = lipgloss.Color("#FFFFFF")
	Black = lipgloss.Color("#000000")

	// Anthropic Orange - Brand accent color
	Orange = lipgloss.Color("#E07A3D")

	// Gray scale - systematic hierarchy
	Gray100 = lipgloss.Color("#F5F5F5")
	Gray200 = lipgloss.Color("#E0E0E0")
	Gray300 = lipgloss.Color("#BDBDBD")
	Gray400 = lipgloss.Color("#9E9E9E")
	Gray500 = lipgloss.Color("#757575")
	Gray600 = lipgloss.Color("#616161")
	Gray700 = lipgloss.Color("#424242")
	Gray800 = lipgloss.Color("#212121")
	Gray900 = lipgloss.Color("#121212")

	// Semantic colors
	Success = lipgloss.Color("#4CAF50")
	Warning = Orange
	Error   = Orange
	Info    = lipgloss.Color("#FFFFFF")

	// Aliases
	LightGray    = Gray400
	DimGray      = Gray500
	DarkGray     = Gray700
	Accent       = Orange
	AccentBg     = Gray900
	BrightPurple = Orange
	NASARed      = Orange // Legacy alias
)
