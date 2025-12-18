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

var (
	appStyle     = lipgloss.NewStyle().Margin(1, 2)
	headerStyle  = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230")).Padding(0, 1).Bold(true)
	pendingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("211")).Bold(true)
	doneStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Strikethrough(true)
	boxStyle     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62")).Padding(1).Width(45).Height(20)
	focusedBox   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("205")).Padding(1).Width(45).Height(20)
	selStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
)

type focus int

const (
	focusTodos focus = iota
	focusNotes
	focusInput
)

type vimFinishedMsg struct {
	content string
	noteID  int
	err     error
}

type Model struct {
	db        *sql.DB
	todos     []storage.Todo
	notes     []storage.Note
	cursor    int
	focus     focus
	input     textinput.Model
	inputMode string
}

func NewModel(db *sql.DB) Model {
	ti := textinput.New()
	m := Model{db: db, focus: focusTodos, input: ti}
	m.refreshData()
	return m
}

func (m *Model) refreshData() {
	m.todos, _ = storage.GetTodos(m.db)
	m.notes, _ = storage.GetNotes(m.db)
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
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
			case "v":
				if m.focus == focusNotes && len(m.notes) > 0 {
					return m, m.openVimAction(m.notes[m.cursor])
				}
			case "x":
				if m.focus == focusTodos && len(m.todos) > 0 {
					storage.DeleteTodo(m.db, m.todos[m.cursor].ID)
				} else if m.focus == focusNotes && len(m.notes) > 0 {
					storage.DeleteNote(m.db, m.notes[m.cursor].ID)
				}
				m.refreshData()
				if m.cursor > 0 { m.cursor-- }
			case "t":
				m.focus = focusInput
				m.inputMode = "todo"
				m.input.Placeholder = "Task..."
				m.input.SetValue("")
				m.input.Focus()
				return m, textinput.Blink
			case "n":
				m.focus = focusInput
				m.inputMode = "note"
				m.input.Placeholder = "Note Title..."
				m.input.SetValue("")
				m.input.Focus()
				return m, textinput.Blink
			}
		} else {
			switch msg.String() {
			case "esc":
				m.focus = focusTodos
			case "enter":
				val := m.input.Value()
				if val == "" { return m, nil }
				if m.inputMode == "todo" {
					storage.AddTodo(m.db, val, time.Now().Add(1*time.Hour))
					m.refreshData()
					m.focus = focusTodos
					return m, nil
				} else {
					// 1. Save Title immediately so it appears in the list
					storage.AddNote(m.db, val, "") 
					m.refreshData()
					// 2. Open Vim for the newly created note (which will be at index 0)
					newNote := m.notes[0]
					m.focus = focusNotes
					return m, m.openVimAction(newNote)
				}
			}
		}

	case vimFinishedMsg:
		if msg.err == nil {
			storage.UpdateNoteContent(m.db, msg.noteID, msg.content)
		}
		m.refreshData()
		return m, nil
	}

	if m.focus == focusInput {
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) openVimAction(n storage.Note) tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" { editor = "vim" }

	f, _ := os.CreateTemp("", "mynote-*.md")
	f.Write([]byte(n.Content))
	f.Close()

	c := exec.Command(editor, f.Name())
	return tea.ExecProcess(c, func(err error) tea.Msg {
		content, _ := os.ReadFile(f.Name())
		os.Remove(f.Name())
		return vimFinishedMsg{content: string(content), noteID: n.ID, err: err}
	})
}

func (m Model) View() string {
	if m.focus == focusInput {
		return appStyle.Render(fmt.Sprintf("%s\n\n%s\n\n[Enter] Confirm  [Esc] Cancel", headerStyle.Render(" NEW "+m.inputMode), m.input.View()))
	}

	tView := headerStyle.Render(" TODOS ") + "\n\n" + pendingStyle.Render(" PENDING") + "\n"
	for i, t := range m.todos {
		if t.IsDone { continue }
		p, l := "  ", "[ ] "+t.Task
		if m.focus == focusTodos && m.cursor == i { p, l = selStyle.Render("> "), selStyle.Render(l) }
		tView += p + l + "\n"
	}
	tView += "\n" + doneStyle.Render(" COMPLETED") + "\n"
	for i, t := range m.todos {
		if !t.IsDone { continue }
		p, l := "  ", doneStyle.Render("[x] "+t.Task)
		if m.focus == focusTodos && m.cursor == i { p = selStyle.Render("> ") }
		tView += p + l + "\n"
	}

	nView := headerStyle.Render(" NOTES ") + "\n\n"
	for i, n := range m.notes {
		p, t := "  ", n.Title
		if m.focus == focusNotes && m.cursor == i { p, t = selStyle.Render("> "), selStyle.Render(t) }
		nView += p + t + "\n"
	}

	tBox, nBox := boxStyle, boxStyle
	if m.focus == focusTodos { tBox = focusedBox }
	if m.focus == focusNotes { nBox = focusedBox }

	return appStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, tBox.Render(tView), nBox.Render(nView)) +
		"\n [Tab] Switch  [n] Note  [t] Task  [v] View/Edit  [x] Delete")
}
