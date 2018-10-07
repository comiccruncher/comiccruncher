package cmd

import (
	"github.com/spf13/cobra"
	"github.com/aimeelaplant/comiccruncher/internal/log"
	"go.uber.org/zap"
	"os"
)

// The root command.
var RootCmd = &cobra.Command{
	Use:   "web",
	Short: "The application for starting the web application API.",
}

// Execution of the root command.
func Exec() {
	if err := RootCmd.Execute(); err != nil {
		log.WEB().Error("received execution error", zap.Error(err))
		os.Exit(-1)
	}
}
