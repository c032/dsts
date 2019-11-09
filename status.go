package dsts

// Provider is a function that pushes updates to `ch`.
//
// It's normal for a Provider to block forever, e.g. due to an infinite loop.
type Provider func(ch chan<- I3Status)

// I3Status is the output of a component for some specific moment in time.
type I3Status struct {
	FullText string `json:"full_text"`
	Color    string `json:"color,omitempty"`
}
