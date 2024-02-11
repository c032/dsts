package time

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/c032/dsts"
)

var _ dsts.Notifier = (*Notifier)(nil)

type Notifier struct {
	sync.Mutex

	Context context.Context

	ticker    *time.Ticker
	callbacks sync.Map
	nextKey   int64

	StatusUnix     atomic.Pointer[dsts.StatusLineBlock]
	StatusDateTime atomic.Pointer[dsts.StatusLineBlock]
}

func (tn *Notifier) runCallbacks() {
	tn.callbacks.Range(func(key any, value any) bool {
		const shouldContinue = true

		// We want this to panic if the conversion can't be done, because that
		// would mean there's a bug somewhere.
		callback := value.(dsts.NotifierCallbackFunc)

		callback()

		return shouldContinue
	})
}

func (tn *Notifier) update(now time.Time) {
	const (
		suffix = " · "
		color  = dsts.DefaultStatusColor
	)

	tn.StatusUnix.Store(&dsts.StatusLineBlock{
		FullText: fmt.Sprintf("@%d", now.Unix()),
		Color:    color,
	})

	tn.StatusDateTime.Store(&dsts.StatusLineBlock{
		FullText: now.Format("2006-01-02 15:04:05") + suffix,
		Color:    color,
	})
}

func (tn *Notifier) init() {
	if tn.Context == nil {
		tn.Context = context.Background()
	}

	ctx := tn.Context

	if tn.ticker == nil {
		tn.ticker = time.NewTicker(200 * time.Millisecond)

		go func(ctx context.Context, tick <-chan time.Time) {
			firstTick := make(chan struct{})
			go func() {
				firstTick <- struct{}{}
			}()

			onTick := func() {
				now := time.Now()
				tn.update(now)
				tn.runCallbacks()
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
		}(ctx, tn.ticker.C)
	}
}

func (tn *Notifier) OnUpdate(callback dsts.NotifierCallbackFunc) dsts.RemoveCallbackFunc {
	tn.Lock()
	defer tn.Unlock()

	tn.init()

	key := tn.nextKey
	tn.nextKey++
	tn.callbacks.Store(key, callback)

	removeCallback := func() {
		tn.callbacks.Delete(key)
	}

	return removeCallback
}