package web

import (
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/search"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// App is the struct for the web app with echo and the controllers.
type App struct {
	echo           *echo.Echo
	searchCtrlr    *SearchController
	characterCtrlr *CharacterController
	statsCtrlr     *StatsController
	publisherCtrlr *PublisherController
	trendingCtrlr  *TrendingController
}

// Run runs the web application from the specified port. Logs and exits if there is an error.
func (a App) Run(port string) error {
	a.echo.Use(middleware.Recover())
	a.echo.HTTPErrorHandler = ErrorHandler
	a.echo.Use(middleware.CSRF())
	// TODO: This is temporary until the site is ready.
	a.echo.Use(RequireAuthentication)
	a.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		// TODO: allow appropriate access-control-allow-origin
		AllowHeaders: []string{"application/json"},
	}))

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

	// trending
	a.echo.GET("trending/marvel", a.trendingCtrlr.Marvel)
	a.echo.GET("trending/dc", a.trendingCtrlr.DC)

	// Start the server.
	return a.echo.Start(":" + port)
}

// Close closes the app server.
func (a App) Close() error {
	return a.echo.Close()
}

// NewApp creates a new app from the parameters.
func NewApp(
	expandedSvc comic.ExpandedServicer,
	searcher search.Searcher,
	statsRepository comic.StatsRepository,
	rankedSvc comic.RankedServicer) *App {
	return &App{
		echo:           echo.New(),
		statsCtrlr:     NewStatsController(statsRepository),
		searchCtrlr:    NewSearchController(searcher),
		characterCtrlr: NewCharacterController(expandedSvc, rankedSvc),
		publisherCtrlr: NewPublisherController(rankedSvc),
		trendingCtrlr:  NewTrendingController(rankedSvc),
	}
}
