package web

import (
	"github.com/labstack/echo"
	"net/http"
)

var (
	// InternalServerMessage is a message string for an internal server error.
	InternalServerMessage = "Internal server error."
	// NotFoundMessage is a message string for when something wasn't found.
	NotFoundMessage = "The result was not found."
)

// Meta is the meta struct for an HTTP JSON response.
type Meta struct {
	StatusCode int         `json:"status_code"`
	Error      *string     `json:"error"`
	Pagination *Pagination `json:"pagination"`
}

// DetailView is a detail view for a single object.
type DetailView struct {
	Meta `json:"meta"`
	Data interface{} `json:"data"`
}

// ListView is a list view for multiple, similar objects.
type ListView struct {
	Meta `json:"meta"`
	Data []interface{} `json:"data"`
}

// JSONServerError returns a server error view.
func JSONServerError(e echo.Context) error {
	return e.JSONPretty(http.StatusInternalServerError, NewDetailViewServerError(), "  ")
}

// JSONNotFound returns a not found view.
func JSONNotFound(e echo.Context) error {
	return e.JSONPretty(http.StatusNotFound, NewDetailViewNotFound(), "  ")
}

//JSONBadRequest returns a bad request view.
func JSONBadRequest(e echo.Context, message string) error {
	return e.JSONPretty(http.StatusBadRequest, NewDetailViewBadRequest(message), "  ")
}

// JSONDetailViewOK returns a detail view.
func JSONDetailViewOK(ctx echo.Context, data interface{}) error {
	return ctx.JSONPretty(http.StatusOK, NewDetailViewOK(data), "  ")
}

// JSONListViewOK returns a list view.
func JSONListViewOK(ctx echo.Context, data []interface{}, itemsPerPage int) error {
	pagination, err := CreatePagination(ctx, data, itemsPerPage)
	if err != nil {
		return JSONBadRequest(ctx, err.Error())
	}
	if len(data) > itemsPerPage {
		return ctx.JSONPretty(http.StatusOK, NewListViewOK(data[:len(data)-1], pagination), "  ")
	}
	return ctx.JSONPretty(http.StatusOK, NewListViewOK(data, pagination), "  ")
}

// JSONDetailViewUnauthorized returns a detail view for unauthorized requests.
func JSONDetailViewUnauthorized(ctx echo.Context) error {
	return ctx.JSONPretty(http.StatusUnauthorized, NewDetailViewUnauthorized("Invalid credentials"), "  ")
}

// NewDetailViewUnauthorized creates a new detail view for unauthorized.
func NewDetailViewUnauthorized(message string) DetailView {
	dv := DetailView{}
	dv.Meta = Meta{StatusCode: http.StatusUnauthorized, Error: &message}
	return dv
}

// NewDetailViewBadRequest creates a new detail view with a bad request.
func NewDetailViewBadRequest(message string) DetailView {
	dv := DetailView{}
	dv.Meta = Meta{StatusCode: http.StatusBadRequest, Error: &message}
	return dv
}

// NewDetailViewServerError creates a new detail view with a server error.
func NewDetailViewServerError() DetailView {
	return DetailView{Meta: Meta{StatusCode: 500, Error: &InternalServerMessage}}
}

// NewDetailViewNotFound creates a new detail view with a server error.
func NewDetailViewNotFound() DetailView {
	return DetailView{Meta: Meta{StatusCode: http.StatusNotFound, Error: &NotFoundMessage}}
}

// NewDetailViewOK returns a new detail view with a 200.
func NewDetailViewOK(data interface{}) DetailView {
	return DetailView{Meta: Meta{StatusCode: http.StatusOK}, Data: data}
}

// NewListViewOK returns a new list view with 200.
func NewListViewOK(data []interface{}, pagination *Pagination) ListView {
	return ListView{Meta: Meta{StatusCode: http.StatusOK, Pagination: pagination}, Data: data}
}
