package ui

import (
	"github.com/PranavDesai-Git/qLog/ui/chat"
	"github.com/PranavDesai-Git/qLog/ui/manage"
	"github.com/PranavDesai-Git/qLog/ui/menu"
	"github.com/PranavDesai-Git/qLog/ui/settings"
	tea "github.com/charmbracelet/bubbletea"
)

type sessionState int

const (
	menuView sessionState = iota
	chatView
	manageView
	settingsView
)

type Manager struct {
	state      sessionState
	prevStates []sessionState
	menu       menu.Model
	chat       chat.Model
	manage     manage.Model
	settings   settings.Model
}

func New() *Manager {
	return &Manager{
		state:      menuView,
		prevStates: make([]sessionState, 0, 8),
		menu:       menu.New(),
		chat:       chat.New(),
		manage:     manage.New(),
		settings:   settings.New(),
	}
}

func (m *Manager) Init() tea.Cmd {
	return nil
}

func (m *Manager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch tmsg := msg.(type) {
	case menu.SelectChatMessage:
		m.prevStates = append(m.prevStates, m.state)
		m.state = chatView
		return m, nil
	case menu.SelectManageMessage:
		m.prevStates = append(m.prevStates, m.state)
		m.state = manageView
		return m, nil
	case menu.SelectSettingsMessage:
		m.prevStates = append(m.prevStates, m.state)
		m.state = settingsView
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
	case chatView:
		newChat, newCmd := m.chat.Update(msg)
		m.chat = newChat.(chat.Model)
		cmd = newCmd
	case manageView:
		newManage, newCmd := m.manage.Update(msg)
		m.manage = newManage.(manage.Model)
		cmd = newCmd
	case settingsView:
		newSettings, newCmd := m.settings.Update(msg)
		m.settings = newSettings.(settings.Model)
		cmd = newCmd
	}
	return m, cmd
}

func (m *Manager) View() string {
	switch m.state {
	case menuView:
		return m.menu.View()
	case chatView:
		return m.chat.View()
	case manageView:
		return m.manage.View()
	case settingsView:
		return m.settings.View()
	}
	return ""
}
