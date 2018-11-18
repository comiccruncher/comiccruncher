package web

import (
	"bytes"
	"github.com/labstack/echo"
	"strconv"
)

// Page represents a page number and link.
type Page struct {
	Number int    `json:"number"`
	Link   string `json:"link"`
}

// Pagination is a view that displays pagination info.
type Pagination struct {
	PerPage      int   `json:"per_page"`
	PreviousPage string `json:"previous_page"`
	CurrentPage  string  `json:"current_page"`
	NextPage     string `json:"next_page"`
}

// CreatePagination creates a new pagination. TODO: clean this crap up.
func CreatePagination(ctx echo.Context, data []interface{}, itemsPerPage int) (*Pagination, error) {
	page, err := parsePageNumber(ctx)
	if err != nil {
		return nil, err
	}
	// Start with default
	pagination := &Pagination{
		PerPage:     itemsPerPage,
		CurrentPage: fullPath(ctx.Request().URL.EscapedPath(), ctx.QueryString()),
	}
	if page > 0 {
		pagination.CurrentPage = fullPath(ctx.Request().URL.EscapedPath(), ctx.QueryString())
		pagination.PreviousPage, err = previousPage(ctx)
		if err != nil {
			return nil, ErrInvalidPageParameter
		}
	}
	if len(data) > itemsPerPage {
		pagination.NextPage, err = nextPage(ctx)
		if err != nil {
			return nil, ErrInvalidPageParameter
		}
	}
	return pagination, nil
}

// Gets the previous page from the current page and context.
func previousPage(ctx echo.Context) (string, error) {
	pageNum, err := parsePageNumber(ctx)
	if err != nil {
		return "", err
	}
	if pageNum != 0 && pageNum > 1 {
		prevPageNum := pageNum - 1
		queryParams := ctx.QueryParams()
		queryParams.Set("page", strconv.Itoa(prevPageNum))
		return fullPath(ctx.Request().URL.EscapedPath(), queryParams.Encode()), nil
	}
	return "", nil
}

func nextPage(ctx echo.Context) (string, error) {
	pageNum, err := parsePageNumber(ctx)
	if err != nil {
		return "", err
	}
	nextPageNumber := pageNum + 1
	queryParams := ctx.QueryParams()
	queryParams.Set("page", strconv.Itoa(nextPageNumber))
	return fullPath(ctx.Request().URL.EscapedPath(), queryParams.Encode()), nil
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

func parsePageNumber(ctx echo.Context) (int, error) {
	page := ctx.QueryParam("page")
	if page == "" || page == "0" {
		page = "1"
	}
	pageNum, err := strconv.Atoi(page)
	if err != nil {
		return 0, ErrInvalidPageParameter
	}
	return pageNum, nil
}
