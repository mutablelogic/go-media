package googleclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type APIError struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type OAuthError struct {
	Status  string `json:"error"`
	Message string `json:"error_description"`
}

type APIErrors struct {
	APIError `json:"error"`
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewError(resp *http.Response) error {
	var apierror APIErrors
	var oautherror OAuthError

	// Decode any error in the body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return APIError{Code: resp.StatusCode, Message: resp.Status}
	}
	if err := json.Unmarshal(body, &apierror); err == nil {
		return apierror.APIError
	}
	if err := json.Unmarshal(body, &oautherror); err == nil {
		return oautherror
	}
	return ErrUnexpectedResponse.With(resp.Status)
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (e APIError) Error() string {
	if e.Status != "" && e.Message != "" {
		return fmt.Sprint(e.Status, ": ", e.Message)
	}
	if e.Status != "" {
		return e.Status
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Code != 0 {
		return fmt.Sprint("Code ", e.Code)
	}
	return ErrUnexpectedResponse.Error()
}

func (e OAuthError) Error() string {
	if e.Status != "" && e.Message != "" {
		return fmt.Sprint(e.Status, ": ", e.Message)
	}
	if e.Status != "" {
		return e.Status
	}
	if e.Message != "" {
		return e.Message
	}
	return ErrUnexpectedResponse.Error()
}
