package ui

import (
	"fmt"
	"log"
	"strings"

	"charm.land/bubbles/v2/cursor"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// agentResponse now carries both the response text and any title update.
type agentResponse struct {
	Text  string
	Title string
}

func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Viewport.SetWidth(msg.Width)
		m.textarea.SetWidth(msg.Width)
		m.Viewport.SetHeight(msg.Height - m.textarea.Height())

		if len(m.messages) > 0 {
			// Wrap content before setting it.
			m.Viewport.SetContent(lipgloss.NewStyle().Width(m.Viewport.Width()).Render(strings.Join(m.messages, "\n")))
		}
		m.Viewport.GotoBottom()
	case agentResponse:
		if msg.Text == "" {
			return m, m.agentReponseCmd("")
		}
		m.messages = append(m.messages, m.senderStyle.Render("Agent: ")+msg.Text)
		// if msg.Title != "" {
		// 	m.title = msg.Title
		// }
		m.Viewport.SetContent(lipgloss.NewStyle().Width(m.Viewport.Width()).Render(strings.Join(m.messages, "\n")))
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
			m.Viewport.SetContent(lipgloss.NewStyle().Width(m.Viewport.Width()).Render(strings.Join(m.messages, "\n")))
			m.textarea.Reset()
			m.Viewport.GotoBottom()

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
	m.Viewport, cmd = m.Viewport.Update(msg)

	return m, cmd
}

func (m *MainModel) agentReponseCmd(prompt string) tea.Cmd {
	return func() tea.Msg {
		resp, tuiUpdates, err := m.agent.MakeResponse(prompt, m.ctx)
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
				m.title = val.(string)
				// title = val.(string)
			}
		}
		return agentResponse{Text: resp, Title: title}
	}
}
