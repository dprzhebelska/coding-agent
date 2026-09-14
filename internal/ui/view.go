package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m *MainModel) View() tea.View {
	viewportView := m.Viewport.View()
	v := tea.NewView(viewportView + "\n" + m.textarea.View())
	c := m.textarea.Cursor()
	if c != nil {
		c.Y += lipgloss.Height(viewportView)
	}
	v.Cursor = c
	v.AltScreen = m.altscreenEnabled
	v.MouseMode = tea.MouseModeCellMotion
	v.WindowTitle = fmt.Sprintf("Coding agent - %s", m.title)
	return v
}
