package main

import (
	"context"
	"fmt"
	"os"

	"github.com/PranavDesai-Git/qLog/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ollama/ollama/api"
)

func main() {
	modelName := "gemma3:1b"

	go func() {
		client, err := api.ClientFromEnvironment()
		if err == nil {
			req := &api.GenerateRequest{
				Model: modelName,
			}
			_ = client.Generate(context.Background(), req, func(resp api.GenerateResponse) error {
				return nil
			})
		}
	}()

	m := ui.New(modelName)
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(1)
	}
}
