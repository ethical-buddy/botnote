package storage

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Todo struct {
	ID        int
	Task      string
	IsDone    bool
	DueAt     time.Time
	AlertSent bool
}

type Note struct {
	ID      int
	Content string
}

func InitDB() (*sql.DB, error) {
	home, _ := os.UserHomeDir()
	dbPath := filepath.Join(home, ".local", "share", "mynotes")
	os.MkdirAll(dbPath, 0755)

	db, err := sql.Open("sqlite3", filepath.Join(dbPath, "data.db"))
	if err != nil {
		return nil, err
	}

	query := `
	CREATE TABLE IF NOT EXISTS todos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task TEXT,
		is_done BOOLEAN DEFAULT 0,
		due_at DATETIME,
		alert_sent BOOLEAN DEFAULT 0
	);
	CREATE TABLE IF NOT EXISTS notes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		content TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(query)
	return db, err
}

// --- Todo Operations ---

func AddTodo(db *sql.DB, task string, due time.Time) error {
	_, err := db.Exec("INSERT INTO todos (task, due_at) VALUES (?, ?)", task, due)
	return err
}

func GetTodos(db *sql.DB) ([]Todo, error) {
	rows, err := db.Query("SELECT id, task, is_done, due_at FROM todos ORDER BY is_done ASC, due_at ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		rows.Scan(&t.ID, &t.Task, &t.IsDone, &t.DueAt)
		todos = append(todos, t)
	}
	return todos, nil
}

func ToggleTodo(db *sql.DB, id int, currentStatus bool) error {
	_, err := db.Exec("UPDATE todos SET is_done = ? WHERE id = ?", !currentStatus, id)
	return err
}

func GetPendingAlerts(db *sql.DB) ([]Todo, error) {
	rows, err := db.Query("SELECT id, task FROM todos WHERE due_at <= ? AND alert_sent = 0 AND is_done = 0", time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		rows.Scan(&t.ID, &t.Task)
		todos = append(todos, t)
	}
	return todos, nil
}

func MarkAlerted(db *sql.DB, id int) error {
	_, err := db.Exec("UPDATE todos SET alert_sent = 1 WHERE id = ?", id)
	return err
}

// --- Note Operations ---

func AddNote(db *sql.DB, content string) error {
	_, err := db.Exec("INSERT INTO notes (content) VALUES (?)", content)
	return err
}

func GetNotes(db *sql.DB) ([]Note, error) {
	rows, err := db.Query("SELECT id, content FROM notes ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		rows.Scan(&n.ID, &n.Content)
		notes = append(notes, n)
	}
	return notes, nil
}
