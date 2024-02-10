package dsts

import (
	"errors"
)

var ErrInvalidColor = errors.New("invalid color")

const (
	DefaultStatusColor      = "#999999"
	DefaultStatusColorError = "#e20024"
)

func isNumber(c rune) bool {
	return c >= '0' && c <= '9'
}

func IsValidColor(color string) bool {
	if len(color) != len("#000") && len(color) != len("#000000") {
		return false
	}
	if color[0] != '#' {
		return false
	}

	for i, c := range color {
		if i == 0 {
			continue
		}
		if !isNumber(c) {
			return false
		}
	}

	return true
}
