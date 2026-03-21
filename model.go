package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	exit             bool
	width            int
	height           int
	setup            *Setup
	store            *Store
	templates        []TemplateInfo
	projects         []Project
	selected         int
	previewOutput    string
	previewErr       string
	previewReadme    string
	previewReadmeErr string
	previewProject   uint
	previewLoading   bool
	confirmDelete    bool
	deleteCandidate  Project
	status           string
	errMsg           string
	wizard           WizardState
}

func (m *Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	if m.setup == nil {
		m.setup = NewSetup()
	}
	if m.store == nil {
		m.store = NewStore()
	}

	templates, err := m.setup.Templates.Templates()
	if err != nil {
		m.errMsg = fmt.Sprintf("load templates: %v", err)
	} else {
		m.templates = templates
	}

	if err := m.store.Init(); err != nil {
		m.errMsg = fmt.Sprintf("init store: %v", err)
	} else if err := m.refreshProjects(); err != nil {
		m.errMsg = fmt.Sprintf("load projects: %v", err)
	}

	if cmd := m.loadPreviewForSelected(); cmd != nil {
		cmds = append(cmds, cmd)
	}

	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.exit {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if m.wizard.active {
			return m.updateWizard(msg)
		}
		if m.confirmDelete {
			return m.updateDeleteConfirm(msg)
		}
		cmd := m.Key(msg.String())
		return m, cmd
	case setupResultMsg:
		m.wizard.active = false
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			m.status = ""
		} else {
			m.errMsg = ""
			m.status = "project created"
			if err := m.addProject(msg.project); err != nil {
				m.errMsg = err.Error()
				m.status = ""
			} else if cmd := m.loadPreviewForSelected(); cmd != nil {
				return m, cmd
			}
		}
	case projectPreviewMsg:
		if current, ok := m.selectedProject(); ok && current.ID == msg.projectID {
			m.previewLoading = false
			if msg.treeErr != nil {
				m.previewErr = msg.treeErr.Error()
				if msg.treeOutput != "" {
					m.previewOutput = msg.treeOutput
				} else {
					m.previewOutput = ""
				}
			} else {
				m.previewErr = ""
				m.previewOutput = msg.treeOutput
			}
			if msg.readmeErr != nil {
				m.previewReadme = ""
				m.previewReadmeErr = msg.readmeErr.Error()
			} else {
				m.previewReadme = msg.readme
				m.previewReadmeErr = ""
			}
		}
	case deleteResultMsg:
		m.status = ""
		if msg.err != nil {
			m.errMsg = msg.err.Error()
		} else {
			m.errMsg = ""
			m.status = "project deleted"
			if err := m.refreshProjects(); err != nil {
				m.errMsg = err.Error()
			} else if cmd := m.loadPreviewForSelected(); cmd != nil {
				return m, cmd
			}
		}
	case editorResultMsg:
		m.status = ""
		if msg.err != nil {
			m.errMsg = fmt.Sprintf("open editor: %v", msg.err)
		} else {
			m.errMsg = ""
		}
	}

	if m.exit {
		return m, tea.Quit
	}

	return m, nil
}

func (m *Model) Key(msg string) tea.Cmd {
	switch msg {
	case "q", "esc":
		m.exit = true // exit the program
		return nil    // return from the function
	case "n":
		m.startWizard()
		return nil
	case "up", "k":
		if m.selected > 0 {
			m.selected--
			return m.loadPreviewForSelected()
		}
		return nil
	case "down", "j":
		if m.selected < len(m.projects)-1 {
			m.selected++
			return m.loadPreviewForSelected()
		}
		return nil
	case "d":
		if project, ok := m.selectedProject(); ok {
			m.confirmDelete = true
			m.deleteCandidate = project
		}
		return nil
	case "e":
		return m.openEditor()
	}

	return nil
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
