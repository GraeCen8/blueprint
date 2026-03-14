package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	exit   bool
	width  int
	height int
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		m.Key(msg.String())
	}

	if m.exit {
		return m, tea.Quit
	}

	return m, nil
}

func (m *Model) Key(msg string) {
	switch msg {
	case "q", "esc":
		m.exit = true // exit the program
		return        // return from the function
	}
}
