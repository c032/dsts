package dsts

// StatusLineBlock is one block of an i3 status line.
type StatusLineBlock struct {
	FullText string `json:"full_text"`
	Color    string `json:"color,omitempty"`
}
