// go:build js

package main

import (
	"fmt"
	"unsafe"

	. "github.com/fmarmol/x-gojs"
)

type Counter struct {
	*Val
	Count string
}

func (c *Counter) View() *Val {

	div := Div().C(
		Div().State(c, "Count").Text(func() string { return fmt.Sprint(c.Count) }),
		Button().C(
			Text(String("inc")),
		).OnClick(func() {
			c.Count += "1"
			Update(unsafe.Pointer(&c.Count))
		}),
	)
	otherDiv := Div().C(
		Button().C(Text(String("outside"))).OnClick(func() { c.Count += "2"; Update(unsafe.Pointer(&c.Count)) }),
	)
	return Div().C(div, otherDiv)
}

func main() {

	stop := make(chan struct{})
	c := new(Counter)

	Init(c.View())

	<-stop

}
