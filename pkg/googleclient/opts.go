package googleclient

import (
	"net/url"
)

type ClientOpt func(v url.Values)
