package web

import (
	"errors"
	"github.com/comiccruncher/comiccruncher/internal/log"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"go.uber.org/zap"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	// ErrInternalServerError is for when something bad happens internally.
	ErrInternalServerError = errors.New("internal server error")
)

// JWTConfig is a struct for configuration.
type JWTConfig struct {
	SecretSigningKey string
}

// JWTMiddlewareWithConfig creates a new middleware func from the specified configuration.
func JWTMiddlewareWithConfig(config JWTConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			_, err := validateToken(c.Request(), config.SecretSigningKey)
			if err != nil {
				return err
			}
			return next(c)
		}
	}
}

// RequireCheapAuthentication is a cheap, temporary authentication middleware for handling requests.
func RequireCheapAuthentication(next echo.HandlerFunc) echo.HandlerFunc {
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
	echoErr, ok := err.(*echo.HTTPError)
	if !ok {
		logContext(err, ctx)
		NewJSONErrorView(ctx, ErrInternalServerError.Error(), http.StatusInternalServerError)
		return
	}
	if echoErr.Code != http.StatusNotFound && echoErr.Code != http.StatusBadRequest {
		logContext(err, ctx)
	}
	NewJSONErrorView(ctx, echoErr.Message.(string), echoErr.Code)
}

// NewJWTConfigFromEnvironment creates a new configuration struct from environment variables.
func NewJWTConfigFromEnvironment() JWTConfig {
	return JWTConfig{
		SecretSigningKey: os.Getenv("CC_JWT_SIGNING_SECRET"),
	}
}

// NewDefaultJWTMiddleware creates the default JWT middleware with the default config.
func NewDefaultJWTMiddleware() echo.MiddlewareFunc {
	return JWTMiddlewareWithConfig(NewJWTConfigFromEnvironment())
}

// logContext logs the error and context from the request.
func logContext(err error, ctx echo.Context) {
	req := ctx.Request()
	log.WEB().Error("error",
		zap.Error(err),
		zap.Time("time", time.Now()),
		zap.String("Method", req.Method),
		zap.String("URL", req.URL.String()),
		zap.String("Host", req.Host),
		zap.String("RemoteAddr", req.RemoteAddr),
		zap.String("RealIP", ctx.RealIP()),
		zap.String("User Agent", req.UserAgent()),
		zap.String("Referer", req.Referer()),
		zap.String("X-VISITOR-ID", req.Header.Get("X-VISITOR-ID")),
	)
}

func validateToken(req *http.Request, secret string) (map[string]interface{}, error) {
	val, err := parseAuthorizationBearer(req.Header)
	if err != nil {
		return nil, err
	}
	token, err := jwt.Parse(val, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		log.WEB().Error("error parsing token", zap.Error(err))
		return nil, echo.ErrUnauthorized
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, echo.ErrUnauthorized
}

func parseAuthorizationBearer(h http.Header) (string, error) {
	val := h.Get("Authorization")
	if val == "" {
		return "", echo.ErrUnauthorized
	}
	tokenString := strings.SplitAfter(val, "Bearer ")
	if len(tokenString) < 2 {
		return "", echo.ErrUnauthorized
	}
	return tokenString[1], nil
}
