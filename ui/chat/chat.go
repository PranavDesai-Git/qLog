package chat

import (
	"context"
	"fmt"
	bchat "github.com/PranavDesai-Git/qLog/src/chat"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/wordwrap"
	"github.com/ollama/ollama/api"
	"os"
	"os/exec"
	"strings"
)

type editorMsg struct {
	content string
	err     error
}

type chunkMsg string
type streamDoneMsg struct {
	err error
}

type Model struct {
	viewport     viewport.Model
	messages     []string
	aiHistory    []api.Message
	client       *bchat.OllamaClient
	streamChan   chan string
	isGenerating bool
}

func New(client *bchat.OllamaClient) Model {
	vp := viewport.New(60, 15)
	return Model{
		viewport:  vp,
		messages:  []string{},
		aiHistory: []api.Message{},
		client:    client,
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

	const devider = "--------CONTEXT ONLY ANYTHING FROM HERE WILL BE DISCARDED--------\n"

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

func startStreamCmd(client *bchat.OllamaClient, history []api.Message, sub chan string) tea.Cmd {
	return func() tea.Msg {
		err := client.ChatStream(context.Background(), history, func(chunk string) error {
			sub <- chunk
			return nil
		})
		close(sub)
		return streamDoneMsg{err: err}
	}
}

func waitForChunk(sub chan string) tea.Cmd {
	return func() tea.Msg {
		chunk, ok := <-sub
		if !ok {
			return nil
		}
		return chunkMsg(chunk)
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 5
		m.viewport.Style = m.viewport.Style.Width(msg.Width)
		m.updateViewportContent()
	case tea.KeyMsg:
		switch msg.String() {
		case "i":
			if m.isGenerating {
				return m, nil
			}
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

		finalPrompt := bchat.BuildPrompt(input)

		m.messages = append(m.messages, "\nYou:"+finalPrompt)
		m.aiHistory = append(m.aiHistory, api.Message{
			Role:    "user",
			Content: finalPrompt,
		})
		m.messages = append(m.messages, "AI: ")
		m.updateViewportContent()
		m.viewport.GotoBottom()

		m.isGenerating = true
		m.streamChan = make(chan string)
		return m, tea.Batch(
			startStreamCmd(m.client, m.aiHistory, m.streamChan),
			waitForChunk(m.streamChan),
		)
	case chunkMsg:
		last := len(m.messages) - 1
		m.messages[last] += string(msg)

		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()

		return m, waitForChunk(m.streamChan)
	case streamDoneMsg:
		m.isGenerating = false
		if msg.err != nil {
			m.messages = append(m.messages, fmt.Sprintf("API ERR:%v", msg.err))
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
		} else {
			last := len(m.messages) - 1
			finalText := strings.TrimPrefix(m.messages[last], "AI: ")

			m.aiHistory = append(m.aiHistory, api.Message{
				Role:    "assistant",
				Content: finalText,
			})
		}
		return m, nil
	}
	return m, vpCmd
}

func (m Model) updateViewportContent() {
	fullText := strings.Join(m.messages, "\n")
	wrapWidth := m.viewport.Width - 2
	if wrapWidth < 1 {
		wrapWidth = 1
	}
	wrappedText := wordwrap.String(fullText, wrapWidth)
	m.viewport.SetContent(wrappedText)
}

func (m Model) View() string {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano"
	}
	infoMsg := fmt.Sprintf("press [i] to open %s and write a message. [esc] to go to prev screen", editor)
	return fmt.Sprintf("%s\n%s", infoMsg, m.viewport.View())
}
