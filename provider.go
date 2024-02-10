package dsts

import "context"

// Provider is a function that pushes updates to `ch`.
//
// A provider is expected to block until `ctx` is cancelled.
type Provider func(ctx context.Context, ch chan<- I3Status) error
