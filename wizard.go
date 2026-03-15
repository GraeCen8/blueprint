package main

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type WizardStep int

const (
	stepProjectDir WizardStep = iota
	stepGitConfirm
	stepGitUser
	stepGitRepo
	stepLanguage
	stepTemplate
	stepConfirm
)

type WizardState struct {
	active     bool
	step       WizardStep
	input      textinput.Model
	options    []string
	selected   int
	projectDir string
	gitEnabled bool
	gitUser    string
	gitRepo    string
	language   string
	template   string
	errMsg     string
}

type setupResultMsg struct {
	err     error
	project Project
}

func (m *Model) updateWizard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		m.wizard.active = false
		m.wizard.errMsg = ""
		return m, nil
	}

	switch m.wizard.step {
	case stepProjectDir:
		return m.updateWizardInput(msg, func(value string) (WizardStep, bool) {
			if strings.TrimSpace(value) == "" {
				m.wizard.errMsg = "project directory is required"
				return stepProjectDir, false
			}
			m.wizard.projectDir = strings.TrimSpace(value)
			m.wizard.errMsg = ""
			return stepGitConfirm, true
		})
	case stepGitConfirm:
		return m.updateWizardOptions(msg, func(choice string) (WizardStep, bool, tea.Cmd) {
			m.wizard.gitEnabled = choice == "yes"
			if m.wizard.gitEnabled {
				return stepGitUser, true, nil
			}
			return stepLanguage, true, nil
		})
	case stepGitUser:
		return m.updateWizardInput(msg, func(value string) (WizardStep, bool) {
			if strings.TrimSpace(value) == "" {
				m.wizard.errMsg = "git username is required"
				return stepGitUser, false
			}
			m.wizard.gitUser = strings.TrimSpace(value)
			m.wizard.errMsg = ""
			return stepGitRepo, true
		})
	case stepGitRepo:
		return m.updateWizardInput(msg, func(value string) (WizardStep, bool) {
			if strings.TrimSpace(value) == "" {
				m.wizard.errMsg = "git repo name is required"
				return stepGitRepo, false
			}
			m.wizard.gitRepo = strings.TrimSpace(value)
			m.wizard.errMsg = ""
			return stepLanguage, true
		})
	case stepLanguage:
		return m.updateWizardOptions(msg, func(choice string) (WizardStep, bool, tea.Cmd) {
			m.wizard.language = choice
			return stepTemplate, true, nil
		})
	case stepTemplate:
		return m.updateWizardOptions(msg, func(choice string) (WizardStep, bool, tea.Cmd) {
			m.wizard.template = choice
			return stepConfirm, true, nil
		})
	case stepConfirm:
		return m.updateWizardOptions(msg, func(choice string) (WizardStep, bool, tea.Cmd) {
			if choice == "cancel" {
				m.wizard.active = false
				return stepConfirm, false, nil
			}
			opts := SetupOptions{
				ProjectDir: m.wizard.projectDir,
				Git:        m.wizard.gitEnabled,
				GitUser:    m.wizard.gitUser,
				GitRepo:    m.wizard.gitRepo,
				Template:   m.wizard.template,
				Language:   ParseLanguage(m.wizard.language),
			}
			project := Project{
				Name:     filepath.Base(m.wizard.projectDir),
				Path:     m.wizard.projectDir,
				Language: strings.ToLower(m.wizard.language),
				Template: m.wizard.template,
				Git:      m.wizard.gitEnabled,
				GitUser:  m.wizard.gitUser,
				GitRepo:  m.wizard.gitRepo,
			}
			m.wizard.active = false
			m.status = "creating project..."
			return stepConfirm, true, func() tea.Msg {
				err := m.setup.Setup(opts)
				return setupResultMsg{err: err, project: project}
			}
		})
	}

	return m, nil
}

func (m *Model) updateWizardInput(msg tea.KeyMsg, onEnter func(string) (WizardStep, bool)) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.wizard.input, cmd = m.wizard.input.Update(msg)

	if msg.String() == "enter" {
		next, advance := onEnter(m.wizard.input.Value())
		if advance {
			m.wizard.step = next
			m.prepareWizardStep(next)
		}
	}

	return m, cmd
}

func (m *Model) updateWizardOptions(msg tea.KeyMsg, onEnter func(string) (WizardStep, bool, tea.Cmd)) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.wizard.selected > 0 {
			m.wizard.selected--
		}
	case "down", "j":
		if m.wizard.selected < len(m.wizard.options)-1 {
			m.wizard.selected++
		}
	case "enter":
		if len(m.wizard.options) == 0 {
			return m, nil
		}
		next, advance, cmd := onEnter(m.wizard.options[m.wizard.selected])
		if advance {
			m.wizard.step = next
			m.prepareWizardStep(next)
		}
		return m, cmd
	}

	return m, nil
}

func (m *Model) prepareWizardStep(step WizardStep) {
	m.wizard.errMsg = ""
	switch step {
	case stepProjectDir:
		m.wizard.input = newTextInput("project directory", m.wizard.projectDir)
	case stepGitUser:
		m.wizard.input = newTextInput("git username", m.wizard.gitUser)
	case stepGitRepo:
		m.wizard.input = newTextInput("git repo name", m.wizard.gitRepo)
	case stepGitConfirm:
		m.ensureOptions([]string{"yes", "no"})
	case stepLanguage:
		m.ensureOptions(m.languageOptions())
	case stepTemplate:
		m.ensureOptions(m.templateOptions(m.wizard.language))
	case stepConfirm:
		m.ensureOptions([]string{"create", "cancel"})
	}
}

func (m *Model) ensureOptions(options []string) {
	m.wizard.options = options
	m.wizard.selected = 0
}

func (m *Model) languageOptions() []string {
	unique := map[string]bool{}
	var languages []string
	for _, t := range m.templates {
		if t.Language == "" {
			continue
		}
		key := strings.ToLower(t.Language)
		if !unique[key] {
			unique[key] = true
			languages = append(languages, key)
		}
	}
	if len(languages) == 0 {
		return []string{"other"}
	}
	sort.Strings(languages)
	return languages
}

func (m *Model) templateOptions(language string) []string {
	language = strings.ToLower(language)
	var names []string
	for _, t := range m.templates {
		if language == "" || t.Language == "" || strings.EqualFold(t.Language, language) {
			names = append(names, t.Name)
		}
	}
	if len(names) == 0 {
		for _, t := range m.templates {
			names = append(names, t.Name)
		}
	}
	sort.Strings(names)
	return names
}
