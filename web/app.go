package web

import (
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/internal/log"
	"github.com/aimeelaplant/comiccruncher/search"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"go.uber.org/zap"
)

// App is the struct for the web app with echo and the controllers.
type App struct {
	echo                *echo.Echo
	searchController    SearchController
	characterController CharacterController
	statsController     StatsController
}

// MustRun runs the web application from the specified port. Logs and exits if there is an error.
func (a App) MustRun(port string) {
	// TODO: This is temporary until the site is ready.
	a.echo.Use(middleware.Recover())
	a.echo.Use(middleware.CSRF())
	a.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		// TODO: allow appropriate access-control-allow-origin
		AllowHeaders: []string{"application/json"},
	}))
	a.echo.Use(RequireAuthentication)
	a.echo.HTTPErrorHandler = ErrorHandler
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

// NewApp creates a new app from the parameters.
func NewApp(
	characterSvc comic.CharacterServicer,
	searcher search.Searcher,
	statsRepository comic.StatsRepository) App {
	return App{
		echo:                echo.New(),
		statsController:     NewStatsController(statsRepository),
		searchController:    NewSearchController(searcher),
		characterController: NewCharacterController(characterSvc),
	}
}
