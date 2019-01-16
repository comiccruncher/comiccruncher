package web

import (
	"github.com/labstack/echo"
	"net/http"
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

// JSONDetailViewOK returns a detail view for 200 responses.
func JSONDetailViewOK(ctx echo.Context, data interface{}) error {
	return JSONDetailView(ctx, data, http.StatusOK)
}

// JSONDetailView returns a detail view.
func JSONDetailView(ctx echo.Context, data interface{}, statusCode int) error {
	return ctx.JSONPretty(statusCode, NewDetailView(data, statusCode), "  ")
}

// JSONListViewOK returns a list view.
func JSONListViewOK(ctx echo.Context, data []interface{}, itemsPerPage int) error {
	pagination, err := CreatePagination(ctx, data, itemsPerPage)
	if err != nil {
		return err
	}
	if len(data) > itemsPerPage {
		return ctx.JSONPretty(http.StatusOK, NewListViewOK(data[:len(data)-1], pagination), "  ")
	}
	return ctx.JSONPretty(http.StatusOK, NewListViewOK(data, pagination), "  ")
}

// NewJSONErrorView returns a new view JSON view with an error message and status code.
func NewJSONErrorView(ctx echo.Context, err string, statusCode int) error {
	return ctx.JSONPretty(statusCode, DetailView{Meta: Meta{StatusCode: statusCode, Error: &err}}, "  ")
}

// NewDetailViewOK returns a new detail view with a 200.
func NewDetailViewOK(data interface{}) DetailView {
	return NewDetailView(data, http.StatusOK)
}

// NewDetailViewOK returns a new detail view with a 200.
func NewDetailView(data interface{}, statusCode int) DetailView {
	return DetailView{Meta: Meta{StatusCode: statusCode}, Data: data}
}

// NewListViewOK returns a new list view with 200.
func NewListViewOK(data []interface{}, pagination *Pagination) ListView {
	return ListView{Meta: Meta{StatusCode: http.StatusOK, Pagination: pagination}, Data: data}
}
