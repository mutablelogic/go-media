package media

import (
	"fmt"
	"net/http"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Err uint

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	ErrBadParameter   Err = http.StatusBadRequest
	ErrInternalError  Err = http.StatusInternalServerError
	ErrNotImplemented Err = http.StatusNotImplemented
)

///////////////////////////////////////////////////////////////////////////////
// ERROR

func (code Err) Code() uint {
	return uint(code)
}

func (code Err) Error() string {
	switch code {
	case ErrBadParameter:
		return "bad parameter"
	case ErrInternalError:
		return "internal error"
	case ErrNotImplemented:
		return "not implemented"
	default:
		return fmt.Sprintf("error code %d", code.Code())
	}
}

func (code Err) With(args ...interface{}) error {
	return fmt.Errorf("%w: %s", code, fmt.Sprint(args...))
}

func (code Err) Withf(format string, args ...interface{}) error {
	return fmt.Errorf("%w: %s", code, fmt.Sprintf(format, args...))
}
