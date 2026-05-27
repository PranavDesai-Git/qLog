package ui

import (
	bchat "github.com/PranavDesai-Git/qLog/src/chat"
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
	previewView
	manageView
	settingsView
)

type Manager struct {
	state      sessionState
	prevStates []sessionState
	menu       menu.Model
	chat       *chat.Model
	preview    *chat.PreviewModel
	manage     manage.Model
	settings   settings.Model
}

func New(clientName string) *Manager {
	ollamaClient, err := bchat.NewOllamaClient(clientName)
	if err != nil {
		panic(err)
	}
	return &Manager{
		state:      menuView,
		prevStates: make([]sessionState, 0, 8),
		menu:       menu.New(),
		chat:       chat.New(ollamaClient, "New Project"),
		preview:    nil, // Dynamically instantiated when triggered
		manage:     manage.New(),
		settings:   settings.New(),
	}
}

func (m *Manager) Init() tea.Cmd {
	return nil
}

func (m *Manager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var activeCmd tea.Cmd

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

	case chat.RequestPreviewMessage:
		// Pivot directly into document preview view
		m.prevStates = append(m.prevStates, m.state)
		m.state = previewView
		m.preview = chat.NewPreview(m.chat.Client(), m.chat.ProjectName())
		return m, m.preview.Load()

	case tea.KeyMsg:
		switch tmsg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			if len(m.prevStates) > 0 {
				lastElement := len(m.prevStates) - 1
				m.state = m.prevStates[lastElement]
				m.prevStates = m.prevStates[:lastElement]
			}
			return m, nil
		case "shift+esc":
			m.prevStates = m.prevStates[:0]
			m.state = menuView
			return m, nil
		}
	}

	switch m.state {
	case menuView:
		newMenu, newCmd := m.menu.Update(msg)
		m.menu = newMenu.(menu.Model)
		activeCmd = newCmd
	case chatView:
		newChat, newCmd := m.chat.Update(msg)
		m.chat = newChat.(*chat.Model)
		activeCmd = newCmd
	case previewView:
		if m.preview != nil {
			newPreview, newCmd := m.preview.Update(msg)
			m.preview = newPreview.(*chat.PreviewModel)
			activeCmd = newCmd
		}
	case manageView:
		newManage, newCmd := m.manage.Update(msg)
		m.manage = newManage.(manage.Model)
		activeCmd = newCmd
	case settingsView:
		newSettings, newCmd := m.settings.Update(msg)
		m.settings = newSettings.(settings.Model)
		activeCmd = newCmd
	}

	return m, activeCmd
}

func (m *Manager) View() string {
	switch m.state {
	case menuView:
		return m.menu.View()
	case chatView:
		return m.chat.View()
	case previewView:
		if m.preview != nil {
			return m.preview.View()
		}
		return ""
	case manageView:
		return m.manage.View()
	case settingsView:
		return m.settings.View()
	}
	return ""
}
