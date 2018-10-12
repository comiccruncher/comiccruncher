package cmd

import (
	"github.com/aimeelaplant/comiccruncher/internal/log"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// RootCmd is the the root command for cerebro.
var RootCmd = &cobra.Command{
	Use:   "cerebro",
	Short: "The application for importing resources from external sources.",
}

// Execution of the root command.
func Exec() {
	if err := RootCmd.Execute(); err != nil {
		log.CEREBRO().Fatal("received execution error", zap.Error(err))
	}
}
