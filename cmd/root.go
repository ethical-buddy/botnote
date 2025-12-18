package cmd

import (
	"log"
	"mynotes/internal/storage"
	"mynotes/internal/ui"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mynotes",
	Short: "Notes and Todo CLI",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := storage.InitDB()
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		p := tea.NewProgram(ui.NewModel(db), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
