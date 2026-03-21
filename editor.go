package main

import (
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type editorResultMsg struct {
	err error
}

func (m *Model) openEditor() tea.Cmd {
	project, ok := m.selectedProject()
	if !ok {
		m.status = ""
		m.errMsg = "no project selected"
		return nil
	}
	path := strings.TrimSpace(project.Path)
	if path == "" {
		m.status = ""
		m.errMsg = "project path is required"
		return nil
	}

	cmd := exec.Command("nvim", path)
	cmd.Dir = path
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	m.status = "opening editor..."
	m.errMsg = ""
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return editorResultMsg{err: err}
	})
}
