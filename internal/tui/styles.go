package tui

import "github.com/charmbracelet/lipgloss"

// Styles
var (
	// Base colors
	primaryColor   = lipgloss.Color("#00D4AA") // Bright teal
	secondaryColor = lipgloss.Color("#FF6B9D") // Pink
	accentColor    = lipgloss.Color("#FFA500") // Orange
	successColor   = lipgloss.Color("#00E676") // Green
	warningColor   = lipgloss.Color("#FFD700") // Gold
	errorColor     = lipgloss.Color("#FF5252") // Red
	mutedColor     = lipgloss.Color("#6C757D") // Gray
	fgColor        = lipgloss.Color("#C0CAF5") // Light fg

	appStyle = lipgloss.NewStyle().Padding(1, 2)

	// Title and header styles
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(primaryColor).
			Bold(true).
			Padding(0, 1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)

	breadcrumbStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	// Box and border styles
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2).
			MarginTop(1)

	// Mailbox Styles
	mailboxStyle = lipgloss.NewStyle().
			Foreground(fgColor).
			PaddingLeft(2)

	selectedMailboxStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000000")).
				Background(primaryColor).
				Bold(true).
				PaddingLeft(2).
				PaddingRight(2)

	// List row styles (emails, events, contacts)
	emailItemStyle = lipgloss.NewStyle().
			PaddingLeft(1)

	selectedEmailItemStyle = lipgloss.NewStyle().
				PaddingLeft(1).
				Foreground(lipgloss.Color("#000000")).
				Background(primaryColor).
				Bold(true)

	unreadStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	readStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	emailFromStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)

	emailSubjectStyle = lipgloss.NewStyle().
				Foreground(fgColor)

	emailDateStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	// Contact Styles
	contactNameStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true)

	contactEmailStyle = lipgloss.NewStyle().
				Foreground(accentColor)

	contactFieldLabelStyle = lipgloss.NewStyle().
				Foreground(secondaryColor).
				Bold(true)

	contactFieldValueStyle = lipgloss.NewStyle().
				Foreground(fgColor)

	// Calendar Styles
	eventTitleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	eventTimeStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	eventDateHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000000")).
				Background(secondaryColor).
				Bold(true).
				Padding(0, 1).
				MarginTop(1)

	todayBadgeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(warningColor).
			Bold(true).
			Padding(0, 1)

	// Status and badge styles
	statusStyle = lipgloss.NewStyle().
			Foreground(fgColor).
			Background(lipgloss.Color("#2A2B3C")).
			Padding(0, 1)

	badgeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(accentColor).
			Bold(true).
			Padding(0, 1).
			MarginLeft(1)

	warningStyle = lipgloss.NewStyle().
			Foreground(warningColor)

	errorBadgeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(errorColor).
			Bold(true).
			Padding(0, 1)

	// Help text style
	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	keyStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	// Divider style
)
