package web_test

import (
	"errors"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/internal/mocks/comic"
	"github.com/aimeelaplant/comiccruncher/internal/mocks/search"
	"github.com/aimeelaplant/comiccruncher/web"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStatsControllerStatsReturnsOK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	stats := comic.Stats{
		TotalCharacters:  1,
		TotalAppearances: 1,
		MinYear:          1,
		MaxYear:          10,
		TotalIssues:      1,
	}
	statsMock := mock_comic.NewMockStatsRepository(ctrl)
	statsMock.EXPECT().Stats().Return(stats, nil)
	statsCtrl := web.NewStatsController(statsMock)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := statsCtrl.Stats(c)
	header := c.Response().Header()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)
	assert.Equal(t, "application/json; charset=UTF-8", header.Get("Content-Type"))
}

func TestSearchControllerSearchCharactersNotEmptyQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSearch := mock_search.NewMockSearcher(ctrl)

	characters := []*comic.Character{mockCharacter()}
	mockSearch.EXPECT().Characters("emma-frost", 5, 0).Return(characters, nil)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/search?query=emma-frost", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	searchCtrl := web.NewSearchController(mockSearch)
	err := searchCtrl.SearchCharacters(c)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)
}

func TestSearchControllerSearchCharactersEmptyQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSearch := mock_search.NewMockSearcher(ctrl)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/search", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	header := c.Response().Header()

	searchCtrl := web.NewSearchController(mockSearch)
	err := searchCtrl.SearchCharacters(c)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)
	assert.Equal(t, "application/json; charset=UTF-8", header.Get("Content-Type"))
}

func TestSearchControllerSearchCharactersError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSearch := mock_search.NewMockSearcher(ctrl)
	mockSearch.EXPECT().Characters("emma-frost", 5, 0).Return(nil, errors.New("something happened"))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/search?query=emma-frost", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	searchCtrl := web.NewSearchController(mockSearch)
	err := searchCtrl.SearchCharacters(c)

	assert.Error(t, err)
	// no middleware to handle error in tests
	assert.Equal(t, 0, c.Response().Status)
}

