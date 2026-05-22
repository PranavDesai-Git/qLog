package menu

import (
	"github.com/charmbracelet/bubbletea"
)

type SelectChatMessage struct{}

type Model struct {
}

func New() Model {
	return Model{}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	return "menu"
}
