package time

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/c032/dsts"
)

type Source struct {
	sync.RWMutex

	Context context.Context

	ticker    *time.Ticker
	callbacks sync.Map
	nextKey   int64

	StatusUnix     atomic.Pointer[dsts.I3Status]
	StatusDateTime atomic.Pointer[dsts.I3Status]
}

func (t *Source) runCallbacks() {
	t.callbacks.Range(func(key any, value any) bool {
		const shouldContinue = true

		// We want this to panic if the conversion can't be done, because that
		// would mean there's a bug somewhere.
		callback := value.(dsts.OnUpdateCallbackFunc)

		callback()

		return shouldContinue
	})
}

func (t *Source) update(now time.Time) {
	const (
		suffix = " · "
		color  = dsts.DefaultStatusColor
	)

	t.StatusUnix.Store(&dsts.I3Status{
		FullText: fmt.Sprintf("@%d", now.Unix()),
		Color:    color,
	})

	t.StatusDateTime.Store(&dsts.I3Status{
		FullText: now.Format("2006-01-02 15:04:05") + suffix,
		Color:    color,
	})
}

func (t *Source) init() {
	if t.Context == nil {
		t.Context = context.Background()
	}

	ctx := t.Context

	if t.ticker == nil {
		t.ticker = time.NewTicker(200 * time.Millisecond)

		go func(ctx context.Context, tick <-chan time.Time) {
			firstTick := make(chan struct{})
			go func() {
				firstTick <- struct{}{}
			}()

			onTick := func() {
				now := time.Now()
				t.update(now)
				t.runCallbacks()
			}

			for {
				select {
				case <-ctx.Done():
					return
				case <-firstTick:
					onTick()
				case <-tick:
					onTick()
				}
			}
		}(ctx, t.ticker.C)
	}
}

func (t *Source) OnUpdate(callback dsts.OnUpdateCallbackFunc) dsts.RemoveOnUpdateCallbackFunc {
	t.Lock()
	defer t.Unlock()

	t.init()

	key := t.nextKey
	t.nextKey++
	t.callbacks.Store(key, callback)

	removeCallback := func() {
		t.callbacks.Delete(key)
	}

	return removeCallback
}
