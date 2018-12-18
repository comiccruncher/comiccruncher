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
		redis := rediscache.Instance()
		container := comic.NewPGRepositoryContainer(instance)
		apps := comic.NewRedisAppearancesPerYearRepository(redis)
		ctr := comic.NewRedisCharacterThumbRepository(redis)
		expandedSvc := comic.NewExpandedService(container.CharacterRepository(), apps, redis, container.CharacterSyncLogRepository(), ctr)
		searchSvc := search.NewSearchService(instance)
		statsRepository := comic.NewPGStatsRepository(instance)
		rankedSvc := comic.NewRankedService(comic.NewPGPopularRepository(instance, comic.NewRedisCharacterThumbRepository(redis)))
		app := web.NewApp(expandedSvc, searchSvc, statsRepository, rankedSvc)
		port := cmd.Flag("port")
		if err = app.Run(port.Value.String()); err != nil {
			log.WEB().Fatal("error starting web service. closed it.", zap.Error(err), zap.Error(app.Close()))
		}
	},
}

// Init scripts.
func init() {
	startCmd.Flags().IntP("port", "p", 8001, "Choose the port to start the web application on.")
	RootCmd.AddCommand(startCmd)
}
