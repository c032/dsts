package dsts

// Notifier can run callbacks when there's a new update on some data, so that
// the i3 status line can be redrawn.
type Notifier interface {
	// OnUpdate registers `callback` to be called when there's an update.
	//
	// It returns a function that, when called, should un-register `callback`
	// so that it's no longer called on updates.
	OnUpdate(callback OnUpdateCallbackFunc) RemoveOnUpdateCallbackFunc
}

type OnUpdateCallbackFunc func()
type RemoveOnUpdateCallbackFunc func()

type NotifierFunc func(callback OnUpdateCallbackFunc) RemoveOnUpdateCallbackFunc

func (fn NotifierFunc) OnUpdate(callback OnUpdateCallbackFunc) RemoveOnUpdateCallbackFunc {
	return fn(callback)
}
