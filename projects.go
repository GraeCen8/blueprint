package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type projectTreeMsg struct {
	projectID uint
	output    string
	err       error
}

type deleteResultMsg struct {
	projectID uint
	err       error
}

func (m *Model) refreshProjects() error {
	projects, err := m.store.GetProjects()
	if err != nil {
		return err
	}
	m.projects = projects
	if len(m.projects) == 0 {
		m.selected = 0
		m.previewOutput = ""
		m.previewErr = ""
		m.previewLoading = false
		return nil
	}
	if m.selected >= len(m.projects) {
		m.selected = len(m.projects) - 1
	}
	return nil
}

func (m *Model) selectedProject() (Project, bool) {
	if len(m.projects) == 0 || m.selected < 0 || m.selected >= len(m.projects) {
		return Project{}, false
	}
	return m.projects[m.selected], true
}

func (m *Model) loadPreviewForSelected() tea.Cmd {
	project, ok := m.selectedProject()
	if !ok {
		m.previewOutput = ""
		m.previewErr = ""
		m.previewLoading = false
		return nil
	}
	m.previewErr = ""
	m.previewOutput = ""
	m.previewLoading = true
	m.previewProject = project.ID
	return loadProjectTreeCmd(project)
}

func loadProjectTreeCmd(project Project) tea.Cmd {
	return func() tea.Msg {
		if project.Path == "" {
			return projectTreeMsg{projectID: project.ID, err: errors.New("project path is required")}
		}
		cmd := exec.Command(
			"eza",
			"--tree",
			"--level=2",
			"--long",
			"--icons",
			"--git",
			"--color=always",
			project.Path,
		)
		cmd.Env = append(os.Environ(), "TERM=xterm-256color", "CLICOLOR=1")
		output, err := cmd.CombinedOutput()
		return projectTreeMsg{
			projectID: project.ID,
			output:    string(output),
			err:       err,
		}
	}
}

func (m *Model) addProject(project Project) error {
	if m.store == nil {
		return errors.New("store not initialized")
	}
	if project.Path == "" {
		return errors.New("project path is required")
	}
	absPath, err := filepath.Abs(project.Path)
	if err == nil {
		project.Path = absPath
	}
	if project.Name == "" {
		project.Name = filepath.Base(project.Path)
	}
	project.Language = strings.ToLower(strings.TrimSpace(project.Language))
	project.Template = strings.TrimSpace(project.Template)
	project.GitUser = strings.TrimSpace(project.GitUser)
	project.GitRepo = strings.TrimSpace(project.GitRepo)

	saved, err := m.store.AddProject(project)
	if err != nil {
		return err
	}
	if err := m.refreshProjects(); err != nil {
		return err
	}
	for i, p := range m.projects {
		if p.ID == saved.ID || p.Path == saved.Path {
			m.selected = i
			break
		}
	}
	return nil
}

func (m *Model) updateDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		m.confirmDelete = false
		target := m.deleteCandidate
		m.deleteCandidate = Project{}
		m.status = "deleting project..."
		return m, deleteProjectCmd(m.store, target)
	case "n", "esc":
		m.confirmDelete = false
		m.deleteCandidate = Project{}
		return m, nil
	}
	return m, nil
}

func deleteProjectCmd(store *Store, project Project) tea.Cmd {
	return func() tea.Msg {
		if store == nil {
			return deleteResultMsg{projectID: project.ID, err: errors.New("store not initialized")}
		}
		if project.Path == "" {
			return deleteResultMsg{projectID: project.ID, err: errors.New("project path is required")}
		}
		if err := removeProjectDir(project.Path); err != nil {
			return deleteResultMsg{projectID: project.ID, err: err}
		}
		if err := store.DeleteProject(project.ID); err != nil {
			return deleteResultMsg{projectID: project.ID, err: err}
		}
		return deleteResultMsg{projectID: project.ID, err: nil}
	}
}

func removeProjectDir(path string) error {
	clean := filepath.Clean(path)
	if clean == "" || clean == "." {
		return errors.New("refusing to delete empty project path")
	}
	if filepath.Dir(clean) == clean {
		return fmt.Errorf("refusing to delete root directory: %s", clean)
	}
	info, err := os.Stat(clean)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("project path is not a directory: %s", clean)
	}
	return os.RemoveAll(clean)
}
