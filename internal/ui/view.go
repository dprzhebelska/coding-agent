package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m *MainModel) View() tea.View {
	//m.updateTextAreaBoarder()
	viewportView := m.Viewport.View()
	content := viewportView + "\n" + m.getTextAreaView()
	v := tea.NewView(content)
	c := m.textarea.Cursor()
	if c != nil {
		c.Y += lipgloss.Height(viewportView) + 1
		c.X += 1
	}
	v.Cursor = c
	v.AltScreen = m.altscreenEnabled
	v.MouseMode = tea.MouseModeCellMotion
	v.WindowTitle = fmt.Sprintf("Coding agent - %s", m.title)
	v.KeyboardEnhancements.ReportEventTypes = true
	return v
}

func (m *MainModel) getTextAreaView() string {
	modelTitle := fmt.Sprintf(" model: %s ", m.agent.Model)

	modelTitleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#749ef4"))
	dirTitleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#6eca72"))
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#f474e5"))

	modelText := modelTitleStyle.Render(modelTitle)
	dash := borderStyle.Render("──")
	dirText := ""
	if m.currentDirectory != "" {
		dirText = " " + dirTitleStyle.Render(m.currentDirectory) + " "
	}

	width := m.windowWidth
	if m.windowWidth < 50 {
		width = 50
	}

	m.textarea.SetWidth(width - 5)

	rawTopLineContent := lipgloss.PlaceHorizontal(
		width-3,       // Total width minus the 2 corner characters
		lipgloss.Left, // Align titles to the left
		modelText+dash+dirText,
		lipgloss.WithWhitespaceChars("─"),
		// Lipgloss v2 expects a standard style block for whitespace configuration:
		lipgloss.WithWhitespaceStyle(
			lipgloss.NewStyle().Foreground(lipgloss.Color("#f474e5")),
		),
	)

	// 5. Construct the perfect top line manually
	topLeftCorner := borderStyle.Render("╭─")
	topRightCorner := borderStyle.Render("╮")

	topLine := topLeftCorner + rawTopLineContent + topRightCorner

	bodyStyle := lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.Border{
			Bottom:      "─",
			Left:        "│",
			Right:       "│",
			BottomLeft:  "╰",
			BottomRight: "╯",
		}, false, true, true, true). // explicitly disable the top border rule line
		BorderForeground(lipgloss.Color("#f474e5")).
		Padding(0, 0)

	// Remove cursor line styling
	s := textarea.DefaultStyles(true)

	s.Focused.Base = lipgloss.NewStyle().PaddingLeft(1)

	// Color the user's typed text
	s.Focused.Text = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))

	// De-emphasize placeholder text
	s.Focused.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	// Accentuate the active row the cursor is resting on
	s.Focused.CursorLine = lipgloss.NewStyle().Background(lipgloss.Color("235"))
	m.textarea.SetStyles(s)

	renderedTextArea := m.textarea.View()
	renderedBody := bodyStyle.Render(renderedTextArea)
	return lipgloss.JoinVertical(0, topLine, renderedBody)
}

func getCurrentDirWithTilde() (string, error) {
	// Get the absolute path of the current working directory
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Get the current user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		// If home dir cannot be inferred, return the full path anyway
		return wd, nil
	}

	// Check if the current working directory is inside the home directory
	if wd == home {
		return "~", nil
	} else if strings.HasPrefix(wd, home+string(filepath.Separator)) {
		return "~" + strings.TrimPrefix(wd, home), nil
	}

	// Return full path if it's outside the home directory
	return wd, nil
}
