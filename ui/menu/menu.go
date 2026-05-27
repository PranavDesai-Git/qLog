package menu

import (
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type SelectChatMessage struct{}
type SelectManageMessage struct{}
type SelectSettingsMessage struct{}

type Model struct {
	choices       []string
	width         int
	height        int
	agendaContent string
}

func readAgenda() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "_could not find home directory_"
	}
	path := filepath.Join(home, "agenda.md")
	bytes, err := os.ReadFile(path)
	if err != nil {
		return "_no agenda.md found — run `qlog agenda` to generate one_"
	}
	return string(bytes)
}

func New() Model {
	return Model{
		choices: []string{
			"[n]ew Project",
			"[m]anage projects",
			"[s]ettings",
		},
		agendaContent: readAgenda(),
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "n":
			return m, func() tea.Msg { return SelectChatMessage{} }
		case "m":
			return m, func() tea.Msg { return SelectManageMessage{} }
		case "s":
			return m, func() tea.Msg { return SelectSettingsMessage{} }
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	styledLogo := renderLogo()
	logoWidth := lipgloss.Width(styledLogo)

	subtitle := subtitleStyle.Width(logoWidth).MarginBottom(1).Render("plan your day not like a pleb")

	var renderedItems []string
	for _, choice := range m.choices {
		renderedItems = append(renderedItems, itemStyle.Width(logoWidth).Render("› " + choice))
	}
	itemsBlock := lipgloss.JoinVertical(lipgloss.Left, renderedItems...)

	agendaInnerWidth := logoWidth - 4
	if agendaInnerWidth < 20 {
		agendaInnerWidth = 20
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(transparentStyle),
		glamour.WithWordWrap(agendaInnerWidth),
	)

	var renderedAgenda string
	if err != nil {
		renderedAgenda = m.agendaContent
	} else {
		renderedAgenda, err = renderer.Render(m.agendaContent)
		if err != nil {
			renderedAgenda = m.agendaContent
		}
	}

	agendaBox := agendaBoxStyle.Width(logoWidth).MarginBottom(1).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			agendaTitleStyle.Width(agendaInnerWidth).Render("📋 Today's Agenda"),
			strings.TrimSpace(renderedAgenda),
		),
	)

	content := containerStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			styledLogo,
			subtitle,
			agendaBox,
			itemsBlock,
		),
	)

	return placeStyle.
		Width(m.width).
		Height(m.height).
		Render(
			lipgloss.Place(m.width, m.height,
				lipgloss.Center, lipgloss.Center,
				content,
				lipgloss.WithWhitespaceBackground(lipgloss.Color(bgColor)),
			),
		)
}
