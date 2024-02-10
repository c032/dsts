package dsts

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"sync/atomic"
)

var (
	ErrNilStatusLineBlock = errors.New("nil status line block")
	ErrNilNotifier        = errors.New("nil notifier")
)

type I3StatusLine interface {
	AddNotifier(notifier Notifier) error
	AddStatusLineBlocks(block ...*atomic.Pointer[StatusLineBlock]) error
	AddStatusLineBlockProvider(p StatusLineBlockProvider) error

	Run() error
}

func NewI3StatusLine(ctx context.Context, w io.Writer) (I3StatusLine, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if w == nil {
		w = ioutil.Discard
	}

	i3sl := &i3StatusLine{
		ctx: ctx,
		w:   w,
	}

	return i3sl, nil
}

type i3StatusLine struct {
	ctx context.Context
	w   io.Writer

	UpdateNotifiers []Notifier

	// StatusItems contains the blocks of the status line, in the same order
	// that they will be displayed in the i3bar.
	//
	// Providers with a lower index will be displayed somewhere to the left of
	// providers with a higher index.
	StatusLineBlocks []*atomic.Pointer[StatusLineBlock]
}

func (i3sl *i3StatusLine) AddNotifier(notifier Notifier) error {
	if notifier == nil {
		return ErrNilNotifier
	}

	i3sl.UpdateNotifiers = append(i3sl.UpdateNotifiers, notifier)

	return nil
}

func (i3sl *i3StatusLine) AddStatusLineBlocks(blocks ...*atomic.Pointer[StatusLineBlock]) error {
	for _, block := range blocks {
		if block == nil {
			return ErrNilStatusLineBlock
		}
	}

	i3sl.StatusLineBlocks = append(i3sl.StatusLineBlocks, blocks...)

	return nil
}

func (i3sl *i3StatusLine) AddStatusLineBlockProvider(p StatusLineBlockProvider) error {
	notifier, slb := slbpToNotifier(i3sl.ctx, p)

	var err error

	err = i3sl.AddNotifier(notifier)
	if err != nil {
		return err
	}

	err = i3sl.AddStatusLineBlocks(slb)
	if err != nil {
		return err
	}

	return nil
}

func (i3sl *i3StatusLine) Run() error {
	ctx := i3sl.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	notifiers := i3sl.UpdateNotifiers
	mutableStatusItems := i3sl.StatusLineBlocks

	// tick is only used to notify when the status line must be updated.
	tick := make(chan struct{})

	for _, source := range notifiers {
		_ = source.OnUpdate(func() {
			tick <- struct{}{}
		})
	}

	w := i3sl.w

	w.Write([]byte(`{"version":1}[[]`))

	enc := json.NewEncoder(w)

	// visibleStatusItems is what will be serialized as JSON and printed to
	// standard output.
	visibleStatusItems := make([]StatusLineBlock, len(mutableStatusItems))

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

		w.Write([]byte{','})

		_ = enc.Encode(visibleStatusItems)
	}

	<-ctx.Done()

	err := ctx.Err()
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			return nil
		}

		cause := context.Cause(ctx)
		if cause != nil && cause != err {
			return cause
		}
	}

	return nil
}
