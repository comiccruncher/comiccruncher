package web_test

import (
	"errors"
	"github.com/aimeelaplant/comiccruncher/web"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrorHandler(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	web.ErrorHandler(web.ErrInternalServerError, c)
	assert.Equal(t, http.StatusInternalServerError, c.Response().Status)

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	web.ErrorHandler(web.InvalidPageErr, c)
	assert.Equal(t, http.StatusBadRequest, c.Response().Status)

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	web.ErrorHandler(web.NewNotFoundError("Not found"), c)
	assert.Equal(t, http.StatusNotFound, c.Response().Status)

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	web.ErrorHandler(nil, c)
	assert.Equal(t, 0, c.Response().Status)

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	web.ErrorHandler(errors.New("an error"), c)
	assert.Equal(t, http.StatusInternalServerError, c.Response().Status)
}
