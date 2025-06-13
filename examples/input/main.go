// go:build js

package main

import (
	"syscall/js"
	"unsafe"

	. "github.com/fmarmol/x-gojs"
)

type InputExample struct {
	Text string
}

func (c *InputExample) View() *Val {

	div := Div().C(
		Input().Attr("type", String("text")).OnInput(func(this js.Value, args []js.Value) any {
			event := args[0]
			c.Text = event.Get("target").Get("value").String()
			Update[string](unsafe.Pointer(&c.Text))
			return nil
		}).Attr("value", func() string { return c.Text }),
		Div().Text(func() string { return c.Text }),
	)
	State[string](div, c, "Text")
	return div
}

func main() {

	stop := make(chan struct{})
	c := &InputExample{Text: "Hello world"}

	Init(c.View())

	<-stop

}
