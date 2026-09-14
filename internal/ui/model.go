package ui

import (
	"context"
	"fmt"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/dprzhebelska/coding-agent/internal/ai"
)

type MainModel struct {
	Viewport         viewport.Model
	altscreenEnabled bool
	messages         []string
	textarea         textarea.Model
	senderStyle      lipgloss.Style
	agent            *ai.Agent
	ctx              context.Context
	title            string
	err              error
}

func InitialModel(ctx context.Context) *MainModel {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.SetVirtualCursor(false)
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 1000

	ta.SetWidth(30)
	ta.SetHeight(3)

	// Remove cursor line styling
	s := ta.Styles()
	s.Focused.CursorLine = lipgloss.NewStyle()
	ta.SetStyles(s)

	ta.ShowLineNumbers = false

	agent := ai.NewAgent(ctx)

	vp := viewport.New(viewport.WithWidth(30), viewport.WithHeight(5))
	vp.SetContent(fmt.Sprintf(`Welcome to the Coding Agent! Model selected: %s
Type a prompt and press Enter to send.`, agent.Model))
	vp.KeyMap.Left.SetEnabled(false)
	vp.KeyMap.Right.SetEnabled(false)
	vp.MouseWheelEnabled = true

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return &MainModel{
		textarea:         ta,
		altscreenEnabled: true,
		messages:         []string{},
		Viewport:         vp,
		senderStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		ctx:              ctx,
		agent:            agent,
		title:            "New conversation",
		err:              nil,
	}
}

func (m *MainModel) Init() tea.Cmd {
	return textarea.Blink
}
