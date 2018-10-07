package api

import (
	"github.com/labstack/echo"
	"net/http"
)

var (
	InternalServerMessage = "Internal server error."
	NotFoundMessage       = "The result was not found."
)

// The meta struct for an HTTP JSON response.
type Meta struct {
	StatusCode int         `json:"status_code"`
	Error      *string     `json:"error"`
	Pagination *Pagination `json:"pagination"`
}

// A detail view for a single object.
type DetailView struct {
	Meta `json:"meta"`
	Data interface{} `json:"data"`
}

// A list view for multiple objects.
type ListView struct {
	Meta `json:"meta"`
	Data []interface{} `json:"data"`
}

// Returns a server error view.
func JSONServerError(e echo.Context) error {
	return e.JSONPretty(http.StatusInternalServerError, NewDetailViewServerError(), "  ")
}

// Returns a not found view.
func JSONNotFound(e echo.Context) error {
	return e.JSONPretty(http.StatusNotFound, NewDetailViewNotFound(), "  ")
}

// Returns a bad request view.
func JSONBadRequest(e echo.Context, message string) error {
	return e.JSONPretty(http.StatusBadRequest, NewDetailViewBadRequest(message), "  ")
}

// Returns a detail view.
func JSONDetailViewOK(ctx echo.Context, data interface{}) error {
	return ctx.JSONPretty(http.StatusOK, NewDetailViewOK(data), "  ")
}

// Returns a list view.
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

// Creates a new detail view with a bad request.
func NewDetailViewBadRequest(message string) DetailView {
	dv := DetailView{}
	dv.Meta = Meta{StatusCode: http.StatusBadRequest, Error: &message}
	return dv
}

// Creates a new detail view with a server error.
func NewDetailViewServerError() DetailView {
	return DetailView{Meta: Meta{StatusCode: 500, Error: &InternalServerMessage}}
}

// Creates a new detail view with a server error.
func NewDetailViewNotFound() DetailView {
	return DetailView{Meta: Meta{StatusCode: http.StatusNotFound, Error: &NotFoundMessage}}
}

// Returns a new detail view with a 200.
func NewDetailViewOK(data interface{}) DetailView {
	return DetailView{Meta: Meta{StatusCode: http.StatusOK}, Data: data}
}

// Returns a new list view with 200.
func NewListViewOK(data []interface{}, pagination *Pagination) ListView {
	return ListView{Meta: Meta{StatusCode: http.StatusOK, Pagination: pagination}, Data: data}
}
