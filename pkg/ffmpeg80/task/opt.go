package task

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Opt is a functional option for configuring the Manager
type Opt func(*opts) error

type opts struct {
	chromaprintKey string
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// WithChromaprintKey sets the AcoustID API key for chromaprint operations
func WithChromaprintKey(key string) Opt {
	return func(o *opts) error {
		o.chromaprintKey = key
		return nil
	}
}
