package web_test

import (
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/web"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestNewJSONErrorView(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	j := web.NewJSONErrorView(c, "there was an error", http.StatusBadGateway)
	assert.Nil(t, j)
	assert.Equal(t, http.StatusBadGateway, rec.Result().StatusCode)

	f, err := os.Open("testdata/error502.json")
	if err != nil {
		panic(err)
	}
	b2, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(rec.Result().Body)
	assert.Nil(t, err)
	assert.Equal(t, b2, b)
}

func TestJSONDetailViewOK(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	j := web.JSONDetailViewOK(c, struct{}{})
	assert.Nil(t, j)
	assert.Equal(t, http.StatusOK, c.Response().Status)
	b, err := ioutil.ReadAll(rec.Result().Body)
	if err != nil {
		panic(err)
	}

	f, err := os.Open("testdata/detailok.json")
	if err != nil {
		panic(err)
	}
	b2, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, b2, b)
}

func TestJSONListViewOK(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	data := make([]interface{}, 10)
	j := web.JSONListViewOK(c, data, 10)

	assert.Nil(t, j)
	assert.Equal(t, http.StatusOK, c.Response().Status)
	b, err := ioutil.ReadAll(rec.Result().Body)
	if err != nil {
		panic(err)
	}

	f, err := os.Open("testdata/listok.json")
	if err != nil {
		panic(err)
	}
	b2, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, b2, b)
}

func TestNewDetailView(t *testing.T) {
	data := comic.Character{}
	view := web.NewDetailView(data, http.StatusNotModified)
	assert.NotNil(t, view)
	assert.Equal(t, http.StatusNotModified, view.StatusCode)
	assert.Equal(t, data, view.Data)
}

func TestNewDetailViewOK(t *testing.T) {
	data := comic.Character{}
	view := web.NewDetailViewOK(data)
	assert.Equal(t, view.StatusCode, http.StatusOK)
	assert.Equal(t, view.Data, data)
}

func TestNewListViewOk(t *testing.T) {
	data := make([]interface{}, 25)
	view := web.NewListViewOK(data, &web.Pagination{})
	assert.Equal(t, view.StatusCode, http.StatusOK)
	assert.Equal(t, view.Data, data)
}
