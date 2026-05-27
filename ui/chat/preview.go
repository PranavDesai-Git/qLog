package chat

import (
	"context"
	"fmt"
	"os"
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

type fileLoadedMsg struct {
	content string
	err     error
}

type fileModifyDoneMsg struct {
	content string
	err     error
}

type PreviewModel struct {
	projectName string
	width       int
	height      int
	viewport    viewport.Model
	fileContent string
	isSaving    bool
	saveStatus  string
	spinner     spinner.Model
	renderer    *glamour.TermRenderer
	client      *bchat.OllamaClient
}

func NewPreview(client *bchat.OllamaClient, projectName string) *PreviewModel {
	vp := viewport.New(60, 15)
	r, _ := glamour.NewTermRenderer(glamour.WithStyles(transparentStyle))
	s := spinner.New(spinner.WithSpinner(spinner.Dot), spinner.WithStyle(spinnerStyle))

	return &PreviewModel{
		projectName: projectName,
		viewport:    vp,
		spinner:     s,
		renderer:    r,
		client:      client,
	}
}

func (m *PreviewModel) Resize(width, height int) {
	m.width = width
	m.height = height
	m.viewport.Width = m.width - 4
	m.viewport.Height = m.height - 5
	m.viewport.Style = lipgloss.NewStyle().Width(m.viewport.Width).Background(lipgloss.Color(bgColor))
	m.renderer, _ = glamour.NewTermRenderer(
		glamour.WithStyles(transparentStyle),
		glamour.WithWordWrap(m.viewport.Width-2),
	)
	m.updateViewportContent()
}

func (m *PreviewModel) Load() tea.Cmd {
	return readProjectLogCmd(m.projectName)
}

func (m *PreviewModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *PreviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Resize(msg.Width, msg.Height)

	case spinner.TickMsg:
		var spinCmd tea.Cmd
		m.spinner, spinCmd = m.spinner.Update(msg)
		cmds = append(cmds, spinCmd)
		if m.isSaving {
			m.updateViewportContent()
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "i":
			if m.isSaving {
				return m, tea.Batch(cmds...)
			}
			return m, tea.Batch(openEditorCmd(m.fileContent))
		}

	case fileLoadedMsg:
		m.isSaving = false
		m.saveStatus = ""
		if msg.err != nil {
			m.saveStatus = fmt.Sprintf("❌ Error: %v", msg.err)
		} else {
			m.fileContent = msg.content
		}
		m.updateViewportContent()
		m.viewport.GotoTop()

	case fileModifyDoneMsg:
		m.isSaving = false
		if msg.err != nil {
			m.saveStatus = fmt.Sprintf("❌ Update Error: %v", msg.err)
		} else {
			m.fileContent = msg.content
			m.saveStatus = "💾 AI updated file and saved modifications!"
		}
		m.updateViewportContent()
		m.viewport.GotoTop()

		cmds = append(cmds, tea.Tick(time.Second*4, func(t time.Time) tea.Msg {
			return "clear_preview_save_status"
		}))

	case editorMsg:
		if msg.err != nil {
			m.saveStatus = fmt.Sprintf("❌ Editor Error: %v", msg.err)
			m.updateViewportContent()
			return m, tea.Batch(cmds...)
		}
		input := strings.TrimSpace(msg.content)
		if input == "" {
			return m, tea.Batch(cmds...)
		}
		m.isSaving = true
		m.updateViewportContent()
		return m, tea.Batch(modifyProjectLogCmd(m.client, m.projectName, m.fileContent, input), m.spinner.Tick)

	case string:
		if msg == "clear_preview_save_status" {
			m.saveStatus = ""
			m.updateViewportContent()
		}
	}

	return m, tea.Batch(cmds...)
}

func readProjectLogCmd(projectName string) tea.Cmd {
	return func() tea.Msg {
		home, err := os.UserHomeDir()
		if err != nil {
			return fileLoadedMsg{err: err}
		}
		filePath := filepath.Join(home, ".local", "qlog", "projects", fmt.Sprintf("%s.md", projectName))
		bytes, err := os.ReadFile(filePath)
		if err != nil {
			return fileLoadedMsg{err: fmt.Errorf("could not read log file: %w (did you save it using ctrl+s first?)", err)}
		}
		return fileLoadedMsg{content: string(bytes), err: nil}
	}
}

func modifyProjectLogCmd(client *bchat.OllamaClient, projectName, currentFileContent, instructions string) tea.Cmd {
	return func() tea.Msg {
		prompt := fmt.Sprintf("You are an automated technical documentation editing system. Here is the current Markdown project log file:\n\n```md\n%s\n```\n\nThe user wants to apply the following alterations/fixes:\n\"%s\"\n\nProcess the file and output the entirely updated document containing those additions or revisions. Maintain existing file parts unless targeted. Output only the updated raw markdown content—no conversations, no markdown wrapper strings.", currentFileContent, instructions)

		messages := []api.Message{{Role: "user", Content: prompt}}

		var sb strings.Builder
		err := client.ChatStream(context.Background(), messages, func(chunk string) error {
			sb.WriteString(chunk)
			return nil
		})
		if err != nil {
			return fileModifyDoneMsg{err: err}
		}

		// Use the new code-boundary scanner to drop conversational noise from the model
		updatedContent := cleanMarkdownFences(sb.String())

		home, err := os.UserHomeDir()
		if err != nil {
			return fileModifyDoneMsg{err: err}
		}
		filePath := filepath.Join(home, ".local", "qlog", "projects", fmt.Sprintf("%s.md", projectName))
		if err := os.WriteFile(filePath, []byte(updatedContent), 0644); err != nil {
			return fileModifyDoneMsg{err: err}
		}

		return fileModifyDoneMsg{content: updatedContent, err: nil}
	}
}

func (m *PreviewModel) updateViewportContent() {
	var fullContent strings.Builder
	header := fmt.Sprintf("📖 **LOG FILE PREVIEW: %s.md**\n---\n", m.projectName)
	renderedHeader, _ := m.renderer.Render(header)
	fullContent.WriteString(userMsgStyle.Render(strings.TrimRight(renderedHeader, "\n")) + "\n\n")

	renderedFile, err := m.renderer.Render(m.fileContent)
	if err == nil {
		fullContent.WriteString(renderedFile)
	} else {
		fullContent.WriteString(m.fileContent)
	}

	if m.isSaving {
		spinnerFrame := m.spinner.View()
		fullContent.WriteString("\n\n" + infoStyle.Render(spinnerFrame+" AI is processing modifications and rewriting file..."))
	} else if m.saveStatus != "" {
		fullContent.WriteString("\n\n" + infoStyle.Render(m.saveStatus))
	}

	m.viewport.SetContent(fullContent.String())
}

func (m *PreviewModel) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	infoMsg := "Press [i] to type alterations | [esc] Return to Chat"

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
