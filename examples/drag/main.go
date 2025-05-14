//go:build js

package main

import (
	"syscall/js"

	. "github.com/fmarmol/x-gojs"
)

func main() {
	stop := make(chan struct{})
	allowDrop := func(this js.Value, args []js.Value) any {
		event := args[0]
		event.Call("preventDefault")
		return nil

	}

	dragStart := func(this js.Value, args []js.Value) any {
		event := args[0]
		event.Get("dataTransfer").Call("setData", "text", event.Get("target").Get("id").String())
		return nil

	}

	drop := func(this js.Value, args []js.Value) any {
		event := args[0]
		event.Call("preventDefault")
		data := event.Get("dataTransfer").Call("getData", "text").String()

		child := GetElementById(data)
		event.Get("target").Call("appendChild", child.Value)
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
