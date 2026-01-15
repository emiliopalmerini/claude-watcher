package theme

import "github.com/charmbracelet/lipgloss"

// Monochrome grayscale palette
// High contrast typography-focused hierarchy
var (
	// Primary - Maximum contrast
	White = lipgloss.Color("#FFFFFF")
	Black = lipgloss.Color("#000000")

	// Gray scale - carefully tuned for text hierarchy
	Gray50  = lipgloss.Color("#FAFAFA") // Near white
	Gray100 = lipgloss.Color("#F5F5F5") // Backgrounds
	Gray200 = lipgloss.Color("#EEEEEE") // Subtle borders
	Gray300 = lipgloss.Color("#E0E0E0") // Disabled
	Gray400 = lipgloss.Color("#BDBDBD") // Tertiary text
	Gray500 = lipgloss.Color("#9E9E9E") // Secondary text
	Gray600 = lipgloss.Color("#757575") // Body text
	Gray700 = lipgloss.Color("#616161") // Strong secondary
	Gray800 = lipgloss.Color("#424242") // Borders on dark
	Gray900 = lipgloss.Color("#212121") // Near black

	// Semantic - grayscale only
	Success = Gray600
	Warning = Gray500
	Error   = White
	Info    = Gray500

	// Aliases
	LightGray    = Gray400
	DimGray      = Gray600
	DarkGray     = Gray800
	Accent       = White
	AccentBg     = Gray900
	BrightPurple = White
	Orange       = White // Legacy
	NASARed      = White // Legacy
)
