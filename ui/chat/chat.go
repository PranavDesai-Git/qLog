package chat

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	bchat "github.com/PranavDesai-Git/qLog/src/chat"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
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

type startGenerationMsg struct{}

type Model struct {
	width           int
	height          int
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

	r, err := glamour.NewTermRenderer(glamour.WithStyles(transparentStyle))
	if err != nil {
		panic(err)
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	m := &Model{
		viewport:        vp,
		messages:        []string{},
		renderedHistory: []string{},
		aiHistory:       []api.Message{},
		client:          client,
		spinner:         s,
		renderer:        r,
	}

	// Initialize the viewport with the onboarding message right away
	m.updateViewportContent()
	return m
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
		defer os.Remove(tempFile.Name())
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
	logEvent("INIT", "chat model initialized")
	return m.spinner.Tick
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		logEvent("RESIZE", fmt.Sprintf("w=%d h=%d", m.width, m.height))

		m.viewport.Width = m.width - 4
		m.viewport.Height = m.height - 5
		m.viewport.Style = lipgloss.NewStyle().Width(m.viewport.Width).Background(lipgloss.Color(bgColor))

		m.renderer, _ = glamour.NewTermRenderer(
			glamour.WithStyles(transparentStyle),
			glamour.WithWordWrap(m.viewport.Width-2),
		)
		if !m.isGenerating {
			m.rebuildCache()
		}
		m.updateViewportContent()

	case spinner.TickMsg:
		var spinCmd tea.Cmd
		m.spinner, spinCmd = m.spinner.Update(msg)
		cmds = append(cmds, spinCmd)

		if m.isAiLoading {
			m.updateViewportContent()
		}

	case tea.KeyMsg:
		logEvent("KEY", fmt.Sprintf("key=%q isGenerating=%v", msg.String(), m.isGenerating))
		switch msg.String() {
		case "i":
			if m.isGenerating {
				logEvent("KEY:i BLOCKED", "isGenerating=true, ignoring")
				return m, tea.Batch(cmds...)
			}
			var prevMessage string
			if len(m.messages) != 0 {
				prevMessage = m.messages[len(m.messages)-1]
			}
			logEvent("KEY:i", "opening editor", fmt.Sprintf("prevMessage_len=%d", len(prevMessage)))
			cmds = append(cmds, openEditorCmd(prevMessage))
			return m, tea.Batch(cmds...)
		}

	case editorMsg:
		if msg.err != nil {
			logEvent("EDITOR ERR", msg.err.Error())
			m.appendAndCacheError(fmt.Sprintf("Error running editor: %v", msg.err))
			return m, tea.Batch(cmds...)
		}

		input := strings.TrimSpace(msg.content)
		logEvent("EDITOR DONE", fmt.Sprintf("raw_len=%d trimmed_len=%d", len(msg.content), len(input)))

		if input == "" {
			logEvent("EDITOR EMPTY", "user submitted nothing, aborting")
			return m, tea.Batch(cmds...)
		}

		logEvent("EDITOR CONTENT PREVIEW", fmt.Sprintf("%q", func() string {
			if len(input) > 80 {
				return input[:80] + "..."
			}
			return input
		}()))

		finalPrompt := bchat.BuildPrompt(input)
		userMsg := "**You:**\n" + finalPrompt

		m.messages = append(m.messages, userMsg)
		renderedUser, err := m.renderer.Render(userMsg)
		if err != nil {
			logEvent("RENDER ERR", fmt.Sprintf("failed to render user msg: %v", err))
			m.renderedHistory = append(m.renderedHistory, userMsg+"\n")
		} else {
			m.renderedHistory = append(m.renderedHistory, renderedUser)
		}

		logEvent("USER MSG CACHED", fmt.Sprintf("messages=%d renderedHistory=%d", len(m.messages), len(m.renderedHistory)))

		m.aiHistory = append(m.aiHistory, api.Message{
			Role:    "user",
			Content: finalPrompt,
		})

		m.isGenerating = true

		m.updateViewportContent()
		m.viewport.GotoBottom()
		logEvent("DISPATCHING startGenerationMsg")
		cmds = append(cmds, tea.Tick(time.Millisecond*16, func(t time.Time) tea.Msg {
			return startGenerationMsg{}
		}))
		return m, tea.Batch(cmds...)

	case startGenerationMsg:
		logEvent("GENERATION START", "spinner on, stream starting")
		m.isAiLoading = true
		m.streamChan = make(chan string)
		return m, tea.Batch(
			startStreamCmd(m.client, m.aiHistory, m.streamChan),
			waitForChunk(m.streamChan),
			m.spinner.Tick,
		)

	case chunkMsg:
		logEvent("CHUNK", fmt.Sprintf("len=%d isAiLoading=%v", len(msg), m.isAiLoading))
		if m.isAiLoading {
			logEvent("CHUNK FIRST", "spinner off, starting AI message")
			m.isAiLoading = false
			m.messages = append(m.messages, "**AI:**\n"+string(msg))
		} else {
			last := len(m.messages) - 1
			m.messages[last] += string(msg)
		}

		m.updateViewportContent()
		if m.viewport.ScrollPercent() > 0.9 {
			m.viewport.GotoBottom()
		}

		cmds = append(cmds, waitForChunk(m.streamChan))

	case streamDoneMsg:
		logEvent("STREAM DONE", fmt.Sprintf("err=%v messages=%d renderedHistory=%d", msg.err, len(m.messages), len(m.renderedHistory)))
		m.isGenerating = false

		if msg.err != nil {
			m.isAiLoading = false
			logEvent("STREAM ERR", msg.err.Error())
			m.appendAndCacheError(fmt.Sprintf("\nAPI ERR: %v\n", msg.err))
		} else {
			if m.isAiLoading {
				logEvent("STREAM DONE EMPTY", "no chunks received, rendering empty response")
				m.isAiLoading = false
				m.messages = append(m.messages, "**AI:**\n*(Empty response)*")
			}

			last := len(m.messages) - 1
			if last >= 0 && len(m.messages) > len(m.renderedHistory) {
				finalText := m.messages[last]
				logEvent("CACHING AI MSG", fmt.Sprintf("msg_len=%d", len(finalText)))

				renderedAI, err := m.renderer.Render(finalText)
				if err != nil {
					logEvent("RENDER ERR", fmt.Sprintf("failed to render AI msg: %v", err))
					m.renderedHistory = append(m.renderedHistory, finalText+"\n")
				} else {
					m.renderedHistory = append(m.renderedHistory, renderedAI)
				}

				cleanHistoryText := strings.TrimPrefix(finalText, "**AI:**\n")
				m.aiHistory = append(m.aiHistory, api.Message{
					Role:    "assistant",
					Content: cleanHistoryText,
				})
				logEvent("AI MSG CACHED", fmt.Sprintf("messages=%d renderedHistory=%d aiHistory=%d", len(m.messages), len(m.renderedHistory), len(m.aiHistory)))
			}
		}
		m.updateViewportContent()
		m.viewport.GotoBottom()
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) rebuildCache() {
	logEvent("REBUILD CACHE", fmt.Sprintf("rebuilding %d messages", len(m.messages)))
	m.renderedHistory = make([]string, 0, len(m.messages))
	for i, msg := range m.messages {
		rendered, err := m.renderer.Render(msg)
		if err != nil {
			logEvent("REBUILD CACHE ERR", fmt.Sprintf("msg[%d]: %v", i, err))
			m.renderedHistory = append(m.renderedHistory, msg+"\n")
		} else {
			m.renderedHistory = append(m.renderedHistory, rendered)
		}
	}
	logEvent("REBUILD CACHE DONE", fmt.Sprintf("renderedHistory=%d", len(m.renderedHistory)))
}

func (m *Model) appendAndCacheError(errMsg string) {
	logEvent("ERROR APPENDED", fmt.Sprintf("%q", errMsg))
	m.messages = append(m.messages, errMsg)
	m.renderedHistory = append(m.renderedHistory, errMsg+"\n")
	m.updateViewportContent()
}

func (m *Model) updateViewportContent() {
	logEvent("RENDER",
		fmt.Sprintf("msgs=%d rendered=%d isGenerating=%v isAiLoading=%v scrollPct=%.2f",
			len(m.messages), len(m.renderedHistory), m.isGenerating, m.isAiLoading, m.viewport.ScrollPercent()))

	var fullContent strings.Builder

	// Display Onboarding/Help Text box when there are no active chat messages
	if len(m.messages) == 0 && !m.isGenerating && !m.isAiLoading {
		welcomeMarkdown := "💡 **Assistant Core Ready**\n\n" +
			"• press `i` to chat\n" +
			"• `esc` to go back\n" +
			"• `shift+esc` to go back to menu\n" +
			"• `ctrl+c` to exit"

		rendered, err := m.renderer.Render(welcomeMarkdown)
		if err == nil {
			fullContent.WriteString(aiMsgStyle.Render(strings.TrimRight(rendered, "\n")) + "\n\n")
		} else {
			fullContent.WriteString(aiMsgStyle.Render(welcomeMarkdown) + "\n\n")
		}
	}

	for i, rendered := range m.renderedHistory {
		rawMsg := m.messages[i]
		isUser := strings.HasPrefix(rawMsg, "**You:**")

		style := aiMsgStyle
		if isUser {
			style = userMsgStyle
		}

		cleanRendered := strings.TrimRight(rendered, "\n")
		fullContent.WriteString(style.Render(cleanRendered) + "\n\n")
	}

	if m.isGenerating && !m.isAiLoading && len(m.messages) > len(m.renderedHistory) {
		currentMsg := m.messages[len(m.messages)-1]
		rendered, err := m.renderer.Render(currentMsg)

		isUser := strings.HasPrefix(currentMsg, "**You:**")
		style := aiMsgStyle
		if isUser {
			style = userMsgStyle
		}

		cleanRendered := currentMsg
		if err == nil {
			cleanRendered = strings.TrimRight(rendered, "\n")
		}

		fullContent.WriteString(style.Render(cleanRendered) + "\n\n")
	}

	if m.isAiLoading {
		spinnerFrame := m.spinner.View()
		fullContent.WriteString(aiMsgStyle.Render(spinnerFrame + " Thinking...") + "\n\n")
	}

	m.viewport.SetContent(fullContent.String())
}

func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano"
	}

	infoMsg := fmt.Sprintf("Press [i] to open %s and write a message. [esc] to go to prev screen", editor)

	content := containerStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			m.viewport.View(),
			infoStyle.Render(infoMsg),
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
