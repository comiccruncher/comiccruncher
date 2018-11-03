package web

import (
	"errors"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/search"
	"github.com/labstack/echo"
	"strconv"
)

// Pagination limit.
const pageLimit = 24

var (
	// ErrInvalidPageParameter is for when an invalid page parameter is received.
	ErrInvalidPageParameter = errors.New("invalid page parameter")
	// ErrInternalServerError is for when something bad happens internally.
	ErrInternalServerError = errors.New("internal server error")
	// ErrNotFound is for when something can't be found.
	ErrNotFound = errors.New("page cannot be found")
)

// StatsController is the controller for stats about comic cruncher.
type StatsController struct {
	statsRepository comic.StatsRepository
}

// Stats shows the stats for comic cruncher.
func (c StatsController) Stats(ctx echo.Context) error {
	stats, err := c.statsRepository.Stats()
	if err != nil {
		return err
	}
	return JSONDetailViewOK(ctx, stats)
}

// SearchController is the controller for search.
type SearchController struct {
	searcher search.Searcher
}

// SearchCharacters searches characters with the `query` parameter.
func (c SearchController) SearchCharacters(ctx echo.Context) error {
	var err error
	var results []*comic.Character
	query := ctx.QueryParam("query")
	if query != "" {
		results, err = c.searcher.Characters(ctx.QueryParam("query"), 5, 0)
		if err != nil {
			return err
		}
	}
	var data = make([]interface{}, len(results))
	for i, v := range results {
		data[i] = v
	}
	return JSONListViewOK(ctx, data, 5)
}

// PublisherController is the controller for publishers.
type PublisherController struct {
	rankedSvc comic.RankedServicer
}

// DC gets the publisher's characters with their appearances.
func (c PublisherController) DC(ctx echo.Context) error {
	cr, err := popularCriteria(ctx)
	if err != nil {
		return err
	}
	results, err := c.rankedSvc.DCPopular(cr)
	return JSONListViewOK(ctx, listRanked(results), pageLimit)
}

// Marvel gets the publisher's characters with their appearances.
func (c PublisherController) Marvel(ctx echo.Context) error {
	cr, err := popularCriteria(ctx)
	if err != nil {
		return err
	}
	results, err := c.rankedSvc.MarvelPopular(cr)
	if err != nil {
		return err
	}
	return JSONListViewOK(ctx, listRanked(results), pageLimit)
}

// CharacterController is the character controller.
type CharacterController struct {
	characterSvc comic.CharacterServicer
	rankedSvc comic.RankedServicer
}

// Character gets a character by its slug.
func (c CharacterController) Character(ctx echo.Context) error {
	slug := comic.CharacterSlug(ctx.Param("slug"))
	character, err := c.characterSvc.Character(slug)
	if err != nil {
		return err
	}
	if character == nil {
		return ErrNotFound
	}
	// TODO: Query for ranked character instead.
	apps, err := c.characterSvc.ListAppearances(character.Slug)
	if err != nil {
		return err
	}
	ch := NewCharacter(*character, apps)
	if err != nil {
		return err
	}
	return JSONDetailViewOK(ctx, ch)
}

// Characters lists the characters.
func (c CharacterController) Characters(ctx echo.Context) error {
	cr, err := popularCriteria(ctx)
	if err != nil {
		return err
	}
	results, err := c.rankedSvc.AllPopular(cr)
	if err != nil {
		return err
	}
	return JSONListViewOK(ctx, listRanked(results), pageLimit)
}

// Gets the page number from the query parameter `page` with default value if empty.
func pageNumber(ctx echo.Context) (int, error) {
	query := ctx.QueryParam("page")
	if query != "" {
		page, err := strconv.Atoi(query)
		if err != nil {
			return 1, ErrInvalidPageParameter
		}
		return page, nil
	}
	return 1, nil
}

// Gets a popular criteria struct based on the context.
func popularCriteria(ctx echo.Context) (comic.PopularCriteria, error) {
	page, err := pageNumber(ctx)
	if err != nil {
		return comic.PopularCriteria{}, err
	}
	sortBy := comic.MostIssues
	sortReq := ctx.QueryParam("sort")
	if sortReq == "average" {
		sortBy = comic.AverageIssuesPerYear
	}
	appearanceType := comic.Main | comic.Alternate
	typeReq := ctx.QueryParam("type")
	switch typeReq {
	case "main":
		appearanceType = comic.Main
		break
	case "alternate":
		appearanceType = comic.Alternate
		break
	}
	return comic.PopularCriteria{
		SortBy: sortBy,
		AppearanceType: appearanceType,
		Limit: pageLimit+1,
		Offset: (page-1)*pageLimit,
	}, nil
}

// Transforms ranked characters into an interface for pagination.
func listRanked(results []*comic.RankedCharacter) []interface{} {
	var data = make([]interface{}, len(results))
	for i, v := range results {
		data[i] = v
	}
	return data
}

// NewCharacterController creates a new character controller.
func NewCharacterController(service comic.CharacterServicer, rankedSvc comic.RankedServicer) CharacterController {
	return CharacterController{
		characterSvc: service,
		rankedSvc: rankedSvc,
	}
}

// NewSearchController creates a new search controller.
func NewSearchController(searcher search.Searcher) SearchController {
	return SearchController{
		searcher: searcher,
	}
}

// NewStatsController creates a new stats controller.
func NewStatsController(repository comic.StatsRepository) StatsController {
	return StatsController{
		statsRepository: repository,
	}
}

// NewPublisherController creates a new publisher controller.
func NewPublisherController(s comic.RankedServicer) PublisherController {
	return PublisherController{
		rankedSvc: s,
	}
}
