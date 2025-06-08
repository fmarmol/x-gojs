// go:build js

package main

import (
	"fmt"
	"unsafe"

	. "github.com/fmarmol/x-gojs"
)

type Counter struct {
	*Val
	Count int
}

func (c *Counter) View() *Val {

	div := Div().C(
		Div().Text(func() string { return fmt.Sprint(c.Count) }),
	)
	button := Div().C(
		Button().C(Text(String("inc"))).
			OnClick(func() {
				c.Count += 1
				Update[int](unsafe.Pointer(&c.Count))
			}),
	)
	State2[int](div, c, "Count")
	return Div().C(div, button)
}

func main() {

	stop := make(chan struct{})
	c := new(Counter)

	Init(c.View())

	<-stop

}
