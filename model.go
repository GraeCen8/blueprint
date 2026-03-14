package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	exit      bool
	width     int
	height    int
	setup     *Setup
	templates []TemplateInfo
	status    string
	errMsg    string
	wizard    WizardState
}

func (m *Model) Init() tea.Cmd {
	if m.setup == nil {
		m.setup = NewSetup()
	}

	templates, err := m.setup.Templates.Templates()
	if err != nil {
		m.errMsg = fmt.Sprintf("load templates: %v", err)
	} else {
		m.templates = templates
	}

	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if m.wizard.active {
			return m.updateWizard(msg)
		}
		m.Key(msg.String())
	case setupResultMsg:
		m.wizard.active = false
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			m.status = ""
		} else {
			m.errMsg = ""
			m.status = "project created"
		}
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
	case "n":
		m.startWizard()
		return
	}
}

func (m *Model) startWizard() {
	m.wizard = WizardState{
		active: true,
		step:   stepProjectDir,
	}
	m.wizard.input = newTextInput("project directory", "")
}

func newTextInput(placeholder, value string) textinput.Model {
	input := textinput.New()
	input.Placeholder = placeholder
	input.SetValue(value)
	input.Focus()
	input.CharLimit = 120
	input.Width = 40
	return input
}
