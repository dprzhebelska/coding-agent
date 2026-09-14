package main

// A simple program demonstrating the text area component from the Bubbles
// component library.

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"charm.land/bubbles/v2/cursor"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func main() {
	f, err := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	log.Println("Starting bubble tea")

	ctx := context.Background()
	p := tea.NewProgram(initialModel(ctx))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Oof: %v\n", err)
	}
}

type model struct {
	viewport         viewport.Model
	altscreenEnabled bool
	messages         []string
	textarea         textarea.Model
	senderStyle      lipgloss.Style
	agent            *agent
	ctx              context.Context
	title            string
	err              error
}

// agentResponse now carries both the response text and any title update.
type agentResponse struct {
	Text  string
	Title string
}

func initialModel(ctx context.Context) model {
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

	agent := newAgent(ctx)

	vp := viewport.New(viewport.WithWidth(30), viewport.WithHeight(5))
	vp.SetContent(fmt.Sprintf(`Welcome to the Coding Agent! Model selected: %s
Type a prompt and press Enter to send.`, agent.model))
	vp.KeyMap.Left.SetEnabled(false)
	vp.KeyMap.Right.SetEnabled(false)
	vp.MouseWheelEnabled = true

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return model{
		textarea:         ta,
		altscreenEnabled: true,
		messages:         []string{},
		viewport:         vp,
		senderStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		ctx:              ctx,
		agent:            agent,
		title:            "New conversation",
		err:              nil,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.SetWidth(msg.Width)
		m.textarea.SetWidth(msg.Width)
		m.viewport.SetHeight(msg.Height - m.textarea.Height())

		if len(m.messages) > 0 {
			// Wrap content before setting it.
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width()).Render(strings.Join(m.messages, "\n")))
		}
		m.viewport.GotoBottom()
	case agentResponse:
		if msg.Text == "" {
			return m, m.agentReponseCmd("")
		}
		m.messages = append(m.messages, m.senderStyle.Render("Agent: ")+msg.Text)
		if msg.Title != "" {
			m.title = msg.Title
		}
		m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width()).Render(strings.Join(m.messages, "\n")))
		return m, nil
	case tea.PasteMsg:
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			fmt.Println(m.textarea.Value())
			m.altscreenEnabled = false
			return m, tea.Quit
		case "enter":
			message := m.textarea.Value()
			if message == "" {
				return m, nil
			}
			if message == "exit" || message == "bye" {
				m.altscreenEnabled = false
				return m, tea.Quit
			}
			m.messages = append(m.messages, m.senderStyle.Render("You: ")+message)
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width()).Render(strings.Join(m.messages, "\n")))
			m.textarea.Reset()
			m.viewport.GotoBottom()

			return m, m.agentReponseCmd(message)
		default:
			// Send all other keypresses to the textarea.
			var cmd tea.Cmd
			m.textarea, cmd = m.textarea.Update(msg)
			return m, cmd
		}

	case cursor.BlinkMsg:
		// Textarea should also process cursor blinks.
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)

	return m, cmd
}

func (m model) View() tea.View {
	viewportView := m.viewport.View()
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

func (m model) agentReponseCmd(prompt string) tea.Cmd {
	return func() tea.Msg {
		resp, tuiUpdates, err := m.agent.makeResponse(prompt, m.ctx)
		if err != nil {
			if strings.Contains(err.Error(), "429") {
				resp = "Rate limit exceeded, please wait a few minutes and try again"
			} else {
				resp = "Model encountered an error... please retry"
			}
		}
		var title string
		for key, val := range tuiUpdates {
			if key == "title" && val != "" {
				log.Printf("Title update received: %s", val)
				title = val.(string)
			}
		}
		return agentResponse{Text: resp, Title: title}
	}
}
