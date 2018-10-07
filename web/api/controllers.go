package api

import (
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/internal/log"
	"github.com/aimeelaplant/comiccruncher/search"
	"github.com/labstack/echo"
	"go.uber.org/zap"
	"strconv"
)

// The controller for /stats
type StatsController struct {
	statsRepository comic.StatsRepository
}

// Shows the stats for comic cruncher.
func (c StatsController) Stats(ctx echo.Context) error {
	stats, err := c.statsRepository.Stats()
	if err != nil {
		return JSONServerError(ctx)
	}
	return JSONDetailViewOK(ctx, stats)
}

// The controller for search.
type SearchController struct {
	searcher search.Searcher
}

// Searches characters with the `query` parameter.
func (c SearchController) SearchCharacters(ctx echo.Context) error {
	var err error
	var results []*comic.Character
	query := ctx.QueryParam("query")
	if query != "" {
		results, err = c.searcher.Characters(ctx.QueryParam("query"), 5, 0)
		if err != nil {
			log.WEB().Error("error", zap.String("query", query), zap.Error(err))
			return JSONServerError(ctx)
		}
	}
	var data = make([]interface{}, len(results))
	for i, v := range results {
		data[i] = v
	}
	return JSONListViewOK(ctx, data, 5)
}

// The character controller.
type CharacterController struct {
	characterSvc    comic.CharacterServicer
}

// Gets a character by its slug.
func (c CharacterController) Character(ctx echo.Context) error {
	slug := comic.CharacterSlug(ctx.Param("slug"))
	character, err := c.characterSvc.Character(slug)
	if err != nil {
		return JSONServerError(ctx)
	}
	if character == nil {
		return JSONNotFound(ctx)
	}
	apps, err := c.characterSvc.ListAppearances(slug)
	if err != nil {
		return JSONServerError(ctx)
	}
	characterModel := NewCharacter(*character, apps)
	return JSONDetailViewOK(ctx, characterModel)
}

// Lists characters and can filter by publisher with `?publisher=marvel`.
func (c CharacterController) Characters(ctx echo.Context) error {
	var results []*comic.Character
	publisher := comic.PublisherSlug(ctx.QueryParam("publisher"))
	var err error
	pageNumber := 1
	if pageNumber, err = strconv.Atoi(ctx.QueryParam("page")); pageNumber != 0 && err != nil {
		return JSONBadRequest(ctx, "malformed `page` parameter")
	}
	var slugs []comic.PublisherSlug
	if publisher != "" {
		slugs = []comic.PublisherSlug{publisher}
	}
	results, err = c.characterSvc.CharactersByPublisher(slugs, true, 25+1, pageNumber * 25)
	if err != nil {
		return JSONServerError(ctx)
	}
	var data = make([]interface{}, len(results))
	for i, v := range results {
		data[i] = v
	}
	return JSONListViewOK(ctx, data, 25)
}

// Creates a new character controller.
func NewCharacterController(service comic.CharacterServicer) CharacterController {
	return CharacterController{
		characterSvc: service,
	}
}

// Creates a new search controller.
func NewSearchController(searcher search.Searcher) SearchController {
	return SearchController{
		searcher: searcher,
	}
}

// Creates a new stats controller.
func NewStatsController(repository comic.StatsRepository) StatsController {
	return StatsController{
		statsRepository: repository,
	}
}
