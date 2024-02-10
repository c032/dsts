package dsts

type Source interface {
	OnUpdate(callback OnUpdateCallbackFunc) RemoveOnUpdateCallbackFunc
}

type OnUpdateCallbackFunc func()

type RemoveOnUpdateCallbackFunc func()

type sourceOnUpdateFunc func(callback OnUpdateCallbackFunc) RemoveOnUpdateCallbackFunc

func (fn sourceOnUpdateFunc) OnUpdate(callback OnUpdateCallbackFunc) RemoveOnUpdateCallbackFunc {
	return fn(callback)
}
