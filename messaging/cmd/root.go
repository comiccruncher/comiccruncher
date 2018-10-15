package cmd

import (
	"github.com/aimeelaplant/comiccruncher/internal/log"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
)

// RootCmd is the command for enqueue.
var RootCmd = &cobra.Command{
	Use:   "enqueue",
	Short: "The command for sending a message to the queue service.",
}

// Exec executes the root command.
func Exec() {
	if err := RootCmd.Execute(); err != nil {
		log.QUEUE().Error("got an error", zap.Error(err))
		os.Exit(-1)
	}
}
