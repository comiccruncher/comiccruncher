package web_test

import (
	"testing"
	"github.com/labstack/echo"
	"net/http/httptest"
	"net/http"
	"github.com/aimeelaplant/comiccruncher/web"
	"github.com/stretchr/testify/assert"
	"errors"
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
	web.ErrorHandler(web.ErrInvalidPageParameter, c)
	assert.Equal(t, http.StatusBadRequest, c.Response().Status)

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	web.ErrorHandler(web.ErrNotFound, c)
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
