package cmd

import (
	"fmt"
	"log"
	"mynotes/internal/storage"
	"mynotes/internal/ui"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mynotes",
	Short: "A CLI Todo and Notes Manager",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := storage.InitDB()
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		p := tea.NewProgram(ui.NewModel(db), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
