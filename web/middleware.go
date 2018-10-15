package web

import (
	"github.com/labstack/echo"
	"os"
)

// RequireAuthentication is a cheap, temporary authentication middleware for handling requests.
func RequireAuthentication(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if os.Getenv("CC_ENVIRONMENT") == "development" {
			return next(c)
		}
		token := c.QueryParam("key")
		if token != "" && token == os.Getenv("CC_AUTH_TOKEN") {
			return next(c)
		}
		return JSONDetailViewUnauthorized(c)
	}
}
