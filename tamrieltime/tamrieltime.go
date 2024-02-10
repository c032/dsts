// Package tamrieltime implements functions for displaying current time and
// date, formatted as in Skyrim.
//
// See <https://en.uesp.net/wiki/Lore:Calendar>.
package tamrieltime

import (
	"context"
	"fmt"
	"time"

	"github.com/c032/dsts"
)

var weekdays = map[time.Weekday]string{
	time.Sunday:    "Sundas",
	time.Monday:    "Morndas",
	time.Tuesday:   "Tirdas",
	time.Wednesday: "Middas",
	time.Thursday:  "Turdas",
	time.Friday:    "Fredas",
	time.Saturday:  "Loredas",
}

var months = map[time.Month]string{
	time.January:   "Morning Star",
	time.February:  "Sun's Dawn",
	time.March:     "First Seed",
	time.April:     "Rain's Hand",
	time.May:       "Second Seed",
	time.June:      "Mid Year",
	time.July:      "Sun's Height",
	time.August:    "Last Seed",
	time.September: "Heartfire",
	time.October:   "Frostfall",
	time.November:  "Sun's Dusk",
	time.December:  "Evening Star",
}

var suffixes = map[int]string{
	1:  "st",
	2:  "nd",
	3:  "rd",
	21: "st",
	22: "nd",
	23: "rd",
	31: "st",
}

// Format formats `time.Time` using the same format as Skyrim.
func Format(t time.Time) string {
	timeStr := t.Format("03:04:05 PM")

	day := t.Day()
	daySuffix := "th"

	_, ok := suffixes[day]
	if ok {
		daySuffix = suffixes[day]
	}

	return fmt.Sprintf("%s, %s, %d%s of %s", weekdays[t.Weekday()], timeStr, day, daySuffix, months[t.Month()])
}

var _ dsts.StatusProviderFunc = TamrielTime

// TamrielTime is a `dsts.StatusProviderFunc` for displaying the current date
// and time in the format used by Skyrim.
func TamrielTime(ctx context.Context, ch chan<- dsts.I3Status) error {
	firstTick := make(chan struct{})
	go func() {
		firstTick <- struct{}{}
	}()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-firstTick:
			ch <- dsts.I3Status{
				FullText: Format(time.Now()),
				Color:    "#999999",
			}
		case <-time.After(500 * time.Millisecond):
			ch <- dsts.I3Status{
				FullText: Format(time.Now()),
				Color:    "#999999",
			}
		}
	}
}
