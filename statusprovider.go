package dsts

import (
	"context"
	"sync/atomic"
)

type StatusProviderFunc func(ctx context.Context, ch chan<- I3Status) error

func StatusProvider(ctx context.Context, p StatusProviderFunc) (Source, *atomic.Pointer[I3Status]) {
	statusProvider := &atomic.Pointer[I3Status]{}

	source := sourceOnUpdateFunc(func(callback OnUpdateCallbackFunc) RemoveOnUpdateCallbackFunc {
		ch := make(chan I3Status)

		ctxProvider, cancel := context.WithCancelCause(ctx)

		go func(ctxProvider context.Context, ch chan<- I3Status) {
			err := p(ctxProvider, ch)
			if err != nil {
				cancel(err)
			}

			cancel(nil)
		}(ctxProvider, ch)

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case status := <-ch:
					statusProvider.Store(&status)
					callback()
				}
			}
		}()

		remove := func() {
			cancel(nil)
		}

		return remove
	})

	return source, statusProvider
}
