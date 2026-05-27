package chat

import (
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
}

var (
	placeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(bgColor))

	containerStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Background(lipgloss.Color(bgColor))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BD93F9")).
			Background(lipgloss.Color(bgColor)).
			Italic(true).
			MarginTop(1)

	spinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF79C6")).
			Background(lipgloss.Color(bgColor))

	userMsgStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.Color("#8BE9FD")).
			PaddingLeft(1)

	aiMsgStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.Color("#FF79C6")).
			PaddingLeft(1)
)
