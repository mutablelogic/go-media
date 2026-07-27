package schema

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Progress reports how far a running task has got. Total is 0 if not known in
// advance, in which case a percentage can't be computed.
type Progress struct {
	Current uint64 `json:"current" help:"How far the task has got, in task-defined units (e.g. bytes, frames, streams)." example:"512"`
	Total   uint64 `json:"total,omitempty" help:"Total amount of work for the task, in the same units as current; 0 if not known in advance." example:"1024"`
}

func (p Progress) Valid() bool {
	return p.Total > 0
}

func (p Progress) Percent() float64 {
	if !p.Valid() {
		return 0
	}
	return float64(p.Current) / float64(p.Total) * 100
}
