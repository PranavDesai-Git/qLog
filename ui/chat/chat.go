package chat

import (
	"fmt"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"os/exec"
	"strings"
)

type editorMsg struct {
	content string
	err     error
}

type Model struct {
	viewport viewport.Model
	messages []string
}

func New() Model {
	vp := viewport.New(60, 15)
	return Model{
		viewport: vp,
		messages: []string{},
	}
}

func openEditorCmd(prevMessage string) tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano"
	}
	tempFile, err := os.CreateTemp("", "qLog-chat-*")
	if err != nil {
		return func() tea.Msg { return editorMsg{err: err} }
	}

	const devider = "\n\n--------CONTEXT ONLY ANYTHING FROM HERE WILL BE DISCARDED--------\n"

	var initialContent strings.Builder

	initialContent.WriteString("\n\n")

	if prevMessage != "" {
		initialContent.WriteString(devider)
		initialContent.WriteString("PREVIOUS MESSAGE:\n")
		initialContent.WriteString(prevMessage + "\n")
	}

	if _, err := tempFile.WriteString(initialContent.String()); err != nil {
		return func() tea.Msg { return editorMsg{err: err} }
	}

	tempFile.Close()

	c := exec.Command(editor, tempFile.Name())
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return editorMsg{err: err}
		}

		bytes, readErr := os.ReadFile(tempFile.Name())
		defer os.Remove(tempFile.Name())
		if readErr != nil {
			return editorMsg{err: readErr}
		}

		fileContent := string(bytes)
		parts := strings.Split(fileContent, devider)
		userInput := strings.TrimSpace(parts[0])

		return editorMsg{content: userInput, err: nil}
	})
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "i":
			var prevMessage string
			if len(m.messages) != 0 {
				prevMessage = m.messages[len(m.messages)-1]
			}
			return m, openEditorCmd(prevMessage)
		}
	case editorMsg:
		if msg.err != nil {
			m.messages = append(m.messages, fmt.Sprintf("Error running editor: %v", msg.err))
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			return m, nil
		}

		input := strings.TrimSpace(msg.content)

		if input == "" {
			return m, nil
		}
		m.messages = append(m.messages, input)
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()
		return m, nil
	}
	return m, vpCmd
}

func (m Model) View() string {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano"
	}
	infoMsg := fmt.Sprintf("press [i] to open %s and write a message. [esc] to go to prev screen", editor)
	return fmt.Sprintf("%s\n%s",infoMsg, m.viewport.View())
}
