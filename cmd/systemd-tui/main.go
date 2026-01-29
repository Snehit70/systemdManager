package main

import (
	"fmt"
	"os"
	"systemd-tui/internal/service"
	"systemd-tui/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	client := service.NewSystemdClient()

	m := ui.NewMainModel(client)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
