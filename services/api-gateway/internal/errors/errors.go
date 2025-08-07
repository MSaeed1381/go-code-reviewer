package errors

type HttpError struct {
	IsUserError bool
	Description string
	StatusCode  int
}

func (e *HttpError) Error() string {
	return e.Description
}
