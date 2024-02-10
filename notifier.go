package dsts

type Notifier interface {
	OnUpdate(callback OnUpdateCallbackFunc) RemoveOnUpdateCallbackFunc
}

type OnUpdateCallbackFunc func()

type RemoveOnUpdateCallbackFunc func()

type NotifierFunc func(callback OnUpdateCallbackFunc) RemoveOnUpdateCallbackFunc

func (fn NotifierFunc) OnUpdate(callback OnUpdateCallbackFunc) RemoveOnUpdateCallbackFunc {
	return fn(callback)
}
