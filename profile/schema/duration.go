package schema

import (
	"encoding/json"
	"time"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Duration is a time.Duration that marshals to/from its String() form (e.g.
// "1h2m3s") in JSON, rather than time.Duration's default of a raw count of
// nanoseconds.
type Duration time.Duration

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (d Duration) String() string {
	return time.Duration(d).String()
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - MARSHALING

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" {
		*d = 0
		return nil
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(parsed)
	return nil
}
