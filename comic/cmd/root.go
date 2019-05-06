package cmd

import (
	"github.com/comiccruncher/comiccruncher/internal/log"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// RootCmd is the the root command for cerebro.
var RootCmd = &cobra.Command{
	Use:   "comic",
	Short: "The application for automating comic-package related tasks.",
}

// Exec executes the root command.
func Exec() {
	if err := RootCmd.Execute(); err != nil {
		log.COMIC().Fatal("received execution error", zap.Error(err))
	}
}
