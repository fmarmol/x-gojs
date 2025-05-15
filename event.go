//go:build js

package gojs

import "syscall/js"

type Event struct {
	js.Value
}

func (e Event) PreventDefault() {
	e.Call("preventDefault")
}

func (e Event) Target() *Val {
	return &Val{Value: e.Get("target")}
}

type DataTransfer struct {
	js.Value
}

func (dt DataTransfer) GetData(key string) string {
	return dt.Call("getData", key).String()
}

func (dt DataTransfer) SetData(key string, value string) {
	dt.Call("setData", key, value)
}

type DragEvent struct {
	Event
}

func NewDragEvent(v js.Value) DragEvent {
	return DragEvent{Event: Event{Value: v}}
}

func (e DragEvent) DataTransfert() DataTransfer {
	return DataTransfer{
		Value: e.Get("dataTransfer"),
	}

}
