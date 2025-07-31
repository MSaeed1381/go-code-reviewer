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
	ErrNoResponseChoice = &HttpError{
		IsUserError: false,
		Description: "no response choices from LLM",
		StatusCode:  http.StatusInternalServerError,
	}
	ErrUnknownIntent = &HttpError{
		IsUserError: false,
		Description: "unknown intent",
		StatusCode:  http.StatusInternalServerError,
	}
	ErrNoSnippetFound = &HttpError{
		IsUserError: true,
		Description: "no snippet found",
		StatusCode:  http.StatusBadRequest,
	}
)
