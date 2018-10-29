package web

import (
	"errors"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/search"
	"github.com/labstack/echo"
	"strconv"
)

var (
	// ErrInvalidPageParameter is for when an invalid page parameter is received.
	ErrInvalidPageParameter = errors.New("invalid page parameter")
	// ErrInternalServerError is for when something bad happens internally.
	ErrInternalServerError = errors.New("internal server error")
	// ErrNotFound is for when something can't be found.
	ErrNotFound = errors.New("object cannot be found")
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

// CharacterController is the character controller.
type CharacterController struct {
	characterSvc comic.CharacterServicer
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
	cStrct, err := c.withAppearances(character)
	if err != nil {
		return err
	}
	return JSONDetailViewOK(ctx, cStrct)
}

// Characters lists the characters.
func (c CharacterController) Characters(ctx echo.Context) error {
	var results []*comic.Character
	page, err := pageNumber(ctx)
	if err != nil {
		return err
	}
	var slugs []comic.PublisherSlug
	results, err = c.characterSvc.CharactersByPublisher(slugs, true, 25+1, (page-1)*25)
	if err != nil {
		return err
	}
	var data = make([]interface{}, len(results))
	for i, v := range results {
		character, err := c.withAppearances(v)
		if err != nil {
			return err
		}
		data[i] = character
	}
	return JSONListViewOK(ctx, data, 25)
}

// Gets a character struct with the appearances attached.
func (c CharacterController) withAppearances(character *comic.Character) (Character, error) {
	apps, err := c.characterSvc.ListAppearances(character.Slug)
	if err != nil {
		return Character{}, err
	}
	return NewCharacter(*character, apps), nil
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

// NewCharacterController creates a new character controller.
func NewCharacterController(service comic.CharacterServicer) CharacterController {
	return CharacterController{
		characterSvc: service,
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
