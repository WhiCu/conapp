package main

import (
	"conapp/internal"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(internal.InitialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("could not start program:", err)
	}
}
