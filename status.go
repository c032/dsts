package dsts

// I3Status is the output of a component for some specific moment in time.
type I3Status struct {
	FullText string `json:"full_text"`
	Color    string `json:"color,omitempty"`
}
