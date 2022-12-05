package googleclient

import (
	"encoding/json"
	"os"
	"path/filepath"

	// Packages
	"golang.org/x/oauth2"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	tokenPath = "token.json"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (client *Client) ReadToken() (*oauth2.Token, error) {
	var result oauth2.Token

	path := filepath.Join(client.CacheDir, tokenPath)
	if stat, err := os.Stat(path); os.IsNotExist(err) {
		// Return nil token if one doesn't exist, but no error
		return nil, nil
	} else if err != nil {
		return nil, err
	} else if stat.Mode().IsRegular() == false {
		return nil, ErrUnexpectedResponse
	}

	// Read token
	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	if err := json.NewDecoder(r).Decode(&result); err != nil {
		return nil, err
	}

	// Return success
	return &result, nil
}

func (client *Client) WriteToken(token *oauth2.Token) error {
	path := filepath.Join(client.CacheDir, tokenPath)

	// Write token file
	w, err := os.Create(path)
	if err != nil {
		return err
	}
	defer w.Close()
	if err := json.NewEncoder(w).Encode(token); err != nil {
		return err
	}

	// Return success
	return nil
}
