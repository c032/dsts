// dsts prints status updates to standard output, as expected by i3bar.
//
// See <https://i3wm.org/docs/i3bar-protocol.html>.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync/atomic"

	"github.com/c032/dsts"
	dststime "github.com/c032/dsts/time"
)

func main() {
	ctx := context.Background()

	timeSource := &dststime.Source{}

	sources := []dsts.Source{
		timeSource,
	}

	// Active providers, in the same order that they will be displayed in the
	// i3bar.
	//
	// Providers with a lower index will be displayed somewhere to the left of
	// providers with a higher index.
	mutableStatusItems := []*atomic.Pointer[dsts.I3Status]{
		&timeSource.StatusUnix,
		&timeSource.StatusDateTime,
	}

	// tick is only used to notify when the status line must be updated.
	tick := make(chan struct{})

	for _, source := range sources {
		_ = source.OnUpdate(func() {
			tick <- struct{}{}
		})
	}

	enc := json.NewEncoder(os.Stdout)

	os.Stdout.Write([]byte(`{"version":1}[[]`))

	// visibleStatusItems is what will be serialized as JSON and printed to
	// standard output.
	visibleStatusItems := make([]dsts.I3Status, len(mutableStatusItems))

	// Wait for a refresh to trigger, and update the status line.
	for range tick {
		for i, itemPtr := range mutableStatusItems {
			if itemPtr == nil {
				continue
			}

			v := itemPtr.Load()
			if v == nil {
				continue
			}

			visibleStatusItems[i] = *v
		}

		os.Stdout.Write([]byte{','})

		_ = enc.Encode(visibleStatusItems)
	}

	<-ctx.Done()

	err := ctx.Err()
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			os.Exit(1)
		}

		cause := context.Cause(ctx)
		if cause != nil && cause != err {
			os.Exit(1)
		}
	}
}
