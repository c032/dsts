package dsts

import (
	"context"
	"sync/atomic"
)

// StatusLineBlockProvider is a simplified way of producing status line blocks.
//
// The function can run a loop and pass any updates to `ch`. It's meant to be
// called in its own goroutine, so it can block as long as needed, returning
// only when there's an error or when `ctx` is done.
type StatusLineBlockProvider func(ctx context.Context, ch chan<- StatusLineBlock) error

func slbpToNotifier(ctx context.Context, p StatusLineBlockProvider) (Notifier, *atomic.Pointer[StatusLineBlock]) {
	statusProvider := &atomic.Pointer[StatusLineBlock]{}

	source := NotifierFunc(func(callback OnUpdateCallbackFunc) RemoveOnUpdateCallbackFunc {
		ch := make(chan StatusLineBlock)

		ctxProvider, cancel := context.WithCancelCause(ctx)

		go func(ctxProvider context.Context, ch chan<- StatusLineBlock) {
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
