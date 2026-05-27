package menu

import (
	"strings"

	ansi "github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/lipgloss"
)

const bgColor = "#000000"

func stringPtr(s string) *string { return &s }
func boolPtr(b bool) *bool       { return &b }
func uintPtr(u uint) *uint       { return &u }

var transparentStyle = ansi.StyleConfig{
	Document: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color:           stringPtr("#FF79C6"),
			BackgroundColor: stringPtr(bgColor),
		},
		Margin: uintPtr(0),
	},
	BlockQuote: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color:           stringPtr("#BD93F9"),
			BackgroundColor: stringPtr(bgColor),
			Italic:          boolPtr(true),
		},
	},
	Paragraph: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color:           stringPtr("#FF79C6"),
			BackgroundColor: stringPtr(bgColor),
		},
	},
	List: ansi.StyleList{
		LevelIndent: 2,
	},
	Heading: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color:           stringPtr("#BD93F9"),
			BackgroundColor: stringPtr(bgColor),
			Bold:            boolPtr(true),
		},
	},
	H1: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color:           stringPtr("#FF79C6"),
			BackgroundColor: stringPtr(bgColor),
			Bold:            boolPtr(true),
		},
	},
	H2: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color:           stringPtr("#BD93F9"),
			BackgroundColor: stringPtr(bgColor),
			Bold:            boolPtr(true),
		},
	},
	H3: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color:           stringPtr("#FF79C6"),
			BackgroundColor: stringPtr(bgColor),
			Bold:            boolPtr(true),
		},
	},
	H4: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color:           stringPtr("#BD93F9"),
			BackgroundColor: stringPtr(bgColor),
			Bold:            boolPtr(true),
		},
	},
	H5: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color:           stringPtr("#FF79C6"),
			BackgroundColor: stringPtr(bgColor),
			Bold:            boolPtr(true),
		},
	},
	H6: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color:           stringPtr("#BD93F9"),
			BackgroundColor: stringPtr(bgColor),
			Bold:            boolPtr(true),
		},
	},
	Text: ansi.StylePrimitive{
		Color:           stringPtr("#FF79C6"),
		BackgroundColor: stringPtr(bgColor),
	},
	Emph: ansi.StylePrimitive{
		Color:           stringPtr("#BD93F9"),
		BackgroundColor: stringPtr(bgColor),
		Italic:          boolPtr(true),
	},
	Strong: ansi.StylePrimitive{
		Color:           stringPtr("#FF79C6"),
		BackgroundColor: stringPtr(bgColor),
		Bold:            boolPtr(true),
	},
	Item: ansi.StylePrimitive{
		Color:           stringPtr("#FF79C6"),
		BackgroundColor: stringPtr(bgColor),
	},
	Enumeration: ansi.StylePrimitive{
		Color:           stringPtr("#FF79C6"),
		BackgroundColor: stringPtr(bgColor),
	},
	Link: ansi.StylePrimitive{
		Color:           stringPtr("#BD93F9"),
		BackgroundColor: stringPtr(bgColor),
	},
	LinkText: ansi.StylePrimitive{
		Color:           stringPtr("#FF79C6"),
		BackgroundColor: stringPtr(bgColor),
		Bold:            boolPtr(true),
	},
	Code: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color:           stringPtr("#BD93F9"),
			BackgroundColor: stringPtr(bgColor),
		},
	},
	CodeBlock: ansi.StyleCodeBlock{
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:           stringPtr("#BD93F9"),
				BackgroundColor: stringPtr(bgColor),
			},
			Margin: uintPtr(0),
		},
	},
	HorizontalRule: ansi.StylePrimitive{
		Color:           stringPtr("#BD93F9"),
		BackgroundColor: stringPtr(bgColor),
	},
	Strikethrough: ansi.StylePrimitive{
		Color:           stringPtr("#FF79C6"),
		BackgroundColor: stringPtr(bgColor),
	},
}

var (
	logoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF79C6")).
			Background(lipgloss.Color(bgColor)).
			Bold(true)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BD93F9")).
			Background(lipgloss.Color(bgColor))

	itemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF79C6")).
			Background(lipgloss.Color(bgColor))

	agendaBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#BD93F9")).
			Background(lipgloss.Color(bgColor)).
			Padding(0, 1)

	agendaTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF79C6")).
				Background(lipgloss.Color(bgColor)).
				Bold(true).
				MarginBottom(1)

	containerStyle = lipgloss.NewStyle().
			Padding(2, 4).
			Background(lipgloss.Color(bgColor))

	placeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(bgColor))
)

var logo = `
 ██████╗ ██╗      ██████╗  ██████╗ 
██╔═══██╗██║     ██╔═══██╗██╔════╝ 
██║   ██║██║     ██║   ██║██║  ███╗
██║▄▄ ██║██║     ██║   ██║██║   ██║
╚██████╔╝███████╗╚██████╔╝╚██████╔╝
 ╚══▀▀═╝ ╚══════╝ ╚═════╝  ╚═════╝ `

func renderLogo() string {
	raw := strings.Trim(logo, "\n")
	width := lipgloss.Width(raw)
	return logoStyle.Width(width).Render(raw)
}
