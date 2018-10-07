package web

import (
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/internal/log"
	"github.com/aimeelaplant/comiccruncher/search"
	"github.com/aimeelaplant/comiccruncher/web/api"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"go.uber.org/zap"
)

// The struct for the web app with echo and the controllers.
type App struct {
	echo                *echo.Echo
	searchController    api.SearchController
	characterController api.CharacterController
	statsController     api.StatsController
}

// Run the web application from the specified port. Logs and exits if there is an error.
func (a App) MustRun(port string) {
	a.echo.Use(middleware.Recover())
	a.echo.Use(middleware.CSRF())
	a.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		// TODO: allow appropriate access-control-allow-origin
		AllowHeaders: []string{"application/json"},
	}))
	// Stats
	a.echo.GET("/stats", a.statsController.Stats)
	// Search
	a.echo.GET("/search/characters", a.searchController.SearchCharacters)
	// Characters
	a.echo.GET("/characters", a.characterController.Characters)
	a.echo.GET("/characters/:slug", a.characterController.Character)

	// Start the server.
	// Important to listen on localhost only so it binds to only localhost interface.
	if err := a.echo.Start("127.0.0.1:" + port); err != nil {
		log.WEB().Fatal("error starting server", zap.Error(err))
	}
}

func NewApp(
	characterSvc comic.CharacterServicer,
	searcher search.Searcher,
	statsRepository comic.StatsRepository) App {
	return App{
		echo:                echo.New(),
		statsController:     api.NewStatsController(statsRepository),
		searchController:    api.NewSearchController(searcher),
		characterController: api.NewCharacterController(characterSvc),
	}
}
