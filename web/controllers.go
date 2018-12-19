package web

import (
	"errors"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/search"
	"github.com/labstack/echo"
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
	ctr       comic.CharacterThumbRepository
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
	if len(results) > 0 {
		slugs := make([]comic.CharacterSlug, len(results))
		for i, ch := range results {
			slugs[i] = ch.Slug
		}
		thumbs, err := c.ctr.AllThumbnails(slugs...)
		if err != nil {
			return err
		}
		for i, v := range results {
			data[i] = NewCharacter(v, thumbs[v.Slug])
		}
	}
	return JSONListViewOK(ctx, data, 5)
}

// PublisherController is the controller for publishers.
type PublisherController struct {
	rankedSvc comic.RankedServicer
}

// DC gets the publisher's characters with their appearances.
func (c PublisherController) DC(ctx echo.Context) error {
	cr, err := decodeCriteria(ctx)
	if err != nil {
		return err
	}
	results, err := c.rankedSvc.DCPopular(cr)
	if err != nil {
		return err
	}
	return JSONListViewOK(ctx, listRanked(results), pageLimit)
}

// Marvel gets the publisher's characters with their appearances.
func (c PublisherController) Marvel(ctx echo.Context) error {
	cr, err := decodeCriteria(ctx)
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
	rankedSvc   comic.RankedServicer
	expandedSvc comic.ExpandedServicer
}

// Character gets a character by its slug.
func (c CharacterController) Character(ctx echo.Context) error {
	slug := comic.CharacterSlug(ctx.Param("slug"))
	character, err := c.expandedSvc.Character(slug)
	if err != nil {
		return err
	}
	if character == nil {
		return ErrNotFound
	}
	return JSONDetailViewOK(ctx, character)
}

// Characters lists the characters.
func (c CharacterController) Characters(ctx echo.Context) error {
	cr, err := decodeCriteria(ctx)
	if err != nil {
		return err
	}
	results, err := c.rankedSvc.AllPopular(cr)
	if err != nil {
		return err
	}
	return JSONListViewOK(ctx, listRanked(results), pageLimit)
}

// TrendingController is the controller for trending characters.
type TrendingController struct {
	svc comic.RankedServicer
}

// Marvel gets the trending characters for Marvel.
func (c *TrendingController) Marvel(ctx echo.Context) error {
	page, err := parsePageNumber(ctx)
	if err != nil {
		return err
	}
	results, err := c.svc.MarvelTrending(pageLimit+1, (page-1)*pageLimit)
	if err != nil {
		return err
	}
	return JSONListViewOK(ctx, listRanked(results), pageLimit)
}

// DC gets the trending characters for DC.
func (c *TrendingController) DC(ctx echo.Context) error {
	page, err := parsePageNumber(ctx)
	if err != nil {
		return err
	}
	results, err := c.svc.DCTrending(pageLimit+1, (page-1)*pageLimit)
	if err != nil {
		return err
	}
	return JSONListViewOK(ctx, listRanked(results), pageLimit)
}

// Gets a popular criteria struct based on the context.
func decodeCriteria(ctx echo.Context) (comic.PopularCriteria, error) {
	page, err := parsePageNumber(ctx)
	if err != nil {
		return comic.PopularCriteria{}, err
	}
	sortBy := comic.MostIssues
	sortReq := ctx.QueryParam("sort")
	if sortReq == "average" {
		sortBy = comic.AverageIssuesPerYear
	}
	appearanceType := comic.Main | comic.Alternate
	typeReq := ctx.QueryParam("category")
	switch typeReq {
	case "main":
		appearanceType = comic.Main
		break
	case "alternate":
		appearanceType = comic.Alternate
		break
	}
	return comic.PopularCriteria{
		SortBy:         sortBy,
		AppearanceType: appearanceType,
		Limit:          pageLimit + 1,
		Offset:         (page - 1) * pageLimit,
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
func NewCharacterController(eSvc comic.ExpandedServicer, rSvc comic.RankedServicer) *CharacterController {
	return &CharacterController{
		expandedSvc: eSvc,
		rankedSvc:   rSvc,
	}
}

// NewSearchController creates a new search controller.
func NewSearchController(searcher search.Searcher, ctr comic.CharacterThumbRepository) *SearchController {
	return &SearchController{
		searcher: searcher,
		ctr: ctr,
	}
}

// NewStatsController creates a new stats controller.
func NewStatsController(repository comic.StatsRepository) *StatsController {
	return &StatsController{
		statsRepository: repository,
	}
}

// NewPublisherController creates a new publisher controller.
func NewPublisherController(s comic.RankedServicer) *PublisherController {
	return &PublisherController{
		rankedSvc: s,
	}
}

// NewTrendingController creates a new trending controller.
func NewTrendingController(s comic.RankedServicer) *TrendingController {
	return &TrendingController{
		svc: s,
	}
}