func TestCharacterControllerCharacter(t *testing.T) {
	file, err := ioutil.ReadFile("./testdata/character.json")
	assert.Nil(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	characterSvc := mock_comic.NewMockCharacterServicer(ctrl)
	characterSvc.EXPECT().Character(gomock.Any()).Return(mockCharacter(), nil)
	characterSvc.EXPECT().ListAppearances(gomock.Any()).Return([]comic.AppearancesByYears{}, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/characters/emma-frost", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	header := c.Response().Header()

	rankedSvc := mock_comic.NewMockRankedServicer(ctrl)
	characterCtrl := web.NewCharacterController(characterSvc, rankedSvc)
	err = characterCtrl.Character(c)
	read, err := ioutil.ReadAll(rec.Body)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)
	assert.True(t, c.Response().Committed)
	assert.Equal(t, "application/json; charset=UTF-8", header.Get("Content-Type"))
	assert.Nil(t, err)
	assert.Equal(t, file, read)
}

func TestCharacterControllerCharacterNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	characterSvc := mock_comic.NewMockCharacterServicer(ctrl)
	characterSvc.EXPECT().Character(gomock.Any()).Return(nil, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/characters/emma-frost", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	rankedSvc := mock_comic.NewMockRankedServicer(ctrl)
	characterCtrl := web.NewCharacterController(characterSvc, rankedSvc)
	err := characterCtrl.Character(c)

	assert.Equal(t, web.ErrNotFound.Error(), err.Error())
}

func TestCharacterControllerCharacters(t *testing.T) {
	file, err := ioutil.ReadFile("./testdata/characters.json")
	assert.Nil(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	aggs := []comic.YearlyAggregate{
		{Year: 2016, Count: 5},
		{Year: 2017, Count: 10},
	}
	apps := []comic.AppearancesByYears{
		comic.NewAppearancesByYears("test", comic.Main, aggs),
	}
	p := comic.Publisher{ID: 1, Slug: "marvel", Name: "Marvel"}
	rankedChrs := []*comic.RankedCharacter{
		{ID: 1, PublisherID: 1, Publisher: p, AvgRank: 2, AvgRankID: 1, IssueCount: 10, IssueCountRankID: 1, Name: "Test", Slug: "test", Appearances: apps},
		{ID: 2, PublisherID: 1, Publisher: p, AvgRank: 2, AvgRankID: 2, IssueCount: 5, IssueCountRankID: 2, Name: "Test2", Slug: "test2"},
	}
	characterSvc := mock_comic.NewMockCharacterServicer(ctrl)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/characters?page=1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	header := c.Response().Header()

	rankedSvc := mock_comic.NewMockRankedServicer(ctrl)
	rankedSvc.EXPECT().AllPopular(gomock.Any()).Return(rankedChrs, nil)
	characterCtrl := web.NewCharacterController(characterSvc, rankedSvc)
	// make the call
	err = characterCtrl.Characters(c)
	assert.Nil(t, err)

	read, err := ioutil.ReadAll(rec.Body)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)
	assert.True(t, c.Response().Committed)
	assert.Equal(t, "application/json; charset=UTF-8", header.Get("Content-Type"))
	assert.Nil(t, err)
	assert.Equal(t, file, read)
}

func TestPublisherControllerDC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/characters?page=1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	header := c.Response().Header()

	aggs := []comic.YearlyAggregate{
		{Year: 2016, Count: 5},
		{Year: 2017, Count: 10},
	}
	apps := []comic.AppearancesByYears{
		comic.NewAppearancesByYears("test", comic.Main, aggs),
	}
	rankedChrs := []*comic.RankedCharacter{
		{ID: 1, PublisherID: 1, AvgRank: 2, AvgRankID: 1, IssueCount: 10, IssueCountRankID: 1, Name: "Test", Slug: "test", Appearances: apps},
		{ID: 2, PublisherID: 1, AvgRank: 2, AvgRankID: 2, IssueCount: 5, IssueCountRankID: 2, Name: "Test2", Slug: "test2"},
	}
	rankedSvc := mock_comic.NewMockRankedServicer(ctrl)
	rankedSvc.EXPECT().DCPopular(gomock.Any()).Return(rankedChrs, nil)

	publisherCtrlr := web.NewPublisherController(rankedSvc)
	err := publisherCtrlr.DC(c)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json; charset=UTF-8", header.Get("Content-Type"))
}

func TestPublisherControllerMarvel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/characters?page=1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	header := c.Response().Header()

	aggs := []comic.YearlyAggregate{
		{Year: 2016, Count: 5},
		{Year: 2017, Count: 10},
	}
	apps := []comic.AppearancesByYears{
		comic.NewAppearancesByYears("test", comic.Main, aggs),
	}
	rankedChrs := []*comic.RankedCharacter{
		{ID: 1, PublisherID: 1, AvgRank: 2, AvgRankID: 1, IssueCount: 10, IssueCountRankID: 1, Name: "Test", Slug: "test", Appearances: apps},
		{ID: 2, PublisherID: 1, AvgRank: 2, AvgRankID: 2, IssueCount: 5, IssueCountRankID: 2, Name: "Test2", Slug: "test2"},
	}
	rankedSvc := mock_comic.NewMockRankedServicer(ctrl)
	rankedSvc.EXPECT().MarvelPopular(gomock.Any()).Return(rankedChrs, nil)

	publisherCtrlr := web.NewPublisherController(rankedSvc)
	err := publisherCtrlr.Marvel(c)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json; charset=UTF-8", header.Get("Content-Type"))
}

func mockCharacter() *comic.Character {
	publisher := comic.Publisher{Name: "Marvel", Slug: "marvel", ID: 1}
	return &comic.Character{
		ID:          1,
		Slug:        "emma-frost",
		Name:        "Emma-Frost",
		Description: "Blah",
		PublisherID: 1,
		Publisher:   publisher,
		IsDisabled:  false,
	}
}
