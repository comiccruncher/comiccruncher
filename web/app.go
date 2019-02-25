package web

import (
	"context"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/search"
	"github.com/go-pg/pg"
	"github.com/go-redis/redis"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"os"
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
	e := a.echo
	e.Use(middleware.Recover())
	e.HTTPErrorHandler = ErrorHandler

	allowedOrigins := []string{"*"}
	if os.Getenv("CC_ENVIRONMENT") == "production" {
		allowedOrigins = []string{"https://comiccruncher.com"}
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			echo.HeaderVary,
			"X-VISITOR-ID",
		},
		// AllowCredentials: true,
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{"HEAD", "GET", "OPTIONS"},
		MaxAge: 86400,
	}))

	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection: "1; mode=block",
		ContentTypeNosniff: "nosniff",
		XFrameOptions: "SAMEORIGIN",
		HSTSMaxAge: 31536000,
	}))
	// TODO: Use when ready.
	// e.POST("/authenticate", a.authCtrlr.Authenticate)
	// jwtMiddleware := NewDefaultJWTMiddleware()

	// Stats
	e.GET("/stats", a.statsCtrlr.Stats)

	// Search
	s := e.Group("/search")
	s.GET("/characters", a.searchCtrlr.SearchCharacters)

	// Characters
	c := e.Group("/characters")
	c.GET("", a.characterCtrlr.Characters)
	c.GET("/:slug", a.characterCtrlr.Character)

	// Publishers
	p := e.Group("/publishers")
	p.GET("/dc", a.publisherCtrlr.DC)
	p.GET("/marvel", a.publisherCtrlr.Marvel)

	// trending
	t := e.Group("/trending")
	t.GET("/marvel", a.trendingCtrlr.Marvel)
	t.GET("/dc", a.trendingCtrlr.DC)

	// Start the server.
	return e.Start(":" + port)
}

// Close closes the app server.
func (a *App) Close() error {
	return a.echo.Close()
}

// Shutdown gracefully shuts down the application.
func (a *App) Shutdown(ctx context.Context) error {
	return a.echo.Shutdown(ctx)
}

// NewApp creates a new app from the parameters.
func NewApp(
	expandedSvc comic.ExpandedServicer,
	searcher search.Searcher,
	statsRepository comic.StatsRepository,
	rankedSvc comic.RankedServicer,
	ctr comic.CharacterThumbRepository) *App {
	return &App{
		echo:           echo.New(),
		statsCtrlr:     NewStatsController(statsRepository),
		searchCtrlr:    NewSearchController(searcher, ctr),
		characterCtrlr: NewCharacterController(expandedSvc, rankedSvc),
		publisherCtrlr: NewPublisherController(rankedSvc),
		trendingCtrlr:  NewTrendingController(rankedSvc),
	}
}

// NewAppFactory creates a new app with minimal dependencies.
func NewAppFactory(db *pg.DB, redis *redis.Client) *App {
	return NewApp(
		comic.NewExpandedServiceFactory(db, redis),
		search.NewSearchService(db),
		comic.NewPGStatsRepository(db),
		comic.NewRankedServiceFactory(db, redis),
		comic.NewRedisCharacterThumbRepository(redis))
}
