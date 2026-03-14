package main

import (
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
	content := contentStyle.Render("hello world welcome to my app : blueprint")
	footer := footerStyle.Render("↑/↓ navigate • enter select • q quit")
	sidebar := sidebarStyle.Render(" ")
	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content)

	text := lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
	return text
}
