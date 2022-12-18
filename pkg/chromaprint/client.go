package chromaprint

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Key     string        `yaml:"key"`     // AcuostId Web Service Key
	Timeout time.Duration `yaml:"timeout"` // AcoustId Client timeout
	Rate    uint          `yaml:"rate"`    // Maximum requests per second
	Base    string        `yaml:"base"`    // Base URL
}

type Client struct {
	Config
	*http.Client
	*url.URL
	last time.Time
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// Register a clientId: https://acoustid.org/login
	defaultClientId = "-YectPtndAM"

	// Timeout requests after 15 seconds
	defaultTimeout = 15 * time.Second

	// The API endpoint
	baseUrl = "https://api.acoustid.org/v2"

	// defaultQps rate limits number of requests per second
	defaultQps = 3
)

var (
	ErrQueryRateExceeded = errors.New("query rate exceeded")
)

var (
	DefaultConfig = Config{
		Key:     defaultClientId,
		Timeout: defaultTimeout,
		Rate:    defaultQps,
		Base:    baseUrl,
	}
)

////////////////////////////////////////////////////////////////////////////////
// NEW

func NewClient(key string) *Client {
	config := DefaultConfig
	if key != "" {
		config.Key = os.ExpandEnv(key)
	}
	if client, err := NewClientWithConfig(config); err != nil {
		return nil
	} else {
		return client
	}
}

func NewClientWithConfig(cfg Config) (*Client, error) {
	client := new(Client)
	client.Config = cfg

	// Set parameters
	if client.Config.Timeout == 0 {
		client.Config.Timeout = DefaultConfig.Timeout
	}
	if client.Key == "" {
		client.Key = os.ExpandEnv(DefaultConfig.Key)
	}
	if client.Base == "" {
		client.Base = DefaultConfig.Base
	}
	if client.Rate == 0 {
		client.Rate = DefaultConfig.Rate
	}

	// Create HTTP client
	client.Client = &http.Client{
		Timeout: client.Config.Timeout,
	}
	url, err := url.Parse(client.Base)
	if err != nil {
		return nil, err
	} else {
		client.URL = url
	}
	// Fudge key into URL
	v := url.Query()
	v.Set("client", client.Key)
	url.RawQuery = v.Encode()

	// Return success
	return client, nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (client *Client) String() string {
	str := "<chromaprint"
	if u := client.URL; u != nil {
		str += fmt.Sprintf(" url=%q", u.String())
	}
	if rate := client.Rate; rate > 0 {
		str += fmt.Sprintf(" rate=%dops/s", rate)
	}
	if timeout := client.Client.Timeout; timeout > 0 {
		str += fmt.Sprintf(" timeout=%v", timeout)
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// LOOKUP

func (client *Client) Lookup(fingerprint string, duration time.Duration, flags Meta) ([]*ResponseMatch, error) {
	// Check incoming parameters
	if fingerprint == "" || duration == 0 || flags == META_NONE {
		return nil, ErrBadParameter.With("Lookup")
	}

	// Check Qps
	if client.last.IsZero() {
		if time.Since(client.last) < (time.Second / defaultQps) {
			return nil, ErrQueryRateExceeded
		}
	}

	// Set URL parameters
	params := url.Values{}
	params.Set("fingerprint", fingerprint)
	params.Set("duration", fmt.Sprint(duration.Truncate(time.Second).Seconds()))
	params.Set("meta", flags.String())
	url := client.requestUrl("lookup", params)
	if url == nil {
		return nil, ErrBadParameter.With("Lookup")
	}

	//fmt.Println(url.String())

	// Perform request
	now := time.Now()
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}
	response, err := client.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// Decode response body
	var r Response
	if mimeType, _, err := mime.ParseMediaType(response.Header.Get("Content-type")); err != nil {
		return nil, ErrUnexpectedResponse.With(err)
	} else if mimeType != "application/json" {
		return nil, ErrUnexpectedResponse.With(mimeType)
	} else if err := json.Unmarshal(body, &r); err != nil {
		return nil, ErrUnexpectedResponse.With(err)
	}

	// Check for errors
	if r.Status != "ok" {
		return nil, ErrBadParameter.Withf("%v (code %v)", r.Error.Message, r.Error.Code)
	} else if response.StatusCode != http.StatusOK {
		return nil, ErrBadParameter.Withf("%v (code %v)", response.Status, response.StatusCode)
	}

	// Set response time for calculating qps
	client.last = now

	// Return success
	return r.Results, nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (client *Client) requestUrl(path string, v url.Values) *url.URL {
	url, err := url.Parse(client.URL.String())
	if err != nil {
		return nil
	}
	// Copy params
	params := client.URL.Query()
	for k := range v {
		params[k] = v[k]
	}
	url.RawQuery = params.Encode()

	// Set path
	url.Path = filepath.Join(url.Path, path)

	// Return URL
	return url
}
