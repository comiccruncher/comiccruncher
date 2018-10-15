package web

import (
	"bytes"
	"errors"
	"github.com/labstack/echo"
	"strconv"
)

var (
	// ErrInvalidPageParameter is an error for an invalid page parameter from a request.
	ErrInvalidPageParameter = errors.New("invalid `page` parameter")
)

// Page represents a page number and link.
type Page struct {
	Number int    `json:"number"`
	Link   string `json:"link"`
}

// Gets the previous page from the current page and context.
func (p Page) previousPage(ctx echo.Context) *Page {
	if ctx.QueryParam("page") != "" && p.Number > 1 {
		prevPageNumber := p.Number - 1
		queryParams := ctx.QueryParams()
		queryParams.Set("page", strconv.Itoa(prevPageNumber))
		return &Page{Number: prevPageNumber, Link: fullPath(ctx.Request().URL.EscapedPath(), queryParams.Encode())}
	}
	return nil
}

// Gets the next page from the current page and context.
func (p Page) nextPage(ctx echo.Context) *Page {
	nextPageNumber := p.Number + 1
	queryParams := ctx.QueryParams()
	queryParams.Set("page", strconv.Itoa(nextPageNumber))
	return &Page{Number: nextPageNumber, Link: fullPath(ctx.Request().URL.EscapedPath(), queryParams.Encode())}
}

// Pagination is a view that displays pagination info.
type Pagination struct {
	PerPage      int   `json:"per_page"`
	PreviousPage *Page `json:"previous_page"`
	CurrentPage  Page  `json:"current_page"`
	NextPage     *Page `json:"next_page"`
}

// CreatePagination creates a new pagination.
func CreatePagination(ctx echo.Context, data []interface{}, itemsPerPage int) (*Pagination, error) {
	requestedPageParam := ctx.QueryParam("page")
	// Start with default
	pagination := &Pagination{
		PerPage:     itemsPerPage,
		CurrentPage: Page{Link: fullPath(ctx.Request().URL.EscapedPath(), ctx.QueryString()), Number: 1},
	}
	requestedPageNumber, err := strconv.Atoi(requestedPageParam)
	if requestedPageParam != "" && err != nil {
		return nil, ErrInvalidPageParameter
	}
	if requestedPageParam != "" {
		if requestedPageNumber > 0 {
			pagination.CurrentPage = Page{Link: fullPath(ctx.Request().URL.EscapedPath(), ctx.QueryString()), Number: requestedPageNumber}
			pagination.PreviousPage = pagination.CurrentPage.previousPage(ctx)
		} else {
			return nil, ErrInvalidPageParameter
		}
	}
	if len(data) > itemsPerPage {
		pagination.NextPage = pagination.CurrentPage.nextPage(ctx)
	}

	return pagination, nil
}

// Returns the full path given the path and query string.
func fullPath(path, querystring string) string {
	var buffer bytes.Buffer
	buffer.WriteString(path)
	if querystring != "" {
		buffer.WriteString("?")
	}
	buffer.WriteString(querystring)
	return buffer.String()
}
