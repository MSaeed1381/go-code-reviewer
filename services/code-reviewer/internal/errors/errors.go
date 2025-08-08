package errors

import "net/http"

type HttpError struct {
	IsUserError bool
	Description string
	StatusCode  int
}

func (e *HttpError) Error() string {
	return e.Description
}

var (
	ErrNoSnippetFound = &HttpError{
		IsUserError: true,
		Description: "no snippet found",
		StatusCode:  http.StatusBadRequest,
	}
)
