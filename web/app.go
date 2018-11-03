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
	echo           *echo.Echo
	searchCtrlr    SearchController
	characterCtrlr CharacterController
	statsCtrlr     StatsController
	publisherCtrlr PublisherController
}

// MustRun runs the web application from the specified port. Logs and exits if there is an error.
func (a App) MustRun(port string) {
	a.echo.Use(middleware.Recover())
	a.echo.Use(middleware.CSRF())
	a.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		// TODO: allow appropriate access-control-allow-origin
		AllowHeaders: []string{"application/json"},
	}))
	// TODO: This is temporary until the site is ready.
	a.echo.Use(RequireAuthentication)

	// Set error handler.
	a.echo.HTTPErrorHandler = ErrorHandler

	// Stats
	a.echo.GET("/stats", a.statsCtrlr.Stats)
	// Search
	a.echo.GET("/search/characters", a.searchCtrlr.SearchCharacters)
	// Characters
	a.echo.GET("/characters", a.characterCtrlr.Characters)
	a.echo.GET("/characters/:slug", a.characterCtrlr.Character)
	// Publishers
	a.echo.GET("/publishers/dc", a.publisherCtrlr.DC)
	a.echo.GET("/publishers/marvel", a.publisherCtrlr.Marvel)

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
	statsRepository comic.StatsRepository,
	rankedSvc comic.RankedServicer) App {
	return App{
		echo:           echo.New(),
		statsCtrlr:     NewStatsController(statsRepository),
		searchCtrlr:    NewSearchController(searcher),
		characterCtrlr: NewCharacterController(characterSvc, rankedSvc),
		publisherCtrlr: NewPublisherController(rankedSvc),
	}
}
