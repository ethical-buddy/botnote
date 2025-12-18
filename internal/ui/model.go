package ui

import (
	"database/sql"
	"fmt"
	"mynotes/internal/storage"
//	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- Styles ---
var (
	appStyle = lipgloss.NewStyle().Margin(1, 2)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			Bold(true)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205")).
				Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1).
			BorderForeground(lipgloss.Color("63"))
)

type focus int

const (
	focusTodos focus = iota
	focusNotes
	focusInput
)

type Model struct {
	db        *sql.DB
	todos     []storage.Todo
	notes     []storage.Note
	cursor    int
	focus     focus
	input     textinput.Model
	inputMode string // "todo" or "note"
	err       error
}

func NewModel(db *sql.DB) Model {
	ti := textinput.New()
	ti.Placeholder = "Type here..."
	ti.CharLimit = 156
	ti.Width = 40

	m := Model{
		db:    db,
		focus: focusTodos,
		input: ti,
	}
	m.refreshData()
	return m
}

func (m *Model) refreshData() {
	m.todos, _ = storage.GetTodos(m.db)
	m.notes, _ = storage.GetNotes(m.db)
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.focus != focusInput {
				return m, tea.Quit
			}
		case "tab":
			if m.focus == focusTodos {
				m.focus = focusNotes
				m.cursor = 0
			} else {
				m.focus = focusTodos
				m.cursor = 0
			}

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			maxLen := len(m.todos)
			if m.focus == focusNotes {
				maxLen = len(m.notes)
			}
			if m.cursor < maxLen-1 {
				m.cursor++
			}

		case "enter":
			if m.focus == focusInput {
				val := m.input.Value()
				if val != "" {
					if m.inputMode == "todo" {
						// Defaulting to due in 1 hour for simplicity
						storage.AddTodo(m.db, val, time.Now().Add(1*time.Hour))
					} else {
						storage.AddNote(m.db, val)
					}
					m.refreshData()
					m.input.SetValue("")
					m.focus = focusTodos
				}
			} else if m.focus == focusTodos && len(m.todos) > 0 {
				storage.ToggleTodo(m.db, m.todos[m.cursor].ID, m.todos[m.cursor].IsDone)
				m.refreshData()
			}

		case "n":
			if m.focus != focusInput {
				m.focus = focusInput
				m.inputMode = "note"
				m.input.Placeholder = "New Note..."
				m.input.Focus()
			}
		case "t":
			if m.focus != focusInput {
				m.focus = focusInput
				m.inputMode = "todo"
				m.input.Placeholder = "New Task (Due 1h)..."
				m.input.Focus()
			}
		case "esc":
			m.focus = focusTodos
			m.input.Blur()
		}
	}

	if m.focus == focusInput {
		m.input, cmd = m.input.Update(msg)
	}

	return m, cmd
}

func (m Model) View() string {
	if m.focus == focusInput {
		return boxStyle.Render(fmt.Sprintf(
			"Adding new %s:\n\n%s\n\n(Enter to save, Esc to cancel)",
			m.inputMode,
			m.input.View(),
		))
	}

	// Render Todos
	todoView := headerStyle.Render(" TODOS ") + "\n\n"
	for i, t := range m.todos {
		cursor := " "
		if m.focus == focusTodos && m.cursor == i {
			cursor = ">"
		}

		checked := "[ ]"
		if t.IsDone {
			checked = "[x]"
		}

		line := fmt.Sprintf("%s %s %s", cursor, checked, t.Task)
		if m.focus == focusTodos && m.cursor == i {
			line = selectedItemStyle.Render(line)
		}
		todoView += line + "\n"
	}

	// Render Notes
	noteView := headerStyle.Render(" NOTES ") + "\n\n"
	for i, n := range m.notes {
		cursor := " "
		if m.focus == focusNotes && m.cursor == i {
			cursor = ">"
		}
		line := fmt.Sprintf("%s • %s", cursor, n.Content)
		if m.focus == focusNotes && m.cursor == i {
			line = selectedItemStyle.Render(line)
		}
		noteView += line + "\n"
	}

	// Join them side by side
	return appStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			boxStyle.Width(40).Render(todoView),
			boxStyle.Width(40).Render(noteView),
		) + "\n\n[t] New Todo | [n] New Note | [Tab] Switch | [q] Quit",
	)
}
