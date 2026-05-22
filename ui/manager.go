package ui

import (
	"github.com/PranavDesai-Git/qLog/ui/chat"
	"github.com/PranavDesai-Git/qLog/ui/menu"
	"github.com/charmbracelet/bubbletea"
)

type sessionState int

// STATE ENUM
const (
	menuView sessionState = iota
	chatView
)

type Manager struct {
	state sessionState
	menu  menu.Model
	chat  chat.Model
}

func New() Manager {
	return Manager{
		state: menuView,
		menu:  menu.New(),
		chat:  chat.New(),
	}
}

func (m Manager) Init() tea.Cmd {
	return nil
}

func (m Manager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch tmsg := msg.(type) {
	case menu.SelectChatMessage:
		m.state = chatView
		return m, nil
	case tea.KeyMsg:
		switch tmsg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	switch m.state {
	case menuView:
		newMenu, newCmd := m.menu.Update(msg)
		m.menu = newMenu.(menu.Model)
		cmd = newCmd
	}
	return m, cmd
}

func (m Manager) View() string {
	switch m.state {
	case menuView:
		return m.menu.View()
	case chatView:
		return m.chat.View()
	}
	return ""
}
