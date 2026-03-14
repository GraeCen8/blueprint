package main

import (
	"strings"

	lipgloss "github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#22d3ee")).
		Width(m.width).
		Padding(0, 1).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#1a2332"))

	sidebarStyle := lipgloss.NewStyle().
		Width(m.width/3).
		Height(m.height-3).
		Padding(1).
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(lipgloss.Color("#1a2332"))

	contentStyle := lipgloss.NewStyle().
		Width(m.width-24).
		Height(m.height-3).
		Padding(1, 2)

	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6b7280")).
		Width(m.width).
		Padding(0, 1)

	header := headerStyle.Render("blueprint")
	contentLines := []string{
		"Press n to create a new project.",
		"Press q to quit.",
	}
	if m.status != "" {
		contentLines = append(contentLines, "", "Status: "+m.status)
	}
	if m.errMsg != "" {
		contentLines = append(contentLines, "", "Error: "+m.errMsg)
	}
	content := contentStyle.Render(strings.Join(contentLines, "\n"))
	footer := footerStyle.Render("↑/↓ navigate • enter select • n new project • q quit")
	sidebar := sidebarStyle.Render(" ")
	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content)

	text := lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
	if m.wizard.active {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.wizardView())
	}
	return text
}

func (m Model) wizardView() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#22d3ee")).
		Padding(1, 2).
		Width(60)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#f8fafc"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#94a3b8"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#f87171"))
	highlight := lipgloss.NewStyle().Foreground(lipgloss.Color("#22d3ee")).Bold(true)

	title := titleStyle.Render("New Project")
	question := ""
	body := ""

	switch m.wizard.step {
	case stepProjectDir:
		question = "Project directory"
		body = m.wizard.input.View()
	case stepGitConfirm:
		question = "Initialize git?"
		body = m.renderOptions(highlight)
	case stepGitUser:
		question = "Git username"
		body = m.wizard.input.View()
	case stepGitRepo:
		question = "Git repo name"
		body = m.wizard.input.View()
	case stepLanguage:
		question = "Language"
		body = m.renderOptions(highlight)
	case stepTemplate:
		question = "Template"
		body = m.renderOptions(highlight)
	case stepConfirm:
		question = "Confirm"
		body = m.renderSummary(highlight)
	}

	lines := []string{
		title,
		labelStyle.Render(question),
		body,
		labelStyle.Render("enter confirm • esc cancel"),
	}
	if m.wizard.errMsg != "" {
		lines = append(lines, errorStyle.Render(m.wizard.errMsg))
	}

	return boxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m Model) renderOptions(highlight lipgloss.Style) string {
	lines := make([]string, 0, len(m.wizard.options))
	for i, opt := range m.wizard.options {
		prefix := "  "
		if i == m.wizard.selected {
			prefix = "> "
			lines = append(lines, highlight.Render(prefix+opt))
			continue
		}
		lines = append(lines, prefix+opt)
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) renderSummary(highlight lipgloss.Style) string {
	lines := []string{
		"Project: " + m.wizard.projectDir,
		"Git: " + boolLabel(m.wizard.gitEnabled),
		"Language: " + m.wizard.language,
		"Template: " + m.wizard.template,
	}
	if m.wizard.gitEnabled {
		lines = append(lines, "Git user: "+m.wizard.gitUser, "Git repo: "+m.wizard.gitRepo)
	}
	lines = append(lines, "")
	lines = append(lines, m.renderOptions(highlight))
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func boolLabel(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}
