package cmd

import (
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/internal/log"
	"github.com/aimeelaplant/comiccruncher/internal/pgo"
	"github.com/aimeelaplant/comiccruncher/internal/rediscache"
	"github.com/aimeelaplant/comiccruncher/search"
	"github.com/aimeelaplant/comiccruncher/web"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// The start command.
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "The command for starting the web application.",
	Run: func(cmd *cobra.Command, args []string) {
		instance, err := pgo.Instance()
		if err != nil {
			log.WEB().Fatal("cannot instantiate database connection", zap.Error(err))
		}
		container := comic.NewPGRepositoryContainer(instance)
		characterSvc := comic.NewCharacterServiceWithCache(container, rediscache.Instance())
		searchSvc := search.NewSearchService(instance)
		statsRepository := comic.NewPGStatsRepository(instance)
		app := web.NewApp(characterSvc, searchSvc, statsRepository)
		port := cmd.Flag("port")
		app.MustRun(port.Value.String())
	},
}

// Init scripts.
func init() {
	startCmd.Flags().IntP("port", "p", 8001, "Choose the port to start the web application on.")
	RootCmd.AddCommand(startCmd)
}
