package dsts

// Notifier can run callbacks when there's a new update on some data, so that
// the i3 status line can be redrawn.
type Notifier interface {
	// OnUpdate registers `callback` to be called when there's an update.
	//
	// It returns a function that, when called, should un-register `callback`
	// so that it's no longer called on updates.
	OnUpdate(callback NotifierCallbackFunc) RemoveCallbackFunc
}

type NotifierCallbackFunc func()
type RemoveCallbackFunc func()

// UpdateNotifier is a wrapper for a function with the same signature as
// `Notifier.OnUpdate`, that implements `Notifier` by using itself as the
// `OnUpdate` method.
type UpdateNotifier func(callback NotifierCallbackFunc) RemoveCallbackFunc

func (fn UpdateNotifier) OnUpdate(callback NotifierCallbackFunc) RemoveCallbackFunc {
	return fn(callback)
}
