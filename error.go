package media

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/mutablelogic/go-pg"
	"github.com/mutablelogic/go-server/pkg/httpresponse"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Err uint

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	ErrBadParameter   Err = http.StatusBadRequest
	ErrNotFound       Err = http.StatusNotFound
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
	case ErrNotFound:
		return "not found"
	case ErrNotImplemented:
		return "not implemented"
	default:
		return fmt.Sprintf("error code %d", code.Code())
	}
}

func (e Err) HTTP() httpresponse.Err {
	switch e {
	case ErrNotFound:
		return httpresponse.ErrNotFound
	case ErrBadParameter:
		return httpresponse.ErrBadRequest
	case ErrInternalError:
		return httpresponse.ErrInternalError
	case ErrNotImplemented:
		return httpresponse.ErrNotImplemented
	default:
		return httpresponse.ErrInternalError
	}
}

func (code Err) With(args ...any) error {
	return fmt.Errorf("%w: %s", code, fmt.Sprint(args...))
}

func (code Err) Withf(format string, args ...any) error {
	return fmt.Errorf("%w: "+format, append([]any{code}, args...)...)
}

func HTTPErr(err error) error {
	if err == nil {
		return nil
	}

	// Check for http error
	var httpErr httpresponse.Err
	if errors.As(err, &httpErr) {
		return err
	}

	// Check for database error
	if pg.IsDatabaseError(err) {
		return pg.HTTPError(err)
	}

	// Check for gomedia error
	var schemaErr Err
	if errors.As(err, &schemaErr) {
		return schemaErr.HTTP().With(err)
	}

	// Return internal error
	return httpresponse.ErrInternalError.With(err)
}
