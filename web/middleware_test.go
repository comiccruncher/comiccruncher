package web_test

import (
	"errors"
	"github.com/comiccruncher/comiccruncher/web"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
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

func TestNewDefaultJWTMiddleware(t *testing.T) {
	m := web.NewDefaultJWTMiddleware()
	assert.NotNil(t, m)
}

func TestNewJWTConfigFromEnvironment(t *testing.T) {
	m := web.NewJWTConfigFromEnvironment()
	assert.NotNil(t, m)
	assert.Equal(t, os.Getenv("CC_JWT_SIGNING_SECRET"), m.SecretSigningKey)
}

func TestJWTMiddlewareWithConfigNoHeader(t *testing.T) {
	m := web.JWTMiddlewareWithConfig(web.NewJWTConfigFromEnvironment())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	e := echo.New()
	ctx := e.NewContext(req, rec)

	result := m(ctx.Handler())
	err := result(ctx)

	assert.NotNil(t, err)
	assert.Equal(t, echo.ErrUnauthorized, err)
}

func TestJWTMiddlewareWithConfigWithHeader(t *testing.T) {
	m := web.JWTMiddlewareWithConfig(web.NewJWTConfigFromEnvironment())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// test token
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiJhOTIzYjg4MC0yNWIwLTQ5ZjMtODhjMC1kNjI5NDM4YzFjYjIiLCJwdWJsaWMiOnRydWV9.zLhVMc0fsOQIt2EehDr0aN1BJjsvw5BHgXZkwXaWSfQ")
	rec := httptest.NewRecorder()
	e := echo.New()
	ctx := e.NewContext(req, rec)

	result := m(ctx.Handler())
	err := result(ctx)

	assert.NotNil(t, err)
	//  should return not found to go to next func.
	assert.Equal(t, echo.ErrNotFound, err)
}

func TestJWTMiddlewareWithConfigWithBadHeader(t *testing.T) {
	m := web.JWTMiddlewareWithConfig(web.NewJWTConfigFromEnvironment())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer test")
	rec := httptest.NewRecorder()
	e := echo.New()
	ctx := e.NewContext(req, rec)

	result := m(ctx.Handler())
	err := result(ctx)

	assert.NotNil(t, err)
	assert.Equal(t, echo.ErrUnauthorized, err)
}

func TestRequireCheapAuthenticationPass(t *testing.T) {
	m := web.RequireCheapAuthentication
	req := httptest.NewRequest(http.MethodGet, "/?key="+os.Getenv("CC_AUTH_TOKEN"), nil)
	rec := httptest.NewRecorder()
	e := echo.New()
	ctx := e.NewContext(req, rec)

	result := m(ctx.Handler())
	err := result(ctx)
	assert.NotNil(t, err)
	// it passes, so 404
	assert.Equal(t, echo.ErrNotFound, err)
}

func TestRequireCheapAuthenticationFails(t *testing.T) {
	m := web.RequireCheapAuthentication
	req := httptest.NewRequest(http.MethodGet, "/?key=blah", nil)
	rec := httptest.NewRecorder()
	e := echo.New()
	ctx := e.NewContext(req, rec)

	result := m(ctx.Handler())
	err := result(ctx)

	response := ctx.Response()
	defer response.Flush()
	if os.Getenv("CC_ENVIRONMENT") == "development" {
		assert.Equal(t, 0, response.Status)
	} else {
		assert.Nil(t, err)
		assert.Equal(t, http.StatusUnauthorized, response.Status)
	}
}
