package web

import (
	"github.com/labstack/echo"
	"os"
	"github.com/aimeelaplant/comiccruncher/internal/log"
	"go.uber.org/zap"
	"time"
	"net/http"
)

// RequireAuthentication is a cheap, temporary authentication middleware for handling requests.
func RequireAuthentication(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		if os.Getenv("CC_ENVIRONMENT") == "development" {
			return next(ctx)
		}
		token := ctx.QueryParam("key")
		if token != "" && token == os.Getenv("CC_AUTH_TOKEN") {
			return next(ctx)
		}
		return NewJSONErrorView(ctx, "Invalid credentials", http.StatusUnauthorized)
	}
}

// ErrorHandler logs errors to the logger if there are any and sends the appropriate response back.
func ErrorHandler(err error, ctx echo.Context) {
	if err == nil {
		return
	}
	errInvalidPageStr, errNotFoundStr := ErrInvalidPageParameter.Error(), ErrNotFound.Error()
	var response error
	switch err.Error() {
	case errInvalidPageStr:
		response = NewJSONErrorView(ctx, errInvalidPageStr, http.StatusBadRequest)
		break
	case errNotFoundStr:
		response = NewJSONErrorView(ctx, errNotFoundStr, http.StatusNotFound)
		break
	default: // for ErrInternalServerError or anything else...
		// we want to log these types
		logContext(err, ctx)
		response = NewJSONErrorView(ctx, ErrInternalServerError.Error(), http.StatusInternalServerError)
		break
	}
	if response != nil {
		log.WEB().Panic("error setting response", zap.Error(err))
	}
}

// logContext logs the error and context from the request.
func logContext(err error, ctx echo.Context) {
	req := ctx.Request()
	log.WEB().Error("error",
		zap.Error(err),
		zap.Time("time", time.Now()),
		zap.String("Method", req.Method),
		zap.String("RemoteAddr", req.RemoteAddr),
		zap.String("URL", req.URL.String()))
}
