package task

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Manager struct {
	chromaprintKey string
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewManager(opt ...Opt) *Manager {
	m := &Manager{}

	// Apply options
	o := &opts{}
	for _, fn := range opt {
		if err := fn(o); err != nil {
			// For now, ignore errors during manager creation
			// Could be enhanced to return error in the future
			continue
		}
	}

	m.chromaprintKey = o.chromaprintKey
	return m
}
