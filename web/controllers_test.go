package web_test

import (
	"testing"
	"github.com/aimeelaplant/comiccruncher/internal/mocks/comic"
	"github.com/golang/mock/gomock"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/web"
	"github.com/labstack/echo"
	"net/http/httptest"
	"net/http"
	"github.com/stretchr/testify/assert"
	"github.com/aimeelaplant/comiccruncher/internal/mocks/search"
	"errors"
	"io/ioutil"
)

func TestStatsControllerStatsReturnsOK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	stats := comic.Stats{
		TotalCharacters: 1,
		TotalAppearances: 1,
		MinYear: 1,
		MaxYear: 10,
		TotalIssues: 1,
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

	characterCtrl := web.NewCharacterController(characterSvc)
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

	characterCtrl := web.NewCharacterController(characterSvc)
	err := characterCtrl.Character(c)

	assert.Equal(t, web.ErrNotFound.Error(), err.Error())
}

func TestCharacterControllerCharacters(t *testing.T) {
	file, err := ioutil.ReadFile("./testdata/characters.json")
	assert.Nil(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	characters := []*comic.Character{
		mockCharacter(),
		mockCharacter(),
	}
	characterSvc := mock_comic.NewMockCharacterServicer(ctrl)
	characterSvc.EXPECT().CharactersByPublisher(gomock.Any(), true, gomock.Any(), gomock.Any()).Return(characters, nil)
	characterSvc.EXPECT().ListAppearances(gomock.Any()).Return([]comic.AppearancesByYears{}, nil).Times(2)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/characters?page=1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	header := c.Response().Header()

	characterCtrl := web.NewCharacterController(characterSvc)
	// make the call
	err = characterCtrl.Characters(c)

	read, err := ioutil.ReadAll(rec.Body)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)
	assert.True(t, c.Response().Committed)
	assert.Equal(t, "application/json; charset=UTF-8", header.Get("Content-Type"))
	assert.Nil(t, err)
	assert.Equal(t, file, read)
}

func mockCharacter() *comic.Character {
	publisher := comic.Publisher{Name: "Marvel", Slug: "marvel", ID: 1}
	return &comic.Character{
		ID: 1,
		Slug:  "emma-frost",
		Name: "Emma-Frost",
		Description: "Blah",
		PublisherID: 1,
		Publisher: publisher,
		IsDisabled: false,
	}
}
