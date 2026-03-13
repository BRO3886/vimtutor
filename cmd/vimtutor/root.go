package vimtutor

import (
	"fmt"
	"os"

	"github.com/BRO3886/vimtutor/internal/storage"
	"github.com/BRO3886/vimtutor/internal/ui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "vimtutor",
	Short: "Learn vim interactively, track your progress",
	Long: `vimtutor — an interactive vim tutor with metrics tracking.

Start learning:
  vimtutor              opens the lesson menu
  vimtutor learn        same as above
  vimtutor stats        show your progress dashboard

Navigate the lesson menu with j/k, press Enter to start a lesson.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runApp()
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runApp() error {
	db, err := storage.Open()
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	return ui.Run(db)
}
