package ui

import (
	"database/sql"
	"fmt"
	"mynotes/internal/storage"
	"os"
	"os/exec"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- Styles ---
var (
	appStyle     = lipgloss.NewStyle().Margin(1, 2)
	pendingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("211")).Bold(true) // Pinkish/Red for pending
	doneStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Strikethrough(true)
	headerStyle  = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230")).Padding(0, 1)
	boxStyle     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62")).Padding(1).Width(45).Height(15)
	selStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
)

type focus int

const (
	focusTodos focus = iota
	focusNotes
	focusInput
)

// Msg sent when returning from Vim
type vimFinishedMsg struct {
	content string
	err     error
}

type Model struct {
	db         *sql.DB
	todos      []storage.Todo
	notes      []storage.Note
	cursor     int
	focus      focus
	input      textinput.Model
	inputMode  string // "todo" or "note_title"
	tempTitle  string // Stores note title while vim is open
	terminalW  int
	terminalH  int
}

func NewModel(db *sql.DB) Model {
	ti := textinput.New()
	ti.Focus()

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
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.terminalW = msg.Width
		m.terminalH = msg.Height

	case tea.KeyMsg:
		// Global keys
		if m.focus != focusInput {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "tab":
				if m.focus == focusTodos {
					m.focus = focusNotes
				} else {
					m.focus = focusTodos
				}
				m.cursor = 0
			case "up", "k":
				if m.cursor > 0 { m.cursor-- }
			case "down", "j":
				limit := len(m.todos)
				if m.focus == focusNotes { limit = len(m.notes) }
				if m.cursor < limit-1 { m.cursor++ }
			case "enter":
				if m.focus == focusTodos && len(m.todos) > 0 {
					t := m.todos[m.cursor]
					storage.ToggleTodo(m.db, t.ID, t.IsDone)
					m.refreshData()
				}
			case "x": // DELETE
				if m.focus == focusTodos && len(m.todos) > 0 {
					storage.DeleteTodo(m.db, m.todos[m.cursor].ID)
				} else if m.focus == focusNotes && len(m.notes) > 0 {
					storage.DeleteNote(m.db, m.notes[m.cursor].ID)
				}
				m.refreshData()
			case "t":
				m.focus = focusInput
				m.inputMode = "todo"
				m.input.Placeholder = "Task description..."
				m.input.Focus()
				return m, textinput.Blink
			case "n":
				m.focus = focusInput
				m.inputMode = "note_title"
				m.input.Placeholder = "Note title..."
				m.input.Focus()
				return m, textinput.Blink
			}
		} else {
			// Input mode keys
			switch msg.String() {
			case "esc":
				m.focus = focusTodos
				m.input.Blur()
			case "enter":
				val := m.input.Value()
				if val == "" { return m, nil }
				
				if m.inputMode == "todo" {
					storage.AddTodo(m.db, val, time.Now().Add(1*time.Hour))
					m.refreshData()
					m.focus = focusTodos
					m.input.SetValue("")
				} else if m.inputMode == "note_title" {
					m.tempTitle = val
					m.input.SetValue("")
					// TRIGGER VIM
					return m, m.openVimAction()
				}
			}
		}

	case vimFinishedMsg:
		if msg.err == nil && msg.content != "" {
			storage.AddNote(m.db, m.tempTitle, msg.content)
			m.refreshData()
		}
		m.focus = focusNotes
	}

	if m.focus == focusInput {
		m.input, cmd = m.input.Update(msg)
	}

	return m, cmd
}

// Action to open Vim
func (m Model) openVimAction() tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" { editor = "vim" }

	// Create temp file
	f, _ := os.CreateTemp("", "note-*.md")
	f.Close()

	c := exec.Command(editor, f.Name())
	return tea.ExecProcess(c, func(err error) tea.Msg {
		content, _ := os.ReadFile(f.Name())
		os.Remove(f.Name())
		return vimFinishedMsg{content: string(content), err: err}
	})
}

func (m Model) View() string {
	if m.focus == focusInput {
		return appStyle.Render(fmt.Sprintf(
			"Creating %s\n\n%s\n\n[Enter] Confirm  [Esc] Cancel",
			m.inputMode,
			m.input.View(),
		))
	}

	// --- TODO SECTION ---
	todoContent := headerStyle.Render("TODOS") + "\n\n"
	
	// Pending
	todoContent += pendingStyle.Render("󱎫 PENDING") + "\n"
	for i, t := range m.todos {
		if t.IsDone { continue }
		cursor := "  "
		line := fmt.Sprintf("[ ] %s", t.Task)
		if m.focus == focusTodos && m.cursor == i {
			cursor = selStyle.Render("> ")
			line = selStyle.Render(line)
		}
		todoContent += cursor + line + "\n"
	}

	// Completed
	todoContent += "\n" + doneStyle.Render("󰄬 COMPLETED") + "\n"
	for i, t := range m.todos {
		if !t.IsDone { continue }
		cursor := "  "
		line := doneStyle.Render(fmt.Sprintf("[x] %s", t.Task))
		if m.focus == focusTodos && m.cursor == i {
			cursor = selStyle.Render("> ")
		}
		todoContent += cursor + line + "\n"
	}

	// --- NOTES SECTION ---
	notesContent := headerStyle.Render("NOTES") + "\n\n"
	for i, n := range m.notes {
		cursor := "  "
		title := n.Title
		if m.focus == focusNotes && m.cursor == i {
			cursor = selStyle.Render("> ")
			title = selStyle.Render(title)
		}
		notesContent += fmt.Sprintf("%s%s\n", cursor, title)
	}

	// Layout
	mainView := lipgloss.JoinHorizontal(
		lipgloss.Top,
		boxStyle.Render(todoContent),
		boxStyle.Render(notesContent),
	)

	help := "\n[t] New Task  [n] New Note (Vim)  [x] Delete  [Tab] Switch  [Enter] Toggle  [q] Quit"
	return appStyle.Render(mainView + help)
}
