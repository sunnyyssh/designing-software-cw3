package errs

import (
	"errors"
	"fmt"
)

type HTTPError struct {
	Code    int
	Message string
}

func (e HTTPError) Error() string { return e.Message }

func BadRequest(format string, args ...any) HTTPError {
	return HTTPError{
		Code:    400,
		Message: fmt.Sprintf(format, args...),
	}
}

func NotFound(format string, args ...any) HTTPError {
	return HTTPError{
		Code:    404,
		Message: fmt.Sprintf(format, args...),
	}
}

func IsNotFound(err error) bool {
	var httpErr HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.Code == 404
	}
	return false
}
