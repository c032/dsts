// dsts prints status updates to standard output, as expected by i3bar.
//
// See <https://i3wm.org/docs/i3bar-protocol.html>.
package main

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/c032/dsts"
	"github.com/c032/dsts/mpd"
	"github.com/c032/dsts/tamrieltime"
)

const (
	// mpdAddr is the address where MPD is listening in.
	mpdAddr = ":6600"

	// mpdStatusMaxWidth is the maximum width, in characters, allowed for the
	// MPD status.
	mpdStatusMaxWidth = 80
)

func main() {
	// Active providers, in the same order that they will be displayed in the
	// i3bar.
	//
	// Providers with a lower index will be displayed somewhere to the left of
	// providers with a higher index.
	providers := []dsts.Provider{
		dsts.Wrap(
			mpd.Dial(mpdAddr),
			mpdStatusMaxWidth,
		),
		tamrieltime.TamrielTime,
	}

	enc := json.NewEncoder(os.Stdout)

	os.Stdout.Write([]byte(`{"version":1}[[]`))

	// tick is only used to notify when the status line must be updated.
	tick := make(chan struct{})

	// mu prevents multiple goroutines from modifying `i3sts` at the same time.
	mu := sync.RWMutex{}

	// i3sts is what will be serialized as JSON and printed to standard output.
	i3sts := make([]dsts.I3Status, len(providers))

	// Run each provider on their own goroutine, coordinate updates to the
	// i3status line, and trigger a refresh when necessary.
	for i, p := range providers {
		ch := make(chan dsts.I3Status)

		// Run the provider, usually forever.
		go p(ch)

		// For each update from the provider, update the current status and
		// trigger a refresh.
		go func(i int, p dsts.Provider) {
			for status := range ch {
				mu.Lock()
				i3sts[i] = status
				mu.Unlock()

				tick <- struct{}{}
			}
		}(i, p)
	}

	// Wait for a refresh to trigger, and update the status line.
	for range tick {
		os.Stdout.Write([]byte{','})

		mu.RLock()
		enc.Encode(i3sts)
		mu.RUnlock()
	}
}
