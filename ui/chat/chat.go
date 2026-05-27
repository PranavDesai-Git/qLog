package chat

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	projectName string
	err         error
}

type startGenerationMsg struct{}

type RequestPreviewMessage struct{}

type Model struct {
	projectName     string
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
	isSaving        bool
	saveStatus      string
	spinner         spinner.Model
	renderer        *glamour.TermRenderer
}

func New(client *bchat.OllamaClient, projectName string) *Model {
	vp := viewport.New(60, 15)

	r, err := glamour.NewTermRenderer(glamour.WithStyles(transparentStyle))
	if err != nil {
		panic(err)
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	if projectName == "" {
		projectName = "untitled_project"
	}

	m := &Model{
		projectName:     projectName,
		viewport:        vp,
		messages:        []string{},
		renderedHistory: []string{},
		aiHistory:       []api.Message{},
		client:          client,
		spinner:         s,
		renderer:        r,
	}

	m.updateViewportContent()
	return m
}

func (m *Model) Client() *bchat.OllamaClient { return m.client }
func (m *Model) ProjectName() string         { return m.projectName }

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
		initialContent.WriteString("CONTEXT / PREVIOUS MESSAGE:\n")
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

func cleanMarkdownFences(s string) string {
	s = strings.TrimSpace(s)

	// Scan for the first and last occurrences of backtick arrays
	firstIdx := strings.Index(s, "```")
	lastIdx := strings.LastIndex(s, "```")

	// If there is a distinct pair of code block boundaries anywhere in the output
	if firstIdx != -1 && lastIdx != -1 && firstIdx != lastIdx {
		// Extract only what's inside the bounding code block container
		inner := s[firstIdx:lastIdx]

		// Find the first newline to drop the code block language tag (e.g. ```markdown or ```md)
		firstNewLine := strings.Index(inner, "\n")
		if firstNewLine != -1 {
			s = inner[firstNewLine+1:]
		} else {
			s = strings.TrimPrefix(inner, "```")
			s = strings.TrimPrefix(s, "markdown")
			s = strings.TrimPrefix(s, "md")
		}
	} else if strings.HasPrefix(s, "```") {
		// Fallback clean-up if we only detected a prefix boundary
		firstNewLine := strings.Index(s, "\n")
		if firstNewLine != -1 {
			s = s[firstNewLine+1:]
		}
		if strings.HasSuffix(s, "```") {
			s = s[:len(s)-3]
		}
	}

	return strings.TrimSpace(s)
}

func saveProjectLogCmd(client *bchat.OllamaClient, projectName string, currentHistory []api.Message) tea.Cmd {
	return func() tea.Msg {
		if len(currentHistory) == 0 {
			return streamDoneMsg{err: fmt.Errorf("no conversation history to summarize yet")}
		}

		resolvedProjectName := projectName

		// Generate context-driven hyphenated names if standard templates are matched
		if projectName == "New Project" || projectName == "untitled_project" || projectName == "" {
			namePrompt := []api.Message{
				{
					Role:    "user",
					Content: "You are an automated file naming assistant. Based on our conversation above, generate a very short, slugified descriptive project slug (e.g., 'rest-api-client' or 'sqlite-migrator'). Output ONLY the lowercase hyphenated name. No extensions, no punctuation, no markdown, no fluff.",
				},
			}

			fullContext := append(currentHistory, namePrompt...)
			var nameSb strings.Builder
			err := client.ChatStream(context.Background(), fullContext, func(chunk string) error {
				nameSb.WriteString(chunk)
				return nil
			})
			if err == nil && nameSb.Len() > 0 {
				cleaned := strings.TrimSpace(nameSb.String())
				cleaned = strings.Trim(cleaned, "`'\"")
				cleaned = strings.ReplaceAll(cleaned, " ", "-")
				cleaned = strings.ToLower(cleaned)

				var validRunes []rune
				for _, r := range cleaned {
					if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
						validRunes = append(validRunes, r)
					}
				}
				if len(validRunes) > 0 {
					resolvedProjectName = string(validRunes)
				}
			}
		}

		summaryContext := make([]api.Message, len(currentHistory))
		copy(summaryContext, currentHistory)

		summaryContext = append(summaryContext, api.Message{
			Role:    "user",
			Content: "Analyze our entire dialogue above. Generate a structured, highly professional development log detailing major design decisions, structural changes, and code blocks discussed. You must write this exclusively in clean Markdown (.md) layout. Do not enclose your overall response in wrapping code fences or write back conversational fluff—return only raw markdown syntax elements.",
		})

		var sb strings.Builder
		err := client.ChatStream(context.Background(), summaryContext, func(chunk string) error {
			sb.WriteString(chunk)
			return nil
		})
		if err != nil {
			return streamDoneMsg{err: err}
		}

		// Clean up markdown fences and external chat-bot conversational noise
		cleanedLog := cleanMarkdownFences(sb.String())

		home, err := os.UserHomeDir()
		if err != nil {
			return streamDoneMsg{err: err}
		}

		dirPath := filepath.Join(home, ".local", "qlog", "projects")
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return streamDoneMsg{err: err}
		}

		filePath := filepath.Join(dirPath, fmt.Sprintf("%s.md", resolvedProjectName))
		if err := os.WriteFile(filePath, []byte(cleanedLog), 0644); err != nil {
			return streamDoneMsg{err: err}
		}

		return streamDoneMsg{projectName: resolvedProjectName, err: nil}
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
		m.rebuildCache()
		m.updateViewportContent()

	case spinner.TickMsg:
		var spinCmd tea.Cmd
		m.spinner, spinCmd = m.spinner.Update(msg)
		cmds = append(cmds, spinCmd)

		if m.isAiLoading || m.isSaving {
			m.updateViewportContent()
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+s":
			if m.isGenerating || m.isAiLoading || m.isSaving {
				return m, tea.Batch(cmds...)
			}
			m.isSaving = true
			m.saveStatus = ""
			m.updateViewportContent()
			return m, tea.Batch(saveProjectLogCmd(m.client, m.projectName, m.aiHistory), m.spinner.Tick)

		case "ctrl+o":
			if m.isGenerating || m.isAiLoading || m.isSaving {
				return m, tea.Batch(cmds...)
			}
			return m, func() tea.Msg { return RequestPreviewMessage{} }

		case "i":
			if m.isGenerating || m.isSaving {
				return m, tea.Batch(cmds...)
			}
			var prevMessage string
			if len(m.messages) != 0 {
				prevMessage = m.messages[len(m.messages)-1]
			}
			cmds = append(cmds, openEditorCmd(prevMessage))
			return m, tea.Batch(cmds...)
		}

	case streamDoneMsg:
		m.isSaving = false
		if msg.err != nil {
			m.saveStatus = fmt.Sprintf("❌ Error auto-logging: %v", msg.err)
			m.updateViewportContent()
			cmds = append(cmds, tea.Tick(time.Second*4, func(t time.Time) tea.Msg {
				return "clear_save_status"
			}))
			return m, tea.Batch(cmds...)
		}

		if msg.projectName != "" {
			m.projectName = msg.projectName
		}

		return m, func() tea.Msg { return RequestPreviewMessage{} }

	case string:
		if msg == "clear_save_status" {
			m.saveStatus = ""
			m.updateViewportContent()
		}

	case editorMsg:
		if msg.err != nil {
			m.appendAndCacheError(fmt.Sprintf("Error running editor: %v", msg.err))
			return m, tea.Batch(cmds...)
		}
		input := strings.TrimSpace(msg.content)
		if input == "" {
			return m, tea.Batch(cmds...)
		}

		finalPrompt := bchat.BuildPrompt(input)
		userMsg := "**You:**\n" + finalPrompt

		m.messages = append(m.messages, userMsg)
		renderedUser, err := m.renderer.Render(userMsg)
		if err != nil {
			m.renderedHistory = append(m.renderedHistory, userMsg+"\n")
		} else {
			m.renderedHistory = append(m.renderedHistory, renderedUser)
		}

		m.aiHistory = append(m.aiHistory, api.Message{
			Role:    "user",
			Content: finalPrompt,
		})

		m.isGenerating = true
		m.updateViewportContent()
		m.viewport.GotoBottom()

		cmds = append(cmds, tea.Tick(time.Millisecond*16, func(t time.Time) tea.Msg {
			return startGenerationMsg{}
		}))
		return m, tea.Batch(cmds...)

	case startGenerationMsg:
		m.isAiLoading = true
		m.streamChan = make(chan string)
		return m, tea.Batch(
			startStreamCmd(m.client, m.aiHistory, m.streamChan),
			waitForChunk(m.streamChan),
			m.spinner.Tick,
		)

	case chunkMsg:
		if m.isAiLoading {
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

	case tea.Msg:
		if _, ok := msg.(streamDoneMsg); !ok {
			break
		}
		m.isGenerating = false
		m.isAiLoading = false
		m.updateViewportContent()
		m.viewport.GotoBottom()
	}

	return m, tea.Batch(cmds...)
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

	if len(m.messages) == 0 && !m.isGenerating && !m.isAiLoading {
		welcomeMarkdown := "💡 **Assistant Core Ready**\n\n" +
			"• press `i` to chat\n" +
			"• `ctrl+s` to have AI compile/write markdown logs\n" +
			"• `ctrl+o` to show the output log file window\n" +
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

	if m.isSaving {
		spinnerFrame := m.spinner.View()
		fullContent.WriteString(infoStyle.Render(spinnerFrame + " Writing files..."))
	} else if m.saveStatus != "" {
		fullContent.WriteString(infoStyle.Render(m.saveStatus))
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

	infoMsg := fmt.Sprintf("Press [i] open %s | [ctrl+s] AI Save Log | [ctrl+o] Open File Window", editor)

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

func logEvent(category, message string, extra ...string) {}
