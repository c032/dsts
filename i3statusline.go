package dsts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

	AppendStatusLineBlocks(block ...*atomic.Pointer[StatusLineBlock]) error
	AppendStatusLineBlockProvider(p StatusLineBlockProvider) error

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

	// StatusLineBlocks contains the blocks of the status line, in the same
	// order that they will be displayed in the i3bar.
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

func (i3sl *i3StatusLine) AppendStatusLineBlocks(blocks ...*atomic.Pointer[StatusLineBlock]) error {
	for _, block := range blocks {
		if block == nil {
			return ErrNilStatusLineBlock
		}
	}

	i3sl.StatusLineBlocks = append(i3sl.StatusLineBlocks, blocks...)

	return nil
}

func (i3sl *i3StatusLine) AppendStatusLineBlockProvider(p StatusLineBlockProvider) error {
	notifier, slb := slbpToNotifier(i3sl.ctx, p)

	var err error

	err = i3sl.AddNotifier(notifier)
	if err != nil {
		return err
	}

	err = i3sl.AppendStatusLineBlocks(slb)
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
	mutableBlocks := i3sl.StatusLineBlocks

	// tick is only used to notify when the status line must be updated.
	tick := make(chan struct{})

	for _, notifier := range notifiers {
		unsub := notifier.OnUpdate(func() {
			tick <- struct{}{}
		})

		// Ensure `unsub` is called when the function returns.
		defer unsub()
	}

	w := i3sl.w

	w.Write([]byte(`{"version":1}[[]`))

	enc := json.NewEncoder(w)

	// visibleBlocks is what will be serialized as JSON and printed to the
	// writer.
	visibleBlocks := make([]StatusLineBlock, len(mutableBlocks))

	// Wait for a refresh to trigger, and update the status line.
	for range tick {
		for i, itemPtr := range mutableBlocks {
			if itemPtr == nil {
				continue
			}

			v := itemPtr.Load()
			if v == nil {
				continue
			}

			visibleBlocks[i] = *v
		}

		w.Write([]byte{','})

		_ = enc.Encode(visibleBlocks)
	}

	<-ctx.Done()

	err := ctx.Err()
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			return fmt.Errorf("unexpected error: %w", err)
		}

		cause := context.Cause(ctx)
		if cause != nil && cause != err {
			return cause
		}
	}

	return nil
}
