package vimtutor

import (
	"fmt"
	"os"

	"github.com/BRO3886/vimtutor/internal/storage"
	"github.com/BRO3886/vimtutor/internal/ui"
	"github.com/spf13/cobra"
)

var learnCmd = &cobra.Command{
	Use:     "learn [lesson-id]",
	Aliases: []string{"l"},
	Short:   "Start or resume a lesson",
	Long: `Open the lesson menu, or jump directly to a specific lesson.

Examples:
  vimtutor learn
  vimtutor learn 01_basic_motions`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := storage.Open()
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}
		defer db.Close()

		if len(args) == 1 {
			return ui.RunLesson(db, args[0])
		}
		return ui.Run(db)
	},
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show your progress and stats",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := storage.Open()
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}
		defer db.Close()

		stats, err := db.GetOverallStats()
		if err != nil {
			return err
		}

		if stats == nil {
			fmt.Fprintln(os.Stdout, "No stats yet. Start a lesson with: vimtutor learn")
			return nil
		}

		fmt.Printf("\n  vimtutor stats\n")
		fmt.Printf("  ═══════════════════\n\n")
		fmt.Printf("  Level:    %d — %s\n", stats.Level+1, stats.LevelName)
		fmt.Printf("  Total XP: %d\n", stats.TotalXP)
		fmt.Printf("  Streak:   %d days\n", stats.StreakDays)
		fmt.Printf("  Sessions: %d\n", stats.TotalSessions)
		fmt.Printf("  Accuracy: %.1f%%\n", stats.AvgAccuracy*100)
		fmt.Printf("  Lessons:  %d completed\n\n", stats.LessonsCompleted)

		if len(stats.TopMistakes) > 0 {
			fmt.Printf("  Top mistakes:\n")
			for i, m := range stats.TopMistakes {
				if i >= 5 {
					break
				}
				fmt.Printf("    %-10s %d times\n", m.Key, m.Count)
			}
		}
		fmt.Println()
		return nil
	},
}

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset all progress (cannot be undone)",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print("This will delete ALL your progress. Type 'yes' to confirm: ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
		// For now, just tell user to delete the DB file
		fmt.Println("Delete ~/.vimtutor/vimtutor.db to reset all progress.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(learnCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(resetCmd)
}
