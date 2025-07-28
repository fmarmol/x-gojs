//go:build js

package main

import (
	"syscall/js"

	. "github.com/fmarmol/gojs"
)

func main() {
	stop := make(chan struct{})
	allowDrop := func(this js.Value, args []js.Value) any {
		event := NewDragEvent(args[0])
		event.PreventDefault()
		return nil

	}

	dragStart := func(this js.Value, args []js.Value) any {
		event := NewDragEvent(args[0])
		event.DataTransfert().SetData("text", event.Target().Value.Get("id").String())
		return nil

	}

	drop := func(this js.Value, args []js.Value) any {
		event := NewDragEvent(args[0])
		event.PreventDefault()
		data := event.DataTransfert().GetData("text")
		child := GetElementById(data)
		event.Target().C(child)
		return nil
	}

	div := Div().C(
		Div().Draggable(dragStart).Text(String("drag me 1")).
			Style("width", String("300px")).
			Style("height", String("150px")).
			Style("border", String("1px solid black")),
		Div().Draggable(dragStart).Text(String("drag me 2")).
			Style("width", String("300px")).
			Style("height", String("150px")).
			Style("border", String("1px solid black")),
		Div().OnDrop(drop).OnDragOver(allowDrop).Text(String("drop here")).
			Style("width", String("350px")).
			Style("height", String("200px")).
			Style("border", String("1px solid black")),
	)
	Init(div)

	<-stop
}
