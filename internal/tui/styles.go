package tui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39"))

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	tabStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(0, 1)

	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Bold(true)

	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 1)

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("82"))

	warnStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("214"))

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196"))

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39"))

	sparklineChars = []rune{'\u2581', '\u2582', '\u2583', '\u2584', '\u2585', '\u2586', '\u2587', '\u2588'}
)
