package cmd

import (
	"fmt"
	"github.com/aimeelaplant/comiccruncher/internal/log"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
)

// The root command.
var RootCmd = &cobra.Command{
	Use:   "cerebro",
	Short: "The application for importing resources from external sources.",
}

// Execution of the root command.
func Exec() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		log.CEREBRO().Error("received execution error", zap.Error(err))
		os.Exit(-1)
	}
}
