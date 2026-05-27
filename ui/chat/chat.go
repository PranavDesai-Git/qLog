package chat

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	bchat "github.com/PranavDesai-Git/qLog/src/chat"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/ollama/ollama/api"
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
	viewport        viewport.Model
	messages        []string
	renderedHistory []string
	aiHistory       []api.Message
	client          *bchat.OllamaClient
	streamChan      chan string
	isGenerating    bool
	isAiLoading     bool
	spinner         spinner.Model
	renderer        *glamour.TermRenderer
}

func New(client *bchat.OllamaClient) *Model {
	vp := viewport.New(60, 15)
	r, err := glamour.NewTermRenderer(glamour.WithAutoStyle())
	if err != nil {
		panic(err)
	}
	s := spinner.New()
	s.Spinner = spinner.Dot
	return &Model{
		viewport:        vp,
		messages:        []string{},
		renderedHistory: []string{},
		aiHistory:       []api.Message{},
		client:          client,
		spinner:         s,
		renderer:        r,
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

	const divider = "--------CONTEXT ONLY ANYTHING FROM HERE WILL BE DISCARDED--------\n"

	var initialContent strings.Builder
	initialContent.WriteString("\n\n")

	if prevMessage != "" {
		initialContent.WriteString(divider)
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
		defer os.Remove(tempFile.Name()) // Clean up after reading
		if readErr != nil {
			return editorMsg{err: readErr}
		}

		fileContent := string(bytes)
		parts := strings.Split(fileContent, divider)
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

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)

	var spinCmd tea.Cmd
	if m.isAiLoading {
		m.spinner, spinCmd = m.spinner.Update(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 5
		m.viewport.Style = m.viewport.Style.Width(msg.Width)

		m.renderer, _ = glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(m.viewport.Width-2),
		)
		m.rebuildCache()
		m.updateViewportContent()

	case spinner.TickMsg:
		if m.isAiLoading {
			m.updateViewportContent()
			return m, spinCmd
		}

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
			m.appendAndCacheError(fmt.Sprintf("Error running editor: %v", msg.err))
			return m, nil
		}

		input := strings.TrimSpace(msg.content)
		if input == "" {
			return m, nil
		}

		finalPrompt := bchat.BuildPrompt(input)
		userMsg := "**You:**\n" + finalPrompt
		
		m.messages = append(m.messages, userMsg)
		
		renderedUser, _ := m.renderer.Render(userMsg)
		m.renderedHistory = append(m.renderedHistory, renderedUser)

		m.aiHistory = append(m.aiHistory, api.Message{
			Role:    "user",
			Content: finalPrompt,
		})
		
		m.isGenerating = true
		m.isAiLoading = true
		
		m.messages = append(m.messages, "**AI:**\n")

		m.updateViewportContent()
		m.viewport.GotoBottom()

		m.streamChan = make(chan string)
		return m, tea.Batch(
			startStreamCmd(m.client, m.aiHistory, m.streamChan),
			waitForChunk(m.streamChan),
			m.spinner.Tick,
		)

	case chunkMsg:
		if m.isAiLoading {
			m.isAiLoading = false
		}
		last := len(m.messages) - 1
		m.messages[last] += string(msg)
		m.updateViewportContent()
		
		if m.viewport.ScrollPercent() > 0.9 {
			m.viewport.GotoBottom()
		}
		return m, waitForChunk(m.streamChan)

	case streamDoneMsg:
		m.isGenerating = false
		m.isAiLoading = false
		
		if msg.err != nil {
			m.appendAndCacheError(fmt.Sprintf("API ERR: %v", msg.err))
		} else {
			last := len(m.messages) - 1
			finalText := m.messages[last]
			
			renderedAI, err := m.renderer.Render(finalText)
			if err != nil {
				m.renderedHistory = append(m.renderedHistory, finalText+"\n")
			} else {
				m.renderedHistory = append(m.renderedHistory, renderedAI)
			}

			cleanHistoryText := strings.TrimPrefix(finalText, "**AI:**\n")
			m.aiHistory = append(m.aiHistory, api.Message{
				Role:    "assistant",
				Content: cleanHistoryText,
			})
		}
		m.updateViewportContent()
		return m, nil
	}
	return m, tea.Batch(vpCmd, spinCmd)
}

func (m *Model) rebuildCache() {
	m.renderedHistory = make([]string, 0, len(m.messages))
	for _, msg := range m.messages {
		rendered, err := m.renderer.Render(msg)
		if err != nil {
			m.renderedHistory = append(m.renderedHistory, msg+"\n")
		} else {
			m.renderedHistory = append(m.renderedHistory, rendered)
		}
	}
}

func (m *Model) appendAndCacheError(errMsg string) {
	m.messages = append(m.messages, errMsg)
	m.renderedHistory = append(m.renderedHistory, errMsg+"\n")
	m.updateViewportContent()
}

func (m *Model) updateViewportContent() {
	var fullContent strings.Builder

	for _, rendered := range m.renderedHistory {
		fullContent.WriteString(rendered)
	}

	if m.isGenerating && len(m.messages) > len(m.renderedHistory) {
		currentMsg := m.messages[len(m.messages)-1]
		rendered, err := m.renderer.Render(currentMsg)
		if err != nil {
			fullContent.WriteString(currentMsg + "\n")
		} else {
			fullContent.WriteString(rendered)
		}
	}
	if m.isAiLoading {
		spinnerFrame := m.spinner.View()
		fullContent.WriteString(spinnerFrame + " Thinking...")
	}

	m.viewport.SetContent(fullContent.String())
}

// Changed to pointer receiver for consistency
func (m *Model) View() string {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano"
	}
	infoMsg := fmt.Sprintf("Press [i] to open %s and write a message. [esc] to go to prev screen", editor)
	return fmt.Sprintf("%s\n%s", infoMsg, m.viewport.View())
}
