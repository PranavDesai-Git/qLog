package main

import (
	"fmt"
	"os"

	"github.com/PranavDesai-Git/qLog/ui"
	"github.com/charmbracelet/bubbletea"
)

func main() {
	m := ui.New("gemma3:4B")
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Printf("err:%s", err)
		os.Exit(1)
	}
}
