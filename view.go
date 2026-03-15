package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/glamour"
	lipgloss "github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	sidebarWidth := m.width / 3
	if sidebarWidth < 24 {
		sidebarWidth = 24
	}
	if sidebarWidth > m.width-24 {
		sidebarWidth = maxInt(18, m.width-24)
	}
	contentWidth := m.width - sidebarWidth
	if contentWidth < 20 {
		contentWidth = 20
		sidebarWidth = maxInt(18, m.width-contentWidth)
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#22d3ee")).
		Width(m.width).
		Padding(0, 1).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#1a2332"))

	sidebarStyle := lipgloss.NewStyle().
		Width(sidebarWidth).
		Height(m.height-3).
		Padding(1).
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(lipgloss.Color("#1a2332"))

	contentStyle := lipgloss.NewStyle().
		Width(contentWidth).
		Height(m.height-3).
		Padding(1, 2)

	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6b7280")).
		Width(m.width).
		Padding(0, 1)

	header := headerStyle.Render("blueprint")
	sidebar := sidebarStyle.Render(m.renderProjectList())
	content := contentStyle.Render(m.renderProjectPreview(contentWidth, m.height-3))
	footer := footerStyle.Render(m.footerText())
	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content)

	text := lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
	if m.wizard.active {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.wizardView())
	}
	if m.confirmDelete {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.deleteConfirmView())
	}
	return text
}

func (m Model) renderProjectList() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#f8fafc"))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#64748b"))
	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#cbd5f5"))
	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#0f172a")).
		Background(lipgloss.Color("#22d3ee")).
		Bold(true).
		Padding(0, 1)

	lines := []string{
		titleStyle.Render("Projects"),
		mutedStyle.Render(fmt.Sprintf("%d total", len(m.projects))),
		"",
	}
	if len(m.projects) == 0 {
		lines = append(lines, mutedStyle.Render("No projects yet."))
		lines = append(lines, mutedStyle.Render("Press n to create one."))
		return lipgloss.JoinVertical(lipgloss.Left, lines...)
	}

	for i, project := range m.projects {
		name := project.Name
		if name == "" {
			name = filepath.Base(project.Path)
		}
		if i == m.selected {
			lines = append(lines, selectedStyle.Render(name))
		} else {
			lines = append(lines, itemStyle.Render(name))
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) renderProjectPreview(width, height int) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#f8fafc"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#94a3b8"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#e2e8f0"))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#64748b"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#f87171"))
	innerWidth := maxInt(10, width-4)

	project, ok := m.selectedProject()
	if !ok {
		return mutedStyle.Render("No project selected.")
	}

	name := project.Name
	if name == "" {
		name = filepath.Base(project.Path)
	}
	path := project.Path
	if strings.TrimSpace(path) == "" {
		path = "unknown"
	}

	infoLines := []string{
		titleStyle.Render("Project Overview"),
		labelStyle.Render("Name: ") + valueStyle.Render(name),
		labelStyle.Render("Path: ") + valueStyle.Render(path),
		labelStyle.Render("Language: ") + valueStyle.Render(emptyFallback(project.Language, "unknown")),
		labelStyle.Render("Template: ") + valueStyle.Render(emptyFallback(project.Template, "none")),
		labelStyle.Render("Git: ") + valueStyle.Render(boolLabel(project.Git)),
		labelStyle.Render("Created: ") + valueStyle.Render(project.CreatedAt.Format("2006-01-02 15:04")),
	}
	if project.Git && (project.GitUser != "" || project.GitRepo != "") {
		repo := strings.Trim(project.GitUser+"/"+project.GitRepo, "/")
		infoLines = append(infoLines, labelStyle.Render("Repo: ")+valueStyle.Render(repo))
	}

	treeHeader := titleStyle.Render("Project Tree")
	treeBody := ""
	switch {
	case m.previewLoading:
		treeBody = mutedStyle.Render("Loading project tree...")
	case m.previewErr != "":
		treeBody = errorStyle.Render(m.previewErr)
		if strings.TrimSpace(m.previewOutput) != "" {
			treeBody = lipgloss.JoinVertical(lipgloss.Left, treeBody, m.previewOutput)
		}
	default:
		treeBody = m.previewOutput
		if strings.TrimSpace(treeBody) == "" {
			treeBody = mutedStyle.Render("No output yet.")
		}
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinVertical(lipgloss.Left, infoLines...),
		"",
		treeHeader,
		treeBody,
	)
	readme := strings.TrimSpace(m.previewReadme)
	if readme != "" && shouldShowReadme(treeBody, innerWidth) {
		readmeHeader := titleStyle.Render("README")
		readmeBody := renderMarkdown(readme, innerWidth)
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			"",
			readmeHeader,
			readmeBody,
		)
	}
	return clipLines(content, height-1)
}

func shouldShowReadme(treeBody string, width int) bool {
	if width <= 0 {
		return false
	}
	return maxLineWidth(treeBody) <= width/2
}

func maxLineWidth(text string) int {
	maxWidth := 0
	for _, line := range strings.Split(text, "\n") {
		lineWidth := lipgloss.Width(line)
		if lineWidth > maxWidth {
			maxWidth = lineWidth
		}
	}
	return maxWidth
}

func renderMarkdown(source string, width int) string {
	if width <= 0 {
		return ""
	}
	renderer, err := glamour.NewTermRenderer(
		glamour.WithWordWrap(width),
		glamour.WithStandardStyle("dark"),
	)
	if err != nil {
		return source
	}
	rendered, err := renderer.Render(source)
	if err != nil {
		return source
	}
	return strings.TrimRight(rendered, "\n")
}

func (m Model) deleteConfirmView() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#f87171")).
		Padding(1, 2).
		Width(64)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#f8fafc"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#94a3b8"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#e2e8f0"))
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#fca5a5"))

	project := m.deleteCandidate
	name := project.Name
	if name == "" {
		name = filepath.Base(project.Path)
	}
	path := project.Path
	if strings.TrimSpace(path) == "" {
		path = "unknown"
	}
	lines := []string{
		titleStyle.Render("Delete Project"),
		warnStyle.Render("This will remove the project directory and its record."),
		"",
		labelStyle.Render("Name: ") + valueStyle.Render(name),
		labelStyle.Render("Path: ") + valueStyle.Render(path),
		"",
		labelStyle.Render("Press y to confirm • n or esc to cancel"),
	}
	return boxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m Model) footerText() string {
	footer := "↑/↓ navigate • n new project • d delete • q quit"
	if m.errMsg != "" {
		return footer + " | error: " + m.errMsg
	}
	if m.status != "" {
		return footer + " | " + m.status
	}
	return footer
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

func emptyFallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func clipLines(content string, maxLines int) string {
	if maxLines <= 0 {
		return ""
	}
	lines := strings.Split(content, "\n")
	if len(lines) <= maxLines {
		return content
	}
	lines = lines[:maxLines]
	if maxLines > 1 {
		lines[maxLines-1] = "\x1b[0m..."
	}
	return strings.Join(lines, "\n")
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
