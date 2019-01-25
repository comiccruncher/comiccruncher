package cmd

import (
	"context"
	"github.com/aimeelaplant/comiccruncher/auth"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/internal/log"
	"github.com/aimeelaplant/comiccruncher/internal/pgo"
	"github.com/aimeelaplant/comiccruncher/internal/rediscache"
	"github.com/aimeelaplant/comiccruncher/search"
	"github.com/aimeelaplant/comiccruncher/web"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"time"
)

// The start command.
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "The command for starting the web application.",
	Run: func(cmd *cobra.Command, args []string) {
		instance, err := pgo.Instance()
		if err != nil {
			log.WEB().Fatal("cannot instantiate database", zap.Error(err))
		}
		redis := rediscache.Instance()
		container := comic.NewPGRepositoryContainer(instance)
		apps := comic.NewRedisAppearancesPerYearRepository(redis)
		ctr := comic.NewRedisCharacterThumbRepository(redis)
		expandedSvc := comic.NewExpandedService(container.CharacterRepository(), apps, redis, container.CharacterSyncLogRepository(), ctr)
		searchSvc := search.NewSearchService(instance)
		statsRepository := comic.NewPGStatsRepository(instance)
		rankedSvc := comic.NewRankedService(comic.NewPGPopularRepository(instance, comic.NewRedisCharacterThumbRepository(redis)))
		tr := auth.NewPGTokenRepository(instance)
		app := web.NewApp(expandedSvc, searchSvc, statsRepository, rankedSvc, ctr, tr)
		port := cmd.Flag("port")

		go func() {
			if appErr := app.Run(port.Value.String()); appErr != nil {
				log.WEB().Info("Shutting down web service...")
			}
		}()

		quit := make(chan os.Signal)
		signal.Notify(quit, os.Interrupt)
		<- quit
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if shutErr := app.Shutdown(ctx); err != nil {
			log.WEB().Fatal("error shutting down service", zap.Error(shutErr))
		}
		handleError(redis.Close(), "redis")
		handleError(instance.Close(), "database")
		log.WEB().Info("Gracefully shut down web service.")
	},
}

func handleError(err error, client string) {
	if err != nil {
		log.WEB().Error("error closing connection", zap.String("client", client),  zap.Error(err))
	}
	log.WEB().Info("closed connection", zap.String("client", client))
}

// Init scripts.
func init() {
	startCmd.Flags().IntP("port", "p", 8001, "Choose the port to start the web application on.")
	RootCmd.AddCommand(startCmd)
}
