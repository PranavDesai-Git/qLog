package menu

import (
	"github.com/charmbracelet/bubbletea"
	"strings"
)

type SelectChatMessage struct{}
type SelectManageMessage struct{}
type SelectSettingsMessage struct{}

type Model struct {
	choices []string
	selectd int
}

func New() Model {
	return Model{
		choices: []string{
			"[n]ew Project",
			"[m]anage projects",
			"[s]ettings",
		},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var newCmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "n":
			newCmd = func() tea.Msg { return SelectChatMessage{} }
		case "m":
			newCmd = func() tea.Msg { return SelectManageMessage{} }
		case "s":
			newCmd = func() tea.Msg { return SelectSettingsMessage{} }
		}
	}
	return m, newCmd
}

func (m Model) View() string {
	s := ""
	s += strings.Join(m.choices, "\n")
	return s
}
