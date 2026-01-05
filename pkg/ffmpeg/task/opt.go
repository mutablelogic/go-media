package task

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Opt is a functional option for configuring the Manager
type Opt func(*opts) error

type opts struct {
	verbose        bool
	tracefn        TraceFn
	chromaprintKey string
}

type TraceFn func(v string)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func applyOpts(o *opts, opt ...Opt) error {
	for _, fn := range opt {
		if err := fn(o); err != nil {
			return err
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// WithTraceFn sets a tracing function for logging
func WithTraceFn(fn TraceFn, verbose bool) Opt {
	return func(o *opts) error {
		o.tracefn = fn
		o.verbose = verbose
		return nil
	}
}

// WithChromaprintKey sets the AcoustID API key for chromaprint operations
func WithChromaprintKey(key string) Opt {
	return func(o *opts) error {
		o.chromaprintKey = key
		return nil
	}
}
