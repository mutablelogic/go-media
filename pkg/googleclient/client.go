package googleclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	// Packages
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Name      string        `yaml:"name"`      // Path to application name
	ConfigDir string        `yaml:"configdir"` // Path to configuration path
	CacheDir  string        `yaml:"cachedir"`  // Path to cache path
	Scopes    []string      `yaml:"scopes"`    // Scopes for read, write and share
	Timeout   time.Duration `yaml:"timeout"`   // Client timeout
}

type Client struct {
	Name     string
	CacheDir string
	*oauth2.Config
	*url.URL
	*http.Client
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// Timeout requests after 15 seconds
	defaultTimeout = 15 * time.Second

	// Default perm on cache folder
	defaultCacheFileMode = os.FileMode(0750)
)

const (
	ScopeProfile = "https://www.googleapis.com/auth/userinfo.profile"
	ScopeEmail   = "https://www.googleapis.com/auth/userinfo.email"
)

var (
	DefaultConfig = Config{
		Name:    "",
		Timeout: defaultTimeout,
		Scopes:  []string{ScopeProfile, ScopeEmail},
	}
)

////////////////////////////////////////////////////////////////////////////////
// NEW

// Client with client secret
func NewClientWithClientSecret(cfg Config, path string) (*Client, error) {
	client := new(Client)

	// Set default parameters
	if cfg.Name == "" {
		client.Name = DefaultConfig.Name
	} else {
		client.Name = cfg.Name
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = DefaultConfig.Timeout
	}
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = DefaultConfig.Scopes
	}
	if cfg.ConfigDir == "" {
		if path, err := os.UserConfigDir(); err == nil {
			if cfg.Name != "" {
				path = filepath.Join(path, cfg.Name)
			}
			cfg.ConfigDir = path
		}
	}
	if cfg.CacheDir == "" {
		if path, err := os.UserCacheDir(); err == nil {
			if cfg.Name != "" {
				path = filepath.Join(path, cfg.Name)
			}
			cfg.CacheDir = path
		}
	}

	// Make configuration absolute
	if !filepath.IsAbs(path) {
		path = filepath.Join(cfg.ConfigDir, path)
	}

	// Read client secret
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Retrieve OAuth config from client secret
	if config, err := google.ConfigFromJSON(data, cfg.Scopes...); err != nil {
		return nil, err
	} else {
		client.Config = config
	}

	// Make a client
	client.Client = &(*http.DefaultClient)
	client.Client.Timeout = cfg.Timeout

	// If cache directory doesn't exist, then create it
	if stat, err := os.Stat(cfg.CacheDir); os.IsNotExist(err) {
		if err := os.MkdirAll(cfg.CacheDir, defaultCacheFileMode); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else if stat.IsDir() == false {
		return nil, fmt.Errorf("not a directory: %s", cfg.CacheDir)
	}

	// Set cache directory, which is where OAuth token is stored
	client.CacheDir = cfg.CacheDir

	// Return success
	return client, nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (client *Client) String() string {
	str := "<oauth"
	if name := client.Name; name != "" {
		str += fmt.Sprintf(" name=%q", name)
	}
	if client_id := client.Config.ClientID; client_id != "" {
		str += fmt.Sprintf(" client_id=%q", client_id)
	}
	if path := client.CacheDir; path != "" {
		str += fmt.Sprintf(" cache_dir=%q", path)
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Use a token and endpoint for communication, refresh token when necessary, etc
func (client *Client) Use(ctx context.Context, token *oauth2.Token, endpoint string) error {
	if token == nil {
		return ErrBadParameter
	}
	if url, err := url.Parse(endpoint); err != nil {
		return err
	} else {
		client.Client = client.Config.Client(ctx, token)
		client.URL = url
	}

	// Return success
	return nil
}

// Perform a GET with no body
func (client *Client) Get(path string, out interface{}, opts ...ClientOpt) error {
	// Make a new GET request
	req, err := http.NewRequest(http.MethodGet, client.URL.String(), nil)
	if err != nil {
		return err
	}

	// Do request
	return client.do(req, path, out, opts)
}

// Perform a POST with a JSON body
func (client *Client) Post(path string, in, out interface{}, opts ...ClientOpt) error {
	// Make a reader for the request object
	data := new(bytes.Buffer)
	if err := json.NewEncoder(data).Encode(in); err != nil {
		return err
	}

	// Make a new POST request with JSON payload
	req, err := http.NewRequest(http.MethodPost, client.URL.String(), data)
	if err != nil {
		return err
	} else {
		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	}

	// Do request
	return client.do(req, path, out, opts)
}

// Perform POST with a byte stream
func (client *Client) PostBinary(path string, data io.Reader, out interface{}, opts ...ClientOpt) error {
	// Make a new POST request with byte stream payload
	req, err := http.NewRequest(http.MethodPost, client.URL.String(), data)
	if err != nil {
		return err
	} else {
		req.Header.Set("Content-Type", "application/octet-stream")
	}

	// Do request
	return client.do(req, path, out, opts)
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (client *Client) do(req *http.Request, path string, out interface{}, opts []ClientOpt) error {
	// Process client options
	params := req.URL.Query()
	for _, opt := range opts {
		opt(params, req)
	}

	// Set URL path and query based on parameters and path
	req.URL.RawQuery = params.Encode()
	req.URL.Path = path

	// Perform a request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Return an error if the response is not OK
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return NewError(resp)
	}

	// Decode response
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return err
	}

	// Return success
	return nil
}
