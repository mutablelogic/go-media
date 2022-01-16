package googleclient

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	// Packages
	"golang.org/x/oauth2"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type DeviceAuth struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURL string `json:"verification_url"`
	ExpiresIn       int64  `json:"expires_in"`
	Interval        int64  `json:"interval"`
	r               time.Time
	client          *http.Client
	config          *oauth2.Config
}

type deviceToken struct {
	*oauth2.Token
	Error string `json:"error,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// Default Client Endpoint
	deviceCodeEndpoint  = "https://oauth2.googleapis.com/device/code"
	deviceTokenEndpoint = "https://oauth2.googleapis.com/token"
	deviceGrantType     = "urn:ietf:params:oauth:grant-type:device_code"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Start device authentication
func (client *Client) DeviceAuth() (*DeviceAuth, error) {
	var result DeviceAuth

	// Request the device code
	body := url.Values{
		"client_id": {client.ClientID},
		"scope":     {strings.Join(client.Scopes, " ")},
	}
	resp, err := client.PostForm(deviceCodeEndpoint, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, NewError(resp)
	}

	// Decode the response
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	} else {
		result.r = time.Now()
	}

	// Return success
	return &result, nil
}

// Wait for user to provide authentication response and return an oauth token
func (client *Client) DeviceToken(ctx context.Context, auth *DeviceAuth) (*oauth2.Token, error) {
	var token deviceToken

	// Continue to authenticate until expiry
	ctx, cancel := context.WithDeadline(ctx, auth.ExpiryTime())
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(auth.Interval) * time.Second):
			body := url.Values{
				"client_id":     {client.ClientID},
				"client_secret": {client.ClientSecret},
				"device_code":   {auth.DeviceCode},
				"grant_type":    {deviceGrantType},
			}
			resp, err := client.PostForm(deviceTokenEndpoint, body)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
			if resp.StatusCode == 400 || resp.StatusCode == 401 {
				return nil, NewError(resp)
			}
			if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
				return nil, err
			}
			switch token.Error {
			case "":
				return token.Token, nil
			case "slow_down":
				auth.Interval *= 2
			case "authorization_pending":
				// No-op
			default:
				return nil, ErrBadParameter.With(token.Error)
			}
		}
	}
}

func (code *DeviceAuth) ExpiryTime() time.Time {
	expiry := time.Duration(time.Second * time.Duration(code.ExpiresIn))
	return code.r.Add(expiry)
}
