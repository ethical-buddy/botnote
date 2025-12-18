package cmd

import (
	"log"
	"mynotes/internal/storage"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run the background notification listener",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := storage.InitDB()
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		ticker := time.NewTicker(1 * time.Minute)
		for range ticker.C {
			tasks, err := storage.GetPendingAlerts(db)
			if err != nil {
				continue
			}

			for _, t := range tasks {
				exec.Command("notify-send", "-u", "critical", "Task Due!", t.Task).Run()
				storage.MarkAlerted(db, t.ID)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
}
