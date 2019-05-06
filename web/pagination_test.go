package web_test

import (
	"github.com/comiccruncher/comiccruncher/web"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreatePaginationDataLessThanItemsPerPage(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/characters?page=1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	data := make([]interface{}, 5)
	for k, v := range data {
		data[k] = v
	}

	p, err := web.CreatePagination(c, data, 6)

	assert.Nil(t, err)
	assert.Equal(t, "", p.PreviousPage)
	assert.Equal(t, "/characters?page=1", p.CurrentPage)
	assert.Equal(t, "", p.NextPage)
	assert.Equal(t, 6, p.PerPage)
}

func TestCreatePaginationDataMoreThanItemsPerPage(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/characters?page=1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	data := make([]interface{}, 5)
	for k, v := range data {
		data[k] = v
	}

	p, err := web.CreatePagination(c, data, 4)

	assert.Nil(t, err)
	assert.Equal(t, "", p.PreviousPage)
	assert.Equal(t, "/characters?page=1", p.CurrentPage)
	assert.Equal(t, "/characters?page=2", p.NextPage)
	assert.Equal(t, 4, p.PerPage)
}

func TestCreatePaginationPreviousPageAndNextPage(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/characters?page=5&type=main", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	data := make([]interface{}, 25)
	for k, v := range data {
		data[k] = v
	}

	p, err := web.CreatePagination(c, data, 20)

	assert.Nil(t, err)
	assert.Equal(t, "/characters?page=4&type=main", p.PreviousPage)
	assert.Equal(t, "/characters?page=5&type=main", p.CurrentPage)
	assert.Equal(t, "/characters?page=6&type=main", p.NextPage)
	assert.Equal(t, 20, p.PerPage)
}

func TestCreatePaginationPreviousPage(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/characters?page=5&type=main", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	data := make([]interface{}, 20)
	for k, v := range data {
		data[k] = v
	}

	p, err := web.CreatePagination(c, data, 20)

	assert.Nil(t, err)
	assert.Equal(t, "/characters?page=4&type=main", p.PreviousPage)
	assert.Equal(t, "/characters?page=5&type=main", p.CurrentPage)
	assert.Equal(t, "", p.NextPage)
	assert.Equal(t, 20, p.PerPage)
}

func TestCreatePaginationBadParameter(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/characters?page=xx", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	data := make([]interface{}, 0)
	_, err := web.CreatePagination(c, data, 20)
	assert.Error(t, err)
}
