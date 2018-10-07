package cmd

import (
	"github.com/aimeelaplant/comiccruncher/internal/log"
	"github.com/spf13/cobra"
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
