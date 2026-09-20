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
	windowHeight       int
	windowWidth        int
	toggleIntialScreen bool
	Viewport           viewport.Model
	altscreenEnabled   bool
	messages           []string
	textarea           textarea.Model
	senderStyle        lipgloss.Style
	agentStyle         lipgloss.Style
	agent              *ai.Agent
	currentDirectory   string
	ctx                context.Context
	title              string
	err                error
}

func InitialModel(ctx context.Context) *MainModel {
	agent := ai.NewAgent(ctx)

	currentDir, err := getCurrentDirWithTilde()
	if err != nil {
		currentDir = ""
	}

	ta := textarea.New()
	ta.Placeholder = ""
	ta.DynamicHeight = true
	ta.MinHeight = 1
	ta.MaxHeight = 15
	ta.KeyMap.InsertNewline.SetEnabled(true)
	ta.SetVirtualCursor(false)
	ta.Focus()
	ta.Prompt = ""

	ta.CharLimit = 1000

	ta.SetWidth(28)

	ta.ShowLineNumbers = false

	vp := viewport.New(viewport.WithWidth(28), viewport.WithHeight(5))

	vp.SetContent(renderStartUpScreen(130, agent.Model, currentDir))
	vp.KeyMap.Left.SetEnabled(false)
	vp.KeyMap.Right.SetEnabled(false)
	vp.MouseWheelEnabled = true

	return &MainModel{
		windowHeight:       10,
		windowWidth:        30,
		textarea:           ta,
		altscreenEnabled:   true,
		toggleIntialScreen: true,
		messages:           []string{},
		Viewport:           vp,
		senderStyle:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#f474e5")),
		agentStyle:         lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#749ef4")),
		ctx:                ctx,
		agent:              agent,
		currentDirectory:   currentDir,
		title:              "New conversation",
		err:                nil,
	}
}

func (m *MainModel) Init() tea.Cmd {
	return textarea.Blink
}

func renderStartUpScreen(width int, model string, dir string) string {
	if width < 50 {
		width = 50
	}
	startUpstr := `┏┓   ┓•      ┏┓       
┃ ┏┓┏┫┓┏┓┏┓  ┣┫┏┓┏┓┏┓╋
┗┛┗┛┗┻┗┛┗┗┫  ┛┗┗┫┗ ┛┗┗
          ┛     ┛     `

	startUpCard := lipgloss.NewStyle().
		Width(30).
		Height(10).
		Foreground(lipgloss.Color("#749ef4")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#f474e5")).
		Align(lipgloss.Center, lipgloss.Center).
		Render(startUpstr)

	statsStr := fmt.Sprintf("model selected: %s\ncurrent directory: %s\n\nrecent sessions - coming soon", model, dir)

	statsCard := lipgloss.NewStyle().
		Width(width-30).
		Height(10).
		Foreground(lipgloss.Color("#749ef4")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#6f686e")).
		PaddingLeft(2).
		PaddingRight(0).
		Align(lipgloss.Left, lipgloss.Top).
		Render(statsStr)

	return lipgloss.JoinHorizontal(lipgloss.Left, startUpCard, statsCard)
}
