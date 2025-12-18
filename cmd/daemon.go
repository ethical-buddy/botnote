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
		log.Println("Daemon started...")

		for range ticker.C {
			tasks, err := storage.GetPendingAlerts(db)
			if err != nil {
				log.Println("Error checking tasks:", err)
				continue
			}

			for _, t := range tasks {
				// Send notification using system notify-send
				exec.Command("notify-send", "-u", "critical", "Task Due!", t.Task).Run()
				
				storage.MarkAlerted(db, t.ID)
				log.Printf("Alerted for task: %s\n", t.Task)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
}
